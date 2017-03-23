package p2p

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Connection Routines
	p2pProcessSendsGuage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_p2p_connection_process_sends_routine_guage",
		Help: "Number of current processSend routines",
	})

	p2pProcessReceivesGuage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_p2p_connection_process_receives_routine_guage",
		Help: "Number of current processReceive routines",
	})

	p2pConnectionCommands = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_p2p_connection_commands_gauge",
		Help: "The number of commands in a command queue of a connection",
	})
)

var registered = false

// RegisterPrometheus registers the variables to be exposed. This can only be run once, hence the
// boolean flag to prevent panics if launched more than once. This is called in NetStart
func RegisterPrometheus() {
	if registered {
		return
	}
	registered = true

	// Connection Routines
	prometheus.MustRegister(p2pProcessSendsGuage)    // processSends
	prometheus.MustRegister(p2pProcessReceivesGuage) // processReceives
	prometheus.MustRegister(p2pConnectionCommands)
}
