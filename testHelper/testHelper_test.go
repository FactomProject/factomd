package testHelper_test

import (
	"crypto/rand"
	"github.com/FactomProject/factomd/util"

	"github.com/FactomProject/factomd/engine"

	"github.com/FactomProject/ed25519"
	//"github.com/FactomProject/factomd/common/factoid/wallet"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/testHelper"
	"github.com/stretchr/testify/assert"
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
	state := CreateAndPopulateTestStateAndStartValidator()
	t.Log("Highest Recorded Block: ", state.GetHighestSavedBlk())
}

func TestCreateFullTestBlockSet(t *testing.T) {
	set := CreateFullTestBlockSet()
	if set[BlockCount-1].DBlock.DatabasePrimaryIndex().String() != DBlockHeadPrimaryIndex {
		t.Errorf("Wrong dblock hash - %v vs %v", set[BlockCount-1].DBlock.DatabasePrimaryIndex().String(), DBlockHeadPrimaryIndex)
	}
	if set[BlockCount-1].DBlock.DatabaseSecondaryIndex().String() != DBlockHeadSecondaryIndex {
		t.Errorf("Wrong dblock hash - %v vs %v", set[BlockCount-1].DBlock.DatabaseSecondaryIndex().String(), DBlockHeadSecondaryIndex)
	}

	if set[BlockCount-1].ABlock.DatabasePrimaryIndex().String() != ABlockHeadPrimaryIndex {
		t.Errorf("Wrong ablock hash - %v vs %v", set[BlockCount-1].ABlock.DatabasePrimaryIndex().String(), ABlockHeadPrimaryIndex)
	}
	if set[BlockCount-1].ABlock.DatabaseSecondaryIndex().String() != ABlockHeadSecondaryIndex {
		t.Errorf("Wrong ablock hash - %v vs %v", set[BlockCount-1].ABlock.DatabaseSecondaryIndex().String(), ABlockHeadSecondaryIndex)
	}

	if set[BlockCount-1].ECBlock.DatabasePrimaryIndex().String() != ECBlockHeadPrimaryIndex {
		t.Errorf("Wrong ecblock hash - %v vs %v", set[BlockCount-1].ECBlock.DatabasePrimaryIndex().String(), ECBlockHeadPrimaryIndex)
	}
	if set[BlockCount-1].ECBlock.DatabaseSecondaryIndex().String() != ECBlockHeadSecondaryIndex {
		t.Errorf("Wrong ecblock hash - %v vs %v", set[BlockCount-1].ECBlock.DatabaseSecondaryIndex().String(), ECBlockHeadSecondaryIndex)
	}

	if set[BlockCount-1].FBlock.DatabasePrimaryIndex().String() != FBlockHeadPrimaryIndex {
		t.Errorf("Wrong fblock hash - %v vs %v", set[BlockCount-1].FBlock.DatabasePrimaryIndex().String(), FBlockHeadPrimaryIndex)
	}
	if set[BlockCount-1].FBlock.DatabaseSecondaryIndex().String() != FBlockHeadSecondaryIndex {
		t.Errorf("Wrong fblock hash - %v vs %v", set[BlockCount-1].FBlock.DatabaseSecondaryIndex().String(), FBlockHeadSecondaryIndex)
	}

	if set[BlockCount-1].EBlock.GetChainID().String() != EBlockHeadPrimaryIndex {
		t.Errorf("Wrong eblock hash - %v vs %v", set[BlockCount-1].EBlock.GetChainID().String(), EBlockHeadPrimaryIndex)
	}
	if set[BlockCount-1].EBlock.DatabasePrimaryIndex().String() != EBlockHeadSecondaryIndex {
		t.Errorf("Wrong eblock hash - %v vs %v", set[BlockCount-1].EBlock.DatabasePrimaryIndex().String(), EBlockHeadSecondaryIndex)
	}

	if set[BlockCount-1].AnchorEBlock.GetChainID().String() != AnchorBlockHeadPrimaryIndex {
		t.Errorf("Wrong anblock hash - %v vs %v", set[BlockCount-1].AnchorEBlock.GetChainID().String(), AnchorBlockHeadPrimaryIndex)
	}
	if set[BlockCount-1].AnchorEBlock.DatabasePrimaryIndex().String() != AnchorBlockHeadSecondaryIndex {
		t.Errorf("Wrong anblock hash - %v vs %v", set[BlockCount-1].AnchorEBlock.DatabasePrimaryIndex().String(), AnchorBlockHeadSecondaryIndex)
	}
}

/*
func TestAnchor(t *testing.T) {
	anchor := CreateFirstAnchorEntry()
	t.Errorf("%x", anchor.ChainID.Bytes())
}*/

func TestNewCommitChain(t *testing.T) {
	j := new(entryBlock.EBlock)
	//block 1000
	eblock1kbytes, _ := hex.DecodeString("df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e6041611c693d62887530c5420a48f2ea2d6038745fc493d6b1e531232805dd2149614ef537df0c73df748b12d508b4334fe8d2832a4cd6ea24f64a3363839bd0efa46e835bfed10ded0d756d7ccafd44830cc942799fca43f2505e9d024b0a9dd3c00000221000003e800000002b24d4ee9e2184673a4d7de6fdac1288ea00b7856940341122c34bd50a662340a0000000000000000000000000000000000000000000000000000000000000009")
	j.UnmarshalBinary(eblock1kbytes)
	k := NewCommitChain(j)
	m, _ := k.MarshalBinary()
	//fmt.Printf("%x\n",m)
	anticipated_commit := "010000000000e8aaec8504394192fc7f6129a024ec5919d38a3967955aa7bbb3ac0ff0879266937015fcc30d961f985ab8a8b5132273fa3fc72a246a68bce1d95cc8c6fdeb183cb24d4ee9e2184673a4d7de6fdac1288ea00b7856940341122c34bd50a662340a013b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29eb46be4cfa8f1e056565aae7d600ac1be4f2cc143333cbbff784b9efb1e8c86a79c333c1f748a45cee507cfabc5de59f6a41b7e83a0af0821de564dc99836b0b"
	cf := fmt.Sprintf("%x", m)
	if anticipated_commit != cf {
		t.Errorf("testhelper NewCommitChain comparison failed")
	}
}

func TestFeeTxnCreate(t *testing.T) {
	var oneFct uint64 = 100000000 // Factoshis
	var ecPrice uint64 = 10000

	balance := oneFct
	inUser := "Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK" // FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q
	outAddress := "FA2s2SJ5Cxmv4MzpbGxVS9zbNCjpNRJoTX4Vy7EZaTwLq3YTur4u"

	for i := 0; i < 10; i++ {
		txn, _ := engine.NewTransaction(balance, inUser, outAddress, ecPrice)
		fee, _ := txn.CalculateFee(ecPrice)
		balance = balance - fee
		assert.Equal(t, 12*ecPrice, fee)
	}
}

func TestTxnCreate(t *testing.T) {
	var amt uint64 = 100000000
	var ecPrice uint64 = 10000

	inUser := "Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK" // FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q
	//outUser := "Fs2GCfAa2HBKaGEUWCtw8eGDkN1CfyS6HhdgLv8783shkrCgvcpJ" // FA2s2SJ5Cxmv4MzpbGxVS9zbNCjpNRJoTX4Vy7EZaTwLq3YTur4u
	outAddress := "FA2s2SJ5Cxmv4MzpbGxVS9zbNCjpNRJoTX4Vy7EZaTwLq3YTur4u"

	txn, err := engine.NewTransaction(amt, inUser, outAddress, ecPrice)
	assert.Nil(t, err)

	err = txn.ValidateSignatures()
	assert.Nil(t, err)

	err = txn.Validate(1)
	assert.Nil(t, err)

	if err := txn.Validate(0); err == nil {
		t.Fatalf("expected coinbase txn to error")
	}

	// test that we are sending to the address we thought
	assert.Equal(t, outAddress, txn.Outputs[0].GetUserAddress())

}

// test that we can get the name of our test
func TestGetName(t *testing.T) {
	TestGetFoo := func() string {
		// add extra frame depth
		return GetTestName()
	}
	assert.Equal(t, "TestGetName", TestGetFoo())
}

func TestResetFactomHome(t *testing.T) {
	s := GetSimHome(t)
	t.Logf("simhome: %v", s)

	h, err := ResetFactomHome(t)
	assert.Nil(t, err)

	t.Logf("reset home: %v", h)
	t.Logf("util home: %v", util.GetHomeDir())

	assert.Equal(t, s, h)
}
