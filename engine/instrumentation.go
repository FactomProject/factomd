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

	// Etcd
	EtcdSendOutTime = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_state_etcd_send_ns",
		Help: "Time it takes to complete an etcd send",
	})
	EtcdSendCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_etcd_send_count",
		Help: "Count of all messages sent through etcd",
	})
	EtcdGetCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_etcd_get_count",
		Help: "Count of all messages got through etcd",
	})

	// Non-etcd
	NonEtcdSendCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_non_etcd_send_count",
		Help: "Count of all messages sent NOT through etcd",
	})
	NonEtcdGetCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_non_etcd_get_count",
		Help: "Count of all messages got NOT through etcd",
	})

	// Send/Receive Times
	TotalSendTime = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_total_send_time",
		Help: "Time spent sending (nanoseconds)",
	})
	TotalReceiveTime = prometheus.NewCounter(prometheus.CounterOpts{
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

	// Etcd
	prometheus.MustRegister(EtcdSendOutTime)
	prometheus.MustRegister(EtcdSendCount)
	prometheus.MustRegister(EtcdGetCount)

	// Non-etcd
	prometheus.MustRegister(NonEtcdSendCount)
	prometheus.MustRegister(NonEtcdGetCount)

	// Send/Receive Times
	prometheus.MustRegister(TotalSendTime)
	prometheus.MustRegister(TotalReceiveTime)
}
