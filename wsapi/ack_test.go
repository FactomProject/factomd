package wsapi_test

import (
	"encoding/hex"
	//"fmt"
	"testing"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/testHelper"
	. "github.com/FactomProject/factomd/wsapi"
)

func TestDecodeTransactionToHashes(t *testing.T) {
	blocks := testHelper.CreateFullTestBlockSet()

	for _, block := range blocks {
		for _, tx := range block.Entries {
			txID := tx.GetHash().String()

			h, err := tx.MarshalBinary()
			if err != nil {
				t.Errorf("%v", err)
				continue
			}
			fullTx := hex.EncodeToString(h)

			eTxID, ecTxID := DecodeTransactionToHashes(fullTx)
			if eTxID == "" && ecTxID == "" {
				t.Error("No TxID returned")
				continue
			}
			if ecTxID != "" {
				t.Error("Entry mistaken for EC Transaction")
			}

			if eTxID != txID {
				t.Errorf("Returned wrong Entry hash - %v vs %v", eTxID, txID)
			}
		}

		for _, tx := range block.ECBlock.GetEntries() {
			if tx.ECID() != constants.ECIDChainCommit && tx.ECID() != constants.ECIDEntryCommit {
				continue
			}
			/*if tx.ECID() == constants.ECIDChainCommit {
				fmt.Println("CC!")
			}
			if tx.ECID() == constants.ECIDEntryCommit {
				fmt.Println("EC!")
			}*/
			txID := tx.GetHash().String()
			entryHash := tx.GetEntryHash().String()

			h, err := tx.MarshalBinary()
			if err != nil {
				t.Errorf("%v", err)
				continue
			}
			fullTx := hex.EncodeToString(h)

			eTxID, ecTxID := DecodeTransactionToHashes(fullTx)
			if eTxID == "" && ecTxID == "" {
				t.Error("No TxID returned")
				continue
			}

			if eTxID != entryHash {
				t.Errorf("Returned wrong Entry hash - %v vs %v", eTxID, entryHash)
			}

			if ecTxID != txID {
				t.Errorf("Returned wrong EC TxID - %v vs %v", ecTxID, txID)
			}
		}
	}
}

func TestHandleV2FactoidACK(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	blocks := testHelper.CreateFullTestBlockSet()

	for _, block := range blocks {
		for _, tx := range block.FBlock.GetTransactions() {
			req := AckRequest{}
			txID := tx.GetSigHash().String()
			req.TxID = txID

			r, jError := HandleV2FactoidACK(state, req)

			if jError != nil {
				t.Errorf("%v", jError)
				continue
			}

			resp, ok := r.(*FactoidTxStatus)
			if ok == false {
				t.Error("Invalid response type returned")
				continue
			}

			if resp.TxID != txID {
				t.Error("Invalid TxID returned")
			}
			if resp.Status != AckStatusDBlockConfirmed {
				t.Error("Invalid status returned")
			}

			req = AckRequest{}
			h, err := tx.MarshalBinary()
			if err != nil {
				t.Errorf("%v", err)
				continue
			}
			req.FullTransaction = hex.EncodeToString(h)

			r, jError = HandleV2FactoidACK(state, req)

			if jError != nil {
				t.Errorf("%v", jError)
				continue
			}

			resp, ok = r.(*FactoidTxStatus)
			if ok == false {
				t.Error("Invalid response type returned")
				continue
			}

			if resp.TxID != txID {
				t.Error("Invalid TxID returned")
			}
			if resp.Status != AckStatusDBlockConfirmed {
				t.Error("Invalid status returned")
			}
		}
	}

	for i := 0; i < 10; i++ {
		h := testHelper.NewRepeatingHash(byte(i))

		req := AckRequest{}
		req.TxID = h.String()

		r, jError := HandleV2FactoidACK(state, req)

		if jError != nil {
			t.Errorf("%v", jError)
			continue
		}

		resp, ok := r.(*FactoidTxStatus)
		if ok == false {
			t.Error("Invalid response type returned")
			continue
		}

		if resp.TxID != h.String() {
			t.Error("Invalid TxID returned")
		}
		if resp.Status != AckStatusUnknown {
			t.Error("Invalid status returned")
		}

		req = AckRequest{}
		req.FullTransaction = h.String()

		_, jError = HandleV2FactoidACK(state, req)

		if jError == nil {
			t.Error("Invalid transactions not caught")
			continue
		}
	}
}

func TestHandleV2EntryACK(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	blocks := testHelper.CreateFullTestBlockSet()

	for _, block := range blocks {
		for _, tx := range block.Entries {
			req := AckRequest{}
			txID := tx.GetHash().String()
			req.TxID = txID

			r, jError := HandleV2EntryACK(state, req)

			if jError != nil {
				t.Errorf("%v", jError)
				continue
			}

			resp, ok := r.(*EntryStatus)
			if ok == false {
				t.Error("Invalid response type returned")
				continue
			}

			if resp.EntryHash != txID {
				t.Errorf("Invalid EntryHash returned - %v vs %v", resp.EntryHash, txID)
			}
			if resp.CommitTxID == "" {
				t.Errorf("Invalid CommitTxID returned - %v", resp.CommitTxID)
			}
			if resp.CommitData.Status != AckStatusDBlockConfirmed {
				t.Errorf("Invalid status returned - %v vs %v", resp.CommitData.Status, AckStatusDBlockConfirmed)
			}
			if resp.EntryData.Status != AckStatusDBlockConfirmed {
				t.Errorf("Invalid status returned - %v vs %v", resp.EntryData.Status, AckStatusDBlockConfirmed)
			}

			req = AckRequest{}
			h, err := tx.MarshalBinary()
			if err != nil {
				t.Errorf("%v", err)
				continue
			}
			req.FullTransaction = hex.EncodeToString(h)

			r, jError = HandleV2EntryACK(state, req)

			if jError != nil {
				t.Errorf("%v", jError)
				continue
			}

			resp, ok = r.(*EntryStatus)
			if ok == false {
				t.Error("Invalid response type returned")
				continue
			}

			if resp.EntryHash != txID {
				t.Errorf("Invalid EntryHash returned - %v vs %v", resp.EntryHash, txID)
			}
			if resp.CommitTxID == "" {
				t.Errorf("Invalid CommitTxID returned - %v", resp.CommitTxID)
			}
			if resp.CommitData.Status != AckStatusDBlockConfirmed {
				t.Errorf("Invalid status returned - %v vs %v", resp.CommitData.Status, AckStatusDBlockConfirmed)
			}
			if resp.EntryData.Status != AckStatusDBlockConfirmed {
				t.Errorf("Invalid status returned - %v vs %v", resp.EntryData.Status, AckStatusDBlockConfirmed)
			}
		}

		for _, tx := range block.ECBlock.GetEntries() {
			if tx.ECID() != constants.ECIDChainCommit && tx.ECID() != constants.ECIDEntryCommit {
				continue
			}
			req := AckRequest{}

			txID := tx.GetSigHash().String()
			entryHash := tx.GetEntryHash().String()
			req.TxID = txID

			r, jError := HandleV2EntryACK(state, req)

			if jError != nil {
				t.Errorf("%v", jError)
				continue
			}

			resp, ok := r.(*EntryStatus)
			if ok == false {
				t.Error("Invalid response type returned")
				continue
			}
			t.Logf("resp - %v", resp)

			if resp.CommitTxID != txID {
				t.Errorf("Invalid CommitTxID returned - %v vs %v", resp.CommitTxID, txID)
			}
			if resp.EntryHash != entryHash {
				t.Errorf("Invalid EntryHash returned - %v vs %v", resp.EntryHash, entryHash)
			}
			if resp.CommitData.Status != AckStatusDBlockConfirmed {
				t.Errorf("Invalid status returned - %v vs %v", resp.CommitData.Status, AckStatusDBlockConfirmed)
			}
			if resp.EntryData.Status != AckStatusDBlockConfirmed {
				t.Errorf("Invalid status returned - %v vs %v", resp.EntryData.Status, AckStatusDBlockConfirmed)
			}

			req = AckRequest{}
			h, err := tx.MarshalBinary()
			if err != nil {
				t.Errorf("%v", err)
				continue
			}
			req.FullTransaction = hex.EncodeToString(h)

			r, jError = HandleV2EntryACK(state, req)

			if jError != nil {
				t.Errorf("%v", jError)
				continue
			}

			resp, ok = r.(*EntryStatus)
			if ok == false {
				t.Error("Invalid response type returned")
				continue
			}
			t.Logf("resp - %v", resp)

			if resp.CommitTxID != txID {
				t.Errorf("Invalid CommitTxID returned - %v vs %v", resp.CommitTxID, txID)
			}
			if resp.EntryHash != entryHash {
				t.Errorf("Invalid EntryHash returned - %v vs %v", resp.EntryHash, entryHash)
			}
			if resp.CommitData.Status != AckStatusDBlockConfirmed {
				t.Errorf("Invalid status returned - %v vs %v", resp.CommitData.Status, AckStatusDBlockConfirmed)
			}
			if resp.EntryData.Status != AckStatusDBlockConfirmed {
				t.Errorf("Invalid status returned - %v vs %v", resp.EntryData.Status, AckStatusDBlockConfirmed)
			}
		}
	}

	for i := 0; i < 10; i++ {
		h := testHelper.NewRepeatingHash(byte(i))

		req := AckRequest{}
		req.TxID = h.String()

		r, jError := HandleV2EntryACK(state, req)

		if jError != nil {
			t.Errorf("%v", jError)
			continue
		}

		resp, ok := r.(*EntryStatus)
		if ok == false {
			t.Error("Invalid response type returned")
			continue
		}

		if resp.CommitTxID != "" {
			t.Error("Invalid CommitTxID returned")
		}
		if resp.EntryHash != "" {
			t.Error("Invalid EntryHash returned")
		}
		if resp.CommitData.Status != AckStatusUnknown {
			t.Error("Invalid status returned")
		}
		if resp.EntryData.Status != AckStatusUnknown {
			t.Error("Invalid status returned")
		}

		req = AckRequest{}
		req.FullTransaction = h.String()

		_, jError = HandleV2EntryACK(state, req)

		if jError == nil {
			t.Error("Invalid transactions not caught")
			continue
		}
	}
}
