// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package electionMsgs_test

import (
	"testing"

	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/messages/electionMsgs"
	"github.com/FactomProject/factomd/common/messages/msgsupport"
	"github.com/FactomProject/factomd/common/primitives"
)

func TestUnmarshalfolunteerSyncMsg_test(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(SyncMsg)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Error("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Error("Error is nil when it shouldn't be")
	}
}

func TestMarshalUnmarshalSyncMsg(t *testing.T) {
	test := func(sm *SyncMsg, num string) {
		_, err := sm.JSONString()
		if err != nil {
			t.Error(err)
		}
		hex, err := sm.MarshalBinary()
		if err != nil {
			t.Error(err)
		}

		va2, err := msgsupport.UnmarshalMessage(hex)
		if err != nil {
			t.Error(err)
		}
		_, err = va2.JSONString()
		if err != nil {
			t.Error(err)
		}

		if va2.Type() != constants.SYNC_MSG {
			t.Error(num + " Invalid message type unmarshalled")
		}

		if sm.IsSameAs(va2) == false {
			t.Error(num + " Acks are not the same")
			fmt.Println(sm.String())
			fmt.Println(va2.String())
		}
	}
	sm := new(SyncMsg)
	sm.Minute = 5
	sm.Name = "bob"
	sm.EOM = true
	sm.DBHeight = 10
	sm.ServerID = primitives.Sha([]byte("leader"))
	sm.Weight = primitives.Sha([]byte("Weight"))
	sm.ServerIdx = 3
	sm.TS = primitives.NewTimestampNow()
	test(sm, "1")
}
