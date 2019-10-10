package engine

import "github.com/FactomProject/factomd/telemetry"

var gauge = telemetry.RegisterMetric.Gauge
var counter = telemetry.RegisterMetric.Counter

var (
	// Messages
	RepeatMsgs = counter(
		"factomd_state_msg_replay_toss_total",
		"Number of repeated msgs.",
	)

	BroadInCastQueue = gauge(
		"factomd_state_broadcast_in_current",
		"Number of msgs in broadcastin queue.",
	)

	BroadCastInQueueDrop = counter(
		"factomd_state_broadcast_in_drop_total",
		"How many messages are dropped due to full queues",
	)

	// NetworkReplayFilter
	TotalNetworkReplayFilter = counter(
		"factomd_state_network_replay_filter_total",
		"Tally of total messages gone into NetworkReplayFilter",
	)
	TotalNetworkAckReplayFilter = counter(
		"factomd_state_network_ack_replay_filter_total",
		"Tally of total messages gone into NetworkAckReplayFilter",
	)

	// Network Out Queue
	NetworkOutTotalDequeue = counter(
		"factomd_state_queue_netoutmsg_total_general",
		"Count of all messages being dequeued",
	)

	// Send/Receive Times
	TotalSendTime = gauge(
		"factomd_state_total_send_time",
		"Time spent sending (nanoseconds)",
	)
	TotalReceiveTime = gauge(
		"factomd_state_total_receive_time",
		"Time spent receiving (nanoseconds)",
	)
)
