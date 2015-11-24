package wsapi_test

import (
	"github.com/FactomProject/factomd/testHelper"
	. "github.com/FactomProject/factomd/wsapi"
	"github.com/hoisie/web"
	"testing"
)

func TestHandleDirectoryBlockHead(t *testing.T) {
	context := createWebContext()

	HandleDirectoryBlockHead(context)

	t.Logf("context - %v", context)
	t.Fail()
}

func createWebContext() *web.Context {
	context := new(web.Context)
	context.Server = new(web.Server)

	context.Server.Env = map[string]interface{}{}

	context.Server.Env["state"] = testHelper.CreateAndPopulateTestState()

	return context
}
