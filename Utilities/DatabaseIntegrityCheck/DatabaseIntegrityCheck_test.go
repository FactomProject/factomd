package main

import (
	"testing"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/testHelper"
)

func TestCheckDatabaseFromDBO(t *testing.T) {
	dbo := testHelper.CreateAndPopulateTestDatabaseOverlay()
	CheckDatabase(dbo.DB)
}

func TestCheckDatabaseFromState(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	CheckDatabase(state.DB.DB)
}

func TestCheckDatabaseFromWSAPI(t *testing.T) {
	ctx := testHelper.CreateWebContext()
	state := ctx.Server.Env["state"].(interfaces.IState)
	dbase := state.GetAndLockDB()
	defer state.UnlockDB()

	CheckDatabase(dbase)
}
