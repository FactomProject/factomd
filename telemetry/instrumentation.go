package telemetry

var (
	// Top Level Metrics
	MapSize = NewGaugeVec(
		"map_size",
		"common metric for all queues",
		[]string{"package", "name", "thread", "message"},
	)

	ChannelSize = NewGaugeVec(
		"channel_size",
		"common metric for all channels",
		[]string{"package", "name", "thread", "message"},
	)

	TotalCounter = NewCounterVec(
		"total_count",
		"common metric for counting totals",
		[]string{"package", "name", "thread", "message"},
	)
)
