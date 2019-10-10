package wsapi

import "github.com/FactomProject/factomd/telemetry"

var counter = telemetry.RegisterMetric.Counter
var summary = telemetry.RegisterMetric.Summary

var (
	GensisFblockCall = counter(
		"factomd_wsapi_v2_gensis_fblock_count",
		"Number of times the gensis Fblock is asked for",
	)

	HandleV2APICallGeneral = summary(
		"factomd_wsapi_v2_api_general_call_ns",
		"Time it takes to compelete a call",
	)

	HandleV2APICallChainHead = summary(
		"factomd_wsapi_v2_api_call_chainhead_ns",
		"Time it takes to compelete a chainhead",
	)

	HandleV2APICallCommitChain = summary(
		"factomd_wsapi_v2_api_call_commitchain_ns",
		"Time it takes to compelete a commithcain",
	)

	HandleV2APICallCommitEntry = summary(
		"factomd_wsapi_v2_api_call_commitentry_ns",
		"Time it takes to compelete a commitentry",
	)

	HandleV2APICallDBlock = summary(
		"factomd_wsapi_v2_api_call_dblock_ns",
		"Time it takes to compelete a dblock",
	)

	HandleV2APICallDBlockHead = summary(
		"factomd_wsapi_v2_api_call_dblockhead_ns",
		"Time it takes to compelete a dblockhead",
	)

	HandleV2APICallEblock = summary(
		"factomd_wsapi_v2_api_call_eblock_ns",
		"Time it takes to compelete a eblock",
	)

	HandleV2APICallAblock = summary(
		"factomd_wsapi_v2_api_call_ablock_ns",
		"Time it takes to compelete a eblock",
	)

	HandleV2APICallEntry = summary(
		"factomd_wsapi_v2_api_call_entry_ns",
		"Time it takes to compelete an entry",
	)

	HandleV2APICallECBal = summary(
		"factomd_wsapi_v2_api_call_ecbal_ns",
		"Time it takes to compelete a ecbal",
	)

	HandleV2APICallECRate = summary(
		"factomd_wsapi_v2_api_call_ecrate_ns",
		"Time it takes to compelete a ecrate",
	)

	HandleV2APICallFABal = summary(
		"factomd_wsapi_v2_api_call_fabal_ns",
		"Time it takes to compelete a fabal",
	)

	HandleV2APICallFctTx = summary(
		"factomd_wsapi_v2_api_call_fcttx_ns",
		"Time it takes to compelete a fcttx",
	)

	HandleV2APICallHeights = summary(
		"factomd_wsapi_v2_api_call_heights_ns",
		"Time it takes to compelete a heights",
	)

	HandleV2APICallProp = summary(
		"factomd_wsapi_v2_api_call_prop_ns",
		"Time it takes to compelete a prop",
	)

	HandleV2APICallRawData = summary(
		"factomd_wsapi_v2_api_call_rawdata_ns",
		"Time it takes to compelete a rawdata",
	)

	HandleV2APICallAnchors = summary(
		"factomd_wsapi_v2_api_call_anchors_ns",
		"Time it takes to compelete a ",
	)

	HandleV2APICallReceipt = summary(
		"factomd_wsapi_v2_api_call_receipt_ns",
		"Time it takes to compelete a ",
	)

	HandleV2APICallRevealEntry = summary(
		"factomd_wsapi_v2_api_call_reventry_ns",
		"Time it takes to compelete a revealentry",
	)

	HandleV2APICallFctAck = summary(
		"factomd_wsapi_v2_api_call_fctack_ns",
		"Time it takes to compelete a fctack",
	)

	HandleV2APICallEntryAck = summary(
		"factomd_wsapi_v2_api_call_entryack_ns",
		"Time it takes to compelete a entryack",
	)

	HandleV2APICallPendingEntries = summary(
		"factomd_wsapi_v2_api_call_pendingentries_ns",
		"Time it takes to compelete a pendingentries",
	)

	HandleV2APICallPendingTxs = summary(
		"factomd_wsapi_v2_api_call_pendingtxs_ns",
		"Time it takes to compelete a pendingtxs",
	)

	HandleV2APICallSendRaw = summary(
		"factomd_wsapi_v2_api_call_sendraw_ns",
		"Time it takes to compelete a sendraw",
	)

	HandleV2APICallTransaction = summary(
		"factomd_wsapi_v2_api_call_tx_ns",
		"Time it takes to compelete a tx",
	)

	HandleV2APICallDBlockByHeight = summary(
		"factomd_wsapi_v2_api_call_dblockbyheight_ns",
		"Time it takes to compelete a dblockbyheight",
	)

	HandleV2APICallECBlockByHeight = summary(
		"factomd_wsapi_v2_api_call_ecblockbyheight_ns",
		"Time it takes to compelete a ecblockbyheight",
	)

	HandleV2APICallECBlock = summary(
		"factomd_wsapi_v2_api_call_ecblock_ns",
		"Time it takes to compelete a ecblock",
	)

	HandleV2APICallFblockByHeight = summary(
		"factomd_wsapi_v2_api_call_fblockbyheight_ns",
		"Time it takes to compelete a fblockbyheight",
	)

	HandleV2APICallFblock = summary(
		"factomd_wsapi_v2_api_call_fblock_ns",
		"Time it takes to compelete a fblock",
	)

	HandleV2APICallABlockByHeight = summary(
		"factomd_wsapi_v2_api_call_ablockbyheight_ns",
		"Time it takes to compelete a ablockbyheight",
	)

	HandleV2APICallTpsRate = summary(
		"factomd_wsapi_v2_api_call_tpsrate_ns",
		"Time it takes to compelete a tpsrate",
	)
)
