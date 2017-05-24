package state

import (
	"sync/atomic"
	"time"
)

// IPrometheusRateMethods indicated which prometheus counters/gauges to set
type IPrometheusRateMethods interface {
	// Arrival
	SetArrivalWeightedAvg(v float64)
	SetArrivalTotalAvg(v float64)
	SetArrivalBackup(v float64)

	// Complete
	SetCompleteWeightedAvg(v float64)
	SetCompleteTotalAvg(v float64)
}

// RateCalculator will maintain the rate of msgs arriving and rate of msgs
// leaving a queue. The instant rate is a 2s avg
type RateCalculator struct {
	prometheusMethods IPrometheusRateMethods
	arrival           *int32
	completed         *int32
	line              *int32

	tickerTime time.Duration
}

// NewRateCalculatorTime is good for unit tests, or if you want to change the measureing time
func NewRateCalculatorTime(p IPrometheusRateMethods, td time.Duration) *RateCalculator {
	r := new(RateCalculator)
	r.prometheusMethods = p

	r.arrival = new(int32)
	r.completed = new(int32)
	r.line = new(int32)
	r.tickerTime = td

	return r
}

func NewRateCalculator(p IPrometheusRateMethods) *RateCalculator {
	return NewRateCalculatorTime(p, time.Duration(2*time.Second))
}

// Start begins instrumentation
func (r *RateCalculator) Start() {
	r.StartTime(time.Now())
}

// StartTime is good for unit tests
func (r *RateCalculator) StartTime(start time.Time) {
	var totalArrival int32 = 0
	var totalComplete int32 = 0

	ticker := time.NewTicker(r.tickerTime)
	// Every 2 seconds caluclate the instant rate and adjust the total avg
	for _ = range ticker.C {
		na, nc := int32(0), int32(0)

		//
		// Grab the current values and reset
		ca := atomic.SwapInt32(r.arrival, na)
		cc := atomic.SwapInt32(r.completed, nc)
		cl := atomic.LoadInt32(r.line)

		totalArrival += ca
		totalComplete += cc

		// Calculate Total Avg
		totalTime := time.Since(start).Seconds()
		r.prometheusMethods.SetArrivalTotalAvg(float64(totalArrival) / totalTime)
		r.prometheusMethods.SetCompleteTotalAvg(float64(totalComplete) / totalTime)

		// Calculate 2s Avg
		r.prometheusMethods.SetArrivalWeightedAvg(float64(ca) / r.tickerTime.Seconds())
		r.prometheusMethods.SetCompleteWeightedAvg(float64(cc) / r.tickerTime.Seconds())

		// Set the backup
		r.prometheusMethods.SetArrivalBackup(float64(cl))
	}
}

// Arrival indicates a new item added to the queue
func (r *RateCalculator) Arrival() {
	atomic.AddInt32(r.arrival, 1)
	atomic.AddInt32(r.line, 1)
}

// Complete indicates something left the queue
func (r *RateCalculator) Complete() {
	atomic.AddInt32(r.completed, 1)
	atomic.AddInt32(r.line, -1)
}
