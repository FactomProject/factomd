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

	if strings.Contains(GetBody(context), "2e7b6ef91954a7112f4c089cde9b6b7febe89c74dbca867b967954881ebb0e16") == false {
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
			t.Errorf("GetRaw %v/%v from Hash1 failed - %v", i, len(toTest), GetBody(context))
		}

		clearContextResponseWriter(context)
		HandleGetRaw(context, v.Hash2)

		if strings.Contains(GetBody(context), v.Raw) == false {
			t.Errorf("GetRaw %v/%v from Hash2 failed - %v", i, len(toTest), GetBody(context))
		}
	}
}

func TestHandleDirectoryBlock(t *testing.T) {
	context := createWebContext()
	hash := "95450198260994f250863dc9b25d570f48a61fd7135476f0d391fe78a29250af"

	HandleDirectoryBlock(context, hash)

	if strings.Contains(GetBody(context), "000000000000000000000000000000000000000000000000000000000000000a") == false {
		t.Errorf("%v", GetBody(context))
	}

	if strings.Contains(GetBody(context), "b07a252e7ff13ef3ae6b18356949af34f535eca0383a03f71f5f4c526c58b562") == false {
		t.Errorf("%v", GetBody(context))
	}

	if strings.Contains(GetBody(context), "4bf71c177e71504032ab84023d8afc16e302de970e6be110dac20adbf9a19746") == false {
		t.Errorf("%v", GetBody(context))
	}

	if strings.Contains(GetBody(context), "f282aa3feb35b5922d60ff2e39139a4b8f5eb4ede0844334f36cda9ebeeeeb76") == false {
		t.Errorf("%v", GetBody(context))
	}

	if strings.Contains(GetBody(context), "000000000000000000000000000000000000000000000000000000000000000c") == false {
		t.Errorf("%v", GetBody(context))
	}

	if strings.Contains(GetBody(context), "3b3b6ed470aa64173b62f87ffd35cf3b9df180ae569e800bf05acfe0dd961fad") == false {
		t.Errorf("%v", GetBody(context))
	}

	if strings.Contains(GetBody(context), "000000000000000000000000000000000000000000000000000000000000000f") == false {
		t.Errorf("%v", GetBody(context))
	}

	if strings.Contains(GetBody(context), "915f2d39e09ab51994dc5246628d2dd46e796d7ae65159c72631592d8d10220d") == false {
		t.Errorf("%v", GetBody(context))
	}
}

func TestHandleEntryBlock(t *testing.T) {
	context := createWebContext()
	chain, err := primitives.HexToHash("4bf71c177e71504032ab84023d8afc16e302de970e6be110dac20adbf9a19746")
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

		if strings.Contains(GetBody(context), "4bf71c177e71504032ab84023d8afc16e302de970e6be110dac20adbf9a19746") == false {
			t.Errorf("%v", GetBody(context))
		}

		clearContextResponseWriter(context)
		HandleEntryBlock(context, hash2)

		if strings.Contains(GetBody(context), "4bf71c177e71504032ab84023d8afc16e302de970e6be110dac20adbf9a19746") == false {
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

	if strings.Contains(GetBody(context), "95450198260994f250863dc9b25d570f48a61fd7135476f0d391fe78a29250af") == false {
		t.Errorf("Invalid directory block head: %v", GetBody(context))
	}

	hash = "000000000000000000000000000000000000000000000000000000000000000a"

	clearContextResponseWriter(context)
	HandleChainHead(context, hash)

	if strings.Contains(GetBody(context), "b07a252e7ff13ef3ae6b18356949af34f535eca0383a03f71f5f4c526c58b562") == false {
		t.Errorf("Invalid admin block head: %v", GetBody(context))
	}

	hash = "4bf71c177e71504032ab84023d8afc16e302de970e6be110dac20adbf9a19746"

	clearContextResponseWriter(context)
	HandleChainHead(context, hash)

	if strings.Contains(GetBody(context), "f282aa3feb35b5922d60ff2e39139a4b8f5eb4ede0844334f36cda9ebeeeeb76") == false {
		t.Errorf("Invalid entry block head: %v", GetBody(context))
	}

	hash = "000000000000000000000000000000000000000000000000000000000000000c"

	clearContextResponseWriter(context)
	HandleChainHead(context, hash)

	if strings.Contains(GetBody(context), "3b3b6ed470aa64173b62f87ffd35cf3b9df180ae569e800bf05acfe0dd961fad") == false {
		t.Errorf("Invalid entry credit block head: %v", GetBody(context))
	}

	hash = "000000000000000000000000000000000000000000000000000000000000000f"

	clearContextResponseWriter(context)
	HandleChainHead(context, hash)

	if strings.Contains(GetBody(context), "915f2d39e09ab51994dc5246628d2dd46e796d7ae65159c72631592d8d10220d") == false {
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

	expectedAmount := fmt.Sprintf("%v", uint64(testHelper.BlockCount)*testHelper.DefaultCoinbaseAmount)

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

	json := GetBody(context)
	head := new(CHead)
	err := primitives.DecodeJSONString(json, head)
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

		json = GetBody(context)
		block := new(DBlock)
		err = primitives.DecodeJSONString(json, block)
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

	json := GetBody(context)
	t.Logf("Receipt - %v", json)

	dbo := context.Server.Env["state"].(interfaces.IState).GetDB()

	err := receipts.VerifyFullReceipt(dbo, json)
	if err != nil {
		t.Error(err)
	}
}

//****************************************************************

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
