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
		t.Errorf("Marshaling for a and b are not the same")
	}
}
