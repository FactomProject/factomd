// +build all

package databaseOverlay_test

import (
	"testing"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/testHelper"
)

func TestFetchFactoidTransactionByHash(t *testing.T) {
	blocks := CreateFullTestBlockSet()
	dbo := CreateAndPopulateTestDatabaseOverlay()

	for _, block := range blocks {
		for _, tx := range block.FBlock.GetTransactions() {
			dTx, err := dbo.FetchFactoidTransaction(tx.GetHash())
			if err != nil {
				t.Errorf("%v", err)
			}
			if dTx == nil {
				t.Errorf("Tx %v not found!", tx.GetHash().String())
				continue
			}

			h1, err := tx.MarshalBinary()
			if err != nil {
				t.Errorf("%v", err)
			}

			h2, err := dTx.MarshalBinary()
			if err != nil {
				t.Errorf("%v", err)
			}

			if primitives.AreBytesEqual(h1, h2) == false {
				t.Error("Returned transactions are not equal")
			}
		}
	}
}

func TestFetchECTransactionByHash(t *testing.T) {
	blocks := CreateFullTestBlockSet()
	dbo := CreateAndPopulateTestDatabaseOverlay()

	for _, block := range blocks {
		for _, tx := range block.ECBlock.GetEntries() {
			if tx.ECID() != constants.ECIDChainCommit && tx.ECID() != constants.ECIDEntryCommit || tx.ECID() == constants.ECIDBalanceIncrease {
				continue
			}

			dTx, err := dbo.FetchECTransaction(tx.Hash())
			if err != nil {
				t.Errorf("%v", err)
			}
			if dTx == nil {
				t.Errorf("Tx %v not found!", tx.Hash().String())
				continue
			}

			h1, err := tx.MarshalBinary()
			if err != nil {
				t.Errorf("%v", err)
			}

			h2, err := dTx.MarshalBinary()
			if err != nil {
				t.Errorf("%v", err)
			}

			if primitives.AreBytesEqual(h1, h2) == false {
				t.Error("Returned transactions are not equal")
			}
		}
	}
}
