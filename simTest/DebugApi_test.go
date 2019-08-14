package simtest

import (
	"io/ioutil"
	"testing"

	. "github.com/FactomProject/factomd/testHelper"
)

func TestDebugApi(t *testing.T) {

	state0 := SetupSim("L", map[string]string{"--blktime": "15"}, 20, 0, 0, t)
	_ = state0

	rpc := func (method string,  params string) {
		r, _ := DebugCall(method, params)
		defer r.Body.Close()
		body, _ := ioutil.ReadAll(r.Body)
		t.Logf("BODY: %s", body)
	}

	// TODO: add better responses
	rpc("wait-for-block", `{ "block": 3 }`)
	rpc("wait-for-minute", `{ "minute": 5 }`)
	rpc("wait-blocks", `{ "blocks": 1 }`)
	rpc("wait-minutes", `{ "minutes": 1 }`)

	//TODO test wait-for-past

}
