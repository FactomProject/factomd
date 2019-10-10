package databaseOverlay

import (
	"github.com/FactomProject/factomd/telemetry"
)

var counter = telemetry.RegisterMetric.Counter

var (
	OverlayDBGetsDblock = counter(
		"factomd_database_overlay_gets_dblock",
		"Counts gets from the database",
	)

	OverlayDBGetsDblockNumber = counter(
		"factomd_database_overlay_gets_dblocknum",
		"Counts gets from the database",
	)

	OverlayDBGetsDBlockSecond = counter(
		"factomd_database_overlay_gets_dblocksecondart",
		"Counts gets from the database",
	)

	OverlayDBGetsABlock = counter(
		"factomd_database_overlay_gets_ablock",
		"Counts gets from the database",
	)

	OverlayDBGetsABlockNumber = counter(
		"factomd_database_overlay_gets_ablocknum",
		"Counts gets from the database",
	)

	OverlayDBGetsABlockSecondary = counter(
		"factomd_database_overlay_gets_ablocksecondary",
		"Counts gets from the database",
	)

	OverlayDBGetsFBlock = counter(
		"factomd_database_overlay_gets_fblock",
		"Counts gets from the database",
	)

	OverlayDBGetsFBlockNumber = counter(
		"factomd_database_overlay_gets_fblocknumber",
		"Counts gets from the database",
	)

	OverlayDBGetsFBlockSecondary = counter(
		"factomd_database_overlay_gets_fblocksecondary",
		"Counts gets from the database",
	)

	OverlayDBGetsEC = counter(
		"factomd_database_overlay_gets_ec",
		"Counts gets from the database",
	)

	OverlayDBGetsECNumber = counter(
		"factomd_database_overlay_gets_ecnumber",
		"Counts gets from the database",
	)

	OverlayDBGetsECSecondary = counter(
		"factomd_database_overlay_gets_ecsecondary",
		"Counts gets from the database",
	)

	OverlayDBGetsChainHead = counter(
		"factomd_database_overlay_gets_chainhead",
		"Counts gets from the database",
	)

	OverlayDBGetsEBlock = counter(
		"factomd_database_overlay_gets_eblock",
		"Counts gets from the database",
	)

	OverlayDBGetsEBlockNumber = counter(
		"factomd_database_overlay_gets_eblocknumber",
		"Counts gets from the database",
	)

	OverlayDBGetsEBlockSecondary = counter(
		"factomd_database_overlay_gets_eblocksecondary",
		"Counts gets from the database",
	)

	OverlayDBGetsEntry = counter(
		"factomd_database_overlay_gets_entry",
		"Counts gets from the database",
	)

	OverlayDBGetsDirBlockInfo = counter(
		"factomd_database_overlay_gets_dirblockinfo",
		"Counts gets from the database",
	)

	OverlayDBGetsDirBlockUnconfirmed = counter(
		"factomd_database_overlay_gets_dirblockunconfirmed",
		"Counts gets from the database",
	)

	OverlayDBGetsDirBlockInfoNumber = counter(
		"factomd_database_overlay_gets_dirblockinfonumber",
		"Counts gets from the database",
	)

	OverlayDBGetsDirBlockInfoSecondary = counter(
		"factomd_database_overlay_gets_secondary",
		"Counts gets from the database",
	)

	OverlayDBGetsInvludeIn = counter(
		"factomd_database_overlay_gets_includein",
		"Counts gets from the database",
	)

	OverlayDBGetsPaidFor = counter(
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
