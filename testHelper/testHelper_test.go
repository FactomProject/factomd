package testHelper_test

import (
	"crypto/rand"
	"github.com/FactomProject/ed25519"
	//"github.com/FactomProject/factomd/common/factoid/wallet"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/testHelper"
	"testing"
)

/*
func TestTest(t *testing.T) {
	privKey, pubKey, add := NewFactoidAddressStrings(1)
	t.Errorf("%v, %v, %v", privKey, pubKey, add)
}
*/

func TestSignature(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Error(err)
	}
	t.Logf("priv, pub - %x, %x", *priv, *pub)

	testData := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}

	sig := ed25519.Sign(priv, testData)

	ok := ed25519.Verify(pub, testData, sig)

	if ok == false {
		t.Error("Signature could not be verified")
	}

	pub2, err := primitives.PrivateKeyToPublicKey(priv[:])
	if err != nil {
		t.Error(err)
	}

	t.Logf("pub1 - %x", pub)
	t.Logf("pub2 - %x", pub2)

	if primitives.AreBytesEqual(pub[:], pub2[:]) == false {
		t.Error("Public keys are not equal")
	}
}

func Test(t *testing.T) {
	set := CreateTestBlockSet(nil)
	str, _ := set.ECBlock.JSONString()
	t.Logf("set ECBlock - %v", str)
	str, _ = set.FBlock.JSONString()
	t.Logf("set FBlock - %v", str)
	t.Logf("set Height - %v", set.Height)
}

func Test_DB_With_Ten_Blks(t *testing.T) {
	state := CreateAndPopulateTestState()
	t.Log("Highest Recorded Block: ", state.GetHighestCompletedBlock())
}

/*
func TestAnchor(t *testing.T) {
	anchor := CreateFirstAnchorEntry()
	t.Errorf("%x", anchor.ChainID.Bytes())
}*/
