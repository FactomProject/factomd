package wsapi_test

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/receipts"
	"github.com/FactomProject/factomd/testHelper"
	. "github.com/FactomProject/factomd/wsapi"
)

func TestRegisterPrometheus(t *testing.T) {
	RegisterPrometheus()
	RegisterPrometheus()
}

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

		time.Sleep(time.Millisecond * 100)
		resp, err := v2Request(req)
		assert.Nil(t, err)
		assert.True(t, strings.Contains(resp.String(), v.Raw), "Looking for %v but got %v \nGetRaw %v/%v from Hash1 failed - %v", v.Hash1, v.Raw, i, len(toTest), resp.String())

		data.Hash = v.Hash2
		req = primitives.NewJSON2Request("raw-data", 1, data)
		resp, err = v2Request(req)
		assert.Nil(t, err)
		assert.True(t, strings.Contains(resp.String(), v.Raw), "Looking for %v \nGetRaw %v/%v from Hash2 failed - %v", v.Hash1, i, len(toTest), resp.String())
	}
}

func TestHandleV2CommitChain(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	Start(state)

	msg := new(MessageRequest)
	// Can replace with any Chain message
	msg.Message = "00015507b2f70bd0165d9fa19a28cfaafb6bc82f538955a98c7b7e60d79fbf92655c1bff1c76466cb3bc3f3cc68d8b2c111f4f24c88d9c031b4124395c940e5e2c5ea496e8aaa2f5c956749fc3eba4acc60fd485fb100e601070a44fcce54ff358d606698547340b3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da2946c901273e616bdbb166c535b26d0d446bc69b22c887c534297c7d01b2ac120237086112b5ef34fc6474e5e941d60aa054b465d4d770d7f850169170ef39150b"
	req := primitives.NewJSON2Request("commit-chain", 0, msg)
	resp, err := v2Request(req)
	assert.Nil(t, err)

	respObj := new(CommitChainResponse)
	err = MapToObject(resp.Result, respObj)
	assert.Nil(t, err)

	txID := "76e123d133a841fe3e08c5e3f3d392f8431f2d7668890c03f003f541efa8fc61"
	assert.Equal(t, txID, respObj.TxID, "Error: TxID returned during Commit Chain is incorrect - %v vs %v", respObj.TxID, txID)
}

/*
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

	txID := "8b751bc182766e6187d39b1eca538d9ece0b8ff662e408cd4e45f89359f8c7e7"
	if respObj.TxID != txID {
		t.Errorf("Error: TxID returned during Commit Entry is incorrect - %v vs %v", respObj.TxID, txID)
	}
}
*/
/*
func TestV2HandleEntryCreditBalance(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	eckey := testHelper.NewECAddressPublicKeyString(0)
	req := new(AddressRequest)
	req.Address = eckey

	resp, err := HandleV2EntryCreditBalance(state, req)
	if err != nil {
		t.Errorf("%v", err)
	}

	var expectedAmount int64 = 400

	if resp.(*EntryCreditBalanceResponse).Balance != expectedAmount {
		t.Errorf("Invalid balance returned - %v vs %v", resp.(*EntryCreditBalanceResponse).Balance, expectedAmount)
	}

	eckey = testHelper.NewECAddressString(0)
	req = new(AddressRequest)
	req.Address = eckey

	resp, err = HandleV2EntryCreditBalance(state, req)
	if err != nil {
		t.Errorf("%v", err)
	}

	if resp.(*EntryCreditBalanceResponse).Balance != expectedAmount {
		t.Errorf("Invalid balance returned - %v vs %v", resp.(*EntryCreditBalanceResponse).Balance, expectedAmount)
	}
}
*/
/*
func TestV2HandleFactoidBalance(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	eckey := testHelper.NewFactoidRCDAddressString(0)
	req := new(AddressRequest)
	req.Address = eckey

	resp, err := HandleV2FactoidBalance(state, req)
	if err != nil {
		t.Errorf("%v", err)
	}

	var expectedAmount int64 = 199977800

	if resp.(*FactoidBalanceResponse).Balance != expectedAmount {
		t.Errorf("Invalid balance returned - %v vs %v", resp.(*FactoidBalanceResponse).Balance, expectedAmount)
	}
}
*/

func TestHandleV2GetReceipt(t *testing.T) {
	state := testHelper.CreateAndPopulateTestStateAndStartValidator()

	hashkey := new(HashRequest)
	hashkey.Hash = "be5fb8c3ba92c0436269fab394ff7277c67e9b2de4431b723ce5d89799c0b93a"

	resp, jErr := HandleV2Receipt(state, hashkey)
	assert.Nil(t, jErr)

	dbo := state.GetDB()

	marshalled, err := json.Marshal(resp.(*ReceiptResponse).Receipt)
	assert.Nil(t, err)

	t.Logf("Resp - %s", marshalled)

	err = receipts.VerifyFullReceipt(dbo, string(marshalled))
	assert.Nil(t, err, "receipt - %s", marshalled)
}

func TestHandleV2GetTranasction(t *testing.T) {
	state := testHelper.CreateAndPopulateTestStateAndStartValidator()
	blocks := testHelper.CreateFullTestBlockSet()

	for _, block := range blocks {
		for _, tx := range block.FBlock.GetTransactions() {
			hashkey := new(HashRequest)
			hashkey.Hash = tx.GetFullHash().String()

			resp, jErr := HandleV2GetTranasction(state, hashkey)
			assert.Nil(t, jErr)

			r := resp.(*TransactionResponse)
			assert.Nil(t, r.ECTranasction)
			assert.Nil(t, r.Entry)
			assert.Equal(t, hashkey.Hash, r.FactoidTransaction.GetFullHash().String(), "Got wrong hash for FactoidTransaction")
			assert.Equal(t, block.FBlock.DatabasePrimaryIndex().String(), r.IncludedInTransactionBlock, "Invalid IncludedInTransactionBlock")
			assert.Equal(t, block.DBlock.DatabasePrimaryIndex().String(), r.IncludedInDirectoryBlock, "Invalid IncludedInDirectoryBlock")
			assert.Equal(t, int64(block.DBlock.GetDatabaseHeight()), r.IncludedInDirectoryBlockHeight, "Invalid IncludedInDirectoryBlockHeight")
		}

		for _, h := range block.ECBlock.GetEntryHashes() {
			if h.IsMinuteMarker() == true {
				continue
			}
			hashkey := new(HashRequest)
			hashkey.Hash = h.String()

			resp, jErr := HandleV2GetTranasction(state, hashkey)
			assert.Nil(t, jErr)

			r := resp.(*TransactionResponse)
			assert.Nil(t, r.FactoidTransaction)
			assert.Nil(t, r.Entry)
			assert.Equal(t, hashkey.Hash, r.ECTranasction.Hash().String(), "Got wrong hash for ECTranasction")
			assert.Equal(t, block.ECBlock.DatabasePrimaryIndex().String(), r.IncludedInTransactionBlock, "Invalid IncludedInTransactionBlock")
			assert.Equal(t, block.DBlock.DatabasePrimaryIndex().String(), r.IncludedInDirectoryBlock, "Invalid IncludedInDirectoryBlock")
			assert.Equal(t, int64(block.DBlock.GetDatabaseHeight()), r.IncludedInDirectoryBlockHeight, "Invalid IncludedInDirectoryBlockHeight")
		}

		for _, tx := range block.EBlock.GetEntryHashes() {
			if tx.IsMinuteMarker() == true {
				continue
			}
			hashkey := new(HashRequest)
			hashkey.Hash = tx.String()

			resp, jErr := HandleV2GetTranasction(state, hashkey)
			assert.Nil(t, jErr)

			r := resp.(*TransactionResponse)
			assert.Nil(t, r.ECTranasction)
			assert.Nil(t, r.FactoidTransaction)
			assert.Equal(t, hashkey.Hash, r.Entry.GetHash().String(), "Got wrong hash for Entry")
			assert.Equal(t, block.EBlock.DatabasePrimaryIndex().String(), r.IncludedInTransactionBlock, "Invalid IncludedInTransactionBlock")
			assert.Equal(t, block.DBlock.DatabasePrimaryIndex().String(), r.IncludedInDirectoryBlock, "Invalid IncludedInDirectoryBlock")
			assert.Equal(t, int64(block.DBlock.GetDatabaseHeight()), r.IncludedInDirectoryBlockHeight, "Invalid IncludedInDirectoryBlockHeight")
		}

		for _, tx := range block.AnchorEBlock.GetEntryHashes() {
			if tx.IsMinuteMarker() == true {
				continue
			}
			hashkey := new(HashRequest)
			hashkey.Hash = tx.String()

			resp, jErr := HandleV2GetTranasction(state, hashkey)
			assert.Nil(t, jErr)

			r := resp.(*TransactionResponse)
			assert.Nil(t, r.ECTranasction)
			assert.Nil(t, r.FactoidTransaction)
			assert.Equal(t, hashkey.Hash, r.Entry.GetHash().String(), "Got wrong hash for Entry")
			assert.Equal(t, block.AnchorEBlock.DatabasePrimaryIndex().String(), r.IncludedInTransactionBlock, "Invalid IncludedInTransactionBlock")
			assert.Equal(t, block.DBlock.DatabasePrimaryIndex().String(), r.IncludedInDirectoryBlock, "Invalid IncludedInDirectoryBlock")
			assert.Equal(t, int64(block.DBlock.GetDatabaseHeight()), r.IncludedInDirectoryBlockHeight, "Invalid IncludedInDirectoryBlockHeight")
		}
	}
}

func TestJSONString(t *testing.T) {
	eblock := new(EBlock)
	eblock.Header.BlockSequenceNumber = 5
	eblock.Header.ChainID = "Findthis"

	s, err := eblock.JSONString()
	assert.Nil(t, err)
	assert.True(t, strings.Contains(s, "Findthis"), "Missing chainID")

	e := new(EntryStruct)
	e.ChainID = "Findthis"

	s, err = e.JSONString()
	assert.Nil(t, err)
	assert.True(t, strings.Contains(s, "Findthis"), "Missing chainID")

	c := new(CHead)
	c.ChainHead = "Findthis"

	s, err = e.JSONString()
	assert.Nil(t, err)
	assert.True(t, strings.Contains(s, "Findthis"), "Missing ChainHead")

	d := new(DBlock)
	d.Header.PrevBlockKeyMR = "Findthis"
	s, err = e.JSONString()
	assert.Nil(t, err)
	assert.True(t, strings.Contains(s, "Findthis"), "Missing PrevBlockKeyMR")
}

func Test_ecBlockToResp(t *testing.T) {
	type args struct {
		block interfaces.IEntryCreditBlock
	}
	tests := []struct {
		name  string
		args  args
		want  interface{}
		want1 *primitives.JSONError
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, got1 := ECBlockToResp(tt.args.block)
			assert.True(t, reflect.DeepEqual(got, tt.want), "ecBlockToResp() got = %v, want %v", got, tt.want)
			assert.True(t, reflect.DeepEqual(got1, tt.want1), "ecBlockToResp() got1 = %v, want %v", got1, tt.want1)
		})
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
