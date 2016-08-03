package main

import (
	"github.com/FactomProject/factomd/testHelper"
	"testing"
)

func TestTest(t *testing.T) {
	dbo := testHelper.CreateAndPopulateTestDatabaseOverlay()

	err := ExportDatabaseJSON(dbo, true)
	if err != nil {
		t.Error(err)
	}
}
