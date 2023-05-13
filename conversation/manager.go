package conversation

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/duh-uh/teabot/backend"
	"github.com/duh-uh/teabot/engine"
	"github.com/duh-uh/teabot/globals"
	"github.com/duh-uh/teabot/message"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
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

	commandRegex *regexp.Regexp
}

func NewManager(ctx context.Context, cfg *cli.Context, backend backend.Backender, backendQueues backend.BackendQueues) *Manager {
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

		commandRegex: regexp.MustCompile(fmt.Sprintf(`\b%s(.+)\b`, globals.BotUrlScheme)),
	}

	engine.ConfigInit()
	cm.LoadEngines(ctx, cfg)

	return &cm
}

// Load engines from the config directory
func (cm *Manager) LoadEngines(ctx context.Context, cfg *cli.Context) {
	config_dir := cfg.String("config-directory")

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

		factory, err := cm.registry.GetEngineFactory(ctx, config)
		if err != nil {
			logrus.Warnf("Failed to load engine factory: %s: %v", config_file, err)
			continue
		}
		logrus.Debugf("Loaded engine factory: %#v", factory)

		execEngineNames[config.Name] = true
		logrus.Infof("Loaded bot: %s", config.Name)

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

func (cm *Manager) getEngineEnvironment(m *message.Message, env map[string]string) map[string]string {
	envmap := make(map[string]string)
	prefix := strings.ToUpper(globals.BotName)
	envmap[prefix+"_USER_ID"] = m.BotUserId
	envmap[prefix+"_USER_NAME"] = m.BotUserName
	envmap[prefix+"_CHANNEL"] = m.ChannelName
	envmap[prefix+"_CHANNEL_ID"] = m.ChannelId
	envmap[prefix+"_BACKEND_NAME"] = cm.backend.Name()
	envmap[prefix+"_LOCALE"] = m.Locale

	if m.ThreadId != "" {
		envmap[prefix+"_THREAD"] = m.ThreadId
	}

	for k, v := range env {
		envmap[k] = v
	}

	return envmap
}

func (cm *Manager) cleanupConversation(c *Conversation) {
	// Remove conversation from manager
	if c.conversationType == ConversationTypeThreaded {
		cm.convLock.Lock()
		delete(cm.conversations, c.threadId)
		globals.NumThreadedConversations.Dec()
		globals.NumConversations.Dec()
		cm.convLock.Unlock()
	} else {
		cm.channelConvLock.Lock()
		delete(cm.channelConversations[c.channelId], c.engineName)
		globals.NumChannelConversations.Dec()
		globals.NumConversations.Dec()
		cm.channelConvLock.Unlock()
	}
}

func (cm *Manager) addThreadedConversation(ctx context.Context, c *Conversation, threadId string) {
	c.threadId = threadId
	c.conversationType = ConversationTypeThreaded

	cm.convLock.Lock()
	if _, exists := cm.conversations[threadId]; !exists {
		cm.conversations[threadId] = c
		globals.NumThreadedConversations.Inc()
		globals.NumConversations.Inc()
		cm.convLock.Unlock()

		go func() {
			c.Start(ctx)
			cm.cleanupConversation(c)
		}()
	} else {
		cm.convLock.Unlock()
		logrus.Infof("Race detected: conversation: %#v, thread: %s", c, threadId)
	}
}

func (cm *Manager) addChannelConversation(ctx context.Context, c *Conversation, channelId string) bool {
	c.conversationType = ConversationTypeChannel

	cm.channelConvLock.Lock()
	if _, exists := cm.channelConversations[channelId][c.engineName]; !exists {
		if _, exists := cm.channelConversations[channelId]; !exists {
			cm.channelConversations[channelId] = map[string]*Conversation{}
		}
		cm.channelConversations[channelId][c.engineName] = c
		globals.NumChannelConversations.Inc()
		globals.NumConversations.Inc()
		cm.channelConvLock.Unlock()

		go func() {
			c.Start(ctx)
			cm.cleanupConversation(c)
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
		if c.directMessagesOnly && !m.DirectMessage {
			logrus.Debugf("Conversation is direct messages only, ignoring message: %#v", c)
		} else {
			logrus.Debugf("Matched existing conversation: %#v", c)
			conversations = append(conversations, c)
		}
	}
	cm.convLock.RUnlock()

	// Can't have multple bot conversations in a thread
	// XXX: Or should we allow this?
	if len(conversations) > 0 {
		return conversations
	}

	if !m.InThread {
		cm.channelConvLock.RLock()
		if cc, ok := cm.channelConversations[m.ChannelId]; ok {
			// Found channel conversations for channel ID
			for _, c := range cc {
				if c.directMessagesOnly && !m.DirectMessage {
					logrus.Debugf("Conversation is direct messages only, ignoring message: %#v", c)
				} else {
					logrus.Debugf("Matched existing channel conversation: %#v", c)
					conversations = append(conversations, c)
				}
			}
		}
		cm.channelConvLock.RUnlock()
	}

	cm.triggerLock.RLock()
	for re, efs := range cm.triggers {
		if re.MatchString(m.Text) {
			for _, ef := range efs {
				config := ef.Config()

				// Respond only to direct messages unless trigger is global
				if config.DirectMessageTriggersOnly && !m.DirectMessage {
					logrus.Debugf("Skipping %s trigger for non-direct message", config.Name)
					continue
				}

				// If list of channels is specified, only create conversation
				// if channel is in list
				if len(config.Channels) > 0 {
					found := false
					for _, channel := range config.Channels {
						if channel == m.ChannelName {
							found = true
							break
						}
					}
					if !found {
						logrus.Debugf("Skipping %s trigger for channel %s", config.Name, m.ChannelName)
						continue
					}
				}

				engqs := engine.NewEngineQueues()
				envmap := cm.getEngineEnvironment(m, config.Environment)

				e := ef.Create(envmap)

				c := Conversation{
					channelId:          m.ChannelId,
					channelName:        m.ChannelName,
					manager:            cm,
					engine:             e,
					engineName:         config.Name,
					engineQueues:       engqs,
					prefixUsername:     config.PrefixUsername,
					directMessagesOnly: config.DirectMessagesOnly,
				}

				if config.Threaded {
					cm.addThreadedConversation(ctx, &c, m.ThreadId)
					conversations = append(conversations, &c)
					logrus.Debugf("New threaded conversation with %s: %+v", config.Name, c)
				} else {
					if cm.addChannelConversation(ctx, &c, m.ChannelId) {
						conversations = append(conversations, &c)
						logrus.Debugf("New channel conversation with %s: %+v", c.engineName, c)
					} else {
						logrus.Debugf("Ignoring trigger as bot already active: %s: channel='%s' msg='%s' trigger='%s'",
							c.engineName, c.channelName, m.Text, re.String())
					}
				}

			}
		}
	}
	cm.triggerLock.RUnlock()

	return conversations
}

// Conversation commands
const (
	_ = iota
	ConversationCommandSwitchChannel
	ConversationCommandSwitchThread
)

func (cm *Manager) Post(c *Conversation, m *message.Message) {
	logrus.Debugf("Posting message to backend: %#v", m)

	command := 0
	if strings.Contains(m.Text, globals.BotUrlScheme) {
		matches := cm.commandRegex.FindStringSubmatch(m.Text)
		if len(matches) > 0 {
			switch matches[1] {
			case "switch/channel":
				command = ConversationCommandSwitchChannel
			case "switch/thread":
				command = ConversationCommandSwitchThread
			}

			if command != 0 {
				logrus.Debugf("Matched command: %s", matches[0])

				// Remove command from message text
				m.Text = strings.TrimSpace(strings.Replace(m.Text, matches[0], "", 1))
				if len(m.Text) == 0 {
					m.Text = "_..._"
				}
			} else {
				logrus.Debugf("Ignoring unknown command in message: %s", m.Text)
			}
		}
	}

	if command == ConversationCommandSwitchThread {
		// Need the new thread ID
		m.NeedThreadId = true
		m.ThreadIdChan = make(chan string, 1)
	}

	// Send message to backend
	cm.backendQueues.RespQ <- m

	if m.NeedThreadId {
		// Wait for thread ID
		select {
		case m.ThreadId = <-m.ThreadIdChan:
			logrus.Debugf("Got thread ID: %s", m.ThreadId)
		case <-time.After(5 * time.Second):
			logrus.Warnf("Timeout waiting for thread ID: bot=%s channel=%s",
				c.engineName, c.channelName)
			return
		}

	}

	switch command {
	case ConversationCommandSwitchChannel:
		// Switch to channel conversation
		logrus.Debugf("Switching to channel conversation for %s", c.engineName)
		cm.convLock.Lock()
		cm.channelConvLock.Lock()
		delete(cm.conversations, m.ThreadId)
		cm.channelConversations[m.ChannelId][c.engineName] = c
		c.threadId = ""
		globals.NumThreadedConversations.Dec()
		globals.NumChannelConversations.Inc()
		cm.channelConvLock.Unlock()
		cm.convLock.Unlock()

	case ConversationCommandSwitchThread:
		// Switch to threaded conversation
		logrus.Debugf("Switching to threaded conversation for %s: channel=%s thread=%s",
			c.engineName, m.ChannelId, m.ThreadId)
		cm.convLock.Lock()
		cm.channelConvLock.Lock()
		delete(cm.channelConversations[m.ChannelId], c.engineName)
		cm.conversations[m.ThreadId] = c
		c.threadId = m.ThreadId
		globals.NumChannelConversations.Dec()
		globals.NumThreadedConversations.Inc()
		cm.channelConvLock.Unlock()
		cm.convLock.Unlock()
	}
}
