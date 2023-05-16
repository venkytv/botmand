package backend

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/venkytv/botters/message"
)

const (
	TestEventConnect int = iota
	TestEventChannelJoined
	TestEventMessage
	TestEventDisconnect
)

type TestSlackEvent struct {
	Type        int
	SubType     string
	From        string
	ChannelId   string
	ChannelName string
	Message     string
	Thread      string
	Timestamp   string
}

type TestSlackApi struct {
	ChannelMap   map[string]string
	Events       []TestSlackEvent
	ExpectedMsgs []*message.Message
}

func (s TestSlackApi) ChannelInfo(channel string) *slack.Channel {
	ci, exists := s.ChannelMap[channel]
	if !exists {
		return &slack.Channel{}
	}

	return &slack.Channel{
		GroupConversation: slack.GroupConversation{
			Name: ci,
		},
	}
}

func (s TestSlackApi) genRTMEvent(tse TestSlackEvent) slack.RTMEvent {
	switch tse.Type {
	case TestEventConnect:
		ev := slack.ConnectedEvent{
			Info: &slack.Info{
				User: &slack.UserDetails{
					ID: tse.From,
				},
			},
		}
		return slack.RTMEvent{Data: &ev}

	case TestEventChannelJoined:
		msg := slack.Message{
			Msg: slack.Msg{
				User:            tse.From,
				Text:            tse.Message,
				ThreadTimestamp: tse.Thread,
			},
		}
		ev := slack.ChannelJoinedEvent{
			Channel: slack.Channel{
				GroupConversation: slack.GroupConversation{
					Conversation: slack.Conversation{
						ID:     tse.ChannelId,
						Latest: &msg,
					},
				},
			},
		}
		return slack.RTMEvent{Data: &ev}

	case TestEventMessage:
		ev := slack.MessageEvent{
			Msg: slack.Msg{
				User:    tse.From,
				Channel: tse.ChannelId,
				SubType: tse.SubType,
				Text:    tse.Message,
			},
		}
		if tse.Thread != "" {
			ev.ThreadTimestamp = tse.Thread
		}
		if tse.Timestamp == "" {
			now := float64(time.Now().UnixNano()) / 1000000
			ev.Timestamp = fmt.Sprintf("%f", now)
		} else {
			ev.Timestamp = tse.Timestamp
		}

		return slack.RTMEvent{Data: &ev}

	case TestEventDisconnect:
		ev := slack.DisconnectedEvent{Intentional: true}
		return slack.RTMEvent{Data: &ev}

	default:
		panic(fmt.Sprintf("Unknown event: %#v", tse))
	}
}

func (s TestSlackApi) GetEvents() chan slack.RTMEvent {
	nEvents := len(s.Events)
	logrus.Debugf("Loading %d events", nEvents)

	ch := make(chan slack.RTMEvent, nEvents)

	for _, ev := range s.Events {
		ch <- s.genRTMEvent(ev)
	}
	close(ch)

	return ch
}

func (s TestSlackApi) PostMessage(channel string, msgOptions ...slack.MsgOption) (string, error) {
	return "", nil
}

func (s TestSlackApi) PostTypingIndicator(channel string) {}

func TestRead(t *testing.T) {
	var botUserId = "IAMALITTLESLACKBOT"
	//var myMsgTimestamp = "3344556.77889"

	api := TestSlackApi{
		ChannelMap: map[string]string{
			"C234567": "TestChannel1",
		},
		Events: []TestSlackEvent{
			TestSlackEvent{
				// This message sets up the bot User ID
				Type: TestEventConnect,
				From: botUserId,
			},
			TestSlackEvent{
				// This message should be dropped
				Type:      TestEventMessage,
				Message:   "BotMessage",
				From:      botUserId,
				ChannelId: "C234567",
			},
			TestSlackEvent{
				Type:      TestEventMessage,
				Message:   "TestMessage",
				From:      "U234567",
				ChannelId: "C234567",
			},
		},
		ExpectedMsgs: []*message.Message{
			&message.Message{
				Text:        "TestMessage",
				User:        "U234567",
				ChannelId:   "C234567",
				ChannelName: "TestChannel1",
			},
		},
	}

	backendQs := NewBackendQueues()

	backend := NewSlackBackend(&api, &backendQs)
	go backend.Read()

	for _, m := range api.ExpectedMsgs {
		t.Run(m.Text, func(t *testing.T) {
			select {
			case got := <-backendQs.MesgQ:
				logrus.Debugf("Got message: %#v", got)
				assert.Equal(t, m.Text, got.Text)

				if m.ChannelName != "" {
					assert.Equal(t, m.ChannelName, got.ChannelName)
				}

			case <-time.After(500 * time.Millisecond):
				assert.Failf(t, "Failed to read backend msg", "Was expecting: %#v", m)
			}
		})
	}

	// Ensure there are no other messages in the channel
	t.Run("UnexpectedMessages", func(t *testing.T) {
		select {
		case got := <-backendQs.MesgQ:
			assert.Fail(t, "Received unexpected message", got)
		default:
			break
		}
	})

	t.Run("BotUserID", func(t *testing.T) {
		assert.Equal(t, botUserId, backend.botId)
	})
}

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)

	// Discard log messages during normal testing
	logrus.SetOutput(ioutil.Discard)

	os.Exit(m.Run())
}
