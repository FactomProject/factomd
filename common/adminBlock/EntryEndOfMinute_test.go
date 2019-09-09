package adminBlock_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
)

// TestEndOfMinuteEntryGetHash checks that an empty object has the correct hash
func TestEndOfMinuteEntryGetHash(t *testing.T) {
	a := new(EndOfMinuteEntry)
	h := a.Hash()
	expected := "96a296d224f285c67bee93c30f8a309157f0daa35dc5b87e410b78630a09cfc7"
	if h.String() != expected {
		t.Errorf("Wrong hash returned - %v vs %v", h.String(), expected)
	}
}

// TestEndOfMinuteEntryTypeIDCheck checks that an empty object can be marshalled and unmarshalled correctly, and
// that a corrupted marshalled object throws an error when unmarshalled
func TestEndOfMinuteEntryTypeIDCheck(t *testing.T) {
	a := new(EndOfMinuteEntry)
	b, err := a.MarshalBinary()
	if err != nil {
		t.Errorf("%v", err)
	}
	if b[0] != a.Type() {
		t.Errorf("Invalid byte marshalled")
	}
	a2 := new(EndOfMinuteEntry)
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

// TestUnmarshalNilEndOfMinuteEntry checks that the nil and empty interfaces throw errors when they are unmarshalled
func TestUnmarshalNilEndOfMinuteEntry(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(EndOfMinuteEntry)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

// TestEOMMisc checks that smaller member functions return proper values
func TestEOMMisc(t *testing.T) {
	eom := new(EndOfMinuteEntry)
	if eom.IsInterpretable() != true {
		t.Fail()
	}
	eom.MinuteNumber = 1
	if eom.Interpret() != "End of Minute 1" {
		t.Fail()
	}
	if eom.Type() != constants.TYPE_MINUTE_NUM {
		t.Fail()
	}
}

// TestEOMMarshalUnmarshal checks that marshalled data can be unmarshaled correctly
func TestEOMMarshalUnmarshal(t *testing.T) {
	tmp := []byte{constants.TYPE_MINUTE_NUM, 0x01, 0x02}
	eom := new(EndOfMinuteEntry)
	rest, err := eom.UnmarshalBinaryData(tmp)
	if err != nil {
		t.Errorf("%v", err)
	}
	if len(rest) != 1 {
		t.Errorf("Invalid length - %v", len(rest))
	}
	if rest[0] != 0x02 {
		t.Errorf("Invalid rest")
	}
	if eom.Type() != constants.TYPE_MINUTE_NUM {
		t.Errorf("Invalid type")
	}
	if eom.MinuteNumber != 0x01 {
		t.Errorf("Invalid MinuteNumber")
	}
	tmp2, err := eom.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	if len(tmp2) != 2 {
		t.Errorf("Invalid len")
	}
	if tmp[0] != constants.TYPE_MINUTE_NUM {
		t.Errorf("Invalid tmp")
	}
	if tmp[1] != 0x01 {
		t.Errorf("Invalid tmp")
	}

	eom = new(EndOfMinuteEntry)
	err = eom.UnmarshalBinary(tmp)
	if err != nil {
		t.Error(err)
	}
	if eom.Type() != constants.TYPE_MINUTE_NUM {
		t.Errorf("Invalid type")
	}
	if eom.MinuteNumber != 0x01 {
		t.Errorf("Invalid MinuteNumber")
	}
}

// TestEOMjsonMisc checks that various strings are returns properly from the object
func TestEOMjsonMisc(t *testing.T) {
	a := new(EndOfMinuteEntry)
	if a.String() != "    E:                    EndOfMinuteEntry --            Minute 0" {
		t.Error("Unexpected string:", a.String())
	}
	as, err := a.JSONString()
	if err != nil {
		t.Error(err)
	}
	if as != "{\"adminidtype\":0,\"minutenumber\":0}" {
		t.Error("Unexpected JSON string:", as)
	}
	ab, err := a.JSONByte()
	if err != nil {
		t.Error(err)
	}
	if string(ab) != "{\"adminidtype\":0,\"minutenumber\":0}" {
		t.Error("Unexpected JSON bytes:", string(ab))
	}
}
