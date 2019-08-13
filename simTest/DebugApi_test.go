package simtest

import (
	"io/ioutil"
	"testing"

	. "github.com/FactomProject/factomd/testHelper"
)

func TestDebugApi(t *testing.T) {

	state0 := SetupSim("L", map[string]string{"--blktime": "15"}, 20, 0, 0, t)

	_ = state0
	r, _ := DebugCall("wait-for-block", `{ "block": 3 }`)

	//var data []byte
	//r.Body.Read(data)
	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)
	t.Logf("BODY: %s", body)

}
