package util_test

import (
	"testing"

	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/primitives"
	// "github.com/PaulSnow/factom2d/common/primitives/random"
	. "github.com/PaulSnow/factom2d/util"
)

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
