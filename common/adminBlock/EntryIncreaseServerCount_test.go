package adminBlock_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
)

func TestUnmarshalNilIncreaseServerCount(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(IncreaseServerCount)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestIncreaseServerCountMarshalUnmarshal(t *testing.T) {
	tmp := []byte{constants.TYPE_ADD_SERVER_COUNT, 0x01, 0x02}
	isc := new(IncreaseServerCount)
	rest, err := isc.UnmarshalBinaryData(tmp)
	if err != nil {
		t.Errorf("%v", err)
	}
	if len(rest) != 1 {
		t.Errorf("Invalid length - %v", len(rest))
	}
	if rest[0] != 0x02 {
		t.Errorf("Invalid rest")
	}
	if isc.Type() != constants.TYPE_ADD_SERVER_COUNT {
		t.Errorf("Invalid type")
	}
	if isc.Amount != 0x01 {
		t.Errorf("Invalid Amount")
	}
	tmp2, err := isc.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	if len(tmp2) != 2 {
		t.Errorf("Invalid len")
	}
	if tmp[0] != constants.TYPE_ADD_SERVER_COUNT {
		t.Errorf("Invalid tmp")
	}
	if tmp[1] != 0x01 {
		t.Errorf("Invalid tmp")
	}

	isc = new(IncreaseServerCount)
	err = isc.UnmarshalBinary(tmp)
	if err != nil {
		t.Error(err)
	}
	if isc.Type() != constants.TYPE_ADD_SERVER_COUNT {
		t.Errorf("Invalid type")
	}
	if isc.Amount != 0x01 {
		t.Errorf("Invalid Amount")
	}
}
