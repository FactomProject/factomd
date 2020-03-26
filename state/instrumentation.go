package state

import "github.com/FactomProject/factomd/modules/telemetry"

var (
	// Entry Syncing Controller
	HighestKnown = telemetry.NewGauge(
		"factomd_state_highest_known",
		"Highest known block (which can be different than the highest ack)",
	)
	HighestSaved = telemetry.NewGauge(
		"factomd_state_highest_saved",
		"Highest saved block to the database",
	)
	HighestCompleted = telemetry.NewGauge(
		"factomd_state_highest_completed",
		"Highest completed block, which may or may not be saved to the database",
	)

	// TPS
	TotalTransactionPerSecond = telemetry.NewGauge(
		"factomd_state_txrate_total_tps",
		"Total transactions over life of node",
	)

	InstantTransactionPerSecond = telemetry.NewGauge(
		"factomd_state_txrate_instant_tps",
		"Total transactions over life of node weighted for last 3 seconds",
	)

	// Holding Queue
	TotalHoldingQueueInputs = telemetry.NewCounter(
		"factomd_state_holding_queue_total_inputs",
		"Tally of total messages gone into Holding (useful for rating)",
	)
	TotalHoldingQueueOutputs = telemetry.NewCounter(
		"factomd_state_holding_queue_total_outputs",
		"Tally of total messages drained out of Holding (useful for rating)",
	)
	HoldingQueueDBSigOutputs = telemetry.NewCounter(
		"factomd_state_holding_queue_dbsig_outputs",
		"Tally of DBSig messages drained out of Holding",
	)

	// Acks Queue
	TotalAcksInputs = telemetry.NewCounter(
		"factomd_state_acks_total_inputs",
		"Tally of total messages gone into Acks (useful for rating)",
	)
	TotalAcksOutputs = telemetry.NewCounter(
		"factomd_state_acks_total_outputs",
		"Tally of total messages drained out of Acks (useful for rating)",
	)

	// Commits map
	TotalCommitsOutputs = telemetry.NewCounter(
		"factomd_state_commits_total_outputs",
		"Tally of total messages drained out of Commits (useful for rating)",
	)

	// XReview Queue
	TotalXReviewQueueInputs = telemetry.NewCounter(
		"factomd_state_xreview_queue_total_inputs",
		"Tally of total messages gone into XReview (useful for rating)",
	)

	// Executions
	LeaderExecutions = telemetry.NewCounter(
		"factomd_state_leader_executions",
		"Tally of total messages executed via LeaderExecute",
	)
	FollowerExecutions = telemetry.NewCounter(
		"factomd_state_follower_executions",
		"Tally of total messages executed via FollowerExecute",
	)
	LeaderEOMExecutions = telemetry.NewCounter(
		"factomd_state_leader_eom_executions",
		"Tally of total messages executed via LeaderExecuteEOM",
	)
	FollowerEOMExecutions = telemetry.NewCounter(
		"factomd_state_follower_eom_executions",
		"Tally of total messages executed via FollowerExecuteEOM",
	)

	// ProcessList
	TotalProcessListInputs = telemetry.NewCounter(
		"factomd_state_process_list_inputs",
		"Tally of total messages gone into ProcessLists (useful for rating)",
	)
	TotalProcessListProcesses = telemetry.NewCounter(
		"factomd_state_process_list_processes",
		"Tally of total messages processed from ProcessLists (useful for rating)",
	)
	TotalProcessEOMs = telemetry.NewCounter(
		"factomd_state_process_eom_processes",
		"Tally of EOM messages processed from ProcessLists (useful for rating)",
	)

	// Durations
	TotalReviewHoldingTime = telemetry.NewCounter(
		"factomd_state_review_holding_time",
		"Time spent in ReviewHolding()",
	)
	TotalProcessXReviewTime = telemetry.NewCounter(
		"factomd_state_process_xreview_time",
		"Time spent Processing XReview",
	)
	TotalProcessProcChanTime = telemetry.NewCounter(
		"factomd_state_process_proc_chan_time",
		"Time spent Processing Process Chan",
	)
	TotalEmptyLoopTime = telemetry.NewCounter(
		"factomd_state_empty_loop_time",
		"Time spent in empty loop",
	)
	TotalExecuteMsgTime = telemetry.NewCounter(
		"factomd_state_execute_msg_time",
		"Time spent in executeMsg",
	)
)
