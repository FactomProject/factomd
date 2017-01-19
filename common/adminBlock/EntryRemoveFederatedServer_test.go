package adminBlock_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/testHelper"
)

func TestUnmarshalNilRemoveFederatedServer(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(RemoveFederatedServer)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestRemoveFederatedServerMarshalUnmarshal(t *testing.T) {
	identity := testHelper.NewRepeatingHash(0xAB)
	var dbHeight uint32 = 0xAABBCCDD

	rfs := NewRemoveFederatedServer(identity, dbHeight)
	if rfs.Type() != constants.TYPE_REMOVE_FED_SERVER {
		t.Errorf("Invalid type")
	}
	if rfs.DBHeight != dbHeight {
		t.Errorf("Invalid DBHeight")
	}
	if rfs.IdentityChainID.IsSameAs(identity) == false {
		t.Errorf("Invalid IdentityChainID")
	}
	tmp2, err := rfs.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	rfs = new(RemoveFederatedServer)
	err = rfs.UnmarshalBinary(tmp2)
	if err != nil {
		t.Error(err)
	}
	if rfs.Type() != constants.TYPE_REMOVE_FED_SERVER {
		t.Errorf("Invalid type")
	}
	if rfs.DBHeight != dbHeight {
		t.Errorf("Invalid DBHeight")
	}
	if rfs.IdentityChainID.IsSameAs(identity) == false {
		t.Errorf("Invalid IdentityChainID")
	}
}
