package wire_test

import (
	"testing"

	"github.com/FactomProject/btcd/wire"
)

// TestInvVectStringer tests the stringized output for inventory vector types.
func TestInternalMsg(t *testing.T) {
	testmap := make (map[string]wire.FtmInternalMsg)
	m1 := new(wire.MsgInt_PLI)	
	testmap[m1.Command()] = m1	
	
	m2 := wire.NewMsgCommitChain()
	testmap[m2.Command()] = m2	
	

	for k,v := range testmap {
		t.Log(k)
		t.Log(v.Command())
	}
}
