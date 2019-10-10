package state

import "github.com/FactomProject/factomd/telemetry"

var (
	RegisterPrometheus = telemetry.RegisterPrometheus

	// TODO: refactor to create metrics during initialization

	// Entry Syncing Controller
	ESMissingQueue   = telemetry.ESMissingQueue
	ESMissing        = telemetry.ESMissing
	ESFound          = telemetry.ESFound
	ESAsking         = telemetry.ESAsking
	ESHighestAsking  = telemetry.ESHighestAsking
	ESHighestMissing = telemetry.ESHighestMissing
	ESFirstMissing   = telemetry.ESFirstMissing
	ESDBHTComplete   = telemetry.ESDBHTComplete
	ESAvgRequests    = telemetry.ESAvgRequests
	HighestAck       = telemetry.HighestAck
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
	TotalMsgQueueInputs  = telemetry.TotalMsgQueueInputs
	TotalMsgQueueOutputs = telemetry.TotalMsgQueueOutputs

	// Holding Queue
	TotalHoldingQueueInputs        = telemetry.TotalHoldingQueueInputs
	TotalHoldingQueueOutputs       = telemetry.TotalHoldingQueueOutputs
	TotalHoldingQueueRecycles      = telemetry.TotalHoldingQueueRecycles
	HoldingQueueDBSigInputs        = telemetry.HoldingQueueDBSigInputs
	HoldingQueueDBSigOutputs       = telemetry.HoldingQueueDBSigOutputs
	HoldingQueueCommitEntryInputs  = telemetry.HoldingQueueCommitEntryInputs
	HoldingQueueCommitEntryOutputs = telemetry.HoldingQueueCommitEntryOutputs
	HoldingQueueCommitChainInputs  = telemetry.HoldingQueueCommitChainInputs
	HoldingQueueCommitChainOutputs = telemetry.HoldingQueueCommitChainOutputs
	HoldingQueueRevealEntryInputs  = telemetry.HoldingQueueRevealEntryInputs
	HoldingQueueRevealEntryOutputs = telemetry.HoldingQueueRevealEntryOutputs

	// Acks Queue                          // Acks Queue
	TotalAcksInputs  = telemetry.TotalAcksInputs
	TotalAcksOutputs = telemetry.TotalAcksOutputs

	// Commits map                         // Commits map
	TotalCommitsInputs  = telemetry.TotalCommitsInputs
	TotalCommitsOutputs = telemetry.TotalCommitsOutputs

	// XReview Queue                       // XReview Queue
	TotalXReviewQueueInputs  = telemetry.TotalXReviewQueueInputs
	TotalXReviewQueueOutputs = telemetry.TotalXReviewQueueOutputs

	// Executions                          // Executions
	LeaderExecutions             = telemetry.LeaderExecutions
	FollowerExecutions           = telemetry.FollowerExecutions
	LeaderEOMExecutions          = telemetry.LeaderEOMExecutions
	FollowerEOMExecutions        = telemetry.FollowerEOMExecutions
	FollowerMissingMsgExecutions = telemetry.FollowerMissingMsgExecutions

	// ProcessList                         // ProcessList
	TotalProcessListInputs    = telemetry.TotalProcessListInputs
	TotalProcessListProcesses = telemetry.TotalProcessListProcesses
	TotalProcessEOMs          = telemetry.TotalProcessEOMs

	// Durations                           // Durations
	TotalReviewHoldingTime   = telemetry.TotalReviewHoldingTime
	TotalProcessXReviewTime  = telemetry.TotalProcessXReviewTime
	TotalProcessProcChanTime = telemetry.TotalProcessProcChanTime
	TotalEmptyLoopTime       = telemetry.TotalEmptyLoopTime
	TotalAckLoopTime         = telemetry.TotalAckLoopTime
	TotalExecuteMsgTime      = telemetry.TotalExecuteMsgTime
)
