package state

import (
	"sync"
	"time"
)

// IPrometheusRateMethods indicated which prometheus counters/gauges to set
type IPrometheusRateMethods interface {
	// Arrival
	SetArrivalInstantAvg(v float64)
	SetArrivalTotalAvg(v float64)
	SetArrivalBackup(v float64)
	SetMovingArrival(v float64)

	// Complete
	SetCompleteInstantAvg(v float64)
	SetCompleteTotalAvg(v float64)
	SetMovingComplete(v float64)
}

// RateCalculator will maintain the rate of msgs arriving and rate of msgs
// leaving a queue. The instant rate is a 2s avg
type RateCalculator struct {
	// Accessed on potentially Multiple Threads
	prometheusMethods IPrometheusRateMethods
	arrival           int32
	completed         int32
	line              int32

	// Single threaded
	tickerTime      time.Duration
	rollingArrival  *MovingAverage
	rollingComplete *MovingAverage

	mtx sync.Mutex
}

// NewRateCalculatorTime is good for unit tests, or if you want to change the measureing time
func NewRateCalculatorTime(p IPrometheusRateMethods, td time.Duration) *RateCalculator {
	r := new(RateCalculator)
	r.prometheusMethods = p
	r.tickerTime = td

	r.rollingArrival = NewMovingAverage(10)
	r.rollingComplete = NewMovingAverage(10)

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
	// Every 2 seconds calculate the instant rate and adjust the total avg
	for range ticker.C {
		r.mtx.Lock()

		totalArrival += r.arrival
		totalComplete += r.completed

		r.rollingArrival.Add(float64(r.arrival))
		r.rollingComplete.Add(float64(r.completed))

		// Calculate Total Avg
		totalTime := time.Since(start).Seconds()
		r.prometheusMethods.SetArrivalTotalAvg(float64(totalArrival) / totalTime)
		r.prometheusMethods.SetCompleteTotalAvg(float64(totalComplete) / totalTime)

		// Calculate 2s Avg
		r.prometheusMethods.SetArrivalInstantAvg(float64(r.arrival) / r.tickerTime.Seconds())
		r.prometheusMethods.SetCompleteInstantAvg(float64(r.completed) / r.tickerTime.Seconds())

		// Moving Avg
		r.prometheusMethods.SetMovingArrival(r.rollingArrival.Avg() / r.tickerTime.Seconds())
		r.prometheusMethods.SetMovingComplete(r.rollingComplete.Avg() / r.tickerTime.Seconds())

		// Set the backup
		r.prometheusMethods.SetArrivalBackup(float64(r.line))

		r.arrival = 0
		r.completed = 0

		r.mtx.Unlock()
	}
}

// Arrival indicates a new item added to the queue
func (r *RateCalculator) Arrival() {
	r.mtx.Lock()
	r.arrival++
	r.line++
	r.mtx.Unlock()
}

// Complete indicates something left the queue
func (r *RateCalculator) Complete() {
	r.mtx.Lock()
	r.completed++
	r.line--
	r.mtx.Unlock()
}

type MovingAverage struct {
	Window      int
	values      []float64
	valPos      int
	slotsFilled bool
}

func (ma *MovingAverage) Avg() float64 {
	var sum = float64(0)
	var c = ma.Window - 1

	// Are all slots filled? If not, ignore unused
	if !ma.slotsFilled {
		c = ma.valPos - 1
		if c < 0 {
			// Empty register
			return 0
		}
	}

	// Sum values
	var ic = 0
	for i := 0; i <= c; i++ {
		sum += ma.values[i]
		ic++
	}

	// Finalize average and return
	avg := sum / float64(ic)
	return avg
}

func (ma *MovingAverage) Add(val float64) {
	// Put into values array
	ma.values[ma.valPos] = val

	// Increment value position
	ma.valPos = (ma.valPos + 1) % ma.Window

	if !ma.slotsFilled && ma.valPos == 0 {
		ma.slotsFilled = true
	}
}

func NewMovingAverage(window int) *MovingAverage {
	return &MovingAverage{
		Window:      window,
		values:      make([]float64, window),
		valPos:      0,
		slotsFilled: false,
	}
}
