// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"testing"

	"github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/messages/msgsupport"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/testHelper"
)

func TestUnmarshalNilDBStateMissing(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(DBStateMissing)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestMarshalUnmarshalDBStateMissing(t *testing.T) {
	msg := newDBStateMissing()

	hex, err := msg.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled - %x", hex)

	msg2, err := msgsupport.UnmarshalMessage(hex)
	if err != nil {
		t.Error(err)
	}
	str := msg2.String()
	t.Logf("str - %v", str)

	if msg2.Type() != constants.DBSTATE_MISSING_MSG {
		t.Error("Invalid message type unmarshalled")
	}

	hex2, err := msg2.(*DBStateMissing).MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	if len(hex) != len(hex2) {
		t.Error("Hexes aren't of identical length")
	}
	for i := range hex {
		if hex[i] != hex2[i] {
			t.Error("Hexes do not match")
		}
	}

	if msg.IsSameAs(msg2.(*DBStateMissing)) != true {
		t.Errorf("DBStateMissing messages are not identical")
	}
}

func newDBStateMissing() *DBStateMissing {
	msg := new(DBStateMissing)
	msg.Timestamp = primitives.NewTimestampNow()

	msg.DBHeightStart = 0x01234567
	msg.DBHeightEnd = 0x89012345

	return msg
}

func Testlimits(t *testing.T) {

	for s := uint32(0); s < 100; s++ {
		for e := s; e < s+1000; e++ {
			for in := 0; in < 1000; in++ {
				ns, ne := NewEnd(in, s, e)

				if s != ns {
					t.Errorf(" Failed with e %d s %d and in %d", e, s, in)
				}

				if ne > e {
					t.Errorf(" Failed with e %d s %d and in %d", e, s, in)
				}

				if in > 500 && ne != 0 {
					t.Errorf(" Failed with e %d s %d and in %d", e, s, in)
				}

				if in <= 500 && in > 200 && ne-ns > 50 {
					t.Errorf(" Failed with e %d s %d and in %d", e, s, in)
				}

				if in <= 200 && ne-ns > 200 {
					t.Errorf(" Failed with e %d s %d and in %d", e, s, in)
				}
			}
		}
	}
}

func TestNewDBStateMissing(t *testing.T) {

	s := testHelper.CreateEmptyTestState()

	ndbs := NewDBStateMissing(s, 1, 100)

	//fmt.Printf("asdf: %v\n", ndbs)
	if ndbs.Validate(s) != 1 {
		t.Errorf("NewDBStateMissing is marked invalid when it should be valid")
	}

	ndbs2 := NewDBStateMissing(s, 100, 1)
	if ndbs2.Validate(s) != -1 {
		t.Errorf("NewDBStateMissing is marked valid when it should be invalid")
	}

}
