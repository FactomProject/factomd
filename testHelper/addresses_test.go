package testHelper_test

import (
	//"encoding/hex"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/testHelper"
	"testing"
)

func TestRCDAddress(t *testing.T) {
	priv := NewPrivKey(0)
	rcd := NewFactoidRCDAddress(0)
	pub := rcd.(*factoid.RCD_1).GetPublicKey()
	//pub2 := [32]byte{}
	//copy(pub2[:], pub)
	pub3 := PrivateKeyToEDPub(priv)

	if primitives.AreBytesEqual(pub, pub3) == false {
		t.Error("RCD public keys are not equal")
	}
}
