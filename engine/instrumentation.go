package engine

import "github.com/FactomProject/factomd/modules/telemetry"

var (
	// Messages
	RepeatMsgs = telemetry.NewCounter(
		"factomd_state_msg_replay_toss_total",
		"Number of repeated msgs.",
	)

	BroadInCastQueue = telemetry.NewGauge(
		"factomd_state_broadcast_in_current",
		"Number of msgs in broadcastin queue.",
	)

	BroadCastInQueueDrop = telemetry.NewCounter(
		"factomd_state_broadcast_in_drop_total",
		"How many messages are dropped due to full queues",
	)

	// Network Out Queue
	NetworkOutTotalDequeue = telemetry.NewCounter(
		"factomd_state_queue_netoutmsg_total_general",
		"Count of all messages being dequeued",
	)

	// Send/Read Times
	TotalSendTime = telemetry.NewGauge(
		"factomd_state_total_send_time",
		"Time spent sending (nanoseconds)",
	)
)
