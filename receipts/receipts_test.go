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
	t.Logf("\n\n%v\n", receipt.String())

	err = VerifyFullReceipt(dbo, receipt.String())
	if err != nil {
		t.Error(err)
	}

	receipt.TrimReceipt()
	t.Logf("\n\n%v\n", receipt.String())

	err = VerifyFullReceipt(dbo, receipt.String())
	if err == nil {
		t.Errorf("\n\nError is nil when it shouldn't be for receipt\n\n%v\n\n", receipt)
	}

	err = VerifyMinimalReceipt(dbo, receipt.String())
	if err != nil {
		t.Error(err)
	}

	//t.Fail()
}
