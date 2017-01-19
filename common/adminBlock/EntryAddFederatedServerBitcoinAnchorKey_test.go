package adminBlock_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/testHelper"
)

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
