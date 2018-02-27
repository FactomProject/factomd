package electionMsgs_test

import (
	"fmt"
	"testing"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	. "github.com/FactomProject/factomd/common/messages/electionMsgs"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/elections"
	"github.com/FactomProject/factomd/state"
	. "github.com/FactomProject/factomd/testHelper"
)

var _ = CreateAndPopulateTestDatabaseOverlay

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

	return e
}

/*
type FedVoteVolunteerMsg struct {
	FedVoteMsg
	// Volunteer fields
	EOM        bool             // True if an EOM, false if a DBSig
	Name       string           // Server name
	FedIdx     uint32           // Server faulting
	FedID      interfaces.IHash // Server faulting
	ServerIdx  uint32           // Index of Server replacing
	ServerID   interfaces.IHash // Volunteer Server ChainID
	ServerName string           // Volunteer Name
	Missing    interfaces.IMsg  // The Missing DBSig or EOM
	Ack        interfaces.IMsg  // The acknowledgement for the missing message

	messageHash interfaces.IHash
}

*/

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
