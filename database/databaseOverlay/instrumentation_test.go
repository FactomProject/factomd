// +build all 

package databaseOverlay_test

import (
	"testing"

	. "github.com/FactomProject/factomd/database/databaseOverlay"
)

func TestInstrumentation(t *testing.T) {
	RegisterPrometheus()
	RegisterPrometheus()

	GetBucket(DIRECTORYBLOCK)
	GetBucket(DIRECTORYBLOCK_NUMBER)
	GetBucket(DIRECTORYBLOCK_SECONDARYINDEX)
	GetBucket(ADMINBLOCK)
	GetBucket(ADMINBLOCK_NUMBER)
	GetBucket(ADMINBLOCK_SECONDARYINDEX)
	GetBucket(FACTOIDBLOCK)
	GetBucket(FACTOIDBLOCK_NUMBER)
	GetBucket(FACTOIDBLOCK_SECONDARYINDEX)
	GetBucket(ENTRYCREDITBLOCK)
	GetBucket(ENTRYCREDITBLOCK_NUMBER)
	GetBucket(ENTRYCREDITBLOCK_SECONDARYINDEX)
	GetBucket(CHAIN_HEAD)
	GetBucket(ENTRYBLOCK)
	GetBucket(ENTRYBLOCK_CHAIN_NUMBER)
	GetBucket(ENTRYBLOCK_SECONDARYINDEX)
	GetBucket(ENTRY)
	GetBucket(DIRBLOCKINFO)
	GetBucket(DIRBLOCKINFO_UNCONFIRMED)
	GetBucket(DIRBLOCKINFO_NUMBER)
	GetBucket(DIRBLOCKINFO_SECONDARYINDEX)
	GetBucket(INCLUDED_IN)
	GetBucket(PAID_FOR)
}
