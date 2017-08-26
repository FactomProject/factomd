// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package elections_test

import (
	"testing"

	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/messages"
	. "github.com/FactomProject/factomd/common/messages/elections"
	"github.com/FactomProject/factomd/common/primitives"
)

func TestUnmarshalfolunteerAudit_test(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(VolunteerAudit)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Error("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Error("Error is nil when it shouldn't be")
	}
}

func TestMarshalUnmarshalAck(t *testing.T) {
	test := func(va *VolunteerAudit, num string) {
		_, err := va.JSONString()
		if err != nil {
			t.Error(err)
		}
		hex, err := va.MarshalBinary()
		if err != nil {
			t.Error(err)
		}

		va2, err := messages.UnmarshalMessage(hex)
		if err != nil {
			t.Error(err)
		}
		_, err = va2.JSONString()
		if err != nil {
			t.Error(err)
		}

		if va2.Type() != constants.VOLUNTEERAUDIT {
			t.Error(num + " Invalid message type unmarshalled")
		}

		if va.IsSameAs(va2) == false {
			t.Error(num + " Acks are not the same")
			fmt.Println(va.String())
			fmt.Println(va2.String())
		}
	}
	va := new(VolunteerAudit)
	va.Minute = 5
	va.NName = "bob"
	va.DBHeight = 10
	va.ServerID = primitives.Sha([]byte("leader"))
	va.Weight = primitives.Sha([]byte("Weight"))
	va.ServerIdx = 3
	va.TS = primitives.NewTimestampNow()
	test(va, "1")
}
