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
	t.Logf("\n\n%v\n", receipt.CustomMarshalString())

	err = VerifyFullReceipt(dbo, receipt.CustomMarshalString())
	if err != nil {
		t.Error(err)
	}

	receipt.TrimReceipt()
	t.Logf("\n\n%v\n", receipt.CustomMarshalString())

	err = VerifyFullReceipt(dbo, receipt.CustomMarshalString())
	if err == nil {
		t.Errorf("\n\nError is nil when it shouldn't be for receipt\n\n%v\n\n", receipt)
	}

	err = VerifyMinimalReceipt(dbo, receipt.CustomMarshalString())
	if err != nil {
		t.Error(err)
	}

	//t.Fail()
}

func TestDecodeReceiptString(t *testing.T) {
	receiptStr := `{"BitcoinBlockHash":"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","BitcoinTransactionHash":"0000000000000000000000000000000000000000000000000000000000000000","DirectoryBlockKeyMR":"bdadd16c5335c369a1b784212f80764e1f47805c89d39141bd40d05153edcdf5","Entry":{"Key":"cf9503fad6a6cf3cf6d7a5a491e23d84f9dee6dacb8c12f428633995655bd0d0"},"EntryBlockKeyMR":"905740850540f1d17fcb1fc7fd0c61a33150b2cdc0f88334f6a891ec34bd1cfc","MerkleBranch":[{"Left":"0a2f96c96ea89ee82908be9f5aef2be4b533a32ffb3855aeb3b8327f9e989f3a","Right":"cf9503fad6a6cf3cf6d7a5a491e23d84f9dee6dacb8c12f428633995655bd0d0","Top":"905740850540f1d17fcb1fc7fd0c61a33150b2cdc0f88334f6a891ec34bd1cfc"},{"Left":"6e7e64ac45ff57edbf8537a0c99fba2e9ee351ef3d3f4abd93af9f01107e592c","Right":"905740850540f1d17fcb1fc7fd0c61a33150b2cdc0f88334f6a891ec34bd1cfc","Top":"4f477201a150694ed0f85fee17c41282542f976fae479a4de553a37747b09f41"},{"Left":"4f477201a150694ed0f85fee17c41282542f976fae479a4de553a37747b09f41","Right":"18ab692a40f370e9529c180f2476684ccde4937b9a4b4605805e3f51e592f632","Top":"890003f0db6cceca94031a70745fd83845726987cffa6fc95ddb0e2f6c64b499"},{"Left":"1857570da9a1c93dac4993d3048faa80d1d1d939f4fc44a38e61781fdc123165","Right":"890003f0db6cceca94031a70745fd83845726987cffa6fc95ddb0e2f6c64b499","Top":"4d8ed632f7852a07055a0592c341b957815bdd46e82d2da7bdf58be54fc60bf9"},{"Left":"4d8ed632f7852a07055a0592c341b957815bdd46e82d2da7bdf58be54fc60bf9","Right":"f955a2709628086d656257885bf27b7c054a6acd0b3ebf5b769b3cf036ab04ee","Top":"d6bd24e979e81feddb319483878c678865a80175d1954e5429f2d799eadd1bc9"},{"Left":"49a5c28516f3c4d5e44f5cf0b2e5f5f00ca1187714dd9ee914e7df1eb7702972","Right":"d6bd24e979e81feddb319483878c678865a80175d1954e5429f2d799eadd1bc9","Top":"bdadd16c5335c369a1b784212f80764e1f47805c89d39141bd40d05153edcdf5"}]}`
	receipt, err := DecodeReceiptString(receiptStr)
	if err != nil {
		t.Error(err)
	}
	err = receipt.Validate()
	if err != nil {
		t.Logf("Receipt - %v", receipt.CustomMarshalString())
		t.Error(err)
	}
}
