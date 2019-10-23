package electionMsgs_test

import (
	"testing"

	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	. "github.com/FactomProject/factomd/common/messages/electionMsgs"
	"github.com/FactomProject/factomd/common/messages/msgsupport"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/testHelper"
)

func TestMarshalUnmarshalFedVoteLevel(t *testing.T) {
	messages.General = new(msgsupport.GeneralFactory)
	primitives.General = messages.General
	s := testHelper.CreateAndPopulateTestStateAndStartValidator()

	test := func(va *FedVoteLevelMsg, num string) {
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

		if va2.Type() != constants.VOLUNTEERLEVELVOTE {
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

	var msgs []interfaces.IMsg

	// Have volunteer
	for i := 0; i < 20; i++ {
		v := randomVol(s)
		v.Sign(s)
		l := NewFedVoteLevelMessage(primitives.RandomHash(), *v)
		l.TS = primitives.NewTimestampNow()
		if i%2 == 0 {
			l.Committed = true
		}
		l.Justification = msgs
		err := l.Sign(s)
		if err != nil {
			t.Error(err)
		}
		test(l, fmt.Sprintf("%d", i))
		l.Justification = []interfaces.IMsg{}
		l.Sign(s)
		msgs = append(msgs, l)
	}
}

func randomVol2() *FedVoteVolunteerMsg {
	va := new(FedVoteVolunteerMsg)
	va.Minute = 5
	va.Name = "bob"
	va.DBHeight = 10
	va.ServerID = primitives.RandomHash() // primitives.Sha([]byte("leader"))
	//va.Weight = primitives.Sha([]byte("Weight"))
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

	return va
}
