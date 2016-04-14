package adminBlock

import (
	"testing"
)

func TestEOMMisc(t *testing.T) {
	eom := new(EndOfMinuteEntry)
	if eom.IsInterpretable() != true {
		t.Fail()
	}
	eom.EOM_Type = 1
	if eom.Interpret() != "End of Minute 1" {
		t.Fail()
	}
	eom.EntryType = 3
	if eom.Type() != 3 {
		t.Fail()
	}
}

func TestEOMMarshalUnmarshal(t *testing.T) {
	tmp := []byte{0x01, 0x02, 0x03}
	eom := new(EndOfMinuteEntry)
	rest, err := eom.UnmarshalBinaryData(tmp)
	if len(rest) != 1 {
		t.Fail()
	}
	if rest[0] != 0x03 {
		t.Fail()
	}
	if eom.EntryType != 0x01 {
		t.Fail()
	}
	if eom.EOM_Type != 0x02 {
		t.Fail()
	}
	tmp2, err := eom.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	if len(tmp2) != 2 {
		t.Fail()
	}
	if tmp[0] != 0x01 {
		t.Fail()
	}
	if tmp[1] != 0x02 {
		t.Fail()
	}

	eom = new(EndOfMinuteEntry)
	err = eom.UnmarshalBinary(tmp)
	if err != nil {
		t.Error(err)
	}
	if eom.EntryType != 0x01 {
		t.Fail()
	}
	if eom.EOM_Type != 0x02 {
		t.Fail()
	}
}
