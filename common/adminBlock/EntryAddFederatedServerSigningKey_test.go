package adminBlock_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/testHelper"
)

func TestAddFederatedServerSigningKeyMarshalUnmarshal(t *testing.T) {
	identity := testHelper.NewRepeatingHash(0xAB)
	priv := testHelper.NewPrimitivesPrivateKey(1)
	pub := priv.Pub
	var keyPriority byte = 3

	afsk := NewAddFederatedServerSigningKey(identity, keyPriority, *pub)
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
