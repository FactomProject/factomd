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

	p2pControllerNumConnectionsByAddress = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_p2p_controller_connectionsbyaddress_current",
		Help: "Number of current connections by address",
	})

	SentToPeers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_number_of_peers_broadcast",
		Help: "Number of Peers to which we are broadcasting messages",
	})

	StartingPoint = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_StartingPoint_peers_broadcast",
		Help: "Number of msgs broadcasting",
	})

	DirectSendsVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "factomd_state_direct_sends",
		Help: "Number of messages to sent as directed to peers",
	}, []string{"peer"})

	RandomDirectSendsVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "factomd_state_random_direct_sends",
		Help: "Number of messages to sent as directed to peers",
	}, []string{"peer"})

	ConnecitonSendQueueSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "factomd_state_connection_send_queue",
		Help: "Number of messages in send queue",
	}, []string{"peer"})

	//
	// Connection Routines
	p2pProcessSendsGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_p2p_connection_process_sends_routine_gauge",
		Help: "Number of current processSend routines",
	})

	p2pProcessReceivesGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_p2p_connection_process_receives_routine_gauge",
		Help: "Number of current processReceive routines",
	})

	p2pConnectionsRunLoop = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_p2p_connection_runloop_routine_gauge",
		Help: "The number of connections in runloop",
	})

	p2pConnectionDialLoop = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_p2p_connection_dialloop_routine_gauge",
		Help: "The number of connections in dialloop",
	})

	//
	// Runloops
	p2pConnectionRunLoopInitialized = prometheus.NewGauge(prometheus.GaugeOpts{
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

	//
	// Connections
	p2pConnectionCommonInit = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_p2p_connection_commonInit_calls_total",
		Help: "Number of times the commonInit() is called",
	})

	p2pConnectionOnlineCall = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_p2p_goOnline_total",
		Help: "Number of times we call goOnline()",
	})

	p2pConnectionOfflineCall = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_p2p_goOffline_total",
		Help: "Number of times we call goOffline()",
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
	prometheus.MustRegister(p2pControllerNumConnectionsByAddress)
	prometheus.MustRegister(SentToPeers)
	prometheus.MustRegister(StartingPoint)
	prometheus.MustRegister(DirectSendsVec)
	prometheus.MustRegister(RandomDirectSendsVec)
	prometheus.MustRegister(ConnecitonSendQueueSize)

	// Connection Routines
	prometheus.MustRegister(p2pProcessSendsGauge)    // processSends
	prometheus.MustRegister(p2pProcessReceivesGauge) // processReceives
	prometheus.MustRegister(p2pConnectionsRunLoop)
	prometheus.MustRegister(p2pConnectionDialLoop)
	prometheus.MustRegister(p2pConnectionOnlineCall)
	prometheus.MustRegister(p2pConnectionOfflineCall)

	// RunLoop
	prometheus.MustRegister(p2pConnectionRunLoopInitialized)
	prometheus.MustRegister(p2pConnectionRunLoopOnline)
	prometheus.MustRegister(p2pConnectionRunLoopOffline)
	prometheus.MustRegister(p2pConnectionRunLoopShutdown)

	// Connections
	prometheus.MustRegister(p2pConnectionCommonInit)

}
