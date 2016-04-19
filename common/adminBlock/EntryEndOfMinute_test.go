package adminBlock_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
)

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
