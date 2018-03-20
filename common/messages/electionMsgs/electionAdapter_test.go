package electionMsgs_test

import (
	"fmt"
	"testing"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	. "github.com/FactomProject/factomd/common/messages/electionMsgs"
	"github.com/FactomProject/factomd/common/messages/electionMsgs/electionMsgTesting"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/elections"
	"github.com/FactomProject/factomd/state"
	. "github.com/FactomProject/factomd/testHelper"
)

var _ = CreateAndPopulateTestDatabaseOverlay

func TestElectionAdapterMajority(t *testing.T) {
	c := electionMsgTesting.NewController(3, 3)
	all := []int{0, 1, 2, 1, 0}
	c.RouteVolunteerMessage(0, all)
	c.RouteLeaderSetVoteMessage(all, 0, all)
	c.RouteLeaderVoteMessage(1, 0, []int{0})

	c.RouteLeaderSetLevelMessage(all, 0, all)
	c.RouteLeaderSetLevelMessage(all, 1, all)
	c.RouteLeaderSetLevelMessage(all, 2, all)
	c.RouteLeaderSetLevelMessage(all, 3, all)
	c.RouteLeaderSetLevelMessage(all, 4, all)

	if !c.ElectionAdapters[0].SimulatedElection.Committed {
		t.Errorf("Should be committed , %v", c.GlobalDisplay.String())
	}
}

func TestElectionAuditOrder(t *testing.T) {
	c := electionMsgTesting.NewController(6, 6)
	fmt.Printf("%v\n", c.Elections[0].Audit[1].GetChainID())
	newo := elections.Order(c.Elections[0].Audit, 10, 10, 10)
	fmt.Printf("%v\n", newo)
	fmt.Printf("%v\n", elections.MaxIdx(newo))

}

func TestSimpleSigning(t *testing.T) {
	s := CreateAndPopulateTestState()
	e := NewTestElection()
	v1 := NewTestVolunteerMessage(e, 2, 0)
	err := v1.Sign(s)
	if err != nil {
		t.Error(err)
	}

	data, _ := v1.MarshalForSignature()
	if !v1.Signature.Verify(data) {
		t.Error("Sig did not verify")
	}
}

func TestElectionAdapterSimple(t *testing.T) {
	e := NewTestElection()

	a := NewElectionAdapter(e)
	v1 := NewTestVolunteerMessage(e, 2, 0)
	resp := a.Execute(v1)
	// Verify resp was a vote
	if msg, ok := resp.(*FedVoteProposalMsg); ok {
		if !msg.Signer.IsSameAs(e.FedID) {
			t.Errorf("Message not signed by self")
		}
	} else {
		t.Errorf("Expected a proposal, but did not get one")
	}
}

func NewTestElection() *elections.Elections {
	e := new(elections.Elections)
	e.FedID, _ = primitives.NewShaHashFromStr("888888f0b7e308974afc34b2c7f703f25ed2699cb05f818e84e8745644896c55")
	e.Federated = make([]interfaces.IServer, 3)
	e.Audit = make([]interfaces.IServer, 3)

	for i := range e.Federated {
		s := state.Server{}
		s.ChainID, _ = primitives.HexToHash("888888" + fmt.Sprintf("%058d", i))
		if i == 0 {
			s.ChainID, _ = primitives.NewShaHashFromStr("888888f0b7e308974afc34b2c7f703f25ed2699cb05f818e84e8745644896c55")
		}
		s.Name = fmt.Sprintf("Node%d", i)
		s.Online = true
		e.Federated[i] = &s
	}

	for i := range e.Audit {
		s := state.Server{}
		s.ChainID, _ = primitives.HexToHash("888888" + fmt.Sprintf("%058d", i))
		if i == 0 {
			s.ChainID, _ = primitives.NewShaHashFromStr("888888f0b7e308974afc34b2c7f703f25ed2699cb05f818e84e8745644896c55")
		}
		s.Name = fmt.Sprintf("Node%d", i)
		s.Online = true
		e.Audit[i] = &s
	}

	// Need a majority, so need 2 elections

	return e
}

func NewTestVolunteerMessage(ele *elections.Elections, f, a int) *FedVoteVolunteerMsg {
	v := new(FedVoteVolunteerMsg)
	v.EOM = true
	v.Name = "Leader 2"
	v.FedIdx = uint32(f)
	v.FedID = ele.Federated[f].GetChainID()

	v.ServerIdx = uint32(a)
	v.ServerID = ele.Audit[a].GetChainID()
	v.ServerName = "Volunteer 0"

	v.Missing = new(messages.EOM)
	eom := v.Missing.(*messages.EOM)
	eom.ChainID = primitives.NewHash([]byte("id"))
	eom.LeaderChainID = primitives.NewHash([]byte("leader"))
	eom.Timestamp = primitives.NewTimestampNow()

	v.Ack = new(messages.Ack)
	ack := v.Ack.(*messages.Ack)
	ack.Timestamp = primitives.NewTimestampNow()
	ack.LeaderChainID = primitives.NewHash([]byte("leader"))
	ack.MessageHash = primitives.NewHash([]byte("msg"))
	ack.SerialHash = primitives.NewHash([]byte("serial"))
	v.TS = primitives.NewTimestampNow()
	v.InitFields(ele)

	return v
}
