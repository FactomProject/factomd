package wsapi

import "github.com/FactomProject/factomd/telemetry"

var (
	GensisFblockCall = telemetry.Counter(
		"factomd_wsapi_v2_gensis_fblock_count",
		"Number of times the gensis Fblock is asked for",
	)

	HandleV2APICallGeneral = telemetry.Summary(
		"factomd_wsapi_v2_api_general_call_ns",
		"Time it takes to compelete a call",
	)

	HandleV2APICallChainHead = telemetry.Summary(
		"factomd_wsapi_v2_api_call_chainhead_ns",
		"Time it takes to compelete a chainhead",
	)

	HandleV2APICallCommitChain = telemetry.Summary(
		"factomd_wsapi_v2_api_call_commitchain_ns",
		"Time it takes to compelete a commithcain",
	)

	HandleV2APICallCommitEntry = telemetry.Summary(
		"factomd_wsapi_v2_api_call_commitentry_ns",
		"Time it takes to compelete a commitentry",
	)

	HandleV2APICallDBlock = telemetry.Summary(
		"factomd_wsapi_v2_api_call_dblock_ns",
		"Time it takes to compelete a dblock",
	)

	HandleV2APICallDBlockHead = telemetry.Summary(
		"factomd_wsapi_v2_api_call_dblockhead_ns",
		"Time it takes to compelete a dblockhead",
	)

	HandleV2APICallEblock = telemetry.Summary(
		"factomd_wsapi_v2_api_call_eblock_ns",
		"Time it takes to compelete a eblock",
	)

	HandleV2APICallAblock = telemetry.Summary(
		"factomd_wsapi_v2_api_call_ablock_ns",
		"Time it takes to compelete a eblock",
	)

	HandleV2APICallEntry = telemetry.Summary(
		"factomd_wsapi_v2_api_call_entry_ns",
		"Time it takes to compelete an entry",
	)

	HandleV2APICallECBal = telemetry.Summary(
		"factomd_wsapi_v2_api_call_ecbal_ns",
		"Time it takes to compelete a ecbal",
	)

	HandleV2APICallECRate = telemetry.Summary(
		"factomd_wsapi_v2_api_call_ecrate_ns",
		"Time it takes to compelete a ecrate",
	)

	HandleV2APICallFABal = telemetry.Summary(
		"factomd_wsapi_v2_api_call_fabal_ns",
		"Time it takes to compelete a fabal",
	)

	HandleV2APICallFctTx = telemetry.Summary(
		"factomd_wsapi_v2_api_call_fcttx_ns",
		"Time it takes to compelete a fcttx",
	)

	HandleV2APICallHeights = telemetry.Summary(
		"factomd_wsapi_v2_api_call_heights_ns",
		"Time it takes to compelete a heights",
	)

	HandleV2APICallProp = telemetry.Summary(
		"factomd_wsapi_v2_api_call_prop_ns",
		"Time it takes to compelete a prop",
	)

	HandleV2APICallRawData = telemetry.Summary(
		"factomd_wsapi_v2_api_call_rawdata_ns",
		"Time it takes to compelete a rawdata",
	)

	HandleV2APICallAnchors = telemetry.Summary(
		"factomd_wsapi_v2_api_call_anchors_ns",
		"Time it takes to compelete a ",
	)

	HandleV2APICallReceipt = telemetry.Summary(
		"factomd_wsapi_v2_api_call_receipt_ns",
		"Time it takes to compelete a ",
	)

	HandleV2APICallRevealEntry = telemetry.Summary(
		"factomd_wsapi_v2_api_call_reventry_ns",
		"Time it takes to compelete a revealentry",
	)

	HandleV2APICallFctAck = telemetry.Summary(
		"factomd_wsapi_v2_api_call_fctack_ns",
		"Time it takes to compelete a fctack",
	)

	HandleV2APICallEntryAck = telemetry.Summary(
		"factomd_wsapi_v2_api_call_entryack_ns",
		"Time it takes to compelete a entryack",
	)

	HandleV2APICallPendingEntries = telemetry.Summary(
		"factomd_wsapi_v2_api_call_pendingentries_ns",
		"Time it takes to compelete a pendingentries",
	)

	HandleV2APICallPendingTxs = telemetry.Summary(
		"factomd_wsapi_v2_api_call_pendingtxs_ns",
		"Time it takes to compelete a pendingtxs",
	)

	HandleV2APICallSendRaw = telemetry.Summary(
		"factomd_wsapi_v2_api_call_sendraw_ns",
		"Time it takes to compelete a sendraw",
	)

	HandleV2APICallTransaction = telemetry.Summary(
		"factomd_wsapi_v2_api_call_tx_ns",
		"Time it takes to compelete a tx",
	)

	HandleV2APICallDBlockByHeight = telemetry.Summary(
		"factomd_wsapi_v2_api_call_dblockbyheight_ns",
		"Time it takes to compelete a dblockbyheight",
	)

	HandleV2APICallECBlockByHeight = telemetry.Summary(
		"factomd_wsapi_v2_api_call_ecblockbyheight_ns",
		"Time it takes to compelete a ecblockbyheight",
	)

	HandleV2APICallECBlock = telemetry.Summary(
		"factomd_wsapi_v2_api_call_ecblock_ns",
		"Time it takes to compelete a ecblock",
	)

	HandleV2APICallFblockByHeight = telemetry.Summary(
		"factomd_wsapi_v2_api_call_fblockbyheight_ns",
		"Time it takes to compelete a fblockbyheight",
	)

	HandleV2APICallFblock = telemetry.Summary(
		"factomd_wsapi_v2_api_call_fblock_ns",
		"Time it takes to compelete a fblock",
	)

	HandleV2APICallABlockByHeight = telemetry.Summary(
		"factomd_wsapi_v2_api_call_ablockbyheight_ns",
		"Time it takes to compelete a ablockbyheight",
	)

	HandleV2APICallTpsRate = telemetry.Summary(
		"factomd_wsapi_v2_api_call_tpsrate_ns",
		"Time it takes to compelete a tpsrate",
	)
)
