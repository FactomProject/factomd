package state

import "github.com/FactomProject/factomd/telemetry"

var (
	RegisterPrometheus = telemetry.RegisterPrometheus

	// Entry Syncing Controller
	HighestKnown     = telemetry.HighestKnown
	HighestSaved     = telemetry.HighestSaved
	HighestCompleted = telemetry.HighestCompleted

	// TPS
	TotalTransactionPerSecond   = telemetry.TotalTransactionPerSecond
	InstantTransactionPerSecond = telemetry.InstantTransactionPerSecond

	// Torrent
	stateTorrentSyncingLower = telemetry.StateTorrentSyncingLower
	stateTorrentSyncingUpper = telemetry.StateTorrentSyncingUpper

	// Queues
	CurrentMessageQueueInMsgGeneralVec   = telemetry.CurrentMessageQueueInMsgGeneralVec
	TotalMessageQueueInMsgGeneralVec     = telemetry.TotalMessageQueueInMsgGeneralVec
	CurrentMessageQueueApiGeneralVec     = telemetry.CurrentMessageQueueApiGeneralVec
	TotalMessageQueueApiGeneralVec       = telemetry.TotalMessageQueueApiGeneralVec
	TotalMessageQueueNetOutMsgGeneralVec = telemetry.TotalMessageQueueNetOutMsgGeneralVec

	// MsgQueue chan

	// Holding Queue
	TotalHoldingQueueInputs        = telemetry.TotalHoldingQueueInputs
	TotalHoldingQueueOutputs       = telemetry.TotalHoldingQueueOutputs
	HoldingQueueDBSigOutputs       = telemetry.HoldingQueueDBSigOutputs

	// Acks Queue                          // Acks Queue
	TotalAcksInputs  = telemetry.TotalAcksInputs
	TotalAcksOutputs = telemetry.TotalAcksOutputs

	// Commits map                         // Commits map
	TotalCommitsOutputs = telemetry.TotalCommitsOutputs

	// XReview Queue                       // XReview Queue
	TotalXReviewQueueInputs  = telemetry.TotalXReviewQueueInputs

	// Executions                          // Executions
	LeaderExecutions             = telemetry.LeaderExecutions
	FollowerExecutions           = telemetry.FollowerExecutions
	LeaderEOMExecutions          = telemetry.LeaderEOMExecutions
	FollowerEOMExecutions        = telemetry.FollowerEOMExecutions

	// ProcessList                         // ProcessList
	TotalProcessListInputs    = telemetry.TotalProcessListInputs
	TotalProcessListProcesses = telemetry.TotalProcessListProcesses
	TotalProcessEOMs          = telemetry.TotalProcessEOMs

	// Durations                           // Durations
	TotalReviewHoldingTime   = telemetry.TotalReviewHoldingTime
	TotalProcessXReviewTime  = telemetry.TotalProcessXReviewTime
	TotalProcessProcChanTime = telemetry.TotalProcessProcChanTime
	TotalEmptyLoopTime       = telemetry.TotalEmptyLoopTime
	TotalExecuteMsgTime      = telemetry.TotalExecuteMsgTime
)
