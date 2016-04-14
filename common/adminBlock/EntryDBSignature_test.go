package adminBlock

import (
	"testing"
)

func TestDBSEMisc(t *testing.T) {
	dbse := new(DBSignatureEntry)
	if dbse.IsInterpretable() != false {
		t.Fail()
	}
	if dbse.Interpret() != "" {
		t.Fail()
	}
}
