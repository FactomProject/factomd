// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package electionMsgs_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/messages/electionMsgs"
	"github.com/FactomProject/factomd/common/messages/msgsupport"
	"github.com/FactomProject/factomd/common/primitives"
)

func TestUnmarshalVolunteerSyncMsg_test(t *testing.T) {
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

		_, err = msgsupport.UnmarshalMessage(hex)
		// Expect an error, as sync messages are local only,
		// 		and cannot be unmarshalled.
		if err == nil {
			t.Error(err)
		}
	}
	sm := new(SyncMsg)
	sm.Minute = 5
	sm.Name = "bob"
	sm.SigType = true
	sm.DBHeight = 10
	sm.ServerID = primitives.Sha([]byte("leader"))
	sm.Weight = primitives.Sha([]byte("Weight"))
	sm.ServerIdx = 3
	sm.TS = primitives.NewTimestampNow()
	test(sm, "1")
}
