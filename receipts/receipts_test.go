// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package receipts_test

import (
	. "github.com/FactomProject/factomd/receipts"
	. "github.com/FactomProject/factomd/testHelper"
	"testing"
)

func TestReceipts(t *testing.T) {
	dbo := CreateAndPopulateTestDatabaseOverlay()
	entry := CreateFirstTestEntry()
	receipt, err := CreateFullReceipt(dbo, entry.DatabasePrimaryIndex())
	if err != nil {
		t.Error(err)
	}
	t.Logf("%v", receipt.String())
	t.Fail()
}
