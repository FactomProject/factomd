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

	// Entry Syncing
	stateEntrySyncWriteEntryCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_entrysyncing_write_entry_total", // Name used in Grafana
		Help: "Increment each time an entry is written to the database through " +
			"the 'GoWriteEntries' routine",
	})

	stateEntrySyncRequestEntryCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_entrysyncing_entry_request_total",
		Help: "Increment for each missing entry request sent out.",
	})

	stateEntrySyncMakeMissingEntryRequestsLoopTime = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_state_entrysyncing_missing_entry_loop_ns",
		Help: "Record the time is takes for the main loop in MakeMissingEntryRequests to complete.",
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
	prometheus.MustRegister(stateEntrySyncWriteEntryCounter)
	prometheus.MustRegister(stateEntrySyncRequestEntryCounter)
	prometheus.MustRegister(stateEntrySyncMakeMissingEntryRequestsLoopTime)
}
