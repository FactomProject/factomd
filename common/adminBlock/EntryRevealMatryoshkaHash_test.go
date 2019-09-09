package adminBlock_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/testHelper"
)

// TestRevealMatryoshkaHashGetHash checks that an new, empty RevealMatryoshkaHash has the proper hash
func TestRevealMatryoshkaHashGetHash(t *testing.T) {
	a := new(RevealMatryoshkaHash)
	h := a.Hash()
	expected := "977c6d24ff2b851777af4dce0615e547112c6c0128a37338b3a1db9d055fff09"
	if h.String() != expected {
		t.Errorf("Wrong hash returned - %v vs %v", h.String(), expected)
	}
}

// TestRevealMatryoshkaHashTypeIDCheck checks that an empty RevealMatryoshkaHash can be unmarshaled correctly, and that a corrupted object
// will throw an error when unmarshaled
func TestRevealMatryoshkaHashTypeIDCheck(t *testing.T) {
	a := new(RevealMatryoshkaHash)
	b, err := a.MarshalBinary()
	if err != nil {
		t.Errorf("%v", err)
	}
	if b[0] != a.Type() {
		t.Errorf("Invalid byte marshalled")
	}
	a2 := new(RevealMatryoshkaHash)
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

// TestUnmarshalNilRevealMatryoshkaHash checks that nil and empty interfaces throw errors when unmarshalled
func TestUnmarshalNilRevealMatryoshkaHash(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(RevealMatryoshkaHash)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

// TestRevealMatryoshkaHashMarshalUnmarshal checks that a RevealMatryoshkaHash is correctly created, and can be marshaled and unmarshalled
// correctly
func TestRevealMatryoshkaHashMarshalUnmarshal(t *testing.T) {
	identity := testHelper.NewRepeatingHash(0xAB)
	mhash := testHelper.NewRepeatingHash(0xCD)

	rmh := NewRevealMatryoshkaHash(identity, mhash)
	if rmh.Type() != constants.TYPE_REVEAL_MATRYOSHKA {
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

	rmh = new(RevealMatryoshkaHash)
	err = rmh.UnmarshalBinary(tmp2)
	if err != nil {
		t.Error(err)
	}
	if rmh.Type() != constants.TYPE_REVEAL_MATRYOSHKA {
		t.Errorf("Invalid type")
	}
	if rmh.IdentityChainID.IsSameAs(identity) == false {
		t.Errorf("Invalid IdentityChainID")
	}
	if rmh.MHash.IsSameAs(mhash) == false {
		t.Errorf("Invalid MHash")
	}
}

// TestRevealMhashMisc checks that the various strings and smaller member functions return proper values
func TestRevealMhashMisc(t *testing.T) {
	a := new(RevealMatryoshkaHash)
	if a.String() != "    E:                RevealMatryoshkaHash --   IdentityChainID   000000         Hash 0000000000" {
		t.Error("Unexpected string:", a.String())
	}
	as, err := a.JSONString()
	if err != nil {
		t.Error(err)
	}
	if as != "{\"adminidtype\":2,\"identitychainid\":\"0000000000000000000000000000000000000000000000000000000000000000\",\"mhash\":\"0000000000000000000000000000000000000000000000000000000000000000\"}" {
		t.Error("Unexpected JSON string:", as)
	}
	ab, err := a.JSONByte()
	if err != nil {
		t.Error(err)
	}
	if string(ab) != "{\"adminidtype\":2,\"identitychainid\":\"0000000000000000000000000000000000000000000000000000000000000000\",\"mhash\":\"0000000000000000000000000000000000000000000000000000000000000000\"}" {
		t.Error("Unexpected JSON bytes:", string(ab))
	}

	if a.IsInterpretable() {
		t.Error("IsInterpretable should return false")
	}
	if a.Interpret() != "" {
		t.Error("Interpret should return empty string")
	}
}
