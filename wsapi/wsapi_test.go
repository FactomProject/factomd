package wsapi_test

import (
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
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
func TestHandleCommitChain(t *testing.T) {
	context := createWebContext()

	HandleCommitChain(context)

	if strings.Contains(GetBody(context), "") == false {
		t.Errorf("%v", GetBody(context))
	}
}

func TestHandleRevealChain(t *testing.T) {
	context := createWebContext()

	HandleRevealChain(context)

	if strings.Contains(GetBody(context), "") == false {
		t.Errorf("%v", GetBody(context))
	}
}

func TestHandleCommitEntry(t *testing.T) {
	context := createWebContext()

	HandleCommitEntry(context)

	if strings.Contains(GetBody(context), "") == false {
		t.Errorf("%v", GetBody(context))
	}
}

func TestHandleRevealEntry(t *testing.T) {
	context := createWebContext()

	HandleRevealEntry(context)

	if strings.Contains(GetBody(context), "") == false {
		t.Errorf("%v", GetBody(context))
	}
}

func TestHandleDirectoryBlockHead(t *testing.T) {
	context := createWebContext()

	HandleDirectoryBlockHead(context)

	if strings.Contains(GetBody(context), "2f0fc42094de8172b9a523ba82cef9e517175cac579a70cd473a64a6e277bd6f") == false {
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

	dbEntries := []interfaces.IDBEntry{}
	aBlock := testHelper.CreateTestAdminBlock(nil)
	de := new(directoryBlock.DBEntry)
	de.ChainID, err = primitives.NewShaHash(aBlock.GetChainID())
	if err != nil {
		panic(err)
	}
	de.KeyMR, err = aBlock.GetKeyMR()
	if err != nil {
		panic(err)
	}
	dbEntries = append(dbEntries, de)
	raw := RawData{}
	raw.Hash1 = aBlock.DatabasePrimaryIndex().String()
	raw.Hash2 = aBlock.DatabasePrimaryIndex().String()
	hex, err := aBlock.MarshalBinary()
	if err != nil {
		panic(err)
	}
	raw.Raw = primitives.EncodeBinary(hex)
	toTest = append(toTest, raw)

	eBlock := testHelper.CreateTestEntryBlock(nil)
	de = new(directoryBlock.DBEntry)
	de.ChainID, err = primitives.NewShaHash(eBlock.GetChainID())
	if err != nil {
		panic(err)
	}
	de.KeyMR, err = eBlock.KeyMR()
	if err != nil {
		panic(err)
	}
	dbEntries = append(dbEntries, de)
	raw = RawData{}
	raw.Hash1 = eBlock.DatabasePrimaryIndex().String()
	raw.Hash2 = eBlock.DatabasePrimaryIndex().String()
	hex, err = eBlock.MarshalBinary()
	if err != nil {
		panic(err)
	}
	raw.Raw = primitives.EncodeBinary(hex)
	toTest = append(toTest, raw)

	ecBlock := testHelper.CreateTestEntryCreditBlock(nil)
	de = new(directoryBlock.DBEntry)
	de.ChainID, err = primitives.NewShaHash(ecBlock.GetChainID())
	if err != nil {
		panic(err)
	}
	de.KeyMR, err = ecBlock.HeaderHash()
	if err != nil {
		panic(err)
	}
	dbEntries = append(dbEntries, de)
	raw = RawData{}
	raw.Hash1 = ecBlock.(interfaces.DatabaseBatchable).DatabasePrimaryIndex().String()
	raw.Hash2 = ecBlock.(interfaces.DatabaseBatchable).DatabasePrimaryIndex().String()
	hex, err = ecBlock.MarshalBinary()
	if err != nil {
		panic(err)
	}
	raw.Raw = primitives.EncodeBinary(hex)
	toTest = append(toTest, raw)

	fBlock := testHelper.CreateTestFactoidBlock(nil)
	de = new(directoryBlock.DBEntry)
	de.ChainID, err = primitives.NewShaHash(fBlock.GetChainID())
	if err != nil {
		panic(err)
	}
	de.KeyMR = fBlock.GetKeyMR()
	dbEntries = append(dbEntries, de)
	raw = RawData{}
	raw.Hash1 = fBlock.(interfaces.DatabaseBatchable).DatabasePrimaryIndex().String()
	raw.Hash2 = fBlock.(interfaces.DatabaseBatchable).DatabasePrimaryIndex().String()
	hex, err = fBlock.MarshalBinary()
	if err != nil {
		panic(err)
	}
	raw.Raw = primitives.EncodeBinary(hex)
	toTest = append(toTest, raw)

	dBlock := testHelper.CreateTestDirectoryBlock(nil)
	dBlock.SetDBEntries(dbEntries)
	raw = RawData{}
	raw.Hash1 = dBlock.DatabasePrimaryIndex().String()
	raw.Hash2 = dBlock.DatabasePrimaryIndex().String()
	hex, err = dBlock.MarshalBinary()
	if err != nil {
		panic(err)
	}
	raw.Raw = primitives.EncodeBinary(hex)
	toTest = append(toTest, raw)

	for _, v := range toTest {
		context := createWebContext()
		HandleGetRaw(context, v.Hash1)

		if strings.Contains(GetBody(context), v.Raw) == false {
			t.Errorf("GetRaw from Hash1 failed - %v", GetBody(context))
		}

		context = createWebContext()
		HandleGetRaw(context, v.Hash2)

		if strings.Contains(GetBody(context), v.Raw) == false {
			t.Errorf("GetRaw from Hash2 failed - %v", GetBody(context))
		}
	}
}

func TestHandleDirectoryBlock(t *testing.T) {
	context := createWebContext()
	hash := "c1e0245b28d31cc1163509a0be9e6a5b63e9f8233574b4376c30ca0b9d0cf3e8"

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

	if strings.Contains(GetBody(context), "8a6c19ac1f32c6c36f1134aed634550352485bb140739dda6fe587c6cf91e232") == false {
		t.Errorf("%v", GetBody(context))
	}
}

func TestHandleEntryBlock(t *testing.T) {
	context := createWebContext()
	hash := ""

	HandleEntryBlock(context, hash)

	if strings.Contains(GetBody(context), "") == false {
		t.Errorf("%v", GetBody(context))
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

	if strings.Contains(GetBody(context), "c1e0245b28d31cc1163509a0be9e6a5b63e9f8233574b4376c30ca0b9d0cf3e8") == false {
		t.Errorf("Invalid directory block head: %v", GetBody(context))
	}

	context = createWebContext()
	hash = "000000000000000000000000000000000000000000000000000000000000000a"

	HandleChainHead(context, hash)

	if strings.Contains(GetBody(context), "b07a252e7ff13ef3ae6b18356949af34f535eca0383a03f71f5f4c526c58b562") == false {
		t.Errorf("Invalid admin block head: %v", GetBody(context))
	}

	context = createWebContext()
	hash = "4bf71c177e71504032ab84023d8afc16e302de970e6be110dac20adbf9a19746"

	HandleChainHead(context, hash)

	if strings.Contains(GetBody(context), "f282aa3feb35b5922d60ff2e39139a4b8f5eb4ede0844334f36cda9ebeeeeb76") == false {
		t.Errorf("Invalid entry block head: %v", GetBody(context))
	}

	context = createWebContext()
	hash = "000000000000000000000000000000000000000000000000000000000000000c"

	HandleChainHead(context, hash)

	if strings.Contains(GetBody(context), "3b3b6ed470aa64173b62f87ffd35cf3b9df180ae569e800bf05acfe0dd961fad") == false {
		t.Errorf("Invalid entry credit block head: %v", GetBody(context))
	}

	context = createWebContext()
	hash = "000000000000000000000000000000000000000000000000000000000000000f"

	HandleChainHead(context, hash)

	if strings.Contains(GetBody(context), "8a6c19ac1f32c6c36f1134aed634550352485bb140739dda6fe587c6cf91e232") == false {
		t.Errorf("Invalid factoid block head: %v", GetBody(context))
	}
}

func TestHandleEntryCreditBalance(t *testing.T) {
	context := createWebContext()

	HandleEntryCreditBalance(context)

	if strings.Contains(GetBody(context), "") == false {
		t.Errorf("%v", GetBody(context))
	}
}

func TestHandleFactoidBalance(t *testing.T) {
	context := createWebContext()
	eckey := ""

	HandleFactoidBalance(context, eckey)

	if strings.Contains(GetBody(context), "") == false {
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
