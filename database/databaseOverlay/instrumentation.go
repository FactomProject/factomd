package databaseOverlay

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	OverlayDBGetsDblock = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_database_overlay_gets_dblock",
		Help: "Counts gets from the database",
	})

	OverlayDBGetsDblockNumber = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_database_overlay_gets_dblocknum",
		Help: "Counts gets from the database",
	})

	OverlayDBGetsDBlockSecond = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_database_overlay_gets_dblocksecondart",
		Help: "Counts gets from the database",
	})

	OverlayDBGetsABlock = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_database_overlay_gets_ablock",
		Help: "Counts gets from the database",
	})

	OverlayDBGetsABlockNumber = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_database_overlay_gets_ablocknum",
		Help: "Counts gets from the database",
	})

	OverlayDBGetsABlockSecondary = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_database_overlay_gets_ablocksecondary",
		Help: "Counts gets from the database",
	})

	OverlayDBGetsFBlock = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_database_overlay_gets_fblock",
		Help: "Counts gets from the database",
	})

	OverlayDBGetsFBlockNumber = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_database_overlay_gets_fblocknumber",
		Help: "Counts gets from the database",
	})

	OverlayDBGetsFBlockSecondary = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_database_overlay_gets_fblocksecondary",
		Help: "Counts gets from the database",
	})

	OverlayDBGetsEC = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_database_overlay_gets_ec",
		Help: "Counts gets from the database",
	})

	OverlayDBGetsECNumber = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_database_overlay_gets_ecnumber",
		Help: "Counts gets from the database",
	})

	OverlayDBGetsECSecondary = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_database_overlay_gets_ecsecondary",
		Help: "Counts gets from the database",
	})

	OverlayDBGetsChainHead = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_database_overlay_gets_chainhead",
		Help: "Counts gets from the database",
	})

	OverlayDBGetsEBlock = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_database_overlay_gets_eblock",
		Help: "Counts gets from the database",
	})

	OverlayDBGetsEBlockNumber = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_database_overlay_gets_eblocknumber",
		Help: "Counts gets from the database",
	})

	OverlayDBGetsEBlockSecondary = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_database_overlay_gets_eblocksecondary",
		Help: "Counts gets from the database",
	})

	OverlayDBGetsEntry = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_database_overlay_gets_entry",
		Help: "Counts gets from the database",
	})

	OverlayDBGetsDirBlockInfo = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_database_overlay_gets_dirblockinfo",
		Help: "Counts gets from the database",
	})

	OverlayDBGetsDirBlockUnconfirmed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_database_overlay_gets_dirblockunconfirmed",
		Help: "Counts gets from the database",
	})

	OverlayDBGetsDirBlockInfoNumber = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_database_overlay_gets_dirblockinfonumber",
		Help: "Counts gets from the database",
	})

	OverlayDBGetsDirBlockInfoSecondary = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_database_overlay_gets_secondary",
		Help: "Counts gets from the database",
	})

	OverlayDBGetsInvludeIn = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_database_overlay_gets_includein",
		Help: "Counts gets from the database",
	})

	OverlayDBGetsPaidFor = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_database_overlay_gets_paidfor",
		Help: "Counts gets from the database",
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

	prometheus.MustRegister(OverlayDBGetsDblock)
	prometheus.MustRegister(OverlayDBGetsDblockNumber)
	prometheus.MustRegister(OverlayDBGetsDBlockSecond)
	prometheus.MustRegister(OverlayDBGetsABlock)
	prometheus.MustRegister(OverlayDBGetsABlockNumber)
	prometheus.MustRegister(OverlayDBGetsABlockSecondary)
	prometheus.MustRegister(OverlayDBGetsFBlock)
	prometheus.MustRegister(OverlayDBGetsFBlockNumber)
	prometheus.MustRegister(OverlayDBGetsFBlockSecondary)
	prometheus.MustRegister(OverlayDBGetsEC)
	prometheus.MustRegister(OverlayDBGetsECNumber)
	prometheus.MustRegister(OverlayDBGetsECSecondary)
	prometheus.MustRegister(OverlayDBGetsChainHead)
	prometheus.MustRegister(OverlayDBGetsEBlock)
	prometheus.MustRegister(OverlayDBGetsEBlockNumber)
	prometheus.MustRegister(OverlayDBGetsEBlockSecondary)
	prometheus.MustRegister(OverlayDBGetsEntry)
	prometheus.MustRegister(OverlayDBGetsDirBlockInfo)
	prometheus.MustRegister(OverlayDBGetsDirBlockUnconfirmed)
	prometheus.MustRegister(OverlayDBGetsDirBlockInfoNumber)
	prometheus.MustRegister(OverlayDBGetsDirBlockInfoSecondary)
	prometheus.MustRegister(OverlayDBGetsInvludeIn)
	prometheus.MustRegister(OverlayDBGetsPaidFor)
}

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
