// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package databaseOverlay_test

import (
	"testing"

	"github.com/FactomProject/factomd/common/directoryBlock/dbInfo"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/mapdb"
	"github.com/FactomProject/factomd/testHelper"
)

func TestSaveLoadDirBlockInfo(t *testing.T) {
	b1 := testHelper.CreateTestDirBlockInfo(nil)

	dbo := NewOverlay(new(mapdb.MapDB))
	defer dbo.Close()

	err := dbo.SaveDirBlockInfo(b1)
	if err != nil {
		t.Error(err)
	}

	head, err := dbo.FetchDirBlockInfoByHash(b1.DBHash)
	if err != nil {
		t.Error(err)
	}
	if head == nil {
		t.Error("DirBlockInfo head is nil")
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

	b2 := testHelper.CreateTestDirBlockInfo(b1)

	err = dbo.SaveDirBlockInfo(b2)
	if err != nil {
		t.Error(err)
	}

	head, err = dbo.FetchDirBlockInfoByKeyMR(b1.DBMerkleRoot)
	if err != nil {
		t.Error(err)
	}
	if head == nil {
		t.Error("DirBlockInfo head is nil")
	}

	m1, err = b1.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	m2, err = head.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	if primitives.AreBytesEqual(m1, m2) == false {
		t.Error("Blocks are not equal")
	}

	head, err = dbo.FetchDirBlockInfoByKeyMR(b2.DBMerkleRoot)
	if err != nil {
		t.Error(err)
	}
	if head == nil {
		t.Error("DirBlockInfo head is nil")
	}

	m1, err = b2.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	m2, err = head.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	if primitives.AreBytesEqual(m1, m2) == false {
		t.Error("Blocks are not equal")
	}

	head, err = dbo.FetchDirBlockInfoByHash(b2.DBHash)
	if err != nil {
		t.Error(err)
	}
	if head == nil {
		t.Error("DirBlockInfo head is nil")
	}

	m1, err = b2.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	m2, err = head.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	if primitives.AreBytesEqual(m1, m2) == false {
		t.Error("Blocks are not equal")
	}
}

func TestFetchDirBlockInfoBatches(t *testing.T) {
	dbo := NewOverlay(new(mapdb.MapDB))
	defer dbo.Close()

	blocks := []*dbInfo.DirBlockInfo{}
	max := 10
	var prev *dbInfo.DirBlockInfo = nil
	for i := 0; i < max; i++ {
		prev = testHelper.CreateTestDirBlockInfo(prev)
		blocks = append(blocks, prev)

		err := dbo.SaveDirBlockInfo(prev)
		if err != nil {
			t.Error(err)
		}
	}

	all, err := dbo.FetchAllDirBlockInfos()
	if err != nil {
		t.Error(err)
	}
	if len(all) != max {
		t.Errorf("Returned %d infos, expected %d", len(all), max)
	}

	confirmed, err := dbo.FetchAllConfirmedDirBlockInfos()
	if err != nil {
		t.Error(err)
	}
	if len(confirmed) != max/2 {
		t.Errorf("Returned %d infos, expected %d", len(confirmed), max/2)
	}
	for _, info := range confirmed {
		if info.GetBTCConfirmed() == false {
			t.Error("Confirmed transaction is unconfirmed")
		}
	}

	unconfirmed, err := dbo.FetchAllUnconfirmedDirBlockInfos()
	if err != nil {
		t.Error(err)
	}
	if len(unconfirmed) != max/2 {
		t.Errorf("Returned %d infos, expected %d", len(unconfirmed), max/2)
	}
	for _, info := range unconfirmed {
		if info.GetBTCConfirmed() == true {
			t.Error("Unconfirmed transaction is confirmed")
		}
	}

	for i := 0; i < max; i++ {
		m1, err := blocks[i].MarshalBinary()
		if err != nil {
			t.Error(err)
		}
		m2, err := all[i].MarshalBinary()
		if err != nil {
			t.Error(err)
		}
		if primitives.AreBytesEqual(m1, m2) == false {
			t.Error("Blocks are not equal")
		}

		if i%2 == 0 {
			m2, err = unconfirmed[i/2].MarshalBinary()
			if err != nil {
				t.Error(err)
			}
		} else {
			m2, err = confirmed[i/2].MarshalBinary()
			if err != nil {
				t.Error(err)
			}
		}
		if primitives.AreBytesEqual(m1, m2) == false {
			t.Error("Blocks are not equal")
		}

	}
}

func TestLoadUnknownDirBlockEntries(t *testing.T) {
	dbo := NewOverlay(new(mapdb.MapDB))
	defer dbo.Close()
	for i := 0; i < 10; i++ {
		b := testHelper.IntToByteSlice(i)
		hash, err := primitives.NewShaHash(b)
		if err != nil {
			t.Error(err)
		}
		data, err := dbo.FetchDirBlockInfoByHash(hash)
		if err != nil {
			t.Error(err)
		}
		if data != nil {
			t.Errorf("Fetched entry while we expected nil - %v", data)
		}
		data, err = dbo.FetchDirBlockInfoByKeyMR(hash)
		if err != nil {
			t.Error(err)
		}
		if data != nil {
			t.Errorf("Fetched entry while we expected nil - %v", data)
		}
		all, err := dbo.FetchAllDirBlockInfos()
		if err != nil {
			t.Error(err)
		}
		if len(all) != 0 {
			t.Errorf("Fetched entries while we expected nil - %v", all)
		}
	}
}
