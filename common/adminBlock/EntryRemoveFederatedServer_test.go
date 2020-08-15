package adminBlock_test

import (
	"testing"

	. "github.com/PaulSnow/factom2d/common/adminBlock"
	"github.com/PaulSnow/factom2d/common/constants"
	"github.com/PaulSnow/factom2d/testHelper"
)

func TestRemoveFederatedServerGetHash(t *testing.T) {
	a := new(RemoveFederatedServer)
	h := a.Hash()
	expected := "c5ca81ef33f2130f84a6c939cd31ddd665a194ca7df2620cd8387a31e245e6c7"
	if h.String() != expected {
		t.Errorf("Wrong hash returned - %v vs %v", h.String(), expected)
	}
}

func TestRemoveFederatedServerTypeIDCheck(t *testing.T) {
	a := new(RemoveFederatedServer)
	b, err := a.MarshalBinary()
	if err != nil {
		t.Errorf("%v", err)
	}
	if b[0] != a.Type() {
		t.Errorf("Invalid byte marshalled")
	}
	a2 := new(RemoveFederatedServer)
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

func TestRemoveServerMisc(t *testing.T) {
	a := new(RemoveFederatedServer)
	if a.String() != "    E:             Remove Federated Server --   IdentityChainID   000000     DBHeight        0" {
		t.Error("Unexpected string:", a.String())
	}
	as, err := a.JSONString()
	if err != nil {
		t.Error(err)
	}
	if as != "{\"adminidtype\":7,\"identitychainid\":\"0000000000000000000000000000000000000000000000000000000000000000\",\"dbheight\":0}" {
		t.Error("Unexpected JSON string:", as)
	}
	ab, err := a.JSONByte()
	if err != nil {
		t.Error(err)
	}
	if string(ab) != "{\"adminidtype\":7,\"identitychainid\":\"0000000000000000000000000000000000000000000000000000000000000000\",\"dbheight\":0}" {
		t.Error("Unexpected JSON bytes:", string(ab))
	}

	if a.IsInterpretable() {
		t.Error("IsInterpretable should return false")
	}
	if a.Interpret() != "" {
		t.Error("Interpret should return empty string")
	}
}
