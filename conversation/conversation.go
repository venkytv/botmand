package conversation

import (
	"context"

	"github.com/duh-uh/teabot/engine"
	"github.com/duh-uh/teabot/message"
	"github.com/sirupsen/logrus"
)

type Conversation struct {
	threadId     string
	channelId    string
	channelName  string
	manager      *Manager
	engine       engine.Enginer
	engineQueues engine.EngineQueues
}

func NewConversation(engine engine.Enginer, engineQueues engine.EngineQueues) *Conversation {
	return &Conversation{
		engine:       engine,
		engineQueues: engineQueues,
	}
}

func (c *Conversation) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go c.engine.Start(ctx)

	for {
		select {
		case resp, more := <-c.engineQueues.ReadQ:
			if more {
				m := &message.Message{
					Text:        resp,
					ChannelId:   c.channelId,
					ChannelName: c.channelName,
					ThreadId:    c.threadId,
				}
				c.manager.Post(m)
			} else {
				logrus.Debug("Done with conversation")
				return
			}
		case <-ctx.Done():
			logrus.Debug("Conversation aborted")
			return
		}
	}
}

func (c *Conversation) Post(m *message.Message) {
	c.engineQueues.WriteQ <- m.Text
}
