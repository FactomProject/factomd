package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

var registeredMetrics []prometheus.Collector

// Don't let other packages reference prometheus directly
type Counter = prometheus.Counter
type Gauge = prometheus.Gauge
type GaugeVec = prometheus.GaugeVec

type MetricHandler interface {
	Counter(name string, help string) Counter
	Gauge(name string, help string) Gauge
	GaugeVec(name string, help string, labels []string) *GaugeVec
}

type metric struct {
	MetricHandler
}

var RegisterMetric = metric{}

func (metric) Counter(name string, help string) prometheus.Counter {
	c := prometheus.NewCounter(prometheus.CounterOpts{
		Name: name,
		Help: help,
	})
	registeredMetrics = append(registeredMetrics, c)
	return c
}

func (metric) Gauge(name string, help string) prometheus.Gauge {
	g := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	})
	registeredMetrics = append(registeredMetrics, g)
	return g
}

func (metric) GaugeVec(name string, help string, labels []string) *prometheus.GaugeVec {
	v := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	}, labels)

	registeredMetrics = append(registeredMetrics, v)
	return v
}

var gauge = RegisterMetric.Gauge
var counter = RegisterMetric.Counter
var gaugeVec = RegisterMetric.GaugeVec

var (
	// Entry Syncing Controller
	ESMissingQueue = gauge(
		"factomd_state_es_missing_entry_queue",
		"Number of known missing entries in our queue to find.",
	)
	ESMissing = gauge(
		"factomd_state_es_missing_entries",
		"Number of known missing entries",
	)
	ESFound = gauge(
		"factomd_state_es_found_entries",
		"Number of known missing entries found.",
	)
	ESAsking = gauge(
		"factomd_state_es_asking_missing_entries",
		"Number we are asking for of the known missing entries.",
	)
	ESHighestAsking = gauge(
		"factomd_state_es_highest_asking_entries",
		"Highest entry DBHeight which has has a request made.",
	)
	ESHighestMissing = gauge(
		"factomd_state_es_highest_missing_entries",
		"Highest DBHeight of the entries we know are missing.",
	)
	HighestKnown = gauge(
		"factomd_state_highest_known",
		"Highest known block (which can be different than the highest ack)",
	)
	HighestSaved = gauge(
		"factomd_state_highest_saved",
		"Highest saved block to the database",
	)
	HighestCompleted = gauge(
		"factomd_state_highest_completed",
		"Highest completed block, which may or may not be saved to the database",
	)

	// TPS
	TotalTransactionPerSecond = gauge(
		"factomd_state_txrate_total_tps",
		"Total transactions over life of node",
	)

	InstantTransactionPerSecond = gauge(
		"factomd_state_txrate_instant_tps",
		"Total transactions over life of node weighted for last 3 seconds",
	)

	// Torrent
	StateTorrentSyncingLower = gauge(
		"factomd_state_torrentsync_lower_gauge",
		"The lower limit of torrent sync",
	)

	StateTorrentSyncingUpper = gauge(
		"factomd_state_torrentsync_upper_gauge",
		"The upper limit of torrent sync",
	)

	// Queues
	CurrentMessageQueueInMsgGeneralVec = gaugeVec(
		"factomd_state_queue_current_general_inmsg_vec",
		"Instrumenting the current  inmsg queue ",
		[]string{"message"},
	)

	TotalMessageQueueInMsgGeneralVec = gaugeVec(
		"factomd_state_queue_total_general_inmsg_vec",
		"Instrumenting the inmsg queue ",
		[]string{"message"},
	)

	CurrentMessageQueueApiGeneralVec = gaugeVec(
		"factomd_state_queue_current_general_api_vec",
		"Instrumenting the current API queue ",
		[]string{"message"},
	)

	TotalMessageQueueApiGeneralVec = gaugeVec(
		"factomd_state_queue_total_general_api_vec",
		"Instrumenting the API queue ",
		[]string{"message"},
	)

	TotalMessageQueueNetOutMsgGeneralVec = gaugeVec(
		"factomd_state_queue_total_general_netoutmsg_vec",
		"Instrumenting the netoutmsg queue ",
		[]string{"message"},
	)

	// Holding Queue
	TotalHoldingQueueInputs = counter(
		"factomd_state_holding_queue_total_inputs",
		"Tally of total messages gone into Holding (useful for rating)",
	)
	TotalHoldingQueueOutputs = counter(
		"factomd_state_holding_queue_total_outputs",
		"Tally of total messages drained out of Holding (useful for rating)",
	)
	HoldingQueueDBSigOutputs = counter(
		"factomd_state_holding_queue_dbsig_outputs",
		"Tally of DBSig messages drained out of Holding",
	)

	// Acks Queue
	TotalAcksInputs = counter(
		"factomd_state_acks_total_inputs",
		"Tally of total messages gone into Acks (useful for rating)",
	)
	TotalAcksOutputs = counter(
		"factomd_state_acks_total_outputs",
		"Tally of total messages drained out of Acks (useful for rating)",
	)

	// Commits map
	TotalCommitsOutputs = counter(
		"factomd_state_commits_total_outputs",
		"Tally of total messages drained out of Commits (useful for rating)",
	)

	// XReview Queue
	TotalXReviewQueueInputs = counter(
		"factomd_state_xreview_queue_total_inputs",
		"Tally of total messages gone into XReview (useful for rating)",
	)

	// Executions
	LeaderExecutions = counter(
		"factomd_state_leader_executions",
		"Tally of total messages executed via LeaderExecute",
	)
	FollowerExecutions = counter(
		"factomd_state_follower_executions",
		"Tally of total messages executed via FollowerExecute",
	)
	LeaderEOMExecutions = counter(
		"factomd_state_leader_eom_executions",
		"Tally of total messages executed via LeaderExecuteEOM",
	)
	FollowerEOMExecutions = counter(
		"factomd_state_follower_eom_executions",
		"Tally of total messages executed via FollowerExecuteEOM",
	)

	// ProcessList
	TotalProcessListInputs = counter(
		"factomd_state_process_list_inputs",
		"Tally of total messages gone into ProcessLists (useful for rating)",
	)
	TotalProcessListProcesses = counter(
		"factomd_state_process_list_processes",
		"Tally of total messages processed from ProcessLists (useful for rating)",
	)
	TotalProcessEOMs = counter(
		"factomd_state_process_eom_processes",
		"Tally of EOM messages processed from ProcessLists (useful for rating)",
	)

	// Durations
	TotalReviewHoldingTime = counter(
		"factomd_state_review_holding_time",
		"Time spent in ReviewHolding()",
	)
	TotalProcessXReviewTime = counter(
		"factomd_state_process_xreview_time",
		"Time spent Processing XReview",
	)
	TotalProcessProcChanTime = counter(
		"factomd_state_process_proc_chan_time",
		"Time spent Processing Process Chan",
	)
	TotalEmptyLoopTime = counter(
		"factomd_state_empty_loop_time",
		"Time spent in empty loop",
	)
	TotalExecuteMsgTime = counter(
		"factomd_state_execute_msg_time",
		"Time spent in executeMsg",
	)
)

var registered sync.Once

func RegisterPrometheus() {
	registered.Do(func() {
		prometheus.MustRegister(registeredMetrics...)
	})
}
