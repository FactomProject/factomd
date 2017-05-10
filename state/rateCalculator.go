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
}

func NewRateCalculator(p IPrometheusRateMethods) *RateCalculator {
	r := new(RateCalculator)
	r.prometheusMethods = p

	r.arrival = new(int32)
	r.completed = new(int32)
	r.line = new(int32)
	return r
}

// Start begins instrumentation
func (r *RateCalculator) Start() {
	var totalArrival int32 = 0
	var totalComplete int32 = 0
	start := time.Now()

	// Every 2 seconds caluclate the instant rate and adjust the total avg
	ticker := time.NewTicker(time.Second * 2)
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
		r.prometheusMethods.SetArrivalWeightedAvg(float64(ca) / 2)
		r.prometheusMethods.SetCompleteWeightedAvg(float64(cc) / 2)

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
