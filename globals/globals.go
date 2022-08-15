package globals

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	BotName = "teabot"
)

var (
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
