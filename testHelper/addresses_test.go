package testHelper_test

import (
	//"encoding/hex"
	"testing"

	"github.com/PaulSnow/factom2d/common/factoid"
	"github.com/PaulSnow/factom2d/common/primitives"
	. "github.com/PaulSnow/factom2d/testHelper"
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

func TestPrivPubKeys(t *testing.T) {
	priv := NewPrimitivesPrivateKey(1)
	privStr := priv.PrivateKeyString()
	pubStr := priv.PublicKeyString()

	if privStr+pubStr != "00000000000000000000000000000000000000000000000000000000000000014cb5abf6ad79fbf5abbccafcc269d85cd2651ed4b885b5869f241aedf0a5ba29" {
		t.Errorf("Invalid keys")
	}
}
