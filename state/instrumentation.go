package state

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// 		Example
	//stateRandomCounter = prometheus.NewCounter(prometheus.CounterOpts{
	//	Name: "factomd_state_randomcounter_total",
	//	Help: "Just a basic counter that can only go up",
	//})
	//

	// Entry Syncing Controller
	ESMissingQueue = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_es_missing_entry_queue",
		Help: "Number of known missing entries in our queue to find.",
	})
	ESMissing = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_es_missing_entries",
		Help: "Number of known missing entries",
	})
	ESFound = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_es_found_entries",
		Help: "Number of known missing entries found.",
	})
	ESAsking = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_es_asking_missing_entries",
		Help: "Number we are asking for of the known missing entries.",
	})
	ESHighestAsking = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_es_highest_asking_entries",
		Help: "Highest entry DBHeight which has has a request made.",
	})
	ESHighestMissing = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_es_highest_missing_entries",
		Help: "Highest DBHeight of the entries we know are missing.",
	})
	ESFirstMissing = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_es_first_missing_entries",
		Help: "First DBHeight with a missing entry",
	})
	ESDBHTComplete = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_es_entry_dbheight_complete",
		Help: "First DBHeight with a missing entry",
	})
	ESAvgRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_es_average_requests",
		Help: "Average number of times we have had to request a missing entry",
	})
	HighestAck = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_highest_ack",
		Help: "Acknowledgement with the highest directory block height",
	})
	HighestKnown = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_highest_known",
		Help: "Highest known block (which can be different than the highest ack)",
	})
	HighestSaved = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_highest_saved",
		Help: "Highest saved block to the database",
	})
	HighestCompleted = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_highest_completed",
		Help: "Highest completed block, which may or may not be saved to the database",
	})

	// TPS
	TotalTransactionPerSecond = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_txrate_total_tps",
		Help: "Total transactions over life of node",
	})

	InstantTransactionPerSecond = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_txrate_instant_tps",
		Help: "Total transactions over life of node weighted for last 3 seconds",
	})
)

var registered bool = false

// RegisterPrometheus registers the variables to be exposed. This can only be run once, hence the
// boolean flag to prevent panics if launched more than once. This is called in NetStart
func RegisterPrometheus() {
	if registered {
		return
	}
	registered = true
	// 		Exmaple Cont.
	// prometheus.MustRegister(stateRandomCounter)

	// Entry syncing
	prometheus.MustRegister(ESAsking)
	prometheus.MustRegister(ESHighestAsking)
	prometheus.MustRegister(ESFirstMissing)
	prometheus.MustRegister(ESMissing)
	prometheus.MustRegister(ESFound)
	prometheus.MustRegister(ESDBHTComplete)
	prometheus.MustRegister(ESMissingQueue)
	prometheus.MustRegister(ESHighestMissing)
	prometheus.MustRegister(ESAvgRequests)
	prometheus.MustRegister(HighestAck)
	prometheus.MustRegister(HighestKnown)
	prometheus.MustRegister(HighestSaved)
	prometheus.MustRegister(HighestCompleted)

	// TPS
	prometheus.MustRegister(TotalTransactionPerSecond)
	prometheus.MustRegister(InstantTransactionPerSecond)
}
