package main

import (
	"os"
	"testing"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/leveldb"
	"github.com/FactomProject/factomd/testHelper"
)

func TestCheckDatabaseFromDBO(t *testing.T) {
	dbo := testHelper.CreateAndPopulateTestDatabaseOverlay()
	CheckDatabase(dbo)
}

func TestCheckDatabaseFromState(t *testing.T) {
	state := testHelper.CreateAndPopulateTestStateAndStartValidator()
	CheckDatabase(state.DB.(interfaces.DBOverlay))
}

var dbFilename string = "levelTest.db"

func TestCheckDatabaseForLevelDB(t *testing.T) {
	m, err := leveldb.NewLevelDB(dbFilename, true)
	if err != nil {
		t.Errorf("%v", err)
	}
	defer CleanupLevelDB(t, m)

	dbo := databaseOverlay.NewOverlay(m)
	testHelper.PopulateTestDatabaseOverlay(dbo)

	CheckDatabase(dbo)

}

func CleanupLevelDB(t *testing.T, b interfaces.IDatabase) {
	err := b.Close()
	if err != nil {
		t.Errorf("%v", err)
	}
	err = os.RemoveAll(dbFilename)
	if err != nil {
		t.Errorf("%v", err)
	}
}
