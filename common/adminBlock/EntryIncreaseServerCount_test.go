package adminBlock_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
)

// TestIncreaseServerCountGetHash checks that an empty object has the correct hash
func TestIncreaseServerCountGetHash(t *testing.T) {
	a := new(IncreaseServerCount)
	h := a.Hash()
	expected := "c0ba8a33ac67f44abff5984dfbb6f56c46b880ac2b86e1f23e7fa9c402c53ae7"
	if h.String() != expected {
		t.Errorf("Wrong hash returned - %v vs %v", h.String(), expected)
	}
}

// TestIncreaseServerCountTypeIDCheck checks that an empty object can be marshalled and unmarshalled correctly, and
// that a corrupted marshalled object throws an error when unmarshalled
func TestIncreaseServerCountTypeIDCheck(t *testing.T) {
	a := new(IncreaseServerCount)
	b, err := a.MarshalBinary()
	if err != nil {
		t.Errorf("%v", err)
	}
	if b[0] != a.Type() {
		t.Errorf("Invalid byte marshalled")
	}
	a2 := new(IncreaseServerCount)
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

// TestUnmarshalNilIncreaseServerCount checks that the nil and empty interfaces throw errors when they are unmarshalled
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

// TestIncreaseServerCountMarshalUnmarshal checks that marshalled data can be unmarshaled correctly
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

// TestIncreaseServerMisc checks that the various strings and smaller member functions return proper values
func TestIncreaseServerMisc(t *testing.T) {
	a := new(IncreaseServerCount)
	if a.String() != "    E:               Increase Server Count -- by 0" {
		t.Error("Unexpected string:", a.String())
	}
	as, err := a.JSONString()
	if err != nil {
		t.Error(err)
	}
	if as != "{\"adminidtype\":4,\"amount\":0}" {
		t.Error("Unexpected JSON string:", as)
	}
	ab, err := a.JSONByte()
	if err != nil {
		t.Error(err)
	}
	if string(ab) != "{\"adminidtype\":4,\"amount\":0}" {
		t.Error("Unexpected JSON bytes:", string(ab))
	}

	if a.IsInterpretable() {
		t.Error("IsInterpretable should return false")
	}
	if a.Interpret() != "" {
		t.Error("Interpret should return empty string")
	}
}
