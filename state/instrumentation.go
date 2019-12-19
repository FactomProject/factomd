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
	ESMissingQueue = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_es_missing_entry_queue",
		Help: "Number of known missing entries in our queue to find.",
	})
	ESMissing = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_es_missing_entries",
		Help: "Number of known missing entries",
	})
	ESFound = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_es_found_entries",
		Help: "Number of known missing entries found.",
	})
	ESAsking = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_es_asking_missing_entries",
		Help: "Number we are asking for of the known missing entries.",
	})
	ESHighestAsking = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_es_highest_asking_entries",
		Help: "Highest entry DBHeight which has has a request made.",
	})
	ESHighestMissing = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_es_highest_missing_entries",
		Help: "Highest DBHeight of the entries we know are missing.",
	})
	ESFirstMissing = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_es_first_missing_entries",
		Help: "First DBHeight with a missing entry",
	})
	ESDBHTComplete = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_es_entry_dbheight_complete",
		Help: "First DBHeight with a missing entry",
	})
	ESAvgRequests = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_es_average_requests",
		Help: "Average number of times we have had to request a missing entry",
	})
	HighestAck = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_highest_ack",
		Help: "Acknowledgement with the highest directory block height",
	})
	HighestKnown = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_highest_known",
		Help: "Highest known block (which can be different than the highest ack)",
	})
	HighestSaved = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_highest_saved",
		Help: "Highest saved block to the database",
	})
	HighestCompleted = prometheus.NewGauge(prometheus.GaugeOpts{
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

	// Torrent
	stateTorrentSyncingLower = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_torrentsync_lower_gauge",
		Help: "The lower limit of torrent sync",
	})

	stateTorrentSyncingUpper = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_torrentsync_upper_gauge",
		Help: "The upper limit of torrent sync",
	})

	// Queues
	CurrentMessageQueueInMsgGeneralVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_general_inmsg_vec",
		Help: "Instrumenting the current  inmsg queue ",
	}, []string{"message"})

	TotalMessageQueueInMsgGeneralVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "factomd_state_queue_total_general_inmsg_vec",
		Help: "Instrumenting the inmsg queue ",
	}, []string{"message"})

	CurrentMessageQueueApiGeneralVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "factomd_state_queue_current_general_api_vec",
		Help: "Instrumenting the current API queue ",
	}, []string{"message"})

	TotalMessageQueueApiGeneralVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "factomd_state_queue_total_general_api_vec",
		Help: "Instrumenting the API queue ",
	}, []string{"message"})

	TotalMessageQueueNetOutMsgGeneralVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "factomd_state_queue_total_general_netoutmsg_vec",
		Help: "Instrumenting the netoutmsg queue ",
	}, []string{"message"})

	// MsgQueue chan
	TotalMsgQueueInputs = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_msgqueue_total_inputs",
		Help: "Tally of total messages gone into MsgQueue (useful for rating)",
	})
	TotalMsgQueueOutputs = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_msgqueue_total_outputs",
		Help: "Tally of total messages drained out of MsgQueue (useful for rating)",
	})

	// Holding Queue
	TotalHoldingQueueInputs = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_holding_queue_total_inputs",
		Help: "Tally of total messages gone into Holding (useful for rating)",
	})
	TotalHoldingQueueOutputs = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_holding_queue_total_outputs",
		Help: "Tally of total messages drained out of Holding (useful for rating)",
	})
	TotalHoldingQueueRecycles = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_holding_queue_total_recycles",
		Help: "Tally of total messages recycled thru Holding (useful for rating)",
	})
	HoldingQueueDBSigInputs = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_holding_queue_dbsig_inputs",
		Help: "Tally of DBSig messages gone into Holding (useful for rating)",
	})
	HoldingQueueDBSigOutputs = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_holding_queue_dbsig_outputs",
		Help: "Tally of DBSig messages drained out of Holding",
	})
	HoldingQueueCommitEntryInputs = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_holding_queue_commitentry_inputs",
		Help: "Tally of CommitEntry messages gone into Holding (useful for rating)",
	})
	HoldingQueueCommitEntryOutputs = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_holding_queue_commitentry_outputs",
		Help: "Tally of CommitEntry messages drained out of Holding",
	})
	HoldingQueueCommitChainInputs = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_holding_queue_commitchain_inputs",
		Help: "Tally of CommitChain messages gone into Holding (useful for rating)",
	})
	HoldingQueueCommitChainOutputs = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_holding_queue_commitchain_outputs",
		Help: "Tally of CommitChain messages drained out of Holding",
	})
	HoldingQueueRevealEntryInputs = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_holding_queue_revealentry_inputs",
		Help: "Tally of RevealEntry messages gone into Holding (useful for rating)",
	})
	HoldingQueueRevealEntryOutputs = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_holding_queue_revealentry_outputs",
		Help: "Tally of RevealEntry messages drained out of Holding",
	})

	// Acks Queue
	TotalAcksInputs = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_acks_total_inputs",
		Help: "Tally of total messages gone into Acks (useful for rating)",
	})
	TotalAcksOutputs = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_acks_total_outputs",
		Help: "Tally of total messages drained out of Acks (useful for rating)",
	})

	// Commits map
	TotalCommitsInputs = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_commits_total_inputs",
		Help: "Tally of total messages gone into Commits (useful for rating)",
	})
	TotalCommitsOutputs = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_commits_total_outputs",
		Help: "Tally of total messages drained out of Commits (useful for rating)",
	})

	// XReview Queue
	TotalXReviewQueueInputs = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_xreview_queue_total_inputs",
		Help: "Tally of total messages gone into XReview (useful for rating)",
	})
	TotalXReviewQueueOutputs = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_xreview_queue_total_outputs",
		Help: "Tally of total messages drained out of XReview (useful for rating)",
	})

	// Executions
	LeaderExecutions = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_leader_executions",
		Help: "Tally of total messages executed via LeaderExecute",
	})
	FollowerExecutions = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_follower_executions",
		Help: "Tally of total messages executed via FollowerExecute",
	})
	LeaderEOMExecutions = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_leader_eom_executions",
		Help: "Tally of total messages executed via LeaderExecuteEOM",
	})
	FollowerEOMExecutions = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_follower_eom_executions",
		Help: "Tally of total messages executed via FollowerExecuteEOM",
	})
	FollowerMissingMsgExecutions = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_follower_mm_executions",
		Help: "Tally of total messages executed via FollowerExecuteMissingMsg",
	})

	// ProcessList
	TotalProcessListInputs = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_process_list_inputs",
		Help: "Tally of total messages gone into ProcessLists (useful for rating)",
	})
	TotalProcessListProcesses = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_process_list_processes",
		Help: "Tally of total messages processed from ProcessLists (useful for rating)",
	})
	TotalProcessEOMs = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_state_process_eom_processes",
		Help: "Tally of EOM messages processed from ProcessLists (useful for rating)",
	})

	// Durations
	TotalReviewHoldingTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_review_holding_time",
		Help: "Time spent in ReviewHolding()",
	})
	TotalProcessXReviewTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_process_xreview_time",
		Help: "Time spent Processing XReview",
	})
	TotalProcessProcChanTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_process_proc_chan_time",
		Help: "Time spent Processing Process Chan",
	})
	TotalEmptyLoopTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_empty_loop_time",
		Help: "Time spent in empty loop",
	})
	TotalAckLoopTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_ack_loop_time",
		Help: "Time spent in ack loop",
	})
	TotalExecuteMsgTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_state_execute_msg_time",
		Help: "Time spent in executeMsg",
	})

	//		Eom/DBSig delay
	LeaderSyncAckDelay = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "factomd_state_consensus_sync_delay_ack_vec_sec",
		Help: "Instruments the delay in the time an ack (eom/dbsig only) was signed by a leader, " +
			"and the time it took our node to receive it in nanoseconds.",
	}, []string{"leader"})

	LeaderSyncMsgDelay = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "factomd_state_consensus_sync_delay_msg_vec_sec",
		Help: "Instruments the delay in the time a msg (eom/dbsig only) was signed by a leader, " +
			"and the time it took our node to receive it in nanoseconds.",
	}, []string{"leader"})

	LeaderSyncAckPairDelay = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "factomd_state_consensus_ackpair_delay_vec_sec",
		Help: "Instruments the delay in the time an ack & msg pair was received. " +
			"If there is a delay, it means the ack+msg would benefit from being " +
			"coupled. The delay is measured in seconds.",
	}, []string{"leader"})
)

var registered bool = false

// RegisterPrometheus registers the variables to be exposed. This can only be run once, hence the
// boolean flag to prevent panics if launched more than once. This is called in NetStart
func RegisterPrometheus() {
	if registered {
		return
	}
	registered = true
	// 		Example Cont.
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

	// Torrent
	prometheus.MustRegister(stateTorrentSyncingLower)
	prometheus.MustRegister(stateTorrentSyncingUpper)

	// Queues
	prometheus.MustRegister(CurrentMessageQueueInMsgGeneralVec)
	prometheus.MustRegister(TotalMessageQueueInMsgGeneralVec)
	prometheus.MustRegister(CurrentMessageQueueApiGeneralVec)
	prometheus.MustRegister(TotalMessageQueueApiGeneralVec)
	prometheus.MustRegister(TotalMessageQueueNetOutMsgGeneralVec)

	// MsgQueue chan
	prometheus.MustRegister(TotalMsgQueueInputs)
	prometheus.MustRegister(TotalMsgQueueOutputs)

	// Holding
	prometheus.MustRegister(TotalHoldingQueueInputs)
	prometheus.MustRegister(TotalHoldingQueueOutputs)
	prometheus.MustRegister(HoldingQueueDBSigInputs)
	prometheus.MustRegister(HoldingQueueDBSigOutputs)
	prometheus.MustRegister(HoldingQueueCommitEntryInputs)
	prometheus.MustRegister(HoldingQueueCommitEntryOutputs)
	prometheus.MustRegister(HoldingQueueCommitChainInputs)
	prometheus.MustRegister(HoldingQueueCommitChainOutputs)
	prometheus.MustRegister(HoldingQueueRevealEntryInputs)
	prometheus.MustRegister(HoldingQueueRevealEntryOutputs)

	// Acks
	prometheus.MustRegister(TotalAcksInputs)
	prometheus.MustRegister(TotalAcksOutputs)

	// Execution
	prometheus.MustRegister(LeaderExecutions)
	prometheus.MustRegister(FollowerExecutions)
	prometheus.MustRegister(LeaderEOMExecutions)
	prometheus.MustRegister(FollowerEOMExecutions)
	prometheus.MustRegister(FollowerMissingMsgExecutions)

	// ProcessList
	prometheus.MustRegister(TotalProcessListInputs)
	prometheus.MustRegister(TotalProcessListProcesses)
	prometheus.MustRegister(TotalProcessEOMs)

	// XReview Queue
	prometheus.MustRegister(TotalXReviewQueueInputs)
	prometheus.MustRegister(TotalXReviewQueueOutputs)

	// Commits map
	prometheus.MustRegister(TotalCommitsInputs)
	prometheus.MustRegister(TotalCommitsOutputs)

	// Durations
	prometheus.MustRegister(TotalReviewHoldingTime)
	prometheus.MustRegister(TotalProcessXReviewTime)
	prometheus.MustRegister(TotalProcessProcChanTime)
	prometheus.MustRegister(TotalEmptyLoopTime)
	prometheus.MustRegister(TotalAckLoopTime)
	prometheus.MustRegister(TotalExecuteMsgTime)
	//		Delays
	prometheus.MustRegister(LeaderSyncMsgDelay)
	prometheus.MustRegister(LeaderSyncAckDelay)
	prometheus.MustRegister(LeaderSyncAckPairDelay)
}
