package conversation

import (
	"context"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/duh-uh/teabot/backend"
	"github.com/duh-uh/teabot/engine"
	"github.com/duh-uh/teabot/globals"
	"github.com/duh-uh/teabot/message"
	"github.com/sirupsen/logrus"
	"github.com/venkytv/go-config"
)

type Manager struct {
	registry             engine.EngineRegistry
	backend              backend.Backender
	backendQueues        backend.BackendQueues
	triggers             map[*regexp.Regexp][]engine.EngineFactoryer
	triggerLock          *sync.RWMutex
	conversations        map[string]*Conversation
	convLock             *sync.RWMutex
	channelConversations map[string]map[string]*Conversation
	channelConvLock      *sync.RWMutex
}

func NewManager(ctx context.Context, cfg *config.Config, backend backend.Backender, backendQueues backend.BackendQueues) *Manager {
	engineRegistry := engine.NewEngineRegistry()

	engineRegistry.Register("executable", engine.ExecEngineFactoryLoader{})

	cm := Manager{
		registry:             engineRegistry,
		backend:              backend,
		backendQueues:        backendQueues,
		triggerLock:          &sync.RWMutex{},
		conversations:        make(map[string]*Conversation),
		convLock:             &sync.RWMutex{},
		channelConversations: make(map[string]map[string]*Conversation),
		channelConvLock:      &sync.RWMutex{},
	}

	engine.ConfigInit()
	cm.LoadEngines(ctx, cfg)

	return &cm
}

// Load engines from the config directory
func (cm *Manager) LoadEngines(ctx context.Context, cfg *config.Config) {
	config_dir := cfg.GetString("config-directory")

	config_files, err := filepath.Glob(filepath.Join(config_dir, "*.yaml"))
	if err != nil {
		logrus.Fatalf("Failed to glob config files: %v", err)
	}

	// Load yaml files with ".yml" extension
	yml_config_files, err := filepath.Glob(filepath.Join(config_dir, "*.yml"))
	if err != nil {
		logrus.Fatalf("Failed to glob config files: %v", err)
	}
	config_files = append(config_files, yml_config_files...)

	// Lock the triggers map
	cm.triggerLock.Lock()

	cm.triggers = make(map[*regexp.Regexp][]engine.EngineFactoryer)
	execEngineNames := make(map[string]bool)
	for _, config_file := range config_files {
		logrus.Debugf("Loading config file: %s", config_file)
		config, err := engine.LoadConfig(config_file)
		if err != nil {
			logrus.Warnf("Failed to load config file: %s: %v", config_file, err)
			continue
		}

		// Bail out if config with same name already exists
		if _, ok := execEngineNames[config.Name]; ok {
			logrus.Warnf("Duplicate config name: %s", config.Name)
			continue
		}
		execEngineNames[config.Name] = true

		factory := cm.registry.GetEngineFactory(ctx, config)
		logrus.Debugf("Loaded engine factory: %#v", factory)

		// Add triggers from config to the manager
		for _, pattern := range config.Triggers {
			re, err := regexp.Compile(pattern)
			if err != nil {
				logrus.Warnf("Failed to compile regex: %s: %#v: %v", pattern, factory, err)
				continue
			}

			cm.triggers[re] = append(cm.triggers[re], factory)
		}
	}

	cm.triggerLock.Unlock()

	globals.NumExecEngineFactories.Set(float64(len(execEngineNames)))
	nTriggers := len(cm.triggers)
	globals.NumConversationTriggers.Set(float64(nTriggers))
	if nTriggers < 1 {
		logrus.Warn("No bots loaded!")
	}
	logrus.Debugf("Loaded engine factories: %+v", cm.triggers)
}

func (cm *Manager) Start(ctx context.Context) {
	go cm.backend.Read()
	go cm.backend.Post()

	for {
		select {
		case m := <-cm.backendQueues.MesgQ:
			m = cm.backend.Sanitize(m)

			convs := cm.GetConversations(ctx, m)
			for _, conv := range convs {
				conv.Post(m)
			}
		case <-ctx.Done():
			logrus.Debug("Terminating conversation manager")
			return
		}
	}
}

func (cm *Manager) getEngineEnvironment(m *message.Message) map[string]string {
	envmap := make(map[string]string)
	prefix := strings.ToUpper(globals.BotName)
	envmap[prefix+"_CHANNEL"] = m.ChannelName
	envmap[prefix+"_CHANNEL_ID"] = m.ChannelId
	envmap[prefix+"_BACKEND_NAME"] = cm.backend.Name()
	envmap[prefix+"_LOCALE"] = m.Locale

	if m.ThreadId != "" {
		envmap[prefix+"_THREAD"] = m.ThreadId
	}

	return envmap
}

func (cm *Manager) addThreadedConversation(ctx context.Context, c *Conversation, threadId string) {
	c.threadId = threadId

	cm.convLock.Lock()
	if _, exists := cm.conversations[threadId]; !exists {
		cm.conversations[threadId] = c
		globals.NumThreadedConversations.Inc()
		globals.NumConversations.Inc()
		cm.convLock.Unlock()

		go func() {
			c.Start(ctx)

			// Remove conversation from manager
			cm.convLock.Lock()
			delete(cm.conversations, threadId)
			globals.NumThreadedConversations.Dec()
			globals.NumConversations.Dec()
			cm.convLock.Unlock()
		}()
	} else {
		cm.convLock.Unlock()
		logrus.Infof("Race detected: conversation: %#v, thread: %s", c, threadId)
	}
}

func (cm *Manager) addChannelConversation(ctx context.Context, c *Conversation, channelId string, convId string) bool {
	cm.channelConvLock.Lock()
	if _, exists := cm.channelConversations[channelId][convId]; !exists {
		if _, exists := cm.channelConversations[channelId]; !exists {
			cm.channelConversations[channelId] = map[string]*Conversation{}
		}
		cm.channelConversations[channelId][convId] = c
		globals.NumChannelConversations.Inc()
		globals.NumConversations.Inc()
		cm.channelConvLock.Unlock()

		go func() {
			c.Start(ctx)

			// Remove conversation from manager
			cm.channelConvLock.Lock()
			delete(cm.channelConversations[channelId], convId)
			globals.NumChannelConversations.Dec()
			globals.NumConversations.Dec()
			cm.channelConvLock.Unlock()
		}()

		return true
	} else {
		// Bot already active in channel
		cm.channelConvLock.Unlock()
		return false
	}
}

func (cm *Manager) GetConversations(ctx context.Context, m *message.Message) []*Conversation {
	conversations := []*Conversation{}

	cm.convLock.RLock()
	if c, ok := cm.conversations[m.ThreadId]; ok {
		// Found conversation for message thread
		logrus.Debugf("Matched existing conversation: %#v", c)
		conversations = append(conversations, c)
	}
	cm.convLock.RUnlock()

	if !m.InThread {
		cm.channelConvLock.RLock()
		if cc, ok := cm.channelConversations[m.ChannelId]; ok {
			// Found channel conversations for channel ID
			for _, c := range cc {
				logrus.Debugf("Matched existing channel conversation: %#v", c)
				conversations = append(conversations, c)
			}
		}
		cm.channelConvLock.RUnlock()
	}

	cm.triggerLock.RLock()
	for re, efs := range cm.triggers {
		if re.MatchString(m.Text) {
			for _, ef := range efs {
				// If list of channels is specified, only create conversation
				// if channel is in list
				if len(ef.Config().Channels) > 0 {
					found := false
					for _, channel := range ef.Config().Channels {
						if channel == m.ChannelName {
							found = true
							break
						}
					}
					if !found {
						logrus.Debugf("Skipping %s trigger for channel %s", ef.Config().Name, m.ChannelName)
						continue
					}
				}

				engqs := engine.NewEngineQueues()
				envmap := cm.getEngineEnvironment(m)

				e := ef.Create(envmap, &engqs)
				c := Conversation{
					channelId:     m.ChannelId,
					channelName:   m.ChannelName,
					manager:       cm,
					engine:        e,
					engineFactory: ef,
					engineQueues:  engqs,
				}

				if ef.Config().Threaded {
					cm.addThreadedConversation(ctx, &c, m.ThreadId)
					conversations = append(conversations, &c)
				} else {
					if cm.addChannelConversation(ctx, &c, m.ChannelId, ef.Config().Name) {
						conversations = append(conversations, &c)
					} else {
						logrus.Debugf("Ignoring trigger as bot already active on channel: %s: %s: %#v",
							c.channelName, m.Text, ef)
					}
				}
			}
		}
	}
	cm.triggerLock.RUnlock()

	return conversations
}

func (cm *Manager) Post(m *message.Message) {
	logrus.Debugf("Posting message to backend: %#v", m)
	cm.backendQueues.RespQ <- m
}
