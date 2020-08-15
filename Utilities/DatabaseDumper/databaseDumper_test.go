package main

import (
	"testing"

	"github.com/PaulSnow/factom2d/testHelper"
)

func TestTest(t *testing.T) {
	dbo := testHelper.CreateAndPopulateTestDatabaseOverlay()

	err := ExportDatabaseJSON(dbo, true)
	if err != nil {
		t.Error(err)
	}
}
