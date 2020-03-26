package p2p

import "github.com/FactomProject/factomd/modules/telemetry"

var (
	// Connection Controller
	p2pControllerNumConnections = telemetry.NewGauge(
		"factomd_p2p_controller_connections_current",
		"Number of current connections",
	)

	p2pControllerNumMetrics = telemetry.NewGauge(
		"factomd_p2p_controller_metrics_current",
		"Number of current connection metrics",
	)

	p2pControllerNumConnectionsByAddress = telemetry.NewGauge(
		"factomd_p2p_controller_connectionsbyaddress_current",
		"Number of current connections by address",
	)

	SentToPeers = telemetry.NewGauge(
		"factomd_state_number_of_peers_broadcast",
		"Number of Peers to which we are broadcasting messages",
	)

	//
	// Connection Routines
	p2pProcessSendsGauge = telemetry.NewGauge(
		"factomd_p2p_connection_process_sends_routine_gauge",
		"Number of current processSend routines",
	)

	p2pProcessReceivesGauge = telemetry.NewGauge(
		"factomd_p2p_connection_process_receives_routine_gauge",
		"Number of current processReceive routines",
	)

	p2pConnectionsRunLoop = telemetry.NewGauge(
		"factomd_p2p_connection_runloop_routine_gauge",
		"The number of connections in runloop",
	)

	p2pConnectionDialLoop = telemetry.NewGauge(
		"factomd_p2p_connection_dialloop_routine_gauge",
		"The number of connections in dialloop",
	)

	//
	// Runloops
	p2pConnectionRunLoopInitialized = telemetry.NewGauge(
		"factomd_p2p_connection_runloop_initialized_counter",
		"Numer of runloops that hit initialized",
	)

	p2pConnectionRunLoopOnline = telemetry.NewGauge(
		"factomd_p2p_connection_runloop_online_counter",
		"Numer of runloops that hit online",
	)

	p2pConnectionRunLoopOffline = telemetry.NewGauge(
		"factomd_p2p_connection_runloop_offline_counter",
		"Numer of runloops that hit offline",
	)

	p2pConnectionRunLoopShutdown = telemetry.NewGauge(
		"factomd_p2p_connection_runloop_shutdown_counter",
		"Numer of runloops that hit shutdown",
	)

	//
	// Connections
	p2pConnectionCommonInit = telemetry.NewCounter(
		"factomd_p2p_connection_commonInit_calls_total",
		"Number of times the commonInit() is called",
	)

	p2pConnectionOnlineCall = telemetry.NewCounter(
		"factomd_p2p_goOnline_total",
		"Number of times we call goOnline()",
	)

	p2pConnectionOfflineCall = telemetry.NewCounter(
		"factomd_p2p_goOffline_total",
		"Number of times we call goOffline()",
	)
)
