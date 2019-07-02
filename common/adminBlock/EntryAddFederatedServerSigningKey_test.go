// +build all

package adminBlock_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/testHelper"
)

func TestAddFederatedServerSigningKeyGetHash(t *testing.T) {
	a := new(AddFederatedServerSigningKey)
	h := a.Hash()
	expected := "45147ca74abd1327e991c6e313470b1e6e3355f0e53418565ac67ab2553c25bf"
	if h.String() != expected {
		t.Errorf("Wrong hash returned - %v vs %v", h.String(), expected)
	}
}

func TestAddFederatedServerSigningKeyTypeIDCheck(t *testing.T) {
	a := new(AddFederatedServerSigningKey)
	b, err := a.MarshalBinary()
	if err != nil {
		t.Errorf("%v", err)
	}
	if b[0] != a.Type() {
		t.Errorf("Invalid byte marshalled")
	}
	a2 := new(AddFederatedServerSigningKey)
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

func TestUnmarshalNilAddFederatedServerSigningKey(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(AddFederatedServerSigningKey)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestAddFederatedServerSigningKeyMarshalUnmarshal(t *testing.T) {
	identity := testHelper.NewRepeatingHash(0xAB)
	priv := testHelper.NewPrimitivesPrivateKey(1)
	pub := priv.Pub
	var keyPriority byte = 3

	afsk := NewAddFederatedServerSigningKey(identity, keyPriority, *pub, 0)
	if afsk.Type() != constants.TYPE_ADD_FED_SERVER_KEY {
		t.Errorf("Invalid type")
	}
	if afsk.IdentityChainID.IsSameAs(identity) == false {
		t.Errorf("Invalid IdentityChainID")
	}
	if afsk.KeyPriority != keyPriority {
		t.Errorf("Invalid KeyPriority")
	}
	if afsk.PublicKey.String() != pub.String() {
		t.Errorf("Invalid PublicKey")
	}
	tmp2, err := afsk.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	afsk = new(AddFederatedServerSigningKey)
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
	if afsk.PublicKey.String() != pub.String() {
		t.Errorf("Invalid PublicKey")
	}
}

func TestAddFedServerSignMisc(t *testing.T) {
	a := new(AddFederatedServerSigningKey)
	if a.String() != "    E:        AddFederatedServerSigningKey --   IdentityChainID   000000  KeyPriority        0    PublicKey 00000000     DBHeight 0" {
		t.Error("Unexpected string:", a.String())
	}
	as, err := a.JSONString()
	if err != nil {
		t.Error(err)
	}
	if as != "{\"adminidtype\":8,\"identitychainid\":\"0000000000000000000000000000000000000000000000000000000000000000\",\"keypriority\":0,\"publickey\":\"0000000000000000000000000000000000000000000000000000000000000000\",\"dbheight\":0}" {
		t.Error("Unexpected JSON string:", as)
	}
	ab, err := a.JSONByte()
	if err != nil {
		t.Error(err)
	}
	if string(ab) != "{\"adminidtype\":8,\"identitychainid\":\"0000000000000000000000000000000000000000000000000000000000000000\",\"keypriority\":0,\"publickey\":\"0000000000000000000000000000000000000000000000000000000000000000\",\"dbheight\":0}" {
		t.Error("Unexpected JSON bytes:", string(ab))
	}

	if a.IsInterpretable() {
		t.Error("IsInterpretable should return false")
	}
	if a.Interpret() != "" {
		t.Error("Interpret should return empty string")
	}
}
