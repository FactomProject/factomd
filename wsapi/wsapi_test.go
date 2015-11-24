package wsapi_test

import (
	"github.com/hoisie/web"
	"testing"
)

func TestHandleDirectoryBlock(t *testing.T) {

}

func createWebContext() *web.Context {
	context := new(web.Context)
	context.Server = new(web.Server)

	context.Server.Env["state"] = createState()

	return context
}
