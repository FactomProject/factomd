// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package databaseOverlay_test

import (
	. "github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/mapdb"
	"testing"
)

func TestSaveLoadDBlockHead(t *testing.T) {
	b1 := createTestDirectoryBlock(nil)

	dbo := NewOverlay(new(mapdb.MapDB))

	err := dbo.SaveDirectoryBlockHead(b1)
	if err != nil {
		t.Error(err)
	}

	head, err := dbo.FetchDirectoryBlockHead()
	if err != nil {
		t.Error(err)
	}
	if head == nil {
		t.Error("DBlock head is nil")
	}

	m1, err := b1.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	m2, err := head.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	if AreBytesEqual(m1, m2) == false {
		t.Error("Blocks are not equal")
	}

	b2 := createTestDirectoryBlock(b1)

	err = dbo.SaveDirectoryBlockHead(b2)
	if err != nil {
		t.Error(err)
	}

	head, err = dbo.FetchDirectoryBlockHead()
	if err != nil {
		t.Error(err)
	}
	if head == nil {
		t.Error("DBlock head is nil")
	}

	m1, err = b2.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	m2, err = head.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
}

func createTestDirectoryBlock(prevBlock *DirectoryBlock) *DirectoryBlock {
	dblock := new(DirectoryBlock)

	dblock.SetHeader(createTestDirectoryBlockHeader(prevBlock))

	dblock.SetDBEntries(make([]interfaces.IDBEntry, 0, 5))

	de := new(DBEntry)
	de.ChainID = primitives.NewZeroHash()
	de.KeyMR = primitives.NewZeroHash()

	dblock.SetDBEntries(append(dblock.GetDBEntries(), de))
	dblock.GetHeader().SetBlockCount(uint32(len(dblock.GetDBEntries())))

	return dblock
}

func createTestDirectoryBlockHeader(prevBlock *DirectoryBlock) *DBlockHeader {
	header := new(DBlockHeader)

	header.SetBodyMR(primitives.Sha(primitives.NewZeroHash().Bytes()))
	header.SetBlockCount(0)
	header.SetNetworkID(0xffff)

	if prevBlock == nil {
		header.SetDBHeight(0)
		header.SetPrevLedgerKeyMR(primitives.NewZeroHash())
		header.SetPrevKeyMR(primitives.NewZeroHash())
		header.SetTimestamp(1234)
	} else {
		header.SetDBHeight(prevBlock.Header.GetDBHeight() + 1)
		header.SetPrevLedgerKeyMR(primitives.NewZeroHash())
		keyMR, err := prevBlock.BuildKeyMerkleRoot()
		if err != nil {
			panic(err)
		}
		header.SetPrevKeyMR(keyMR)
		header.SetTimestamp(prevBlock.Header.GetTimestamp() + 1)
	}

	header.SetVersion(1)

	return header
}
