package electionMsgs_test

import (
	"testing"

	"fmt"

	"github.com/FactomProject/factomd/common/messages"
	. "github.com/FactomProject/factomd/common/messages/electionMsgs"
	"github.com/FactomProject/factomd/common/messages/electionMsgs/electionMsgTesting"
	"github.com/FactomProject/factomd/common/messages/msgsupport"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/testHelper"
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

func TestUnmarshalFedVoteVolunteerDBSig(t *testing.T) {

	messages.General = new(msgsupport.GeneralFactory)
	primitives.General = messages.General

	con := electionMsgTesting.NewController(3, 3)
	vol := con.Volunteers[0]

	var err error
	s := testHelper.CreateAndPopulateTestState()
	fmt.Println(s)
	vol.Missing, vol.Ack = s.CreateDBSig(1, 0)
	if err != nil {
		t.Error(err)
	}

	vol.Sign(s)
	data, err := vol.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	v2 := new(FedVoteVolunteerMsg)
	left, err := v2.UnmarshalBinaryData(data)
	if err != nil {
		t.Error(err)
	}
	if len(left) > 0 {
		t.Error("Left over bytes")
	}

}
