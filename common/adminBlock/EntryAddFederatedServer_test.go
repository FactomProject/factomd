package adminBlock_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/testHelper"
)

func TestAddFederatedServerGetHash(t *testing.T) {
	a := new(AddFederatedServer)
	h := a.Hash()
	expected := "2b5ab2a07af807fdeaf12f2b2bfbe9b9db084af3e06f6c29f779e3a490c8f0a6"
	if h.String() != expected {
		t.Errorf("Wrong hash returned - %v vs %v", h.String(), expected)
	}
}

func TestAddFederatedServerTypeIDCheck(t *testing.T) {
	a := new(AddFederatedServer)
	b, err := a.MarshalBinary()
	if err != nil {
		t.Errorf("%v", err)
	}
	if b[0] != a.Type() {
		t.Errorf("Invalid byte marshalled")
	}
	a2 := new(AddFederatedServer)
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

func TestUnmarshalNilAddFederatedServer(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(AddFederatedServer)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

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

func TestAddFedServerMisc(t *testing.T) {
	a := new(AddFederatedServer)
	if a.String() != "    E:                        AddFedServer --   IdentityChainID   000000     DBHeight        0" {
		t.Error("Unexpected string:", a.String())
	}
	as, err := a.JSONString()
	if err != nil {
		t.Error(err)
	}
	if as != "{\"adminidtype\":5,\"identitychainid\":\"0000000000000000000000000000000000000000000000000000000000000000\",\"dbheight\":0}" {
		t.Error("Unexpected JSON string:", as)
	}
	ab, err := a.JSONByte()
	if err != nil {
		t.Error(err)
	}
	if string(ab) != "{\"adminidtype\":5,\"identitychainid\":\"0000000000000000000000000000000000000000000000000000000000000000\",\"dbheight\":0}" {
		t.Error("Unexpected JSON bytes:", string(ab))
	}

	if a.IsInterpretable() {
		t.Error("IsInterpretable should return false")
	}
	if a.Interpret() != "" {
		t.Error("Interpret should return empty string")
	}
}
