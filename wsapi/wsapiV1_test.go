package wsapi_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/receipts"
	"github.com/stretchr/testify/assert"

	"io/ioutil"
	"net/http"

	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/testHelper"
	. "github.com/FactomProject/factomd/wsapi"
)

func TestHandleGetRaw(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	Start(state)

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

	for i, v := range toTest {
		url := fmt.Sprintf("/v1/get-raw-data/%s", v.Hash1)

		rawDataResponse := new(RawDataResponse)
		v1RequestGet(t, url, false, rawDataResponse)

		assert.True(t, strings.Contains(rawDataResponse.Data, v.Raw), "Looking for %v \nGetRaw %v/%v from Hash1 failed - %v", v.Hash1, i, len(toTest), rawDataResponse.Data)

		url = fmt.Sprintf("/v1/get-raw-data/%s", v.Hash2)
		v1RequestGet(t, url, false, rawDataResponse)

		assert.True(t, strings.Contains(rawDataResponse.Data, v.Raw), "Looking for %v \nGetRaw %v/%v from Hash2 failed - %v", v.Hash2, i, len(toTest), rawDataResponse.Data)
	}
}

func TestHandleDirectoryBlock(t *testing.T) {
	//initializing server
	state := testHelper.CreateAndPopulateTestState()
	Start(state)

	hash := testHelper.DBlockHeadPrimaryIndex
	url := fmt.Sprintf("/v1/directory-block-by-keymr/%s", hash)

	dBlock := new(DBlock)
	v1RequestGet(t, url, false, dBlock)

	result, err := dBlock.JSONString()

	assert.Nil(t, err, "HandleDirectoryBlock json not serializable")
	assert.True(t, strings.Contains(result, "000000000000000000000000000000000000000000000000000000000000000a"), "%v", result)
	assert.True(t, strings.Contains(result, testHelper.ABlockHeadPrimaryIndex), "%v", result)
	assert.True(t, strings.Contains(result, "000000000000000000000000000000000000000000000000000000000000000c"), "%v", result)
	assert.True(t, strings.Contains(result, testHelper.ECBlockHeadPrimaryIndex), "%v", result)
	assert.True(t, strings.Contains(result, "000000000000000000000000000000000000000000000000000000000000000f"), "%v", result)
	assert.True(t, strings.Contains(result, testHelper.FBlockHeadPrimaryIndex), "%v", result)
	assert.True(t, strings.Contains(result, "6e7e64ac45ff57edbf8537a0c99fba2e9ee351ef3d3f4abd93af9f01107e592c"), "%v", result)
	assert.True(t, strings.Contains(result, testHelper.EBlockHeadPrimaryIndex), "%v", result)
	assert.True(t, strings.Contains(result, "df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604"), "%v", result)
	assert.True(t, strings.Contains(result, testHelper.AnchorBlockHeadPrimaryIndex), "%v", result)
	assert.True(t, strings.Contains(result, "\"timestamp\":74580"), "%v", result)
}

func TestHandleEntryBlock(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	Start(state)

	chain, err := primitives.HexToHash("df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604")
	assert.Nil(t, err)

	dbo := state.GetDB()

	blocks, err := dbo.FetchAllEBlocksByChain(chain)
	assert.Nil(t, err)

	fetched := 0
	for _, b := range blocks {
		hash := b.(*entryBlock.EBlock).DatabasePrimaryIndex().String()
		hash2 := b.(*entryBlock.EBlock).DatabaseSecondaryIndex().String()

		eBlock := new(EBlock)
		url := fmt.Sprintf("/v1/entry-block-by-keymr/%s", hash)
		v1RequestGet(t, url, false, eBlock)

		assert.Equal(t, "df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604", eBlock.Header.ChainID, "Wrong ChainID in eBlock - %v", eBlock)
		assert.Equal(t, int64(b.(*entryBlock.EBlock).GetHeader().GetDBHeight()), eBlock.Header.DBHeight, "DBHeight is wrong - %v vs %v", eBlock.Header.DBHeight, b.(*entryBlock.EBlock).GetHeader().GetDBHeight())

		url = fmt.Sprintf("/v1/entry-block-by-keymr/%s", hash2)
		v1RequestGet(t, url, false, eBlock)

		assert.Equal(t, "df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604", eBlock.Header.ChainID, "Wrong ChainID in eBlock - %v", eBlock)
		assert.Equal(t, int64(b.(*entryBlock.EBlock).GetHeader().GetDBHeight()), eBlock.Header.DBHeight, "DBHeight is wrong - %v vs %v", eBlock.Header.DBHeight, b.(*entryBlock.EBlock).GetHeader().GetDBHeight())

		fetched++
	}
	assert.Equal(t, testHelper.BlockCount, fetched, "Fetched %v blocks, expected %v", fetched, testHelper.BlockCount)
}

func TestHandleEntryBlockInvalidHash(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	Start(state)

	url := "/v1/entry-block-by-keymr/invalid-hash"

	v1RequestGet(t, url, true, nil)
}

func TestHandleGetFee(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	Start(state)

	url := "/v1/factoid-get-fee/"

	type x struct{ Fee int64 }
	fee := new(x)
	v1RequestGet(t, url, false, fee)

	if fee.Fee < 1 {
		t.Errorf("%v", fee)
	}
}

func TestDBlockList(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	Start(state)

	list := []string{
		"508e19f65a7fc7e9cfa5a73281b5e08115ed25a1af5723350e5c21fc92c39b40", //9
		"aeffa5d5c02498d958b88fab12672054c2729da46621b381793995ad9c47e4d3", //8
		"cd63b26d12e9d397a545fd50e26b53ab8b1fb555f824edb1f71937a6288d5901", //7
		"15f625a8b73f1d3d226ae957728537f084ba8f8d0b0867178a24efb5dc1bdd49", //6
		"c4effa44e5b42d8c4ea78866b9ac99e603d13615780c0f31346fa775fd5cc5f6", //5
		"4f4bbe848b8998f73a4eb940791302cf81733f7f9827865c846fee4f6edd98e2", //4
		"5038b4f268fdc2e779553e70ac6a03a784c2958dbc2489affef52dafbdb073c7", //3
		"f9fac92c710620e1e3dbfeadbc040ad0f2e5cbdd110c65455e168a09c922998f", //2
		"3d451d1aace4dcbaa111106041d956ad3e6973aed945ec8cda5015fa356cf88c", //1
		"dcc95bfa721ebb11297ecd390a5b1c21632b40c00e84ac0729b393b2de7633a7", //0
	}

	dBlock := new(DBlock)
	for _, l := range list {
		url := fmt.Sprintf("/v1/directory-block-by-keymr/%s", l)
		// expect a NewBlockNotFoundError
		v1RequestGet(t, url, true, dBlock)
	}

	hash := "000000000000000000000000000000000000000000000000000000000000000d"

	head := new(CHead)
	url := fmt.Sprintf("/v1/chain-head/%s", hash)
	v1RequestGet(t, url, false, head)

	url = fmt.Sprintf("/v1/directory-block-by-keymr/%s", head.ChainHead)
	v1RequestGet(t, url, false, dBlock)
}

func TestBlockIteration(t *testing.T) {
	//initializing server
	state := testHelper.CreateAndPopulateTestState()
	Start(state)

	hash := "000000000000000000000000000000000000000000000000000000000000000d"

	head := new(CHead)
	url := fmt.Sprintf("/v1/chain-head/%s", hash)
	v1RequestGet(t, url, false, head)

	prev := head.ChainHead
	fetched := 0
	for {
		if prev == "0000000000000000000000000000000000000000000000000000000000000000" || prev == "" {
			break
		}

		block := new(DBlock)
		url := fmt.Sprintf("/v1/directory-block-by-keymr/%s", prev)
		v1RequestGet(t, url, false, block)

		prev = block.Header.PrevBlockKeyMR
		fetched++
	}
	if fetched != testHelper.BlockCount {
		t.Errorf("DBlock only found %v blocks, was expecting %v", fetched, testHelper.BlockCount)
	}
}

func TestHandleGetReceipt(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	Start(state)

	hash := "be5fb8c3ba92c0436269fab394ff7277c67e9b2de4431b723ce5d89799c0b93a"

	receipt := new(ReceiptResponse)
	url := fmt.Sprintf("/v1/get-receipt/%s", hash)
	v1RequestGet(t, url, false, receipt)

	assert.NotNil(t, receipt.Receipt, "Receipt not found!")

	err := receipt.Receipt.Validate()
	assert.Nil(t, err, "failed to validate receipt - %v", receipt)

	dbo := state.GetDB()
	receiptStr, _ := json.Marshal(receipt.Receipt)
	err = receipts.VerifyFullReceipt(dbo, string(receiptStr))
	assert.Nil(t, err, "receipt - %v", receipt)
}

func TestHandleGetUnanchoredReceipt(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	Start(state)

	hash := "68a503bd3d5b87d3a41a737e430d2ce78f5e556f6a9269859eeb1e053b7f92f7"

	receipt := new(ReceiptResponse)
	url := fmt.Sprintf("/v1/get-receipt/%s", hash)
	v1RequestGet(t, url, false, receipt)

	err := receipt.Receipt.Validate()
	assert.Nil(t, err, "failed to validate receipt - %v", receipt)

	dbo := state.GetDB()
	receiptStr, _ := json.Marshal(receipt.Receipt)
	err = receipts.VerifyFullReceipt(dbo, string(receiptStr))
	assert.Nil(t, err, "receipt - %v", receipt)
}

func TestHandleFactoidBalanceUnknownAddress(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	Start(state)

	factoidBalanceResponse := new(FactoidBalanceResponse)
	url := "/v1/factoid-balance/f1ba8879fcf63b596b60ccc4c69c7f6848475ac037fc63b080ba2d9502fe66a4"

	v1RequestGet(t, url, false, factoidBalanceResponse)

	assert.Equal(t, int64(0), factoidBalanceResponse.Balance, "%v", factoidBalanceResponse)
}

func TestHandleProperties(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	Start(state)

	type V1Properties struct {
		Protocol_Version string
		Factomd_Version  string
	}

	properties := new(V1Properties)
	url := "/v1/properties/"
	v1RequestGet(t, url, false, properties)

	assert.Equal(t, properties.Factomd_Version, state.GetFactomdVersion())
	assert.Equal(t, "0.0.0.0", properties.Protocol_Version)
}

func TestHandleHeights(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	Start(state)

	heightsResponse := new(HeightsResponse)
	url := "/v1/heights/"
	v1RequestGet(t, url, false, heightsResponse)

	assert.Equal(t, int64(state.GetTrueLeaderHeight()), heightsResponse.LeaderHeight)
	assert.Equal(t, int64(state.GetHighestSavedBlk()), heightsResponse.DirectoryBlockHeight)
	assert.Equal(t, int64(state.GetHighestSavedBlk()), heightsResponse.EntryBlockHeight)
	assert.Equal(t, int64(state.GetEntryDBHeightComplete()), heightsResponse.EntryHeight)
	assert.Equal(t, int64(state.GetMissingEntryCount()), heightsResponse.MissingEntryCount)
	assert.Equal(t, int64(state.GetEntryBlockDBHeightProcessing()), heightsResponse.EntryBlockDBHeightProcessing)
	assert.Equal(t, int64(state.GetEntryBlockDBHeightComplete()), heightsResponse.EntryBlockDBHeightComplete)
}

func v1RequestGet(t *testing.T, url string, expectederror bool, result interface{}) {
	response, err := http.Get(fmt.Sprintf("http://localhost:8088/%s", url))
	assert.Nil(t, err, "response: %v", response)
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	assert.Nil(t, err, "body: \t%v", string(body))

	if expectederror {
		result = new(primitives.JSONError)
	}

	if len(body) != 0 {
		err := json.Unmarshal(body, result)
		assert.Nil(t, err)
	}
}
