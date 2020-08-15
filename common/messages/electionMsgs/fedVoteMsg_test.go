// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package electionMsgs_test

import (
	"testing"

	"github.com/PaulSnow/factom2d/common/messages"
	. "github.com/PaulSnow/factom2d/common/messages/electionMsgs"
	"github.com/PaulSnow/factom2d/common/messages/msgsupport"
	"github.com/PaulSnow/factom2d/common/primitives"
)

func TestUnmarshalfolunteerAudit_test(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	messages.General = new(msgsupport.GeneralFactory)
	primitives.General = messages.General

	a := new(FedVoteMsg)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Error("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Error("Error is nil when it shouldn't be")
	}
}
