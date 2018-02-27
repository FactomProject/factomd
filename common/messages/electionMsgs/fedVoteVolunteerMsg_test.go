package electionMsgs_test

import (
	"testing"

	"github.com/FactomProject/factomd/common/messages"
	. "github.com/FactomProject/factomd/common/messages/electionMsgs"
	"github.com/FactomProject/factomd/common/messages/msgsupport"
	"github.com/FactomProject/factomd/common/primitives"
)

func TestUnmarshalFedVoteVolunteerEmpty(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	messages.General = new(msgsupport.GeneralFactory)
	primitives.General = messages.General

	a := new(FedVoteVolunteerMsg)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Error("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Error("Error is nil when it shouldn't be")
	}
}
