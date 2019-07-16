package wsapi_test

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
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

func TestHandleV2Requests(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	Start(state)

	cases := map[string]struct {
		Method     string
		Message    interface{}
		StatusCode int
		Expected   map[string]interface{}
		Error      *primitives.JSONError
	}{
		"commit-chain": {
			"commit-chain",
			MessageRequest{Message: "00015507b2f70bd0165d9fa19a28cfaafb6bc82f538955a98c7b7e60d79fbf92655c1bff1c76466cb3bc3f3cc68d8b2c111f4f24c88d9c031b4124395c940e5e2c5ea496e8aaa2f5c956749fc3eba4acc60fd485fb100e601070a44fcce54ff358d606698547340b3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da2946c901273e616bdbb166c535b26d0d446bc69b22c887c534297c7d01b2ac120237086112b5ef34fc6474e5e941d60aa054b465d4d770d7f850169170ef39150b"},
			http.StatusOK,
			map[string]interface{}{
				"message":     "Chain Commit Success",
				"chainidhash": "d0165d9fa19a28cfaafb6bc82f538955a98c7b7e60d79fbf92655c1bff1c7646",
				"entryhash":   "f5c956749fc3eba4acc60fd485fb100e601070a44fcce54ff358d60669854734",
				"txid":        "76e123d133a841fe3e08c5e3f3d392f8431f2d7668890c03f003f541efa8fc61",
			},
			nil,
		},
		"commit-entry": {
			"commit-entry",
			MessageRequest{Message: "00015507C1024BF5C956749FC3EBA4ACC60FD485FB100E601070A44FCCE54FF358D60669854734013B6A27BCCEB6A42D62A3A8D02A6F0D73653215771DE243A63AC048A18B59DA29F4CBD953E6EBE684D693FDCA270CE231783E8ECC62D630F983CD59E559C6253F84D1F54C8E8D8665D493F7B4A4C1864751E3CDEC885A64C2144E0938BF648A00"},
			http.StatusOK,
			map[string]interface{}{
				"message":   "Entry Commit Success",
				"txid":      "8b751bc182766e6187d39b1eca538d9ece0b8ff662e408cd4e45f89359f8c7e7",
				"entryhash": "f5c956749fc3eba4acc60fd485fb100e601070a44fcce54ff358d60669854734",
			},
			nil,
		},
		"commit-entry-invalid": {
			"commit-entry",
			EntryRequest{Entry: "00015507b2f70bd0165d9fa19a28cfaafb6bc82f538955a98c7b7e60d79fbf92655c1bff1c76466cb3bc3f3cc68d8b2c111f4f24c88d9c031b4124395c940e5e2c5ea496e8aaa2f5c956749fc3eba4acc60fd485fb100e601070a44fcce54ff358d606698547340b3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da2946c901273e616bdbb166c535b26d0d446bc69b22c887c534297c7d01b2ac120237086112b5ef34fc6474e5e941d60aa054b465d4d770d7f850169170ef39150b"},
			http.StatusOK,
			map[string]interface{}{},
			&primitives.JSONError{
				Code:    -32602,
				Message: "Invalid params",
				Data:    "Invalid Commit Entry",
			},
		},
		"factoid-balance": {
			"factoid-balance",
			AddressRequest{Address: testHelper.NewFactoidRCDAddressString(0)},
			http.StatusOK,
			map[string]interface{}{
				"balance": number("199977800"),
			},
			nil,
		},
		"factoid-balance-nil": {
			"factoid-balance",
			nil,
			http.StatusOK,
			map[string]interface{}{},
			&primitives.JSONError{Code: -32602, Message: "Invalid params", Data: "Invalid Address"},
		},
		"entry-credit-balance": {
			"entry-credit-balance",
			AddressRequest{Address: testHelper.NewECAddressPublicKeyString(0)},
			http.StatusOK,
			map[string]interface{}{
				"balance": number("196"),
			},
			nil,
		},
		"entry-credit-rate": {
			"entry-credit-rate",
			nil,
			http.StatusOK,
			map[string]interface{}{
				"rate": number("1"),
			},
			nil,
		},
		"entry-credit-rate-params": {
			"entry-credit-rate",
			ReceiptRequest{EntryHash: "be5fb8c3ba92c0436269fab394ff7277c67e9b2de4431b723ce5d89799c0b93a"},
			http.StatusOK,
			map[string]interface{}{
				"rate": number("1"),
			},
			nil,
		},
		"receipt": {
			"receipt",
			ReceiptRequest{EntryHash: "be5fb8c3ba92c0436269fab394ff7277c67e9b2de4431b723ce5d89799c0b93a"},
			http.StatusOK,
			map[string]interface{}{
				"receipt": map[string]interface{}{
					"directoryblockheight": number("8"),
					"directoryblockkeymr":  "0a30d74682b42cc4c3e6c6f8fcc8a353b39c57475cb89f9cdfec7f07cc195b67",
					"entry": map[string]interface{}{
						"entryhash": "be5fb8c3ba92c0436269fab394ff7277c67e9b2de4431b723ce5d89799c0b93a",
						"timestamp": number("75060"),
					},
					"entryblockkeymr": "bcdbea5044d0ab6ca1f961a0075db69248a6d3d19abad612eb79e1ed1533486b",
					"merklebranch": []interface{}{
						map[string]interface{}{
							"left":  "be5fb8c3ba92c0436269fab394ff7277c67e9b2de4431b723ce5d89799c0b93a",
							"right": "0000000000000000000000000000000000000000000000000000000000000009",
							"top":   "87e5cb6bf6aad44b83123f0b79c47a4cf44ee592a05ae30d54dfa54f69754401",
						},
						map[string]interface{}{
							"left":  "53b5bc6f2abc3e7530bae3b5d799f916a1767e041b3240346624c2047be3153b",
							"right": "87e5cb6bf6aad44b83123f0b79c47a4cf44ee592a05ae30d54dfa54f69754401",
							"top":   "bcdbea5044d0ab6ca1f961a0075db69248a6d3d19abad612eb79e1ed1533486b",
						},
						map[string]interface{}{
							"left":  "6e7e64ac45ff57edbf8537a0c99fba2e9ee351ef3d3f4abd93af9f01107e592c",
							"right": "bcdbea5044d0ab6ca1f961a0075db69248a6d3d19abad612eb79e1ed1533486b",
							"top":   "ae2b49bd6c561a3000c9688dc2a1518905520b719defd89e4e53911c6b396231",
						},
						map[string]interface{}{
							"left":  "9f1c92629ac43a9f162dbd0a34ab166b35894178bbedcdcf93ed304619ef509d",
							"right": "ae2b49bd6c561a3000c9688dc2a1518905520b719defd89e4e53911c6b396231",
							"top":   "d3c0718340eab8e5ffbda23adefd6ffc044b4cd5b47bd271ddb7695dab306706",
						},
						map[string]interface{}{
							"left":  "dacc48acfa3421f18ef812af94a02269ad9478c0c8593e7bfdad856a3b71d799",
							"right": "d3c0718340eab8e5ffbda23adefd6ffc044b4cd5b47bd271ddb7695dab306706",
							"top":   "349eb6210c01471dc1772ae334446771415c684faaa01db6a6c03650a2860d7a",
						},
						map[string]interface{}{
							"left":  "349eb6210c01471dc1772ae334446771415c684faaa01db6a6c03650a2860d7a",
							"right": "af7c351e6d8ab20ea21479624f05999b59a187da91fbdd852092292dafd327a1",
							"top":   "5fd3cfa60fc76eb9180528bfc543b6343438c6d4997dd194b72582cef2a5e3b3",
						},
						map[string]interface{}{
							"left":  "b20babddd9be6711c4d707bb7eb16bb1d2042832a1cbf49465903936ee0dff59",
							"right": "5fd3cfa60fc76eb9180528bfc543b6343438c6d4997dd194b72582cef2a5e3b3",
							"top":   "0a30d74682b42cc4c3e6c6f8fcc8a353b39c57475cb89f9cdfec7f07cc195b67",
						},
					},
				},
			},
			nil,
		},
		"entry-block-invalid": {
			"entry-block",
			KeyMRRequest{KeyMR: "f5c956749fc3eba4acc60fd485fb100e601070a44fcce54ff358d60669854734"},
			http.StatusOK,
			map[string]interface{}{},
			&primitives.JSONError{Code: -32008, Message: "Block not found", Data: nil},
		},
		"entry-block": {
			"entry-block",
			KeyMRRequest{KeyMR: "bcdbea5044d0ab6ca1f961a0075db69248a6d3d19abad612eb79e1ed1533486b"},
			http.StatusOK,
			map[string]interface{}{
				"entrylist": []interface{}{
					map[string]interface{}{
						"entryhash": "be5fb8c3ba92c0436269fab394ff7277c67e9b2de4431b723ce5d89799c0b93a",
						"timestamp": number("75060"),
					},
				},
				"header": map[string]interface{}{
					"blocksequencenumber": number("0"),
					"chainid":             "6e7e64ac45ff57edbf8537a0c99fba2e9ee351ef3d3f4abd93af9f01107e592c",
					"dbheight":            number("8"),
					"prevkeymr":           "eb5da48561fbc26def7b6f17ac25af2c297b5b7ace0a57c3e278c05c52a42e3f",
					"timestamp":           number("74520"),
				},
			},
			nil,
		},
		"chain-head": {
			"chain-head",
			ChainIDRequest{ChainID: "6e7e64ac45ff57edbf8537a0c99fba2e9ee351ef3d3f4abd93af9f01107e592c"},
			http.StatusOK,
			map[string]interface{}{
				"chainhead":          "e79fb46ad81f0b4fac7f1e66728b40b390f8fcc3806e93f94550eec041eecff2",
				"chaininprocesslist": false,
			},
			nil,
		},
		"current-minute": {
			"current-minute",
			nil,
			http.StatusOK,
			map[string]interface{}{
				"directoryblockheight":    number("0"),
				"directoryblockinseconds": number("20"),
				"leaderheight":            number("0"),
			},
			nil,
		},
		"diagnostics": {
			"diagnostics",
			nil,
			http.StatusOK,
			map[string]interface{}{
				"balancehash":          "0000f65cd6d3ac9f2f94bd664d0e2a509ae4967b41254ab3ef327841cf759486",
				"lastblockfromdbstate": false,
				"role":                 "Follower",
			},
			nil,
		},
	}

	cases = map[string]struct {
		Method     string
		Message    interface{}
		StatusCode int
		Expected   map[string]interface{}
		Error      *primitives.JSONError
	}{
		"entry-credit-rate-invalid": {
			"entry-credit-rate",
			ReceiptRequest{EntryHash: "be5fb8c3ba92c0436269fab394ff7277c67e9b2de4431b723ce5d89799c0b93a"},
			http.StatusOK,
			map[string]interface{}{},
			&primitives.JSONError{Code: -32600, Message: "Invalid Request"},
		},
	}

	for name, testCase := range cases {
		t.Logf("test case '%s'", name)

		request := primitives.NewJSON2Request(testCase.Method, 0, testCase.Message)
		response, err := v2Request(request)

		assert.Nil(t, err, "test '%s' failed: %v \nresponse: %v", name, err, response)
		assert.Equal(t, testCase.Error, response.Error, "test '%s' failed contains different error that expected: %v", name, response.Error)

		if response.Result != nil {
			result := response.Result.(map[string]interface{})

			for k, v := range testCase.Expected {
				assert.Equal(t, v, result[k], "test case '%s' assertion failed %s: %v != %v\nresponse:%v", name, k, v, result[k], response)
			}
		}
	}
}

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

func TestHandleV2GetTransaction(t *testing.T) {
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

	resp, err := http.Post("http://localhost:8088/v2", "application/json", bytes.NewBuffer(j))
	if err != nil {
		return nil, err
	}

	r := primitives.NewJSON2Response()
	d := json.NewDecoder(resp.Body)
	d.UseNumber()
	if err := d.Decode(r); err != nil {
		return nil, err
	}

	return r, nil
}

func number(n string) json.Number {
	return json.Number(n)
}
