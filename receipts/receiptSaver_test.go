// +build all

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package receipts_test

import (
	"testing"

	. "github.com/FactomProject/factomd/receipts"
	. "github.com/FactomProject/factomd/testHelper"
)

func TestReceiptSaver(t *testing.T) {
	dbo := CreateAndPopulateTestDatabaseOverlay()
	err := ExportAllEntryReceipts(dbo)
	if err != nil {
		t.Error(err)
	}

	//t.Fail()
}
