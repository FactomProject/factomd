package engine

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Messages
	RepeatMsgs = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_msg_replay_toss_total",
		Help: "Number of repeated msgs.",
	})

	BroadInCastQueue = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_broadcast_in_current",
		Help: "Number of msgs in broadcastin queue.",
	})
)

var registered = false

// RegisterPrometheus registers the variables to be exposed. This can only be run once, hence the
// boolean flag to prevent panics if launched more than once. This is called in NetStart
func RegisterPrometheus() {
	if registered {
		return
	}
	registered = true

	prometheus.MustRegister(RepeatMsgs)
	prometheus.MustRegister(BroadInCastQueue)
}
