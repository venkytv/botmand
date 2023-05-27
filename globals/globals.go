package globals

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	BotName      = "botmand"
	BotUrlScheme = BotName + "://"
)

var (
	NumExecEngineFactories = promauto.NewGauge(prometheus.GaugeOpts{
		Name: BotName + "_exec_engines_total",
		Help: "Total number of exec engines.",
	})

	NumConversationTriggers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: BotName + "_conversation_triggers_total",
		Help: "Total number of conversation triggers.",
	})

	NumConversations = promauto.NewGauge(prometheus.GaugeOpts{
		Name: BotName + "_conversations_total",
		Help: "Total number of current conversations.",
	})

	NumThreadedConversations = promauto.NewGauge(prometheus.GaugeOpts{
		Name: BotName + "_threaded_conversations_total",
		Help: "Total number of current threaded conversations.",
	})

	NumChannelConversations = promauto.NewGauge(prometheus.GaugeOpts{
		Name: BotName + "_channel_conversations_total",
		Help: "Total number of current channel conversations.",
	})
)
