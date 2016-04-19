package adminBlock_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/testHelper"
)

func TestAddFederatedServerMarshalUnmarshal(t *testing.T) {
	identity := testHelper.NewRepeatingHash(0xAB)
	var dbHeight uint32 = 0xAABBCCDD

	rfs := NewAddFederatedServer(identity, dbHeight)
	if rfs.Type() != constants.TYPE_ADD_FED_SERVER {
		t.Errorf("Invalid type")
	}
	if rfs.IdentityChainID.IsSameAs(identity) == false {
		t.Errorf("Invalid IdentityChainID")
	}
	if rfs.DBHeight != dbHeight {
		t.Errorf("Invalid DBHeight")
	}
	tmp2, err := rfs.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	rfs = new(AddFederatedServer)
	err = rfs.UnmarshalBinary(tmp2)
	if err != nil {
		t.Error(err)
	}
	if rfs.Type() != constants.TYPE_ADD_FED_SERVER {
		t.Errorf("Invalid type")
	}
	if rfs.IdentityChainID.IsSameAs(identity) == false {
		t.Errorf("Invalid IdentityChainID")
	}
	if rfs.DBHeight != dbHeight {
		t.Errorf("Invalid DBHeight")
	}
}
