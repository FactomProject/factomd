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

	// Queues
	//	InMsg
	TotalMessageQueueInMsgGeneral = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_general_inmsg",
		Help: "Instrumenting the netoutmsg queue",
	})
	CurrentMessageQueueInMsgEOM = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_inmsg_eom",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgACK = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_inmsg_ack",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgAudFault = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_inmsg_audfault",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgFedFault = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_inmsg_fedfault",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgFullFault = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_inmsg_fullfault",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgCommitChain = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_inmsg_commitchain",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgCommitEntry = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_inmsg_commitentry",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgDBSig = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_inmsg_dbsig",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgEOMTimeout = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_inmsg_eomtimeout",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgFactTX = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_inmsg_facttx",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgHeartbeat = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_inmsg_heatbeat",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgEtcdHashPickup = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_inmsg_etcdpickup",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgMissingMsg = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_inmsg_mmsg",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgMissingMsgResp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_inmsg_mmsgresp",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgMissingData = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_inmsg_missingdata",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgMissingDataResp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_inmsg_missingdataresp",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgRevealEntry = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_inmsg_revealentry",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgReqBlock = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_inmsg_reqblock",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgDbStateMissing = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_inmsg_dbstatemissing",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgDbState = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_inmsg_dbstate",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgBounceMsg = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_inmsg_bounce",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgBounceResp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_inmsg_bounceresp",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgMisc = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_inmsg_misc",
		Help: "Instrumenting the inmsg queue",
	})

	TotalMessageQueueInMsgEOM = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_inmsg_eom",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgACK = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_inmsg_ack",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgAudFault = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_inmsg_audfault",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgFedFault = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_inmsg_fedfault",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgFullFault = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_inmsg_fullfault",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgCommitChain = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_inmsg_commitchain",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgCommitEntry = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_inmsg_commitentry",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgDBSig = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_inmsg_dbsig",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgEOMTimeout = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_inmsg_eomtimeout",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgFactTX = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_inmsg_facttx",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgHeartbeat = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_inmsg_heatbeat",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgEtcdHashPickup = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_inmsg_etcdpickup",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgMissingMsg = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_inmsg_mmsg",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgMissingMsgResp = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_inmsg_mmsgresp",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgMissingData = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_inmsg_missingdata",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgMissingDataResp = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_inmsg_missingdataresp",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgRevealEntry = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_inmsg_revealentry",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgReqBlock = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_inmsg_reqblock",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgDbStateMissing = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_inmsg_dbstatemissing",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgDbState = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_inmsg_dbstate",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgBounceMsg = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_inmsg_bounce",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgBounceResp = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_inmsg_bounceresp",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgMisc = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_inmsg_misc",
		Help: "Instrumenting the inmsg queue",
	})

	//	NetworkOutMsg
	TotalMessageQueueNetOutMsgGeneral = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_general_netoutmsg",
		Help: "Instrumenting the netoutmsg queue",
	})
	TotalMessageQueueNetOutMsgEOM = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_netoutmsg_eom",
		Help: "Instrumenting the netoutmsg queue",
	})
	TotalMessageQueueNetOutMsgACK = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_netoutmsg_ack",
		Help: "Instrumenting the netoutmsg queue",
	})
	TotalMessageQueueNetOutMsgAudFault = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_netoutmsg_audfault",
		Help: "Instrumenting the netoutmsg queue",
	})
	TotalMessageQueueNetOutMsgFedFault = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_netoutmsg_fedfault",
		Help: "Instrumenting the netoutmsg queue",
	})
	TotalMessageQueueNetOutMsgFullFault = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_netoutmsg_fullfault",
		Help: "Instrumenting the netoutmsg queue",
	})
	TotalMessageQueueNetOutMsgCommitChain = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_netoutmsg_commitchain",
		Help: "Instrumenting the netoutmsg queue",
	})
	TotalMessageQueueNetOutMsgCommitEntry = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_netoutmsg_commitentry",
		Help: "Instrumenting the netoutmsg queue",
	})
	TotalMessageQueueNetOutMsgDBSig = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_netoutmsg_dbsig",
		Help: "Instrumenting the netoutmsg queue",
	})
	TotalMessageQueueNetOutMsgEOMTimeout = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_netoutmsg_eomtimeout",
		Help: "Instrumenting the netoutmsg queue",
	})
	TotalMessageQueueNetOutMsgFactTX = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_netoutmsg_facttx",
		Help: "Instrumenting the netoutmsg queue",
	})
	TotalMessageQueueNetOutMsgHeartbeat = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_netoutmsg_heatbeat",
		Help: "Instrumenting the netoutmsg queue",
	})
	TotalMessageQueueNetOutMsgEtcdHashPickup = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_netoutmsg_etcdpickup",
		Help: "Instrumenting the netoutmsg queue",
	})
	TotalMessageQueueNetOutMsgMissingMsg = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_netoutmsg_mmsg",
		Help: "Instrumenting the netoutmsg queue",
	})
	TotalMessageQueueNetOutMsgMissingMsgResp = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_netoutmsg_mmsgresp",
		Help: "Instrumenting the netoutmsg queue",
	})
	TotalMessageQueueNetOutMsgMissingData = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_netoutmsg_missingdata",
		Help: "Instrumenting the netoutmsg queue",
	})
	TotalMessageQueueNetOutMsgMissingDataResp = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_netoutmsg_missingdataresp",
		Help: "Instrumenting the netoutmsg queue",
	})
	TotalMessageQueueNetOutMsgRevealEntry = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_netoutmsg_revealentry",
		Help: "Instrumenting the netoutmsg queue",
	})
	TotalMessageQueueNetOutMsgReqBlock = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_netoutmsg_reqblock",
		Help: "Instrumenting the netoutmsg queue",
	})
	TotalMessageQueueNetOutMsgDbStateMissing = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_netoutmsg_dbstatemissing",
		Help: "Instrumenting the netoutmsg queue",
	})
	TotalMessageQueueNetOutMsgDbState = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_netoutmsg_dbstate",
		Help: "Instrumenting the netoutmsg queue",
	})
	TotalMessageQueueNetOutMsgBounceMsg = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_netoutmsg_bounce",
		Help: "Instrumenting the netoutmsg queue",
	})
	TotalMessageQueueNetOutMsgBounceResp = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_netoutmsg_bounceresp",
		Help: "Instrumenting the netoutmsg queue",
	})
	TotalMessageQueueNetOutMsgMisc = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_netoutmsg_misc",
		Help: "Instrumenting the netoutmsg queue",
	})

	// InMsgQueue Rates
	InMsgTotalArrivalQueueRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_arrival_avg_total_inmsg",
		Help: "Total avg of inmsg queue arrival rate",
	})

	InMsgInstantArrivalQueueRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_arrival_avg_instant_inmsg",
		Help: "Instant avg of inmsg queue arrival rate",
	})

	InMsgMovingArrivalQueueRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_arrival_avg_moving_inmsg",
		Help: "Moving avg of inmsg queue arrival rate",
	})

	InMsgTotalCompleteQueueRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_complete_avg_total_inmsg",
		Help: "Total avg of inmsg queue complete rate",
	})

	InMsgInstantCompleteQueueRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_complete_avg_instant_inmsg",
		Help: "Instant avg of inmsg queue complete rate",
	})

	InMsgMovingCompleteQueueRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_complete_avg_moving_inmsg",
		Help: "Moving avg of inmsg queue complete rate",
	})

	InMsgQueueBackupRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_backup_inmsg",
		Help: "Backup of queue",
	})

	// NetOut Rates
	NetOutTotalArrivalQueueRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_arrival_avg_total_netout",
		Help: "Total avg of inmsg queue arrival rate",
	})

	NetOutInstantArrivalQueueRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_arrival_avg_instant_netout",
		Help: "Instant avg of inmsg queue arrival rate",
	})

	NetOutMovingArrivalQueueRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_arrival_avg_moving_netout",
		Help: "Moving avg of inmsg queue arrival rate",
	})

	NetOutTotalCompleteQueueRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_complete_avg_total_netout",
		Help: "Total avg of inmsg queue complete rate",
	})

	NetOutInstantCompleteQueueRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_complete_avg_instant_netout",
		Help: "Instant avg of inmsg queue complete rate",
	})

	NetOutMovingCompleteQueueRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_complete_avg_moving_netout",
		Help: "Moving avg of inmsg queue complete rate",
	})

	NetOutQueueBackupRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_backup_netout",
		Help: "Backup of queue",
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

	// Queues
	//	InMsg Current
	prometheus.MustRegister(CurrentMessageQueueInMsgEOM)
	prometheus.MustRegister(CurrentMessageQueueInMsgACK)
	prometheus.MustRegister(CurrentMessageQueueInMsgAudFault)
	prometheus.MustRegister(CurrentMessageQueueInMsgFedFault)
	prometheus.MustRegister(CurrentMessageQueueInMsgFullFault)
	prometheus.MustRegister(CurrentMessageQueueInMsgCommitChain)
	prometheus.MustRegister(CurrentMessageQueueInMsgCommitEntry)
	prometheus.MustRegister(CurrentMessageQueueInMsgDBSig)
	prometheus.MustRegister(CurrentMessageQueueInMsgEOMTimeout)
	prometheus.MustRegister(CurrentMessageQueueInMsgFactTX)
	prometheus.MustRegister(CurrentMessageQueueInMsgHeartbeat)
	prometheus.MustRegister(CurrentMessageQueueInMsgEtcdHashPickup)
	prometheus.MustRegister(CurrentMessageQueueInMsgMissingMsg)
	prometheus.MustRegister(CurrentMessageQueueInMsgMissingMsgResp)
	prometheus.MustRegister(CurrentMessageQueueInMsgMissingData)
	prometheus.MustRegister(CurrentMessageQueueInMsgMissingDataResp)
	prometheus.MustRegister(CurrentMessageQueueInMsgRevealEntry)
	prometheus.MustRegister(CurrentMessageQueueInMsgReqBlock)
	prometheus.MustRegister(CurrentMessageQueueInMsgDbStateMissing)
	prometheus.MustRegister(CurrentMessageQueueInMsgDbState)
	prometheus.MustRegister(CurrentMessageQueueInMsgBounceMsg)
	prometheus.MustRegister(CurrentMessageQueueInMsgBounceResp)
	prometheus.MustRegister(CurrentMessageQueueInMsgMisc)
	//	InMsg Total
	prometheus.MustRegister(TotalMessageQueueInMsgEOM)
	prometheus.MustRegister(TotalMessageQueueInMsgACK)
	prometheus.MustRegister(TotalMessageQueueInMsgAudFault)
	prometheus.MustRegister(TotalMessageQueueInMsgFedFault)
	prometheus.MustRegister(TotalMessageQueueInMsgFullFault)
	prometheus.MustRegister(TotalMessageQueueInMsgCommitChain)
	prometheus.MustRegister(TotalMessageQueueInMsgCommitEntry)
	prometheus.MustRegister(TotalMessageQueueInMsgDBSig)
	prometheus.MustRegister(TotalMessageQueueInMsgEOMTimeout)
	prometheus.MustRegister(TotalMessageQueueInMsgFactTX)
	prometheus.MustRegister(TotalMessageQueueInMsgHeartbeat)
	prometheus.MustRegister(TotalMessageQueueInMsgEtcdHashPickup)
	prometheus.MustRegister(TotalMessageQueueInMsgMissingMsg)
	prometheus.MustRegister(TotalMessageQueueInMsgMissingMsgResp)
	prometheus.MustRegister(TotalMessageQueueInMsgMissingData)
	prometheus.MustRegister(TotalMessageQueueInMsgMissingDataResp)
	prometheus.MustRegister(TotalMessageQueueInMsgRevealEntry)
	prometheus.MustRegister(TotalMessageQueueInMsgReqBlock)
	prometheus.MustRegister(TotalMessageQueueInMsgDbStateMissing)
	prometheus.MustRegister(TotalMessageQueueInMsgDbState)
	prometheus.MustRegister(TotalMessageQueueInMsgBounceMsg)
	prometheus.MustRegister(TotalMessageQueueInMsgBounceResp)
	prometheus.MustRegister(TotalMessageQueueInMsgMisc)

	// Net Out
	prometheus.MustRegister(TotalMessageQueueNetOutMsgEOM)
	prometheus.MustRegister(TotalMessageQueueNetOutMsgACK)
	prometheus.MustRegister(TotalMessageQueueNetOutMsgAudFault)
	prometheus.MustRegister(TotalMessageQueueNetOutMsgFedFault)
	prometheus.MustRegister(TotalMessageQueueNetOutMsgFullFault)
	prometheus.MustRegister(TotalMessageQueueNetOutMsgCommitChain)
	prometheus.MustRegister(TotalMessageQueueNetOutMsgCommitEntry)
	prometheus.MustRegister(TotalMessageQueueNetOutMsgDBSig)
	prometheus.MustRegister(TotalMessageQueueNetOutMsgEOMTimeout)
	prometheus.MustRegister(TotalMessageQueueNetOutMsgFactTX)
	prometheus.MustRegister(TotalMessageQueueNetOutMsgHeartbeat)
	prometheus.MustRegister(TotalMessageQueueNetOutMsgEtcdHashPickup)
	prometheus.MustRegister(TotalMessageQueueNetOutMsgMissingMsg)
	prometheus.MustRegister(TotalMessageQueueNetOutMsgMissingMsgResp)
	prometheus.MustRegister(TotalMessageQueueNetOutMsgMissingData)
	prometheus.MustRegister(TotalMessageQueueNetOutMsgMissingDataResp)
	prometheus.MustRegister(TotalMessageQueueNetOutMsgRevealEntry)
	prometheus.MustRegister(TotalMessageQueueNetOutMsgReqBlock)
	prometheus.MustRegister(TotalMessageQueueNetOutMsgDbStateMissing)
	prometheus.MustRegister(TotalMessageQueueNetOutMsgDbState)
	prometheus.MustRegister(TotalMessageQueueNetOutMsgBounceMsg)
	prometheus.MustRegister(TotalMessageQueueNetOutMsgBounceResp)
	prometheus.MustRegister(TotalMessageQueueNetOutMsgMisc)

	// InMsgRate
	prometheus.MustRegister(InMsgTotalArrivalQueueRate)
	prometheus.MustRegister(InMsgInstantArrivalQueueRate)
	prometheus.MustRegister(InMsgTotalCompleteQueueRate)
	prometheus.MustRegister(InMsgInstantCompleteQueueRate)
	prometheus.MustRegister(InMsgQueueBackupRate)
	prometheus.MustRegister(InMsgMovingArrivalQueueRate)
	prometheus.MustRegister(InMsgMovingCompleteQueueRate)

	// NetOutRate
	prometheus.MustRegister(NetOutTotalArrivalQueueRate)
	prometheus.MustRegister(NetOutInstantArrivalQueueRate)
	prometheus.MustRegister(NetOutTotalCompleteQueueRate)
	prometheus.MustRegister(NetOutInstantCompleteQueueRate)
	prometheus.MustRegister(NetOutQueueBackupRate)
	prometheus.MustRegister(NetOutMovingArrivalQueueRate)
	prometheus.MustRegister(NetOutMovingCompleteQueueRate)

	prometheus.MustRegister(TotalMessageQueueInMsgGeneral)
	prometheus.MustRegister(TotalMessageQueueNetOutMsgGeneral)
}
