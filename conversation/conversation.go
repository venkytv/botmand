package conversation

import (
	"bufio"
	"context"
	"io"

	"github.com/duh-uh/teabot/engine"
	"github.com/duh-uh/teabot/message"
	"github.com/sirupsen/logrus"
)

// Conversation types
const (
	ConversationTypeThreaded = iota
	ConversationTypeChannel
)

type Conversation struct {
	conversationType   int
	threadId           string
	channelId          string
	channelName        string
	manager            *Manager
	engine             engine.Enginer
	engineName         string
	engineQueues       engine.EngineQueues
	prefixUsername     bool
	directMessagesOnly bool
}

func (c *Conversation) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go c.LaunchEngine(ctx)

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
				c.manager.Post(c, m)
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

func (c *Conversation) LaunchEngine(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer func() {
		logrus.Debug("Closing engine queues")
		close(c.engineQueues.ReadQ)
		close(c.engineQueues.WriteQ)
	}()

	stdin, stdout, stderr, err := c.engine.Setup(ctx)
	if err != nil {
		logrus.Error("Failed to setup engine:", err)
		return
	}

	// Pipe input from WriteQ to the command
	go func() {
		defer stdin.Close()
		for {
			select {
			case t := <-c.engineQueues.WriteQ:
				if t == "" {
					continue
				}
				if _, err := io.WriteString(stdin, t+"\n"); err != nil {
					logrus.Errorf("Failed to post message to command: '%s' (%v)", t, err)
				}
			case <-ctx.Done():
				logrus.Debug("Closing stdin channel")
				return
			}
		}
	}()

	// Pipe output of command to ReadQ
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			t := scanner.Text()
			c.engineQueues.ReadQ <- t
		}
		logrus.Debug("Closing stdout channel")
	}()

	// Log stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			t := scanner.Text()
			logrus.Debug("Engine stderr:", t)
		}
		logrus.Debug("Closing stderr channel")
	}()

	// Start the engine
	if err := c.engine.Start(ctx); err != nil {
		logrus.Error("Failed to start engine:", err)
		return
	}

	// Wait for the engine to finish
	if err := c.engine.Wait(ctx); err != nil {
		logrus.Error("Engine failed:", err)
		return
	}
}

func (c *Conversation) Post(m *message.Message) {
	msg := m.Text
	if c.prefixUsername {
		msg = m.User + ": " + msg
	}
	c.engineQueues.WriteQ <- msg
}
