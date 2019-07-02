// +build all

package wsapi_test

import (
	"encoding/json"
	//"fmt"
	"strings"
	"testing"

	"fmt"

	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/receipts"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
	. "github.com/FactomProject/factomd/wsapi"
)

/*
func TestHandleDirectoryBlockHead(t *testing.T) {
	context := testHelper.CreateWebContext()

	HandleDirectoryBlockHead(context)
	expectedHead := testHelper.DBlockHeadPrimaryIndex

	if strings.Contains(testHelper.GetBody(context), expectedHead) == false {
		t.Errorf("Context does not contain proper DBlock Head - %v vs %v", testHelper.GetBody(context), expectedHead)
	}
}
*/

func TestHandleGetRaw(t *testing.T) {
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

	context := testHelper.CreateWebContext()
	for i, v := range toTest {
		testHelper.ClearContextResponseWriter(context)
		HandleGetRaw(context, v.Hash1)

		if strings.Contains(testHelper.GetBody(context), v.Raw) == false {
			t.Errorf("Looking for %v", v.Hash1)
			t.Errorf("GetRaw %v/%v from Hash1 failed - %v", i, len(toTest), testHelper.GetBody(context))
		}

		testHelper.ClearContextResponseWriter(context)
		HandleGetRaw(context, v.Hash2)

		if strings.Contains(testHelper.GetBody(context), v.Raw) == false {
			t.Errorf("Looking for %v", v.Hash2)
			t.Errorf("GetRaw %v/%v from Hash2 failed - %v", i, len(toTest), testHelper.GetBody(context))
		}
	}
}

func TestHandleDirectoryBlock(t *testing.T) {
	context := testHelper.CreateWebContext()
	hash := testHelper.DBlockHeadPrimaryIndex

	HandleDirectoryBlock(context, hash)

	if testHelper.GetBody(context) == "" {
		t.Errorf("HandleDirectoryBlock returned empty block")
		t.FailNow()
	}

	if strings.Contains(testHelper.GetBody(context), "000000000000000000000000000000000000000000000000000000000000000a") == false {
		t.Errorf("%v", testHelper.GetBody(context))
	}

	if strings.Contains(testHelper.GetBody(context), testHelper.ABlockHeadPrimaryIndex) == false {
		t.Errorf("%v", testHelper.GetBody(context))
	}

	if strings.Contains(testHelper.GetBody(context), "000000000000000000000000000000000000000000000000000000000000000c") == false {
		t.Errorf("%v", testHelper.GetBody(context))
	}

	if strings.Contains(testHelper.GetBody(context), testHelper.ECBlockHeadPrimaryIndex) == false {
		t.Errorf("%v", testHelper.GetBody(context))
	}

	if strings.Contains(testHelper.GetBody(context), "000000000000000000000000000000000000000000000000000000000000000f") == false {
		t.Errorf("%v", testHelper.GetBody(context))
	}

	if strings.Contains(testHelper.GetBody(context), testHelper.FBlockHeadPrimaryIndex) == false {
		t.Errorf("%v", testHelper.GetBody(context))
	}

	if strings.Contains(testHelper.GetBody(context), "6e7e64ac45ff57edbf8537a0c99fba2e9ee351ef3d3f4abd93af9f01107e592c") == false {
		t.Errorf("%v", testHelper.GetBody(context))
	}

	if strings.Contains(testHelper.GetBody(context), testHelper.EBlockHeadPrimaryIndex) == false {
		t.Errorf("%v", testHelper.GetBody(context))
	}

	if strings.Contains(testHelper.GetBody(context), "df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604") == false {
		t.Errorf("%v", testHelper.GetBody(context))
	}

	if strings.Contains(testHelper.GetBody(context), testHelper.AnchorBlockHeadPrimaryIndex) == false {
		t.Errorf("%v", testHelper.GetBody(context))
	}

	if strings.Contains(testHelper.GetBody(context), "\"Timestamp\":74580") == false {
		t.Errorf("%v", testHelper.GetBody(context))
	}

}

func TestHandleEntryBlock(t *testing.T) {
	context := testHelper.CreateWebContext()
	chain, err := primitives.HexToHash("df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604")
	if err != nil {
		t.Error(err)
	}

	dbo := context.Server.Env["state"].(interfaces.IState).GetDB()
	//defer context.Server.Env["state"].(interfaces.IState).UnlockDB()

	blocks, err := dbo.FetchAllEBlocksByChain(chain)
	if err != nil {
		t.Error(err)
	}
	fetched := 0
	for _, b := range blocks {
		hash := b.(*entryBlock.EBlock).DatabasePrimaryIndex().String()
		hash2 := b.(*entryBlock.EBlock).DatabaseSecondaryIndex().String()

		testHelper.ClearContextResponseWriter(context)
		HandleEntryBlock(context, hash)

		eBlock := new(EBlock)

		testHelper.UnmarshalRespDirectly(context, eBlock)

		if eBlock.Header.ChainID != "df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604" {
			t.Errorf("Wrong ChainID - %v", eBlock.Header.ChainID)
			t.Errorf("eBlock - %v", eBlock)
			t.Errorf("%v", testHelper.GetBody(context))
		}

		if eBlock.Header.DBHeight != int64(b.(*entryBlock.EBlock).GetHeader().GetDBHeight()) {
			t.Errorf("DBHeight is wrong - %v vs %v", eBlock.Header.DBHeight, b.(*entryBlock.EBlock).GetHeader().GetDBHeight())
		}

		testHelper.ClearContextResponseWriter(context)
		HandleEntryBlock(context, hash2)

		eBlock = new(EBlock)

		testHelper.UnmarshalRespDirectly(context, eBlock)

		if eBlock.Header.ChainID != "df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604" {
			t.Errorf("Wrong ChainID - %v", eBlock.Header.ChainID)
			t.Errorf("%v", testHelper.GetBody(context))
		}

		if eBlock.Header.DBHeight != int64(b.(*entryBlock.EBlock).GetHeader().GetDBHeight()) {
			t.Errorf("DBHeight is wrong - %v vs %v", eBlock.Header.DBHeight, b.(*entryBlock.EBlock).GetHeader().GetDBHeight())
		}

		fetched++
	}
	if fetched != testHelper.BlockCount {
		t.Errorf("Fetched %v blocks, expected %v", fetched, testHelper.BlockCount)
	}
}

/*
func TestHandleEntry(t *testing.T) {
	context := testHelper.CreateWebContext()
	hash := ""

	HandleEntry(context, hash)

	if strings.Contains(testHelper.GetBody(context), "") == false {
		t.Errorf("%v", testHelper.GetBody(context))
	}
}
*/
func TestHandleChainHead(t *testing.T) {
	context := testHelper.CreateWebContext()
	hash := "000000000000000000000000000000000000000000000000000000000000000d"

	HandleChainHead(context, hash)

	if strings.Contains(testHelper.GetBody(context), testHelper.DBlockHeadPrimaryIndex) == false {
		t.Errorf("Invalid directory block head: %v", testHelper.GetBody(context))
	}

	hash = "000000000000000000000000000000000000000000000000000000000000000a"

	testHelper.ClearContextResponseWriter(context)
	HandleChainHead(context, hash)

	s := context.Server.Env["state"]
	st := s.(*state.State)
	a, _ := st.DB.FetchABlockByHeight(0)
	fmt.Println(a)
	fmt.Println(string(testHelper.GetBody(context)))

	if strings.Contains(testHelper.GetBody(context), testHelper.ABlockHeadPrimaryIndex) == false {
		t.Errorf("Invalid admin block head: %v", testHelper.GetBody(context))
	}

	hash = "000000000000000000000000000000000000000000000000000000000000000c"

	testHelper.ClearContextResponseWriter(context)
	HandleChainHead(context, hash)

	if strings.Contains(testHelper.GetBody(context), testHelper.ECBlockHeadPrimaryIndex) == false {
		t.Errorf("Invalid entry credit block head: %v", testHelper.GetBody(context))
	}

	hash = "000000000000000000000000000000000000000000000000000000000000000f"

	testHelper.ClearContextResponseWriter(context)
	HandleChainHead(context, hash)

	if strings.Contains(testHelper.GetBody(context), testHelper.FBlockHeadPrimaryIndex) == false {
		t.Errorf("Invalid factoid block head: %v", testHelper.GetBody(context))
	}

	hash = "6e7e64ac45ff57edbf8537a0c99fba2e9ee351ef3d3f4abd93af9f01107e592c"

	testHelper.ClearContextResponseWriter(context)
	HandleChainHead(context, hash)

	if strings.Contains(testHelper.GetBody(context), testHelper.EBlockHeadSecondaryIndex) == false {
		t.Errorf("Invalid entry block head: %v", testHelper.GetBody(context))
	}

	hash = "df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604"

	testHelper.ClearContextResponseWriter(context)
	HandleChainHead(context, hash)

	if strings.Contains(testHelper.GetBody(context), testHelper.AnchorBlockHeadSecondaryIndex) == false {
		t.Errorf("Invalid anchor entry block head: %v", testHelper.GetBody(context))
	}
}

/*
func TestHandleEntryCreditBalance(t *testing.T) {
	context := testHelper.CreateWebContext()
	eckey := testHelper.NewECAddressPublicKeyString(0)

	HandleEntryCreditBalance(context, eckey)

	expectedAmount := "400"
	if strings.Contains(testHelper.GetBody(context), expectedAmount) == false {
		t.Errorf("%v", testHelper.GetBody(context))
	}
	testHelper.ClearContextResponseWriter(context)

	eckey = testHelper.NewECAddressString(0)

	HandleEntryCreditBalance(context, eckey)

	if strings.Contains(testHelper.GetBody(context), expectedAmount) == false {
		t.Errorf("%v", testHelper.GetBody(context))
	}
}
*/
/*
func TestHandleFactoidBalance(t *testing.T) {
	context := testHelper.CreateWebContext()
	eckey := testHelper.NewFactoidRCDAddressString(0)

	//t.Logf("%v\n", eckey)

	HandleFactoidBalance(context, eckey)

	//expectedAmount := fmt.Sprintf("%v", uint64(testHelper.BlockCount)*testHelper.DefaultCoinbaseAmount)
	expectedAmount := "199977800"
	if strings.Contains(testHelper.GetBody(context), expectedAmount) == false {
		t.Errorf("%v", testHelper.GetBody(context))
	}
}
*/

func TestHandleGetFee(t *testing.T) {
	context := testHelper.CreateWebContext()

	HandleGetFee(context)

	if strings.Contains(testHelper.GetBody(context), "") == false {
		t.Errorf("%v", testHelper.GetBody(context))
	}
}

func TestDBlockList(t *testing.T) {
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

	context := testHelper.CreateWebContext()
	for i, l := range list {
		testHelper.ClearContextResponseWriter(context)
		HandleDirectoryBlock(context, l)

		j := testHelper.GetRespText(context)
		block := new(DBlock)
		err := primitives.DecodeJSONString(j, block)
		if err != nil {
			t.Errorf("Error loading DBlock %v - %v", i, err)
		}
	}

	hash := "000000000000000000000000000000000000000000000000000000000000000d"

	testHelper.ClearContextResponseWriter(context)
	HandleChainHead(context, hash)

	j := testHelper.GetRespText(context)
	head := new(CHead)
	err := primitives.DecodeJSONString(j, head)
	if err != nil {
		panic(err)
	}

	testHelper.ClearContextResponseWriter(context)
	HandleDirectoryBlock(context, head.ChainHead)

	j = testHelper.GetRespText(context)
	block := new(DBlock)
	err = primitives.DecodeJSONString(j, block)
	if err != nil {
		panic(err)
	}

	//t.Errorf("%s", j)
}

func TestBlockIteration(t *testing.T) {
	context := testHelper.CreateWebContext()
	hash := "000000000000000000000000000000000000000000000000000000000000000d"

	HandleChainHead(context, hash)

	j := testHelper.GetRespText(context)
	head := new(CHead)
	err := primitives.DecodeJSONString(j, head)
	if err != nil {
		panic(err)
	}

	prev := head.ChainHead
	fetched := 0
	for {
		if prev == "0000000000000000000000000000000000000000000000000000000000000000" || prev == "" {
			break
		}
		testHelper.ClearContextResponseWriter(context)
		HandleDirectoryBlock(context, prev)

		j = testHelper.GetRespText(context)
		block := new(DBlock)
		err = primitives.DecodeJSONString(j, block)
		if err != nil {
			panic(err)
		}
		//t.Errorf("\n%v\n", j)
		prev = block.Header.PrevBlockKeyMR
		fetched++
	}
	if fetched != testHelper.BlockCount {
		t.Errorf("DBlock only found %v blocks, was expecting %v", fetched, testHelper.BlockCount)
	}
}

func TestHandleGetReceipt(t *testing.T) {
	context := testHelper.CreateWebContext()
	hash := "be5fb8c3ba92c0436269fab394ff7277c67e9b2de4431b723ce5d89799c0b93a"

	HandleGetReceipt(context, hash)

	j := testHelper.GetRespMap(context)

	if j == nil {
		t.Error("Receipt not found!")
		return
	}

	dbo := context.Server.Env["state"].(interfaces.IState).GetDB()
	//defer context.Server.Env["state"].(interfaces.IState).UnlockDB()

	receipt := j["receipt"].(map[string]interface{})
	marshalled, err := json.Marshal(receipt)
	if err != nil {
		t.Error(err)
	}

	err = receipts.VerifyFullReceipt(dbo, string(marshalled))
	if err != nil {
		t.Logf("receipt - %v", j)
		t.Error(err)
	}
}

func TestHandleGetUnanchoredReceipt(t *testing.T) {
	context := testHelper.CreateWebContext()
	hash := "68a503bd3d5b87d3a41a737e430d2ce78f5e556f6a9269859eeb1e053b7f92f7"

	HandleGetReceipt(context, hash)

	j := testHelper.GetRespMap(context)

	if j == nil {
		t.Error("Receipt not found!")
		return
	}

	dbo := context.Server.Env["state"].(interfaces.IState).GetDB()
	//defer context.Server.Env["state"].(interfaces.IState).UnlockDB()

	receipt := j["receipt"].(map[string]interface{})
	marshalled, err := json.Marshal(receipt)
	if err != nil {
		t.Error(err)
	}

	err = receipts.VerifyFullReceipt(dbo, string(marshalled))
	if err != nil {
		t.Logf("receipt - %v", j)
		t.Error(err)
	}
}
