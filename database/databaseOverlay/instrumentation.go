package databaseOverlay

import (
	"github.com/FactomProject/factomd/telemetry"
)

var (
	OverlayDBGetsDblock = telemetry.Counter(
		"factomd_database_overlay_gets_dblock",
		"Counts gets from the database",
	)

	OverlayDBGetsDblockNumber = telemetry.Counter(
		"factomd_database_overlay_gets_dblocknum",
		"Counts gets from the database",
	)

	OverlayDBGetsDBlockSecond = telemetry.Counter(
		"factomd_database_overlay_gets_dblocksecondart",
		"Counts gets from the database",
	)

	OverlayDBGetsABlock = telemetry.Counter(
		"factomd_database_overlay_gets_ablock",
		"Counts gets from the database",
	)

	OverlayDBGetsABlockNumber = telemetry.Counter(
		"factomd_database_overlay_gets_ablocknum",
		"Counts gets from the database",
	)

	OverlayDBGetsABlockSecondary = telemetry.Counter(
		"factomd_database_overlay_gets_ablocksecondary",
		"Counts gets from the database",
	)

	OverlayDBGetsFBlock = telemetry.Counter(
		"factomd_database_overlay_gets_fblock",
		"Counts gets from the database",
	)

	OverlayDBGetsFBlockNumber = telemetry.Counter(
		"factomd_database_overlay_gets_fblocknumber",
		"Counts gets from the database",
	)

	OverlayDBGetsFBlockSecondary = telemetry.Counter(
		"factomd_database_overlay_gets_fblocksecondary",
		"Counts gets from the database",
	)

	OverlayDBGetsEC = telemetry.Counter(
		"factomd_database_overlay_gets_ec",
		"Counts gets from the database",
	)

	OverlayDBGetsECNumber = telemetry.Counter(
		"factomd_database_overlay_gets_ecnumber",
		"Counts gets from the database",
	)

	OverlayDBGetsECSecondary = telemetry.Counter(
		"factomd_database_overlay_gets_ecsecondary",
		"Counts gets from the database",
	)

	OverlayDBGetsChainHead = telemetry.Counter(
		"factomd_database_overlay_gets_chainhead",
		"Counts gets from the database",
	)

	OverlayDBGetsEBlock = telemetry.Counter(
		"factomd_database_overlay_gets_eblock",
		"Counts gets from the database",
	)

	OverlayDBGetsEBlockNumber = telemetry.Counter(
		"factomd_database_overlay_gets_eblocknumber",
		"Counts gets from the database",
	)

	OverlayDBGetsEBlockSecondary = telemetry.Counter(
		"factomd_database_overlay_gets_eblocksecondary",
		"Counts gets from the database",
	)

	OverlayDBGetsEntry = telemetry.Counter(
		"factomd_database_overlay_gets_entry",
		"Counts gets from the database",
	)

	OverlayDBGetsDirBlockInfo = telemetry.Counter(
		"factomd_database_overlay_gets_dirblockinfo",
		"Counts gets from the database",
	)

	OverlayDBGetsDirBlockUnconfirmed = telemetry.Counter(
		"factomd_database_overlay_gets_dirblockunconfirmed",
		"Counts gets from the database",
	)

	OverlayDBGetsDirBlockInfoNumber = telemetry.Counter(
		"factomd_database_overlay_gets_dirblockinfonumber",
		"Counts gets from the database",
	)

	OverlayDBGetsDirBlockInfoSecondary = telemetry.Counter(
		"factomd_database_overlay_gets_secondary",
		"Counts gets from the database",
	)

	OverlayDBGetsInvludeIn = telemetry.Counter(
		"factomd_database_overlay_gets_includein",
		"Counts gets from the database",
	)

	OverlayDBGetsPaidFor = telemetry.Counter(
		"factomd_database_overlay_gets_paidfor",
		"Counts gets from the database",
	)
)

func GetBucket(bucket []byte) {
	switch string(bucket) {
	case string(DIRECTORYBLOCK):
		OverlayDBGetsDblock.Inc()
	case string(DIRECTORYBLOCK_NUMBER):
		OverlayDBGetsDblockNumber.Inc()
	case string(DIRECTORYBLOCK_SECONDARYINDEX):
		OverlayDBGetsDBlockSecond.Inc()
	case string(ADMINBLOCK):
		OverlayDBGetsABlock.Inc()
	case string(ADMINBLOCK_NUMBER):
		OverlayDBGetsABlockNumber.Inc()
	case string(ADMINBLOCK_SECONDARYINDEX):
		OverlayDBGetsABlockSecondary.Inc()
	case string(FACTOIDBLOCK):
		OverlayDBGetsFBlock.Inc()
	case string(FACTOIDBLOCK_NUMBER):
		OverlayDBGetsFBlockNumber.Inc()
	case string(FACTOIDBLOCK_SECONDARYINDEX):
		OverlayDBGetsFBlockSecondary.Inc()
	case string(ENTRYCREDITBLOCK):
		OverlayDBGetsEC.Inc()
	case string(ENTRYCREDITBLOCK_NUMBER):
		OverlayDBGetsECNumber.Inc()
	case string(ENTRYCREDITBLOCK_SECONDARYINDEX):
		OverlayDBGetsECSecondary.Inc()
	case string(CHAIN_HEAD):
		OverlayDBGetsChainHead.Inc()
	case string(ENTRYBLOCK):
		OverlayDBGetsEBlock.Inc()
	case string(ENTRYBLOCK_CHAIN_NUMBER):
		OverlayDBGetsEBlockNumber.Inc()
	case string(ENTRYBLOCK_SECONDARYINDEX):
		OverlayDBGetsEBlockSecondary.Inc()
	case string(ENTRY):
		OverlayDBGetsEntry.Inc()
	case string(DIRBLOCKINFO):
		OverlayDBGetsDirBlockInfo.Inc()
	case string(DIRBLOCKINFO_UNCONFIRMED):
		OverlayDBGetsDirBlockUnconfirmed.Inc()
	case string(DIRBLOCKINFO_NUMBER):
		OverlayDBGetsDirBlockInfoNumber.Inc()
	case string(DIRBLOCKINFO_SECONDARYINDEX):
		OverlayDBGetsDirBlockInfoSecondary.Inc()
	case string(INCLUDED_IN):
		OverlayDBGetsInvludeIn.Inc()
	case string(PAID_FOR):
		OverlayDBGetsPaidFor.Inc()
	}
}

/*

	DIRECTORYBLOCK
	DIRECTORYBLOCK_NUMBER
	DIRECTORYBLOCK_SECONDARYINDEX
	ADMINBLOCK
	ADMINBLOCK_NUMBER
	ADMINBLOCK_SECONDARYINDEX
	FACTOIDBLOCK
	FACTOIDBLOCK_NUMBER
	FACTOIDBLOCK_SECONDARYINDEX
	ENTRYCREDITBLOCK
	ENTRYCREDITBLOCK_NUMBER
	ENTRYCREDITBLOCK_SECONDARYINDEX
	CHAIN_HEAD
	ENTRYBLOCK
	ENTRYBLOCK_CHAIN_NUMBER
	ENTRYBLOCK_SECONDARYINDEX
	ENTRY
	DIRBLOCKINFO
	DIRBLOCKINFO_UNCONFIRMED
	DIRBLOCKINFO_NUMBER
	DIRBLOCKINFO_SECONDARYINDEX
	INCLUDED_IN
	PAID_FOR
*/
