package adminBlock_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/adminBlock"
)

func TestDBSignatureEntryGetHash(t *testing.T) {
	a := new(DBSignatureEntry)
	h := a.Hash()
	expected := "b84147b0eeb997d0942e214ce03c7889e5653f276830838b91e2dfea9528d46d"
	if h.String() != expected {
		t.Errorf("Wrong hash returned - %v vs %v", h.String(), expected)
	}
}

func TestUnmarshalNilDBSignatureEntry(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(DBSignatureEntry)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestDBSEMisc(t *testing.T) {
	dbse := new(DBSignatureEntry)
	if dbse.IsInterpretable() != false {
		t.Fail()
	}
	if dbse.Interpret() != "" {
		t.Fail()
	}
}
