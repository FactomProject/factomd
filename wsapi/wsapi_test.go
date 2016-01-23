package wsapi_test

import (
	"fmt"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/receipts"
	"github.com/FactomProject/factomd/testHelper"
	. "github.com/FactomProject/factomd/wsapi"
	"github.com/hoisie/web"
	"net/http"
	"strings"
	"testing"
	"encoding/json"
)

/*
func TestHandleFactoidSubmit(t *testing.T) {
	context := createWebContext()

	HandleFactoidSubmit(context)

	if strings.Contains(GetBody(context), "") == false {
		t.Errorf("%v", GetBody(context))
	}
}
*/
/*
func TestHandleCommitChain(t *testing.T) {
	context := createWebContext()

	HandleCommitChain(context)

	if strings.Contains(GetBody(context), "") == false {
		t.Errorf("%v", GetBody(context))
	}
}
*/
/*
func TestHandleRevealChain(t *testing.T) {
	context := createWebContext()

	HandleRevealChain(context)

	if strings.Contains(GetBody(context), "") == false {
		t.Errorf("%v", GetBody(context))
	}
}
*/
/*
func TestHandleCommitEntry(t *testing.T) {
	context := createWebContext()

	HandleCommitEntry(context)

	if strings.Contains(GetBody(context), "") == false {
		t.Errorf("%v", GetBody(context))
	}
}
*/
/*
func TestHandleRevealEntry(t *testing.T) {
	context := createWebContext()

	HandleRevealEntry(context)

	if strings.Contains(GetBody(context), "") == false {
		t.Errorf("%v", GetBody(context))
	}
}
*/

func TestHandleDirectoryBlockHead(t *testing.T) {
	context := createWebContext()

	HandleDirectoryBlockHead(context)

	if strings.Contains(GetBody(context), "043652d5269764b6b82339aa232bf332790ce54a8c574cfc1a8e6ce86dbe1cdf") == false {
		t.Errorf("Context does not contain proper DBlock Head - %v", GetBody(context))
	}
}

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

	context := createWebContext()
	for i, v := range toTest {
		clearContextResponseWriter(context)
		HandleGetRaw(context, v.Hash1)

		if strings.Contains(GetBody(context), v.Raw) == false {
			t.Errorf("Looking for %v", v.Hash1)
			t.Errorf("GetRaw %v/%v from Hash1 failed - %v", i, len(toTest), GetBody(context))
		}

		clearContextResponseWriter(context)
		HandleGetRaw(context, v.Hash2)

		if strings.Contains(GetBody(context), v.Raw) == false {
			t.Errorf("Looking for %v", v.Hash2)
			t.Errorf("GetRaw %v/%v from Hash2 failed - %v", i, len(toTest), GetBody(context))
		}
	}
}

func TestHandleDirectoryBlock(t *testing.T) {
	context := createWebContext()
	hash := "043652d5269764b6b82339aa232bf332790ce54a8c574cfc1a8e6ce86dbe1cdf"

	HandleDirectoryBlock(context, hash)

	if strings.Contains(GetBody(context), "000000000000000000000000000000000000000000000000000000000000000a") == false {
		t.Errorf("%v", GetBody(context))
	}

	if strings.Contains(GetBody(context), "8010d839cd71aa3eb6f4bd9fd474d50fd63853e4aad428b869142410b30c2737") == false {
		t.Errorf("%v", GetBody(context))
	}

	if strings.Contains(GetBody(context), "6e7e64ac45ff57edbf8537a0c99fba2e9ee351ef3d3f4abd93af9f01107e592c") == false {
		t.Errorf("%v", GetBody(context))
	}

	if strings.Contains(GetBody(context), "25c9e5963917c97ed988c571e703104b34d11f2f6241c0c69d9cfd6ad94491db") == false {
		t.Errorf("%v", GetBody(context))
	}

	if strings.Contains(GetBody(context), "df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604") == false {
		t.Errorf("%v", GetBody(context))
	}

	if strings.Contains(GetBody(context), "cc6eda99d3bb94f29539b4567f8577974782328a227c470a301a436fd35522c6") == false {
		t.Errorf("%v", GetBody(context))
	}

	if strings.Contains(GetBody(context), "000000000000000000000000000000000000000000000000000000000000000c") == false {
		t.Errorf("%v", GetBody(context))
	}

	if strings.Contains(GetBody(context), "a6de386509259b4e143fba5b094ec0dd5e71dbe64f4e6976e785e815e8e5a3b3") == false {
		t.Errorf("%v", GetBody(context))
	}

	if strings.Contains(GetBody(context), "000000000000000000000000000000000000000000000000000000000000000f") == false {
		t.Errorf("%v", GetBody(context))
	}

	if strings.Contains(GetBody(context), "484f328de0a33a2451323cd9963ad4cbb93a51357f7b8ad84d92b79efe86d94a") == false {
		t.Errorf("%v", GetBody(context))
	}
}

func TestHandleEntryBlock(t *testing.T) {
	context := createWebContext()
	chain, err := primitives.HexToHash("df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604")
	if err != nil {
		t.Error(err)
	}
	blocks, err := context.Server.Env["state"].(interfaces.IState).GetDB().FetchAllEBlocksByChain(chain)
	if err != nil {
		t.Error(err)
	}
	fetched := 0
	for _, b := range blocks {
		hash := b.(*entryBlock.EBlock).DatabasePrimaryIndex().String()
		hash2 := b.(*entryBlock.EBlock).DatabaseSecondaryIndex().String()

		HandleEntryBlock(context, hash)

		if strings.Contains(GetBody(context), "df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604") == false {
			t.Errorf("%v", GetBody(context))
		}

		clearContextResponseWriter(context)
		HandleEntryBlock(context, hash2)

		if strings.Contains(GetBody(context), "df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604") == false {
			t.Errorf("%v", GetBody(context))
		}

		fetched++
	}
	if fetched != testHelper.BlockCount {
		t.Errorf("Fetched %v blocks, expected %v", fetched, testHelper.BlockCount)
	}
}

func TestHandleEntry(t *testing.T) {
	context := createWebContext()
	hash := ""

	HandleEntry(context, hash)

	if strings.Contains(GetBody(context), "") == false {
		t.Errorf("%v", GetBody(context))
	}
}

func TestHandleChainHead(t *testing.T) {
	context := createWebContext()
	hash := "000000000000000000000000000000000000000000000000000000000000000d"

	HandleChainHead(context, hash)

	if strings.Contains(GetBody(context), "78a3fecb55cdb684d1034d9ff27061a57699ab30f6f5a5a7e2b24465c85ad33f") == false {
		t.Errorf("Invalid directory block head: %v", GetBody(context))
	}

	hash = "000000000000000000000000000000000000000000000000000000000000000a"

	clearContextResponseWriter(context)
	HandleChainHead(context, hash)

	if strings.Contains(GetBody(context), "b07a252e7ff13ef3ae6b18356949af34f535eca0383a03f71f5f4c526c58b562") == false {
		t.Errorf("Invalid admin block head: %v", GetBody(context))
	}

	hash = "6e7e64ac45ff57edbf8537a0c99fba2e9ee351ef3d3f4abd93af9f01107e592c"

	clearContextResponseWriter(context)
	HandleChainHead(context, hash)

	if strings.Contains(GetBody(context), "1127ed78303976572f25dfba2a058e475234c079ea0d0f645280d03caff08347") == false {
		t.Errorf("Invalid entry block head: %v", GetBody(context))
	}

	hash = "000000000000000000000000000000000000000000000000000000000000000c"

	clearContextResponseWriter(context)
	HandleChainHead(context, hash)

	if strings.Contains(GetBody(context), "c5bac69db89fb4eb28aa99a655f1de561442ee461cfa9c2cd9e1321f181ea305") == false {
		t.Errorf("Invalid entry credit block head: %v", GetBody(context))
	}

	hash = "000000000000000000000000000000000000000000000000000000000000000f"

	clearContextResponseWriter(context)
	HandleChainHead(context, hash)

	if strings.Contains(GetBody(context), "067f353dda05ac27261d4b35a09f211f7a4b0182dff0b6098a16ae8659eb7f5f") == false {
		t.Errorf("Invalid factoid block head: %v", GetBody(context))
	}
}

//func TestHandleEntryCreditBalance(t *testing.T) {
//	context := createWebContext()
//
//	HandleEntryCreditBalance(context)
//
//	if strings.Contains(GetBody(context), "") == false {
//		t.Errorf("%v", GetBody(context))
//	}
//}

func TestHandleFactoidBalance(t *testing.T) {
	context := createWebContext()
	eckey := testHelper.NewFactoidRCDAddressString(0)

	fmt.Printf("%v\n", eckey)

	HandleFactoidBalance(context, eckey)

	//expectedAmount := fmt.Sprintf("%v", uint64(testHelper.BlockCount)*testHelper.DefaultCoinbaseAmount)
	expectedAmount := "999889000"
	if strings.Contains(GetBody(context), expectedAmount) == false {
		t.Errorf("%v", GetBody(context))
	}
}

func TestHandleGetFee(t *testing.T) {
	context := createWebContext()

	HandleGetFee(context)

	if strings.Contains(GetBody(context), "") == false {
		t.Errorf("%v", GetBody(context))
	}
}

func TestBlockIteration(t *testing.T) {
	context := createWebContext()
	hash := "000000000000000000000000000000000000000000000000000000000000000d"

	HandleChainHead(context, hash)

	j := GetRespText(context)
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
		clearContextResponseWriter(context)
		HandleDirectoryBlock(context, prev)

		j = GetRespText(context)
		block := new(DBlock)
		err = primitives.DecodeJSONString(j, block)
		if err != nil {
			panic(err)
		}
		prev = block.Header.PrevBlockKeyMR
		fetched++
	}
	if fetched != testHelper.BlockCount {
		t.Errorf("DBlock only found %v blocks, was expecting %v", fetched, testHelper.BlockCount)
	}
}

func TestHandleGetReceipt(t *testing.T) {
	context := createWebContext()
	hash := "cf9503fad6a6cf3cf6d7a5a491e23d84f9dee6dacb8c12f428633995655bd0d0"

	HandleGetReceipt(context, hash)

	j := GetRespText(context)

	dbo := context.Server.Env["state"].(interfaces.IState).GetDB()

	err := receipts.VerifyFullReceipt(dbo, j)
	if err != nil {
		t.Error(err)
	}
}

//****************************************************************

func GetRespText(context *web.Context) string {
	j := GetBody(context)

	unmarshalled:=map[string]interface{}{}
	err:=json.Unmarshal([]byte(j), &unmarshalled)
	if err != nil {
		panic(err)
	}
	marshalled, err:=json.Marshal(unmarshalled["Response"])
	if err != nil {
		panic(err)
	}
	return string(marshalled)
}

func clearContextResponseWriter(context *web.Context) {
	context.ResponseWriter = new(TestResponseWriter)
}

func createWebContext() *web.Context {
	context := new(web.Context)
	context.Server = new(web.Server)
	context.Server.Env = map[string]interface{}{}
	context.Server.Env["state"] = testHelper.CreateAndPopulateTestState()
	context.ResponseWriter = new(TestResponseWriter)

	return context
}

type TestResponseWriter struct {
	HeaderCode int
	Head       map[string][]string
	Body       string
}

var _ http.ResponseWriter = (*TestResponseWriter)(nil)

func (t *TestResponseWriter) Header() http.Header {
	if t.Head == nil {
		t.Head = map[string][]string{}
	}
	return (http.Header)(t.Head)
}

func (t *TestResponseWriter) WriteHeader(h int) {
	t.HeaderCode = h
}

func (t *TestResponseWriter) Write(b []byte) (int, error) {
	t.Body = t.Body + string(b)
	return len(b), nil
}

func GetBody(context *web.Context) string {
	return context.ResponseWriter.(*TestResponseWriter).Body
}
