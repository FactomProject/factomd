package telemetry

var (
	// Top Level Metrics
	MapSize = NewGaugeVec(
		"map_size",
		"common metric for all queues",
		[]string{"path", "message"},
	)

	ChannelSize = NewGaugeVec(
		"channel_size",
		"common metric for all channels",
		[]string{"path", "message"},
	)

	TotalCounter = NewCounterVec(
		"total_count",
		"common metric for counting totals",
		[]string{"path", "message"},
	)
)
