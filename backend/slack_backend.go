package backend

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/duh-uh/teabot/message"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

type SlackApier interface {
	ChannelInfo(channel string) *slack.Channel
	GetEvents() chan slack.RTMEvent
	PostMessage(channel string, msgOptions ...slack.MsgOption) (string, error)
	PostTypingIndicator(channel string)
}

// SlackApi implements the SlackApier interface
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

func (s SlackApi) PostMessage(channel string, msgOptions ...slack.MsgOption) (string, error) {
	_, timestamp, err := s.client.PostMessage(channel, msgOptions...)
	return timestamp, err
}

func (s SlackApi) PostTypingIndicator(channel string) {
	s.rtm.SendMessage(s.rtm.NewTypingMessage(channel))
}

func NewSlackApi(token string, debug bool) SlackApi {
	client := slack.New(
		token,
		slack.OptionDebug(debug),
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

	botId       string
	botName     string
	atMePattern *regexp.Regexp
	chanCache   map[string]*slack.Channel
	sanitiser   func(*message.Message) *message.Message
	msgCache    *bigcache.BigCache
}

func NewSlackBackend(api SlackApier, comm *BackendQueues) *SlackBackend {
	msgCache, err := bigcache.NewBigCache(bigcache.DefaultConfig(1 * time.Minute))
	if err != nil {
		logrus.Fatal("Failed to initialise message cache: ", err)
	}

	return &SlackBackend{
		api:  api,
		comm: comm,

		atMePattern: regexp.MustCompile(`^$`),
		chanCache:   make(map[string]*slack.Channel),
		sanitiser:   func(m *message.Message) *message.Message { return m },
		msgCache:    msgCache,
	}
}

func (s SlackBackend) Name() string {
	return "Slack"
}

func (s SlackBackend) newMessage(ev *slack.MessageEvent, cc *slack.Channel) *message.Message {
	thread := ev.ThreadTimestamp
	inThread := true // Assume we're in a thread unless we're not

	if len(thread) < 1 {
		inThread = false // We're not in a threaded conversation yet
		thread = ev.Timestamp
	}

	return &message.Message{
		Text:          ev.Text,
		User:          ev.User,
		BotUserId:     s.botId,
		BotUserName:   s.botName,
		ChannelId:     ev.Channel,
		ChannelName:   cc.Name,
		ThreadId:      thread,
		InThread:      inThread,
		DirectMessage: s.atMePattern.MatchString(ev.Text),
		Locale:        cc.Locale,
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
			s.botId = ev.Info.User.ID
			s.botName = ev.Info.User.Name
			logrus.Infof("I am %s (%s)", s.botName, s.botId)

			// Set up regex to recognise @mentions of bot
			s.atMePattern = regexp.MustCompile(fmt.Sprintf(`<@%s>`, s.botId))

		case *slack.MessageEvent:
			if s.botId == "" {
				logrus.Debug("Not connected yet!")
				break
			}

			if ev.User == "" {
				logrus.Debugf("Ignoring ghost message: %#v", ev)
				break
			}

			if ev.User == s.botId {
				if _, err := s.msgCache.Get(ev.Timestamp); err != nil {
					// Found message in bot-generated message cache
					logrus.Debugf("Ignoring my message: %#v", ev)
					break
				}
			}

			if ev.User == "USLACKBOT" {
				logrus.Debugf("Ignoring Slackbot message: %#v", ev)
				break
			}

			if ev.SubType == "message_replied" {
				// Ignore message replied notifications
				break
			}

			chanInfo := s.channelInfo(ev.Channel)
			logrus.Debugf("Channel: %#v", chanInfo)

			m := s.newMessage(ev, chanInfo)
			logrus.Debugf("Message: %#v, from event: %#v", m, ev)

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

		if msg.Text == "..." {
			// Send a typing indicator
			s.api.PostTypingIndicator(msg.ChannelId)
			continue
		}

		// Convert embedded \n to actual newlines
		msg.Text = strings.ReplaceAll(msg.Text, `\n`, "\n")

		msgOptions := []slack.MsgOption{
			slack.MsgOptionText(msg.Text, false),
			slack.MsgOptionAsUser(true),
			slack.MsgOptionTS(msg.ThreadId),
		}

		timestamp, err := s.api.PostMessage(msg.ChannelId, msgOptions...)
		if err != nil {
			logrus.Error("PostMessage error: ", err)
		}

		// Cache bot messages for a minute
		err = s.msgCache.Set(timestamp, nil)
		if err != nil {
			logrus.Error("Error caching message timestamp: ", err)
		}
	}
}

func (s SlackBackend) Sanitize(m *message.Message) *message.Message {
	// Do nothing
	return m
}
