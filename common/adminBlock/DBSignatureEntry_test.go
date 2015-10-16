package adminBlock

import (
	"testing"
)

func TestDBSEMarshalledSize(t *testing.T) {
	dbse := new(DBSignatureEntry)
	if dbse.MarshalledSize() != uint64(1+32+32+64) {
		t.Fail()
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
