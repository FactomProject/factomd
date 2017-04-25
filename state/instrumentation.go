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
	CurrentMessageQueueInMsgQueueEOM = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_inmsg_eom",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgQueueACK = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_insmg_ack",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgQueueAudFault = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_insmg_audfault",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgQueueFedFault = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_insmg_fedfault",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgQueueFullFault = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_insmg_fullfault",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgQueueCommitChain = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_insmg_commitchain",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgQueueCommitEntry = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_insmg_commitentry",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgQueueDBSig = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_insmg_dbsig",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgQueueEOMTimeout = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_insmg_eomtimeout",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgQueueFactTX = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_insmg_facttx",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgQueueHeartbeat = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_insmg_heatbeat",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgQueueInvalidDB = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_insmg_invaliddb",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgQueueMissingMsg = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_insmg_mmsg",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgQueueMissingMsgResp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_insmg_mmsgresp",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgQueueMissingData = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_insmg_missingdata",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgQueueMissingDataResp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_insmg_missingdataresp",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgQueueRevealEntry = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_insmg_revealentry",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgQueueReqBlock = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_insmg_reqblock",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgQueueDbStateMissing = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_insmg_dbstatemissing",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgQueueDbState = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_insmg_dbstate",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgQueueBounceMsg = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_insmg_bounce",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgQueueBounceResp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_insmg_bounceresp",
		Help: "Instrumenting the inmsg queue",
	})
	CurrentMessageQueueInMsgQueueMisc = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_insmg_misc",
		Help: "Instrumenting the inmsg queue",
	})

	TotalMessageQueueInMsgQueueEOM = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_inmsg_eom",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgQueueACK = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_insmg_ack",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgQueueAudFault = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_insmg_audfault",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgQueueFedFault = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_insmg_fedfault",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgQueueFullFault = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_insmg_fullfault",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgQueueCommitChain = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_insmg_commitchain",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgQueueCommitEntry = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_insmg_commitentry",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgQueueDBSig = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_insmg_dbsig",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgQueueEOMTimeout = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_insmg_eomtimeout",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgQueueFactTX = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_insmg_facttx",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgQueueHeartbeat = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_insmg_heatbeat",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgQueueInvalidDB = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_insmg_invaliddb",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgQueueMissingMsg = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_insmg_mmsg",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgQueueMissingMsgResp = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_insmg_mmsgresp",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgQueueMissingData = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_insmg_missingdata",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgQueueMissingDataResp = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_insmg_missingdataresp",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgQueueRevealEntry = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_insmg_revealentry",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgQueueReqBlock = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_insmg_reqblock",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgQueueDbStateMissing = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_insmg_dbstatemissing",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgQueueDbState = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_insmg_dbstate",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgQueueBounceMsg = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_insmg_bounce",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgQueueBounceResp = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_insmg_bounceresp",
		Help: "Instrumenting the inmsg queue",
	})
	TotalMessageQueueInMsgQueueMisc = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_queue_total_insmg_misc",
		Help: "Instrumenting the inmsg queue",
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
	//	InMsg
	prometheus.MustRegister(CurrentMessageQueueInMsgQueueEOM)
	prometheus.MustRegister(CurrentMessageQueueInMsgQueueACK)
	prometheus.MustRegister(CurrentMessageQueueInMsgQueueAudFault)
	prometheus.MustRegister(CurrentMessageQueueInMsgQueueFedFault)
	prometheus.MustRegister(CurrentMessageQueueInMsgQueueFullFault)
	prometheus.MustRegister(CurrentMessageQueueInMsgQueueCommitChain)
	prometheus.MustRegister(CurrentMessageQueueInMsgQueueCommitEntry)
	prometheus.MustRegister(CurrentMessageQueueInMsgQueueDBSig)
	prometheus.MustRegister(CurrentMessageQueueInMsgQueueEOMTimeout)
	prometheus.MustRegister(CurrentMessageQueueInMsgQueueFactTX)
	prometheus.MustRegister(CurrentMessageQueueInMsgQueueHeartbeat)
	prometheus.MustRegister(CurrentMessageQueueInMsgQueueInvalidDB)
	prometheus.MustRegister(CurrentMessageQueueInMsgQueueMissingMsg)
	prometheus.MustRegister(CurrentMessageQueueInMsgQueueMissingMsgResp)
	prometheus.MustRegister(CurrentMessageQueueInMsgQueueMissingData)
	prometheus.MustRegister(CurrentMessageQueueInMsgQueueMissingDataResp)
	prometheus.MustRegister(CurrentMessageQueueInMsgQueueRevealEntry)
	prometheus.MustRegister(CurrentMessageQueueInMsgQueueReqBlock)
	prometheus.MustRegister(CurrentMessageQueueInMsgQueueDbStateMissing)
	prometheus.MustRegister(CurrentMessageQueueInMsgQueueDbState)
	prometheus.MustRegister(CurrentMessageQueueInMsgQueueBounceMsg)
	prometheus.MustRegister(CurrentMessageQueueInMsgQueueBounceResp)
	prometheus.MustRegister(CurrentMessageQueueInMsgQueueMisc)

	prometheus.MustRegister(TotalMessageQueueInMsgQueueEOM)
	prometheus.MustRegister(TotalMessageQueueInMsgQueueACK)
	prometheus.MustRegister(TotalMessageQueueInMsgQueueAudFault)
	prometheus.MustRegister(TotalMessageQueueInMsgQueueFedFault)
	prometheus.MustRegister(TotalMessageQueueInMsgQueueFullFault)
	prometheus.MustRegister(TotalMessageQueueInMsgQueueCommitChain)
	prometheus.MustRegister(TotalMessageQueueInMsgQueueCommitEntry)
	prometheus.MustRegister(TotalMessageQueueInMsgQueueDBSig)
	prometheus.MustRegister(TotalMessageQueueInMsgQueueEOMTimeout)
	prometheus.MustRegister(TotalMessageQueueInMsgQueueFactTX)
	prometheus.MustRegister(TotalMessageQueueInMsgQueueHeartbeat)
	prometheus.MustRegister(TotalMessageQueueInMsgQueueInvalidDB)
	prometheus.MustRegister(TotalMessageQueueInMsgQueueMissingMsg)
	prometheus.MustRegister(TotalMessageQueueInMsgQueueMissingMsgResp)
	prometheus.MustRegister(TotalMessageQueueInMsgQueueMissingData)
	prometheus.MustRegister(TotalMessageQueueInMsgQueueMissingDataResp)
	prometheus.MustRegister(TotalMessageQueueInMsgQueueRevealEntry)
	prometheus.MustRegister(TotalMessageQueueInMsgQueueReqBlock)
	prometheus.MustRegister(TotalMessageQueueInMsgQueueDbStateMissing)
	prometheus.MustRegister(TotalMessageQueueInMsgQueueDbState)
	prometheus.MustRegister(TotalMessageQueueInMsgQueueBounceMsg)
	prometheus.MustRegister(TotalMessageQueueInMsgQueueBounceResp)
	prometheus.MustRegister(TotalMessageQueueInMsgQueueMisc)
}
