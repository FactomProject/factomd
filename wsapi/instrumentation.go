package wsapi

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	GensisFblockCall = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_wsapi_v2_gensis_fblock_count",
		Help: "Number of times the gensis Fblock is asked for",
	})

	HandleV2APICallGeneral = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_general_call_ns",
		Help: "Time it takes to compelete a call",
	})

	HandleV2APICallChainHead = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_chainhead_ns",
		Help: "Time it takes to compelete a chainhead",
	})

	HandleV2APICallCommitChain = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_commitchain_ns",
		Help: "Time it takes to compelete a commithcain",
	})

	HandleV2APICallCommitEntry = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_commitentry_ns",
		Help: "Time it takes to compelete a commitentry",
	})

	HandleV2APICallDBlock = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_dblock_ns",
		Help: "Time it takes to compelete a dblock",
	})

	HandleV2APICallDBlockHead = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_dblockhead_ns",
		Help: "Time it takes to compelete a dblockhead",
	})

	HandleV2APICallEblock = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_eblock_ns",
		Help: "Time it takes to compelete a eblock",
	})

	HandleV2APICallAblock = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_ablock_ns",
		Help: "Time it takes to compelete a eblock",
	})

	HandleV2APICallEntry = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_entry_ns",
		Help: "Time it takes to compelete an entry",
	})

	HandleV2APICallECBal = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_ecbal_ns",
		Help: "Time it takes to compelete a ecbal",
	})

	HandleV2APICallECRate = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_ecrate_ns",
		Help: "Time it takes to compelete a ecrate",
	})

	HandleV2APICallFABal = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_fabal_ns",
		Help: "Time it takes to compelete a fabal",
	})

	HandleV2APICallFctTx = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_fcttx_ns",
		Help: "Time it takes to compelete a fcttx",
	})

	HandleV2APICallHeights = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_heights_ns",
		Help: "Time it takes to compelete a heights",
	})

	HandleV2APICallCurrentMinute = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_minute_ns",
		Help: "Time it takes to compelete a minute",
	})

	HandleV2APICallProp = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_prop_ns",
		Help: "Time it takes to compelete a prop",
	})

	HandleV2APICallRawData = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_rawdata_ns",
		Help: "Time it takes to compelete a rawdata",
	})

	HandleV2APICallAnchors = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_anchors_ns",
		Help: "Time it takes to compelete a ",
	})

	HandleV2APICallReceipt = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_receipt_ns",
		Help: "Time it takes to compelete a ",
	})

	HandleV2APICallRevealEntry = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_reventry_ns",
		Help: "Time it takes to compelete a revealentry",
	})

	HandleV2APICallFctAck = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_fctack_ns",
		Help: "Time it takes to compelete a fctack",
	})

	HandleV2APICallEntryAck = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_entryack_ns",
		Help: "Time it takes to compelete a entryack",
	})

	HandleV2APICall = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call__ns",
		Help: "Time it takes to compelete a ",
	})

	HandleV2APICallPendingEntries = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_pendingentries_ns",
		Help: "Time it takes to compelete a pendingentries",
	})

	HandleV2APICallPendingTxs = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_pendingtxs_ns",
		Help: "Time it takes to compelete a pendingtxs",
	})

	HandleV2APICallSendRaw = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_sendraw_ns",
		Help: "Time it takes to compelete a sendraw",
	})

	HandleV2APICallTransaction = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_tx_ns",
		Help: "Time it takes to compelete a tx",
	})

	HandleV2APICallReplayDBFromHeight = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_replay_from_height",
		Help: "Time it takes to replay DBStates from a specific height",
	})

	HandleV2APICallDBlockByHeight = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_dblockbyheight_ns",
		Help: "Time it takes to compelete a dblockbyheight",
	})

	HandleV2APICallECBlockByHeight = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_ecblockbyheight_ns",
		Help: "Time it takes to compelete a ecblockbyheight",
	})

	HandleV2APICallECBlock = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_ecblock_ns",
		Help: "Time it takes to compelete a ecblock",
	})

	HandleV2APICallFblockByHeight = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_fblockbyheight_ns",
		Help: "Time it takes to compelete a fblockbyheight",
	})

	HandleV2APICallFblock = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_fblock_ns",
		Help: "Time it takes to compelete a fblock",
	})

	HandleV2APICallABlockByHeight = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_ablockbyheight_ns",
		Help: "Time it takes to compelete a ablockbyheight",
	})

	HandleV2APICallAuthorities = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_auths_ns",
		Help: "Time it takes to compelete an auths ",
	})

	HandleV2APICallTpsRate = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "factomd_wsapi_v2_api_call_tpsrate_ns",
		Help: "Time it takes to compelete a tpsrate",
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

	prometheus.MustRegister(GensisFblockCall)
	prometheus.MustRegister(HandleV2APICallGeneral)
	prometheus.MustRegister(HandleV2APICallChainHead)
	prometheus.MustRegister(HandleV2APICallCommitChain)
	prometheus.MustRegister(HandleV2APICallCommitEntry)
	prometheus.MustRegister(HandleV2APICallDBlock)
	prometheus.MustRegister(HandleV2APICallDBlockHead)
	prometheus.MustRegister(HandleV2APICallEblock)
	prometheus.MustRegister(HandleV2APICallEntry)
	prometheus.MustRegister(HandleV2APICallECBal)
	prometheus.MustRegister(HandleV2APICallECRate)
	prometheus.MustRegister(HandleV2APICallFABal)
	prometheus.MustRegister(HandleV2APICallFctTx)
	prometheus.MustRegister(HandleV2APICallHeights)
	prometheus.MustRegister(HandleV2APICallProp)
	prometheus.MustRegister(HandleV2APICallRawData)
	prometheus.MustRegister(HandleV2APICallReceipt)
	prometheus.MustRegister(HandleV2APICallRevealEntry)
	prometheus.MustRegister(HandleV2APICallFctAck)
	prometheus.MustRegister(HandleV2APICallEntryAck)
	prometheus.MustRegister(HandleV2APICall)
	prometheus.MustRegister(HandleV2APICallPendingEntries)
	prometheus.MustRegister(HandleV2APICallPendingTxs)
	prometheus.MustRegister(HandleV2APICallSendRaw)
	prometheus.MustRegister(HandleV2APICallTransaction)
	prometheus.MustRegister(HandleV2APICallReplayDBFromHeight)
	prometheus.MustRegister(HandleV2APICallDBlockByHeight)
	prometheus.MustRegister(HandleV2APICallECBlockByHeight)
	prometheus.MustRegister(HandleV2APICallECBlock)
	prometheus.MustRegister(HandleV2APICallFblockByHeight)
	prometheus.MustRegister(HandleV2APICallABlockByHeight)
	prometheus.MustRegister(HandleV2APICallAuthorities)
	prometheus.MustRegister(HandleV2APICallTpsRate)
	prometheus.MustRegister(HandleV2APICallAblock)
	prometheus.MustRegister(HandleV2APICallFblock)
}
