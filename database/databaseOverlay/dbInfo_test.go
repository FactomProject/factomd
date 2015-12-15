// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package databaseOverlay_test

import (
	. "github.com/FactomProject/factomd/testHelper"
	//"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/mapdb"
	"testing"
)

func TestSaveLoadDirBlockInfoHead(t *testing.T) {
	b1 := CreateTestDirBlockInfo(nil)

	dbo := NewOverlay(new(mapdb.MapDB))
	defer dbo.Close()

	err := dbo.SaveDirBlockInfoHead(b1)
	if err != nil {
		t.Error(err)
	}

	head, err := dbo.FetchDirBlockInfoHead()
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

	b2 := CreateTestDirBlockInfo(b1)

	err = dbo.SaveDirBlockInfoHead(b2)
	if err != nil {
		t.Error(err)
	}

	head, err = dbo.FetchDirBlockInfoHead()
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
