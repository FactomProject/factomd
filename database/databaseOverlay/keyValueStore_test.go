package databaseOverlay_test

import (
	"testing"

	"github.com/PaulSnow/factom2d/common/primitives/random"
	. "github.com/PaulSnow/factom2d/database/databaseOverlay"
	"github.com/PaulSnow/factom2d/database/mapdb"
)

func TestSaveLoadDatabaseEntryHeight(t *testing.T) {
	dbo := NewOverlay(new(mapdb.MapDB))
	defer dbo.Close()

	for i := 0; i < 10; i++ {
		height := random.RandUInt32()
		err := dbo.SaveDatabaseEntryHeight(height)
		if err != nil {
			t.Errorf("%v", err)
		}
		height2, err := dbo.FetchDatabaseEntryHeight()
		if err != nil {
			t.Errorf("%v", err)
		}
		if height != height2 {
			t.Errorf("%v != %v", height, height2)
		}
	}
}
