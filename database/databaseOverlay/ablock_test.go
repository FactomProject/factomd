// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package databaseOverlay_test

import (
	. "github.com/FactomProject/factomd/common/adminBlock"
	//"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/mapdb"
	"testing"
)

func TestSaveLoadABlockHead(t *testing.T) {
	b1 := createTestAdminBlock(nil)

	dbo := NewOverlay(new(mapdb.MapDB))
	defer dbo.Close()

	err := dbo.SaveABlockHead(b1)
	if err != nil {
		t.Error(err)
	}

	head, err := dbo.FetchABlockHead()
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

	if primitives.AreBytesEqual(m1, m2) == false {
		t.Error("Blocks are not equal")
	}

	b2 := createTestAdminBlock(b1)

	err = dbo.SaveABlockHead(b2)
	if err != nil {
		t.Error(err)
	}

	head, err = dbo.FetchABlockHead()
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

func TestSaveLoadABlockChain(t *testing.T) {
	blocks := []*AdminBlock{}
	max := 10
	var prev *AdminBlock = nil
	dbo := NewOverlay(new(mapdb.MapDB))
	defer dbo.Close()

	for i := 0; i < max; i++ {
		prev = createTestAdminBlock(prev)
		blocks = append(blocks, prev)
		err := dbo.SaveABlockHead(prev)
		if err != nil {
			t.Error(err)
		}
	}

	current, err := dbo.FetchABlockHead()
	if err != nil {
		t.Error(err)
	}
	zero := primitives.NewZeroHash()
	fetchedCount := 1
	for {
		keyMR := current.GetHeader().GetPrevLedgerKeyMR()
		if keyMR.IsSameAs(zero) {
			break
		}
		t.Logf("KeyMR - %v", keyMR.String())

		current, err = dbo.FetchABlockByKeyMR(keyMR)
		if err != nil {
			t.Error(err)
		}
		if current == nil {
			t.Fatal("Block not found")
		}
		fetchedCount++
		hash, err := current.PartialHash()
		if err != nil {
			t.Error(err)
		}

		byHash, err := dbo.FetchABlockByHash(hash)

		same, err := primitives.AreBinaryMarshallablesEqual(current, byHash)
		if err != nil {
			t.Error(err)
		}
		if same == false {
			t.Error("Blocks fetched by keyMR and hash are not identical")
		}
	}
	if fetchedCount != max {
		t.Errorf("Wrong number of entries fetched - %v vs %v", fetchedCount, max)
	}

	all, err := dbo.FetchAllABlocks()
	if err != nil {
		t.Error(err)
	}
	if len(all) != max {
		t.Error("Wrong number of entries fetched - %v vs %v", len(all), max)
	}
	for i := range all {
		same, err := primitives.AreBinaryMarshallablesEqual(blocks[i], all[i])
		if err != nil {
			t.Error(err)
		}
		if same == false {
			t.Error("Blocks fetched by all and original blocks are not identical")
			t.Logf("\n%v\nvs\n%v", blocks[i].String(), all[i].String())
		}
	}
}

func createTestAdminBlock(prev *AdminBlock) *AdminBlock {
	block := new(AdminBlock)
	block.SetHeader(createTestAdminHeader(prev))
	block.GetHeader().SetMessageCount(uint32(len(block.GetABEntries())))
	return block
}

func createTestAdminHeader(prev *AdminBlock) *ABlockHeader {
	header := new(ABlockHeader)

	if prev == nil {
		header.PrevLedgerKeyMR = primitives.NewZeroHash()
		header.DBHeight = 0
	} else {
		keyMR, err := prev.GetKeyMR()
		if err != nil {
			panic(err)
		}
		header.PrevLedgerKeyMR = keyMR
		header.DBHeight = prev.Header.GetDBHeight() + 1
	}

	header.HeaderExpansionSize = 5
	header.HeaderExpansionArea = []byte{0x00, 0x01, 0x02, 0x03, 0x04}
	header.MessageCount = 0
	header.BodySize = 0

	return header
}
