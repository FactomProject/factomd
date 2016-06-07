package wsapi_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/receipts"
	"github.com/FactomProject/factomd/testHelper"
	. "github.com/FactomProject/factomd/wsapi"
)

func TestHandleV2GetRaw(t *testing.T) {
	type RawData struct {
		Hash1 string
		Hash2 string
		Raw   string
	}

	toTest := []RawData{}
	var err error

	blockSet := testHelper.CreateTestBlockSet(nil)

	aBlock := blockSet.ABlock
	raw := RawData{}
	raw.Hash1 = aBlock.DatabasePrimaryIndex().String()
	raw.Hash2 = aBlock.DatabaseSecondaryIndex().String()
	hex, err := aBlock.MarshalBinary()
	if err != nil {
		panic(err)
	}
	raw.Raw = primitives.EncodeBinary(hex)
	toTest = append(toTest, raw) //1

	eBlock := blockSet.EBlock
	raw = RawData{}
	raw.Hash1 = eBlock.DatabasePrimaryIndex().String()
	raw.Hash2 = eBlock.DatabaseSecondaryIndex().String()
	hex, err = eBlock.MarshalBinary()
	if err != nil {
		panic(err)
	}
	raw.Raw = primitives.EncodeBinary(hex)
	toTest = append(toTest, raw) //2

	ecBlock := blockSet.ECBlock
	raw = RawData{}
	raw.Hash1 = ecBlock.(interfaces.DatabaseBatchable).DatabasePrimaryIndex().String()
	raw.Hash2 = ecBlock.(interfaces.DatabaseBatchable).DatabaseSecondaryIndex().String()
	hex, err = ecBlock.MarshalBinary()
	if err != nil {
		panic(err)
	}
	raw.Raw = primitives.EncodeBinary(hex)
	toTest = append(toTest, raw) //3

	fBlock := blockSet.FBlock
	raw = RawData{}
	raw.Hash1 = fBlock.(interfaces.DatabaseBatchable).DatabasePrimaryIndex().String()
	raw.Hash2 = fBlock.(interfaces.DatabaseBatchable).DatabaseSecondaryIndex().String()
	hex, err = fBlock.MarshalBinary()
	if err != nil {
		panic(err)
	}
	raw.Raw = primitives.EncodeBinary(hex)
	toTest = append(toTest, raw) //4

	dBlock := blockSet.DBlock
	raw = RawData{}
	raw.Hash1 = dBlock.DatabasePrimaryIndex().String()
	raw.Hash2 = dBlock.DatabaseSecondaryIndex().String()
	hex, err = dBlock.MarshalBinary()
	if err != nil {
		panic(err)
	}
	raw.Raw = primitives.EncodeBinary(hex)
	toTest = append(toTest, raw) //5

	//initializing server
	state := testHelper.CreateAndPopulateTestState()
	Start(state)

	for i, v := range toTest {
		data := new(HashRequest)
		data.Hash = v.Hash1
		req := primitives.NewJSON2Request("raw-data", 1, data)

		resp, err := v2Request(req)
		if err != nil {
			t.Errorf("%v", err)
		}

		if strings.Contains(resp.String(), v.Raw) == false {
			t.Errorf("Looking for %v", v.Hash1)
			t.Errorf("GetRaw %v/%v from Hash1 failed - %v", i, len(toTest), resp.String())
		}

		data.Hash = v.Hash2
		req = primitives.NewJSON2Request("raw-data", 1, data)
		resp, err = v2Request(req)
		if err != nil {
			t.Errorf("%v", err)
		}

		if strings.Contains(resp.String(), v.Raw) == false {
			t.Errorf("Looking for %v", v.Hash1)
			t.Errorf("GetRaw %v/%v from Hash2 failed - %v", i, len(toTest), resp.String())
		}
	}
}

func v2Request(req *primitives.JSON2Request) (*primitives.JSON2Response, error) {
	j, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(
		"http://localhost:8088/v2",
		"application/json",
		bytes.NewBuffer(j))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	r := primitives.NewJSON2Response()
	if err := json.Unmarshal(body, r); err != nil {
		return nil, err
	}

	return r, nil
}

func TestHandleV2CommitEntry(t *testing.T) {
	msg := new(EntryRequest)
	// Can replace with any Entry message
	msg.Entry = "00015507C1024BF5C956749FC3EBA4ACC60FD485FB100E601070A44FCCE54FF358D60669854734013B6A27BCCEB6A42D62A3A8D02A6F0D73653215771DE243A63AC048A18B59DA29F4CBD953E6EBE684D693FDCA270CE231783E8ECC62D630F983CD59E559C6253F84D1F54C8E8D8665D493F7B4A4C1864751E3CDEC885A64C2144E0938BF648A00"
	req := primitives.NewJSON2Request("commit-entry", 0, msg)
	resp, err := v2Request(req)
	if err != nil {
		t.Errorf("%v", err)
	}

	respObj := new(CommitEntryResponse)
	if err := MapToObject(resp.Result, respObj); err != nil {
		t.Error(err)
	}

	txID := CheckEntryTransactionID(msg.Entry)
	e := strings.Compare(respObj.TxID, txID)
	if e != 0 {
		t.Error("Error: TxID returned during Commit Entry is incorrect")
	}

}

func CheckEntryTransactionID(entryMsg string) string {
	//entryMsg := msg

	entryBytes, _ := hex.DecodeString(entryMsg)
	// 144 Hex : 72 Bytes
	entryTxID := sha256.Sum256(entryBytes[:72])

	entryTxIDHex := hex.EncodeToString(entryTxID[:])

	return entryTxIDHex
}

func CheckChainTransactionID(chainMsg string) string {
	chainBytes, _ := hex.DecodeString(chainMsg)
	// 272 Hex : 136 Bytes
	chainTxID := sha256.Sum256(chainBytes[:136])

	chainTxIDHex := hex.EncodeToString(chainTxID[:])

	return chainTxIDHex
}

func TestHandleV2CommitChain(t *testing.T) {
	msg := new(MessageRequest)
	// Can replace with any Chain message
	msg.Message = "00015507b2f70bd0165d9fa19a28cfaafb6bc82f538955a98c7b7e60d79fbf92655c1bff1c76466cb3bc3f3cc68d8b2c111f4f24c88d9c031b4124395c940e5e2c5ea496e8aaa2f5c956749fc3eba4acc60fd485fb100e601070a44fcce54ff358d606698547340b3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da2946c901273e616bdbb166c535b26d0d446bc69b22c887c534297c7d01b2ac120237086112b5ef34fc6474e5e941d60aa054b465d4d770d7f850169170ef39150b"
	req := primitives.NewJSON2Request("commit-chain", 0, msg)
	resp, err := v2Request(req)
	if err != nil {
		t.Errorf("%v", err)
	}

	respObj := new(CommitChainResponse)
	if err := MapToObject(resp.Result, respObj); err != nil {
		t.Error(err)
	}

	txID := CheckChainTransactionID(msg.Message)
	e := strings.Compare(respObj.TxID, txID)
	if e != 0 {
		t.Error("Error: TxID returned during Commit Chain is incorrect")
	}
}

func TestHandleV2GetReceipt(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	//Start(state)

	hashkey := new(HashRequest)
	hashkey.Hash = "8c35266c406e5a42fc3ca93f2d850b954bdfa79f49b2ceaf7f7086b691ffc022"

	resp, jErr := HandleV2Receipt(state, hashkey)
	if jErr != nil {
		t.Errorf("%v", jErr)
		return
	}

	dbo := state.GetAndLockDB()
	defer state.UnlockDB()

	marshalled, err := json.Marshal(resp.(*ReceiptResponse).Receipt)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Resp - %s", marshalled)

	err = receipts.VerifyFullReceipt(dbo, string(marshalled))
	if err != nil {
		t.Logf("receipt - %s", marshalled)
		t.Error(err)
	}
}
