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

	BroadCastInQueueDrop = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_broadcast_in_drop_total",
		Help: "How many messages are dropped due to full queues",
	})

	// NetworkReplayFilter
	TotalNetworkReplayFilter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_network_replay_filter_total",
		Help: "Tally of total messages gone into NetworkReplayFilter",
	})
	TotalNetworkAckReplayFilter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_network_ack_replay_filter_total",
		Help: "Tally of total messages gone into NetworkAckReplayFilter",
	})

	// Network Out Queue
	NetworkOutTotalDequeue = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_netoutmsg_total_general",
		Help: "Count of all messages being dequeued",
	})

	// Send/Receive Times
	TotalSendTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_total_send_time",
		Help: "Time spent sending (nanoseconds)",
	})
	TotalReceiveTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_total_receive_time",
		Help: "Time spent receiving (nanoseconds)",
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
	prometheus.MustRegister(BroadCastInQueueDrop)

	// NetworkReplayFilter
	prometheus.MustRegister(TotalNetworkReplayFilter)
	prometheus.MustRegister(TotalNetworkAckReplayFilter)

	// NetOut
	prometheus.MustRegister(NetworkOutTotalDequeue)

	// Send/Receive Times
	prometheus.MustRegister(TotalSendTime)
	prometheus.MustRegister(TotalReceiveTime)
}
