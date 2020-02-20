package testHelper

import (
	"testing"

	"bytes"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives/random"
)

// TestMarshaling will test a marshing and unmarshal operation. Do the comparison of equality yourself. Extrabytes adds a random number
// of bytes up to extrabytes in length
func TestMarshaling(a interfaces.BinaryMarshallable, b interfaces.BinaryMarshallable, extraBytes int, t *testing.T) {
	data, err := a.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	extraData := random.RandByteSliceOfLen(extraBytes)
	extraBytes = len(extraData)
	extraData = append(data, extraData...)

	nd, err := b.UnmarshalBinaryData(extraData)
	if err != nil {
		t.Error(err)
	}

	if len(nd) != extraBytes {
		t.Errorf("Expect %d bytes remain, but %d do", extraBytes, len(nd))
		t.FailNow()
	}

	data2, err := b.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	if bytes.Compare(data, data2) != 0 {
		t.Errorf("Marshaling for a and b are not the same: \n %x\n %x", data, data2)
	}
}

func TestABlockEntryFunctions(a interfaces.IABEntry, b interfaces.IABEntry, t *testing.T) {
	as, err := a.JSONString()
	if err != nil {
		t.Error(err)
	}

	bs, err := b.JSONString()
	if err != nil {
		t.Error(err)
	}

	if as != bs {
		t.Errorf("JSONString() does not match")
	}

	ab, err := a.JSONByte()
	if err != nil {
		t.Error(err)
	}

	bb, err := b.JSONByte()
	if err != nil {
		t.Error(err)
	}

	if bytes.Compare(ab, bb) != 0 {
		t.Errorf("JSONByte() does not match")
	}

	if a.Interpret() != b.Interpret() {
		t.Errorf("Interpret() does not match")
	}

	if a.IsInterpretable() != b.IsInterpretable() {
		t.Errorf("IsInterpretable() does not match")
	}

	if a.String() != b.String() {
		t.Errorf("String() does not match")
	}

	if !a.Hash().IsSameAs(b.Hash()) {
		t.Errorf("Hash() does not match")
	}
}
