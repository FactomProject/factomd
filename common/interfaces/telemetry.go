package interfaces

import "time"

type PollMetricHandler func(*time.Ticker, chan bool)
type MetricHandler func(PollMetricHandler)
