// +build all

package main

import (
	"testing"

	"github.com/FactomProject/factomd/testHelper"
)

func TestTest(t *testing.T) {
	dbo := testHelper.CreateAndPopulateTestDatabaseOverlay()

	err := ExportDatabaseJSON(dbo, true)
	if err != nil {
		t.Error(err)
	}
}
