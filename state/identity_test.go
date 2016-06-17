package state_test

import (
	"testing"

	/*
		"github.com/FactomProject/factomd/common/interfaces"
		"github.com/FactomProject/factomd/common/messages"
	*/

	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/state"
	//"github.com/FactomProject/factomd/testHelper"
)

func TestAppendExtIDs(t *testing.T) {
	ids := [][]byte{
		[]byte{0x01, 0x02},
		[]byte{0x03, 0x04},
		[]byte{0x05, 0x06},
		[]byte{0x07, 0x08},
		[]byte{0x09, 0x0a},
	}

	appended := []byte{
		0x03, 0x04,
		0x05, 0x06,
		0x07, 0x08,
		0x09, 0x0a,
	}

	resp, err := AppendExtIDs(ids, 1, 4)
	if err != nil {
		t.Errorf("%v", err)
	}
	if primitives.AreBytesEqual(resp, appended) == false {
		t.Errorf("AppendExtIDs are not equal - %x vs %x", resp, appended)
	}

	resp, err = AppendExtIDs(ids, 1, 5)
	if err == nil {
		t.Error("Err is nit when it should not be")
	}
	if resp != nil {
		t.Error("Resp is not nil when it should be")
	}
}
