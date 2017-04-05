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
}
