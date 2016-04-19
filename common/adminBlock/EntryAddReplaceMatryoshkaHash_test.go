package adminBlock_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/testHelper"
)

func TestAddReplaceMatryoshkaHashMarshalUnmarshal(t *testing.T) {
	identity := testHelper.NewRepeatingHash(0xAB)
	mhash := testHelper.NewRepeatingHash(0xCD)

	rmh := NewAddReplaceMatryoshkaHash(identity, mhash)
	if rmh.Type() != constants.TYPE_ADD_MATRYOSHKA {
		t.Errorf("Invalid type")
	}
	if rmh.IdentityChainID.IsSameAs(identity) == false {
		t.Errorf("Invalid IdentityChainID")
	}
	if rmh.MHash.IsSameAs(mhash) == false {
		t.Errorf("Invalid MHash")
	}
	tmp2, err := rmh.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	rmh = new(AddReplaceMatryoshkaHash)
	err = rmh.UnmarshalBinary(tmp2)
	if err != nil {
		t.Error(err)
	}
	if rmh.Type() != constants.TYPE_ADD_MATRYOSHKA {
		t.Errorf("Invalid type")
	}
	if rmh.IdentityChainID.IsSameAs(identity) == false {
		t.Errorf("Invalid IdentityChainID")
	}
	if rmh.MHash.IsSameAs(mhash) == false {
		t.Errorf("Invalid MHash")
	}
}
