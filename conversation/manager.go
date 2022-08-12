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
	backend       backend.Backender
	backendQueues backend.BackendQueues
	conversations map[string]*Conversation
	mu            *sync.RWMutex
}

func NewManager(backend backend.Backender, backendQueues backend.BackendQueues) *Manager {
	cm := Manager{
		backend:       backend,
		backendQueues: backendQueues,
		conversations: make(map[string]*Conversation),
		mu:            &sync.RWMutex{},
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

			conv := cm.GetConversation(ctx, m)
			if conv != nil {
				logrus.Debugf("Matched conversation: %#v", conv)
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
	envmap[prefix+"_THREAD"] = m.ThreadId
	envmap[prefix+"_BACKEND_NAME"] = cm.backend.Name()
	envmap[prefix+"_LOCALE"] = m.Locale

	return envmap
}

func (cm *Manager) GetConversation(ctx context.Context, m *message.Message) *Conversation {
	// Read lock the conversations map
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if c, ok := cm.conversations[m.ThreadId]; ok {
		// Found conversation for message thread
		logrus.Debugf("Matched existing conversation: %#v", c)
		return c
	}

	// XXX: Testing; create conversation unconditionally
	if true {
		engqs := engine.NewEngineQueues()
		envmap := cm.getEngineEnvironment(m)
		e := engine.NewExecEngine([]string{"./test.sh"}, envmap, &engqs)
		c := Conversation{
			threadId:     m.ThreadId,
			channelId:    m.ChannelId,
			channelName:  m.ChannelName,
			manager:      cm,
			engine:       e,
			engineQueues: engqs,
		}

		// Upgrade to write lock
		cm.mu.RUnlock()
		cm.mu.Lock()
		if _, exists := cm.conversations[c.threadId]; !exists {
			cm.conversations[c.threadId] = &c
			go func() {
				c.Start(ctx)

				// Remove conversation from manager
				delete(cm.conversations, c.threadId)
			}()
		}

		// Downgrade to read lock
		cm.mu.Unlock()
		cm.mu.RLock()

		return &c
	}

	// No conversation matching message
	return nil
}

func (cm *Manager) Post(m *message.Message) {
	logrus.Debugf("Posting message to backend: %#v", m)
	cm.backendQueues.RespQ <- m
}
