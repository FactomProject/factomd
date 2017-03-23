package p2p

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Connection Controller
	p2pControllerNumConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_p2p_controller_connections_current",
		Help: "Number of current connections",
	})

	p2pControllerNumMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_p2p_controller_metrics_current",
		Help: "Number of current connection metrics",
	})

	// Connection Routines
	p2pProcessSendsGuage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_p2p_connection_process_sends_routine_guage",
		Help: "Number of current processSend routines",
	})

	p2pProcessReceivesGuage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_p2p_connection_process_receives_routine_guage",
		Help: "Number of current processReceive routines",
	})

	p2pConnectionsRunLoop = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_p2p_connection_runloop_routine_gauge",
		Help: "The number of connections in runloop",
	})

	p2pConnectionDialLoop = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_p2p_connection_dialloop_routine_guage",
		Help: "The number of connections in dialloop",
	})

	// Runloops
	p2pConnectionRunLoopInitalized = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_p2p_connection_runloop_initialized_counter",
		Help: "Numer of runloops that hit initialized",
	})

	p2pConnectionRunLoopOnline = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_p2p_connection_runloop_online_counter",
		Help: "Numer of runloops that hit online",
	})

	p2pConnectionRunLoopOffline = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_p2p_connection_runloop_offline_counter",
		Help: "Numer of runloops that hit offline",
	})

	p2pConnectionRunLoopShutdown = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_p2p_connection_runloop_shutdown_counter",
		Help: "Numer of runloops that hit shutdown",
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

	// Controller
	prometheus.MustRegister(p2pControllerNumConnections)
	prometheus.MustRegister(p2pControllerNumMetrics)

	// Connection Routines
	prometheus.MustRegister(p2pProcessSendsGuage)    // processSends
	prometheus.MustRegister(p2pProcessReceivesGuage) // processReceives
	prometheus.MustRegister(p2pConnectionsRunLoop)
	prometheus.MustRegister(p2pConnectionDialLoop)

	// RunLoop
	prometheus.MustRegister(p2pConnectionRunLoopInitalized)
	prometheus.MustRegister(p2pConnectionRunLoopOnline)
	prometheus.MustRegister(p2pConnectionRunLoopOffline)
	prometheus.MustRegister(p2pConnectionRunLoopShutdown)
}
