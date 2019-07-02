// +build all

package adminBlock_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/testHelper"
)

func TestAddReplaceMatryoshkaHashGetHash(t *testing.T) {
	a := new(AddReplaceMatryoshkaHash)
	h := a.Hash()
	expected := "dc48a742ae32cfd66352372d6120ed14d6629fc166246b05ff8b03e23804701f"
	if h.String() != expected {
		t.Errorf("Wrong hash returned - %v vs %v", h.String(), expected)
	}
}

func TestAddReplaceMatryoshkaHashTypeIDCheck(t *testing.T) {
	a := new(AddReplaceMatryoshkaHash)
	b, err := a.MarshalBinary()
	if err != nil {
		t.Errorf("%v", err)
	}
	if b[0] != a.Type() {
		t.Errorf("Invalid byte marshalled")
	}
	a2 := new(AddReplaceMatryoshkaHash)
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

func TestUnmarshalNilAddReplaceMatryoshkaHash(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(AddReplaceMatryoshkaHash)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

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

func TestAddMatryoshkaHashMisc(t *testing.T) {
	a := new(AddReplaceMatryoshkaHash)
	if a.String() != "    E:            AddReplaceMatryoshkaHash --   IdentityChainID   000000        MHash 00000000" {
		t.Error("Unexpected string:", a.String())
	}
	as, err := a.JSONString()
	if err != nil {
		t.Error(err)
	}
	if as != "{\"adminidtype\":3,\"identitychainid\":\"0000000000000000000000000000000000000000000000000000000000000000\",\"mhash\":\"0000000000000000000000000000000000000000000000000000000000000000\"}" {
		t.Error("Unexpected JSON string:", as)
	}
	ab, err := a.JSONByte()
	if err != nil {
		t.Error(err)
	}
	if string(ab) != "{\"adminidtype\":3,\"identitychainid\":\"0000000000000000000000000000000000000000000000000000000000000000\",\"mhash\":\"0000000000000000000000000000000000000000000000000000000000000000\"}" {
		t.Error("Unexpected JSON bytes:", as)
	}

	if a.IsInterpretable() {
		t.Error("IsInterpretable should return false")
	}
	if a.Interpret() != "" {
		t.Error("Interpret should return empty string")
	}
}
