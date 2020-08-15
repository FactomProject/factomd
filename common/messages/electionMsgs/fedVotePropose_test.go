package electionMsgs_test

import (
	"testing"

	"fmt"

	"github.com/PaulSnow/factom2d/common/constants"
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/messages"
	. "github.com/PaulSnow/factom2d/common/messages/electionMsgs"
	"github.com/PaulSnow/factom2d/common/messages/msgsupport"
	"github.com/PaulSnow/factom2d/common/primitives"
	"github.com/PaulSnow/factom2d/testHelper"
)

func init() {
	primitives.General = new(msgsupport.GeneralFactory)
}

func TestMarshalUnmarshalFedVoteProposal(t *testing.T) {
	test := func(va *FedVoteProposalMsg, num string) {
		vas, err := va.JSONString()
		if err != nil {
			t.Error(err)
		}
		var _ = vas
		hex, err := va.MarshalBinary()
		if err != nil {
			t.Error(err)
		}

		va2, err := msgsupport.UnmarshalMessage(hex)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}

		if va2.Type() != constants.VOLUNTEERPROPOSAL {
			t.Error(num + " Invalid message type unmarshalled")
		}

		if va.IsSameAs(va2) == false {
			t.Error(num + " Acks are not the same")
		}

		//va2s, err := va2.JSONString()
		//if va2s != vas {
		//	t.Errorf("Messages are not the json when json marshalled")
		//}
		if err != nil {
			t.Error(err)
		}
	}

	s := testHelper.CreateAndPopulateTestStateAndStartValidator()
	// Have volunteer
	for i := 0; i < 20; i++ {
		p := NewFedProposalMsg(primitives.RandomHash(), *randomVol(s))
		p.Sign(s)
		test(p, fmt.Sprintf("%d", i))
	}
}

func randomVol(s interfaces.IState) *FedVoteVolunteerMsg {
	va := new(FedVoteVolunteerMsg)
	va.Minute = 5
	va.Name = "bob"
	va.DBHeight = 10
	va.ServerID = primitives.RandomHash() // primitives.Sha([]byte("leader"))
	va.Weight = primitives.Sha([]byte("Weight"))
	va.ServerIdx = 3
	va.Missing = new(messages.EOM)
	eom := va.Missing.(*messages.EOM)
	eom.ChainID = primitives.RandomHash()       //primitives.NewHash([]byte("id"))
	eom.LeaderChainID = primitives.RandomHash() //primitives.NewHash([]byte("leader"))
	eom.Timestamp = primitives.NewTimestampNow()

	va.Ack = new(messages.Ack)
	ack := va.Ack.(*messages.Ack)
	ack.Timestamp = primitives.NewTimestampNow()
	ack.LeaderChainID = primitives.RandomHash() //primitives.NewHash([]byte("leader"))
	ack.MessageHash = primitives.RandomHash()   //primitives.NewHash([]byte("msg"))
	ack.SerialHash = primitives.RandomHash()    //primitives.NewHash([]byte("serial"))
	va.TS = primitives.NewTimestampNow()

	va.FedID = primitives.RandomHash()
	va.Sign(s)

	return va
}
