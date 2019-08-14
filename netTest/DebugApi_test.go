package nettest

import (
	"testing"
)

func TestDebugApi(t *testing.T) {

	n := SetupNode(t)
	_ = n

	//rpc("wait-for-block", `{ "block": 3 }`)
	//rpc("wait-for-minute", `{ "minute": 5 }`)
	rpc("wait-blocks", `{ "blocks": 10 }`)
	//rpc("wait-minutes", `{ "minutes": 1 }`)

}
