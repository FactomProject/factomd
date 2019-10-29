package electionMsgTesting

import (
	"fmt"
	"reflect"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/messages/electionMsgs"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/elections"
	"github.com/FactomProject/factomd/electionsCore/election"
	priminterpreter "github.com/FactomProject/factomd/electionsCore/interpreter/primitives"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
)

var _ = fmt.Println

// Controller will be able to route messages to a set of nodes and control various
// communication patterns
type Controller struct {
	Elections        []*elections.Elections
	ElectionAdapters []*electionMsgs.ElectionAdapter
	Volunteers       []*electionMsgs.FedVoteVolunteerMsg

	feds []interfaces.IServer
	auds []interfaces.IServer

	Buffer        *MessageBuffer
	GlobalDisplay *election.Display

	//Router          *Router
	OutputsToRouter bool

	*priminterpreter.Primitives

	PrintingTrace bool

	History []string
}

// NewController creates all the elections and initial volunteer messages
func NewController(feds, auds int) *Controller {
	c := new(Controller)
	c.feds = make([]interfaces.IServer, feds)
	c.auds = make([]interfaces.IServer, auds)
	c.Elections = make([]*elections.Elections, len(c.feds))

	for i := range c.feds {
		e := new(elections.Elections)
		s := state.Server{}
		s.ChainID, _ = primitives.HexToHash("888888" + fmt.Sprintf("%058d", i))
		e.FedID = s.ChainID
		s.Name = fmt.Sprintf("Node%d", i)
		s.Online = true

		c.feds[i] = &s
		e.State = testHelper.CreateAndPopulateTestStateAndStartValidator()
		e.State.SetIdentityChainID(s.ChainID)
		c.Elections[i] = e
	}

	for i := range c.auds {
		s := state.Server{}
		s.ChainID, _ = primitives.HexToHash("888888" + fmt.Sprintf("%058d", i+10))
		s.Name = fmt.Sprintf("Node%d", i)
		s.Online = true
		c.auds[i] = &s
	}

	for _, e := range c.Elections {
		e.Federated = make([]interfaces.IServer, len(c.feds))
		for j := range c.feds {
			e.Federated[j] = c.feds[j]
		}
		e.State.(*state.State).ProcessLists.Get(1).FedServers = e.Federated
	}

	for _, e := range c.Elections {
		e.Audit = make([]interfaces.IServer, len(c.auds))
		for j := range c.auds {
			e.Audit[j] = c.auds[j]
		}
		e.State.(*state.State).ProcessLists.Get(1).AuditServers = e.Audit
	}

	c.ElectionAdapters = make([]*electionMsgs.ElectionAdapter, len(c.Elections))
	for i, e := range c.Elections {
		c.ElectionAdapters[i] = electionMsgs.NewElectionAdapter(e, e.State.GetIdentityChainID())
		c.ElectionAdapters[i].SimulatedElection.AddDisplay(nil)
	}

	c.GlobalDisplay = election.NewDisplay(c.ElectionAdapters[0].SimulatedElection, nil)
	c.GlobalDisplay.Identifier = "Global"

	for _, e := range c.ElectionAdapters {
		e.SimulatedElection.Display.Global = c.GlobalDisplay
	}

	c.Volunteers = make([]*electionMsgs.FedVoteVolunteerMsg, len(c.auds))

	for i, _ := range c.auds {
		c.Volunteers[i] = NewTestVolunteerMessage(c.Elections[0], 2, 0)
	}

	c.Buffer = NewMessageBuffer()

	return c
}

func NewTestVolunteerMessage(ele *elections.Elections, f, a int) *electionMsgs.FedVoteVolunteerMsg {
	v := new(electionMsgs.FedVoteVolunteerMsg)
	v.SigType = true
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

func (c *Controller) ElectionStatus(node int) string {
	if node == -1 {
		return c.GlobalDisplay.String()
	}
	return c.ElectionAdapters[node].SimulatedElection.Display.String()
}

func (c *Controller) RouteLeaderSetLevelMessage(from []int, level int, to []int) bool {
	for _, f := range from {
		if !c.RouteLeaderLevelMessage(f, level, to) {
			return false
		}
	}
	return true
}

func (c *Controller) RouteLeaderLevelMessage(from int, level int, to []int) bool {
	msg := c.Buffer.RetrieveLeaderLevelMessageByLevel(c.indexToFedID(from), level)
	if msg == nil {
		return false
	}
	c.RouteMessage(msg, to)
	return true
}

func (c *Controller) RouteLeaderSetVoteMessage(from []int, vol int, to []int) bool {
	for _, f := range from {
		if !c.RouteLeaderVoteMessage(f, vol, to) {
			return false
		}
	}
	return true
}

func (c *Controller) RouteLeaderVoteMessage(from int, vol int, to []int) bool {
	msg := c.Buffer.RetrieveLeaderVoteMessage(c.indexToFedID(from), c.indexToAudID(vol))
	if msg == nil {
		return false
	}
	c.RouteMessage(msg, to)
	return true
}

func (c *Controller) RouteVolunteerMessage(vol int, nodes []int) {
	c.RouteMessage(c.Volunteers[vol], nodes)
}

// RouteMessage will execute message on all nodes given and add the messages returned to the buffer
func (c *Controller) RouteMessage(msg interfaces.IMsg, nodes []int) {
	for _, n := range nodes {
		c.routeSingleNode(msg, n)
	}
}

func (c *Controller) routeSingleNode(msg interfaces.IMsg, node int) {
	resp := c.ElectionAdapters[node].Execute(msg)
	c.Buffer.Add(resp)
	//
	//if c.PrintingTrace {
	//	str := fmt.Sprintf("L%d: ", node)
	//	str += fmt.Sprintf(" Consumed(%s)", c.Elections[node].Display.FormatMessage(msg))
	//	str += fmt.Sprintf(" Generated(%s) StateChange: %t", c.Elections[node].Display.FormatMessage(resp))
	//	fmt.Println(str)
	//}
}

// indexToAudID will take the human legible "Audit 1" and get the correct identity.
func (c *Controller) indexToAudID(index int) (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("Controller.indexToAudID() saw an interface that was nil")
		}
	}()

	// TODO: Actually implement some logic if this changes
	return c.auds[index].GetChainID()

}

func (c *Controller) fedIDtoIndex(id interfaces.IHash) int {
	for i, f := range c.feds {
		if f.GetChainID().IsSameAs(id) {
			return i
		}
	}
	return -1
}

// indexToFedID will take the human legible "Leader 1" and get the correct identity
func (c *Controller) indexToFedID(index int) (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("Controller.indexToFedID() saw an interface that was nil")
		}
	}()

	// TODO: Actually implement some logic if this changes
	return c.feds[index].GetChainID()
}
