package wsapi_test

import (
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

	if strings.Contains(GetBody(context), "bcdd5686183d351826971f2ee04001ad88ceedb056a44d5a6ea9a3c5bbf1843f") == false {
		t.Errorf("Context does not contain proper DBlock Head - %v", GetBody(context))
	}
}

func TestHandleGetRaw(t *testing.T) {
	context := createWebContext()
	hash := ""

	HandleGetRaw(context, hash)

	if strings.Contains(GetBody(context), "") == false {
		t.Errorf("%v", GetBody(context))
	}
}

func TestHandleDirectoryBlock(t *testing.T) {
	context := createWebContext()
	hash := ""

	HandleDirectoryBlock(context, hash)

	if strings.Contains(GetBody(context), "") == false {
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
	hash := ""

	HandleChainHead(context, hash)

	if strings.Contains(GetBody(context), "") == false {
		t.Errorf("%v", GetBody(context))
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
