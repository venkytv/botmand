package backend

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/duh-uh/teabot/message"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

type SlackApier interface {
	ChannelInfo(channel string) *slack.Channel
	GetEvents() chan slack.RTMEvent
	PostMessage(channel string, msgOptions ...slack.MsgOption) error
}

type SlackApi struct {
	client *slack.Client
	rtm    *slack.RTM
}

func (s SlackApi) ChannelInfo(channel string) *slack.Channel {
	logrus.Debug("Looking up channel info for ", channel)
	ci, err := s.client.GetConversationInfo(channel, true)
	if err != nil {
		logrus.Error("Error looking up channel info: ", channel, err)
		return &slack.Channel{}
	}
	return ci
}

func (s SlackApi) GetEvents() chan slack.RTMEvent {
	go s.rtm.ManageConnection()
	return s.rtm.IncomingEvents
}

func (s SlackApi) PostMessage(channel string, msgOptions ...slack.MsgOption) error {
	_, _, err := s.client.PostMessage(channel, msgOptions...)
	return err
}

func NewSlackApi(configPath string) SlackApi {
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		logrus.Fatal("Failed to open config file:", configPath, err)
	}

	token := strings.TrimSpace(string(content))

	client := slack.New(
		token,
		slack.OptionDebug(false), // XXX: Make option
		slack.OptionLog(log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)),
	)
	rtm := client.NewRTM()

	return SlackApi{
		client: client,
		rtm:    rtm,
	}
}

type SlackBackend struct {
	api  SlackApier
	comm *BackendQueues

	me        string
	chanCache map[string]*slack.Channel
	sanitiser func(*message.Message) *message.Message
}

func NewSlackBackend(api SlackApier, comm *BackendQueues) *SlackBackend {
	return &SlackBackend{
		api:  api,
		comm: comm,

		chanCache: make(map[string]*slack.Channel),
		sanitiser: func(m *message.Message) *message.Message { return m },
	}
}

func (s SlackBackend) newMessage(ev *slack.MessageEvent, cc *slack.Channel) *message.Message {
	thread := ev.ThreadTimestamp
	if len(thread) < 1 {
		thread = ev.Timestamp
	}

	return &message.Message{
		Text:        ev.Text,
		User:        ev.User,
		ChannelId:   ev.Channel,
		ChannelName: cc.Name,
		ThreadId:    thread,
	}
}

func (s SlackBackend) channelInfo(channel string) *slack.Channel {
	if _, exists := s.chanCache[channel]; !exists {
		s.chanCache[channel] = s.api.ChannelInfo(channel)
	}

	return s.chanCache[channel]
}

func (s *SlackBackend) Read() {
	for msg := range s.api.GetEvents() {
		switch ev := msg.Data.(type) {
		case *slack.ConnectedEvent:
			logrus.Debug("Connection counter: ", ev.ConnectionCount)
			s.me = ev.Info.User.ID
			logrus.Info("I am ", s.me)

			// Set up the sanitiser to remove references to bot ID in message
			logrus.Debug("Setting up message sanitiser")
			re := regexp.MustCompile(fmt.Sprintf(`^\s*<@%s>\s+`, s.me))
			s.sanitiser = func(m *message.Message) *message.Message {
				m.Text = re.ReplaceAllString(m.Text, "")
				return m
			}

		case *slack.MessageEvent:
			if s.me == "" {
				logrus.Debug("Not connected yet!")
				break
			}

			if ev.User == "" {
				logrus.Debug("Ignoring ghost message: ", ev)
				break
			}

			if ev.User == s.me {
				logrus.Debug("Ignoring my message: ", ev)
				break
			}

			if ev.SubType == "message_replied" {
				// Ignore message replied notifications
				break
			}

			chanInfo := s.channelInfo(ev.Channel)
			logrus.Debug("Channel: ", chanInfo)

			m := s.newMessage(ev, chanInfo)
			logrus.Debugf("Message: %#v", m)

			s.comm.MesgQ <- m

		case *slack.RTMError:
			logrus.Errorf("RTM error: %s", ev.Error())

		case *slack.InvalidAuthEvent:
			logrus.Fatal("Invalid credentials")

		default:
			// Ignore all other events
			//logrus.Debugf("Ignoring event: %#v", ev)
		}

	}
}

func (s SlackBackend) Post() {
	for {
		msg, more := <-s.comm.RespQ
		if !more {
			logrus.Debug("Shutting down SlackBackend")
			return
		}
		logrus.Debugf("Got response: %#v", msg)

		msgOptions := []slack.MsgOption{
			slack.MsgOptionText(msg.Text, false),
			slack.MsgOptionAsUser(true),
			// TODO: Make this configurable
			slack.MsgOptionTS(msg.ThreadId),
		}

		err := s.api.PostMessage(msg.ChannelId, msgOptions...)
		if err != nil {
			logrus.Error("PostMessage error: ", err)
		}
	}
}

func (s SlackBackend) Sanitize(m *message.Message) *message.Message {
	return s.sanitiser(m)
}
