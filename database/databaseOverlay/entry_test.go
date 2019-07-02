// +build all

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package databaseOverlay_test

import (
	"testing"

	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/mapdb"
	"github.com/FactomProject/factomd/testHelper"
)

func TestSaveLoadEntries(t *testing.T) {
	dbo := NewOverlay(new(mapdb.MapDB))
	defer dbo.Close()

	entries := []*entryBlock.Entry{}

	firstEntry := testHelper.CreateFirstTestEntry()
	err := dbo.InsertEntry(firstEntry)
	if err != nil {
		t.Error(err)
	}
	entries = append(entries, firstEntry)

	max := 10
	for i := 0; i < max; i++ {
		entry := testHelper.CreateTestEntry(uint32(i))
		err = dbo.InsertEntry(entry)
		if err != nil {
			t.Error(err)
		}
		entries = append(entries, entry)
	}

	for _, entry := range entries {
		loaded, err := dbo.FetchEntry(entry.GetHash())
		if err != nil {
			t.Error(err)
		}

		m1, err := entry.MarshalBinary()
		if err != nil {
			t.Error(err)
		}
		m2, err := loaded.MarshalBinary()
		if err != nil {
			t.Error(err)
		}
		if primitives.AreBytesEqual(m1, m2) == false {
			t.Error("Entries are not equal")
		}
	}

	all, err := dbo.FetchAllEntriesByChainID(firstEntry.GetChainIDHash())
	if err != nil {
		t.Error(err)
	}

	if len(entries) != len(all) {
		t.Errorf("Loaded %v out of %v entries", len(all), len(entries))
	}

	foundCount := 0
	for _, entry := range entries {
		found := false
		for _, loaded := range all {
			if entry.GetHash().IsSameAs(loaded.GetHash()) {
				found = true

				m1, err := entry.MarshalBinary()
				if err != nil {
					t.Error(err)
				}
				m2, err := loaded.MarshalBinary()
				if err != nil {
					t.Error(err)
				}
				if primitives.AreBytesEqual(m1, m2) == false {
					t.Error("Entries are not equal")
				}
				break
			}
		}
		if found == false {
			t.Errorf("Entry %v not found", entry)
		} else {
			foundCount++
		}
	}
	if foundCount != len(entries) {
		t.Errorf("Found %v out of %v entries", foundCount, len(entries))
	}
}

func TestLoadUnknownEntries(t *testing.T) {
	dbo := NewOverlay(new(mapdb.MapDB))
	defer dbo.Close()
	for i := 0; i < 10; i++ {
		b := testHelper.IntToByteSlice(i)
		hash, err := primitives.NewShaHash(b)
		if err != nil {
			t.Error(err)
		}
		data, err := dbo.FetchEntry(hash)
		if err != nil {
			t.Error(err)
		}
		if data != nil {
			t.Errorf("Fetched entry while we expected nil - %v", data)
		}
		all, err := dbo.FetchAllEntriesByChainID(hash)
		if err != nil {
			t.Error(err)
		}
		if len(all) != 0 {
			t.Errorf("Fetched entries while we expected nil - %v", all)
		}
	}
}
