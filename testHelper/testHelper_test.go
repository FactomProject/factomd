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
	t.Log("Highest Recorded Block: ", state.GetHighestSavedBlk())
}

func TestCreateFullTestBlockSet(t *testing.T) {
	set := CreateFullTestBlockSet()
	if set[BlockCount-1].DBlock.DatabasePrimaryIndex().String() != "21e5655a7063bb289c09a4569b887dff25c3b43ba156ee9acc8b3b52e6679c04" {
		t.Errorf("Wrong block hash - %v vs %v", set[BlockCount-1].DBlock.DatabasePrimaryIndex().String(), "21e5655a7063bb289c09a4569b887dff25c3b43ba156ee9acc8b3b52e6679c04")
	}
	if set[BlockCount-1].DBlock.DatabaseSecondaryIndex().String() != "1fc0de1370083dcee9d6e7d929d6bcb0efac46ec02ac6f45a5a8c42ece0848d0" {
		t.Errorf("Wrong block hash - %v vs %v", set[BlockCount-1].DBlock.DatabaseSecondaryIndex().String(), "1fc0de1370083dcee9d6e7d929d6bcb0efac46ec02ac6f45a5a8c42ece0848d0")
	}

	if set[BlockCount-1].ABlock.DatabasePrimaryIndex().String() != "956c41312070f58c628ca8027297e0af0aaaa7b8af7f84283fc5ad21a49cc00a" {
		t.Errorf("Wrong block hash - %v vs %v", set[BlockCount-1].ABlock.DatabasePrimaryIndex().String(), "956c41312070f58c628ca8027297e0af0aaaa7b8af7f84283fc5ad21a49cc00a")
	}
	if set[BlockCount-1].ABlock.DatabaseSecondaryIndex().String() != "b07a252e7ff13ef3ae6b18356949af34f535eca0383a03f71f5f4c526c58b562" {
		t.Errorf("Wrong block hash - %v vs %v", set[BlockCount-1].ABlock.DatabaseSecondaryIndex().String(), "b07a252e7ff13ef3ae6b18356949af34f535eca0383a03f71f5f4c526c58b562")
	}

	if set[BlockCount-1].ECBlock.DatabasePrimaryIndex().String() != "99b912e8c705889d3a7295b3583e3675e0bb75c2a47ad7726ac8121b42499831" {
		t.Errorf("Wrong block hash - %v vs %v", set[BlockCount-1].ECBlock.DatabasePrimaryIndex().String(), "99b912e8c705889d3a7295b3583e3675e0bb75c2a47ad7726ac8121b42499831")
	}
	if set[BlockCount-1].ECBlock.DatabaseSecondaryIndex().String() != "5f1da0d95478e989ed19e7f703cb017bc60d4a4e8952db0be6d1864cfc39bc50" {
		t.Errorf("Wrong block hash - %v vs %v", set[BlockCount-1].ECBlock.DatabaseSecondaryIndex().String(), "5f1da0d95478e989ed19e7f703cb017bc60d4a4e8952db0be6d1864cfc39bc50")
	}

	if set[BlockCount-1].FBlock.DatabasePrimaryIndex().String() != "c6cd2ab21d75af1e8589e1eb441411838a508d0674eb294bac4efdc591c3fef4" {
		t.Errorf("Wrong block hash - %v vs %v", set[BlockCount-1].FBlock.DatabasePrimaryIndex().String(), "c6cd2ab21d75af1e8589e1eb441411838a508d0674eb294bac4efdc591c3fef4")
	}
	if set[BlockCount-1].FBlock.DatabaseSecondaryIndex().String() != "e6e8b0a9808bf9ffb53d04acff0dcafc2d5fc7139ef850ab1a5fc94dfd87931e" {
		t.Errorf("Wrong block hash - %v vs %v", set[BlockCount-1].FBlock.DatabaseSecondaryIndex().String(), "e6e8b0a9808bf9ffb53d04acff0dcafc2d5fc7139ef850ab1a5fc94dfd87931e")
	}

	if set[BlockCount-1].EBlock.GetChainID().String() != "6e7e64ac45ff57edbf8537a0c99fba2e9ee351ef3d3f4abd93af9f01107e592c" {
		t.Errorf("Wrong block hash - %v vs %v", set[BlockCount-1].EBlock.GetChainID().String(), "6e7e64ac45ff57edbf8537a0c99fba2e9ee351ef3d3f4abd93af9f01107e592c")
	}
	if set[BlockCount-1].EBlock.DatabasePrimaryIndex().String() != "1127ed78303976572f25dfba2a058e475234c079ea0d0f645280d03caff08347" {
		t.Errorf("Wrong block hash - %v vs %v", set[BlockCount-1].EBlock.DatabasePrimaryIndex().String(), "1127ed78303976572f25dfba2a058e475234c079ea0d0f645280d03caff08347")
	}

	if set[BlockCount-1].AnchorEBlock.GetChainID().String() != "df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604" {
		t.Errorf("Wrong block hash - %v vs %v", set[BlockCount-1].AnchorEBlock.GetChainID().String(), "df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604")
	}
	if set[BlockCount-1].AnchorEBlock.DatabasePrimaryIndex().String() != "b5dfa5186e7d0f3d5be7de2eeeed1137daba08c64cf1b074aabf73648242304e" {
		t.Errorf("Wrong block hash - %v vs %v", set[BlockCount-1].AnchorEBlock.DatabasePrimaryIndex().String(), "b5dfa5186e7d0f3d5be7de2eeeed1137daba08c64cf1b074aabf73648242304e")
	}
}

/*
func TestAnchor(t *testing.T) {
	anchor := CreateFirstAnchorEntry()
	t.Errorf("%x", anchor.ChainID.Bytes())
}*/
