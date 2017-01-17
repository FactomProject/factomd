package util_test

import (
	"testing"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	// "github.com/FactomProject/factomd/common/primitives/random"
	. "github.com/FactomProject/factomd/util"
)

func TestIsInPendingEntryList(t *testing.T) {
	var vector map[int]interfaces.IPendingEntry = map[int]interfaces.IPendingEntry{
		0: *randomIPendingEntry(0),
		1: *randomIPendingEntry(1),
		2: *randomIPendingEntry(2),
		3: *randomIPendingEntry(3),
	}

	var list []interfaces.IPendingEntry
	for i := 0; i < 4; i++ {
		list = append(list, vector[i])
	}

	// This item is not in the list
	eNotIn := randomIPendingEntry(3)
	eNotIn.ChainID = primitives.NewZeroHash()
	eNotIn.EntryHash = primitives.NewZeroHash()

	// Checking nil panics
	for _, e := range vector {
		found := IsInPendingEntryList(list, e)
		if !found && e.ChainID != nil && e.EntryHash != nil {
			t.Error("This entry does exist in the list, it should be found")
		}
	}

	found := IsInPendingEntryList(list, *eNotIn)
	if found {
		t.Error("Should not be found in the list")
	}
}

func TestEntryCost(t *testing.T) {
	var buffer []byte
	var add []byte
	for i := 0; i < 1000; i++ {
		add = append(add, 0x00)
	}

	var i uint8
	for i = 0; i < 1; i++ {
		cost, err := EntryCost(buffer)
		switch {
		case i == 0:
			if cost != 1 {
				t.Error("Entry cost should be 1, size 0 bytes")
			}
		case i > 1 && i <= 10:
			if cost != i {
				t.Errorf("Entry cost should %d. Size is %d bytes, found cost is %d\n", i, len(buffer), cost)
			}
		case i == 11:
			if err == nil {
				t.Errorf("Entry should be too big. Size is %d, cost is %d, but should be an error\n", len(buffer), cost)
			}
		}

		buffer = append(buffer, add...)
	}
}

func randomIPendingEntry(nilSpot int) *interfaces.IPendingEntry {
	entry := new(interfaces.IPendingEntry)
	entry.ChainID = primitives.RandomHash()
	entry.EntryHash = primitives.RandomHash()
	switch nilSpot {
	case 0:
		entry.ChainID = nil
	case 1:
		entry.EntryHash = nil
	case 2:
		entry.ChainID = nil
		entry.EntryHash = nil

	}

	return entry
}
