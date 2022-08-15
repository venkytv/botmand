package conversation

import (
	"context"
	"strings"
	"sync"

	"github.com/duh-uh/teabot/backend"
	"github.com/duh-uh/teabot/engine"
	"github.com/duh-uh/teabot/globals"
	"github.com/duh-uh/teabot/message"
	"github.com/sirupsen/logrus"
)

type Manager struct {
	backend              backend.Backender
	backendQueues        backend.BackendQueues
	conversations        map[string]*Conversation
	convLock             *sync.RWMutex
	channelConversations map[string]map[string]*Conversation
	channelConvLock      *sync.RWMutex
}

func NewManager(backend backend.Backender, backendQueues backend.BackendQueues) *Manager {
	cm := Manager{
		backend:              backend,
		backendQueues:        backendQueues,
		conversations:        make(map[string]*Conversation),
		convLock:             &sync.RWMutex{},
		channelConversations: make(map[string]map[string]*Conversation),
		channelConvLock:      &sync.RWMutex{},
	}

	return &cm
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

func (cm *Manager) addChannelConversation(ctx context.Context, c *Conversation, channelId string, convId string) {
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
	} else {
		cm.channelConvLock.Unlock()
		logrus.Infof("Race detected: conversation: %#v, channelID: %s, convID: %s",
			c, channelId, convId)
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

	cm.channelConvLock.RLock()
	if cc, ok := cm.channelConversations[m.ChannelId]; ok {
		// Found channel conversations for channel ID
		for _, c := range cc {
			logrus.Debugf("Matched existing channel conversation: %#v", c)
			conversations = append(conversations, c)
		}
	}
	cm.channelConvLock.RUnlock()

	// XXX: Testing; create conversation unconditionally
	if len(conversations) < 1 {
		isThreadedConv := true

		engqs := engine.NewEngineQueues()
		envmap := cm.getEngineEnvironment(m)
		e := engine.NewExecEngine([]string{"./test.sh"}, envmap, &engqs)
		c := Conversation{
			channelId:    m.ChannelId,
			channelName:  m.ChannelName,
			manager:      cm,
			engine:       e,
			engineQueues: engqs,
		}

		if isThreadedConv {
			cm.addThreadedConversation(ctx, &c, m.ThreadId)
		} else {
			cm.addChannelConversation(ctx, &c, m.ChannelId, e.Name())
		}

		conversations = append(conversations, &c)
	}

	return conversations
}

func (cm *Manager) Post(m *message.Message) {
	logrus.Debugf("Posting message to backend: %#v", m)
	cm.backendQueues.RespQ <- m
}
