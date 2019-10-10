package p2p

import "github.com/FactomProject/factomd/telemetry"

var gauge = telemetry.RegisterMetric.Gauge
var counter = telemetry.RegisterMetric.Counter

var (
	// Connection Controller
	p2pControllerNumConnections = gauge(
		"factomd_p2p_controller_connections_current",
		"Number of current connections",
	)

	p2pControllerNumMetrics = gauge(
		"factomd_p2p_controller_metrics_current",
		"Number of current connection metrics",
	)

	p2pControllerNumConnectionsByAddress = gauge(
		"factomd_p2p_controller_connectionsbyaddress_current",
		"Number of current connections by address",
	)

	SentToPeers = gauge(
		"factomd_state_number_of_peers_broadcast",
		"Number of Peers to which we are broadcasting messages",
	)

	StartingPoint = gauge(
		"factomd_StartingPoint_peers_broadcast",
		"Number of msgs broadcasting",
	)

	//
	// Connection Routines
	p2pProcessSendsGauge = gauge(
		"factomd_p2p_connection_process_sends_routine_gauge",
		"Number of current processSend routines",
	)

	p2pProcessReceivesGauge = gauge(
		"factomd_p2p_connection_process_receives_routine_gauge",
		"Number of current processReceive routines",
	)

	p2pConnectionsRunLoop = gauge(
		"factomd_p2p_connection_runloop_routine_gauge",
		"The number of connections in runloop",
	)

	p2pConnectionDialLoop = gauge(
		"factomd_p2p_connection_dialloop_routine_gauge",
		"The number of connections in dialloop",
	)

	//
	// Runloops
	p2pConnectionRunLoopInitialized = gauge(
		"factomd_p2p_connection_runloop_initialized_counter",
		"Numer of runloops that hit initialized",
	)

	p2pConnectionRunLoopOnline = gauge(
		"factomd_p2p_connection_runloop_online_counter",
		"Numer of runloops that hit online",
	)

	p2pConnectionRunLoopOffline = gauge(
		"factomd_p2p_connection_runloop_offline_counter",
		"Numer of runloops that hit offline",
	)

	p2pConnectionRunLoopShutdown = gauge(
		"factomd_p2p_connection_runloop_shutdown_counter",
		"Numer of runloops that hit shutdown",
	)

	//
	// Connections
	p2pConnectionCommonInit = counter(
		"factomd_p2p_connection_commonInit_calls_total",
		"Number of times the commonInit() is called",
	)

	p2pConnectionOnlineCall = counter(
		"factomd_p2p_goOnline_total",
		"Number of times we call goOnline()",
	)

	p2pConnectionOfflineCall = counter(
		"factomd_p2p_goOffline_total",
		"Number of times we call goOffline()",
	)
)
