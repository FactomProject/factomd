package entryCreditBlock_test

import (
	"testing"

	. "github.com/PaulSnow/factom2d/common/entryCreditBlock"
)

// TestUnmarshalNilMinuteNumber checks that unmarshalling nil or the empty interface results in the appropriate errors
func TestUnmarshalNilMinuteNumber(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(MinuteNumber)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

// TestMinuteNumberMarshalUnmarshal checks that a new minute number can be marshalled and unmarshalled, and residual
// data is uncorrupted
func TestMinuteNumberMarshalUnmarshal(t *testing.T) {
	mn := NewMinuteNumber(5)
	b, err := mn.MarshalBinary()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(b) != 1 {
		t.Error("Invalid byte length")
	}
	if b[0] != 5 {
		t.Error("Invalid byte")
	}
	mn2 := NewMinuteNumber(0)
	err = mn2.UnmarshalBinary(b)
	if err != nil {
		t.Error(err)
	}
	if mn2.Number != mn.Number {
		t.Error("Invalid data unmarshalled")
	}

	mn3 := new(MinuteNumber)
	remainder, err := mn3.UnmarshalBinaryData([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
	if err != nil {
		t.Error(err)
	}
	if mn3.Number != 1 {
		t.Error("Invalid data unmarshalled")
	}
	if len(remainder) != 4 {
		t.Error("Invalid byte length")
	}
	if remainder[0] != 0x02 || remainder[1] != 0x03 || remainder[2] != 0x04 || remainder[3] != 0x05 {
		t.Error("Wrong remainder returned")
	}
}

// TestInvalidMinuteNumberUnmarshal checks that unmarshalling nil and the empty interface results in errors
func TestInvalidMinuteNumberUnmarshal(t *testing.T) {
	mn := new(MinuteNumber)
	_, err := mn.UnmarshalBinaryData(nil)
	if err == nil {
		t.Error("We expected errors but we didn't get any")
	}
	mn = new(MinuteNumber)
	_, err = mn.UnmarshalBinaryData([]byte{})
	if err == nil {
		t.Error("We expected errors but we didn't get any")
	}

	mn = new(MinuteNumber)
	err = mn.UnmarshalBinary(nil)
	if err == nil {
		t.Error("We expected errors but we didn't get any")
	}
	mn = new(MinuteNumber)
	err = mn.UnmarshalBinary([]byte{})
	if err == nil {
		t.Error("We expected errors but we didn't get any")
	}
}

// TestMinuteNumberMisc checks the various member functions return their expected values.
func TestMinuteNumberMisc(t *testing.T) {
	si := new(MinuteNumber)
	si.Number = 4
	if si.IsInterpretable() == false {
		t.Fail()
	}
	if si.Interpret() != "MinuteNumber 4" {
		t.Fail()
	}
	if si.Hash().String() != "e52d9c508c502347344d8c07ad91cbd6068afc75ff6292f062a09ca381c89e71" {
		t.Fail()
	}
}
