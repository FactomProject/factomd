// +build all 

package state_test

import (
	"fmt"
	"testing"

	"github.com/FactomProject/factomd/common/primitives/random"
	. "github.com/FactomProject/factomd/state"
)

var _ = fmt.Printf

func TestMarshalableUint32_MarshalBinary(t *testing.T) {
	for i := 0; i < 5000; i++ {
		u := random.RandUInt32()
		err := testMarshal(u)
		if err != nil {
			t.Error(err)
		}
	}
}

func testMarshal(u uint32) error {
	var mx = MarshalableUint32(u)
	data, err := (&mx).MarshalBinary()
	if err != nil {
		return err
	}

	var m MarshalableUint32
	err = (&m).UnmarshalBinary(data)
	if err != nil {
		return err
	}

	if uint32(m) != u {
		return fmt.Errorf("Exp %d, got %d", u, uint32(m))
	}

	// Test with normal uint32 to bytes
	data, err = Uint32ToBytes(u)
	var m2 MarshalableUint32
	err = (&m2).UnmarshalBinary(data)
	if err != nil {
		return err
	}

	if uint32(m) != u {
		return fmt.Errorf("Exp %d, got %d", u, uint32(m))
	}
	return nil
}
