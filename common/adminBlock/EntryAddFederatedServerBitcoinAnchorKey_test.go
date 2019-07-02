// +build all

package adminBlock_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/testHelper"
)

func TestAddFederatedServerBitcoinAnchorKeyGetHash(t *testing.T) {
	a := new(AddFederatedServerBitcoinAnchorKey)
	h := a.Hash()
	expected := "251e97a4cd360b93ae13a641ac39603e14f0e75325864c40d19718002be5a8f1"
	if h.String() != expected {
		t.Errorf("Wrong hash returned - %v vs %v", h.String(), expected)
	}
}

func TestAddFederatedServerBitcoinAnchorKeyTypeIDCheck(t *testing.T) {
	a := new(AddFederatedServerBitcoinAnchorKey)
	b, err := a.MarshalBinary()
	if err != nil {
		t.Errorf("%v", err)
	}
	if b[0] != a.Type() {
		t.Errorf("Invalid byte marshalled")
	}
	a2 := new(AddFederatedServerBitcoinAnchorKey)
	err = a2.UnmarshalBinary(b)
	if err != nil {
		t.Errorf("%v", err)
	}

	b[0] = (b[0] + 1) % 255
	err = a2.UnmarshalBinary(b)
	if err == nil {
		t.Errorf("No error caught")
	}
}

func TestUnmarshalNilAddFederatedServerBitcoinAnchorKey(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(AddFederatedServerBitcoinAnchorKey)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestAddFederatedServerBitcoinAnchorKeyMarshalUnmarshal(t *testing.T) {
	identity := testHelper.NewRepeatingHash(0xAB)
	pub := new(primitives.ByteSlice20)
	err := pub.UnmarshalBinary(testHelper.NewRepeatingHash(0xCD).Bytes())
	if err != nil {
		t.Error(err)
	}
	var keyPriority byte = 3
	var keyType byte = 1

	afsk := NewAddFederatedServerBitcoinAnchorKey(identity, keyPriority, keyType, *pub)
	if afsk.Type() != constants.TYPE_ADD_BTC_ANCHOR_KEY {
		t.Errorf("Invalid type")
	}
	if afsk.IdentityChainID.IsSameAs(identity) == false {
		t.Errorf("Invalid IdentityChainID")
	}
	if afsk.KeyPriority != keyPriority {
		t.Errorf("Invalid KeyPriority")
	}
	if afsk.KeyType != keyType {
		t.Errorf("Invalid KeyType")
	}
	if afsk.ECDSAPublicKey.String() != pub.String() {
		t.Errorf("Invalid ECDSAPublicKey")
	}
	tmp2, err := afsk.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	afsk = new(AddFederatedServerBitcoinAnchorKey)
	err = afsk.UnmarshalBinary(tmp2)
	if err != nil {
		t.Error(err)
	}
	if afsk.IdentityChainID.IsSameAs(identity) == false {
		t.Errorf("Invalid IdentityChainID")
	}
	if afsk.KeyPriority != keyPriority {
		t.Errorf("Invalid KeyPriority")
	}
	if afsk.KeyType != keyType {
		t.Errorf("Invalid KeyType")
	}
	if afsk.ECDSAPublicKey.String() != pub.String() {
		t.Errorf("Invalid ECDSAPublicKey")
	}
}

func TestAddFedServerBTCMisc(t *testing.T) {
	a := new(AddFederatedServerBitcoinAnchorKey)
	if a.String() != "    E:  AddFederatedServerBitcoinAnchorKey --   IdentityChainID   000000  KeyPriority        0      KeyType        0 ECDSAPublicKey 00000000" {
		t.Error("Unexpected string:", a.String())
	}
	as, err := a.JSONString()
	if err != nil {
		t.Error(err)
	}
	if as != "{\"adminidtype\":9,\"identitychainid\":\"0000000000000000000000000000000000000000000000000000000000000000\",\"keypriority\":0,\"keytype\":0,\"ecdsapublickey\":\"0000000000000000000000000000000000000000\"}" {
		t.Error("Unexpected JSON string:", as)
	}
	ab, err := a.JSONByte()
	if err != nil {
		t.Error(err)
	}
	if string(ab) != "{\"adminidtype\":9,\"identitychainid\":\"0000000000000000000000000000000000000000000000000000000000000000\",\"keypriority\":0,\"keytype\":0,\"ecdsapublickey\":\"0000000000000000000000000000000000000000\"}" {
		t.Error("Unexpected JSON bytes:", string(ab))
	}

	if a.IsInterpretable() {
		t.Error("IsInterpretable should return false")
	}
	if a.Interpret() != "" {
		t.Error("Interpret should return empty string")
	}
}
