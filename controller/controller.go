package controller

import (
	"fmt"

	"github.com/FactomProject/electiontesting/election"
	"github.com/FactomProject/electiontesting/imessage"
	"github.com/FactomProject/electiontesting/messages"
	"github.com/FactomProject/electiontesting/primitives"
	"github.com/FactomProject/electiontesting/testhelper"
)

var _ = fmt.Println

// Controller will be able to route messages to a set of nodes and control various
// communication patterns
type Controller struct {
	AuthSet *testhelper.AuthSetHelper

	Elections  []*election.Election
	Volunteers []*messages.VolunteerMessage

	feds []primitives.Identity
	auds []primitives.Identity

	// Can be all 0s for now
	primitives.ProcessListLocation

	Buffer        *MessageBuffer
	GlobalDisplay *election.Display
}

// NewController creates all the elections and initial volunteer messages
func NewController(feds, auds int) *Controller {
	c := new(Controller)
	c.AuthSet = testhelper.NewAuthSetHelper(feds, auds)
	fedlist := c.AuthSet.GetFeds()
	c.Elections = make([]*election.Election, len(fedlist))

	for i, f := range fedlist {
		c.Elections[i] = c.newElection(f)
		if i == 0 {
			c.GlobalDisplay = election.NewDisplay(c.Elections[0], nil)
			c.GlobalDisplay.Identifier = "Global"
		}
		c.Elections[i].AddDisplay(c.GlobalDisplay)
	}

	audlist := c.AuthSet.GetAuds()
	c.Volunteers = make([]*messages.VolunteerMessage, len(audlist))

	for i, a := range audlist {
		c.Volunteers[i] = c.newVolunteer(a)
	}

	c.Buffer = NewMessageBuffer()
	c.feds = fedlist
	c.auds = audlist

	return c
}

func (c *Controller) ElectionStatus(node int) string {
	if node == -1 {
		return c.GlobalDisplay.String()
	}
	return c.Elections[node].Display.String()
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
func (c *Controller) RouteMessage(msg imessage.IMessage, nodes []int) {
	for _, n := range nodes {
		c.routeSingleNode(msg, n)
	}
}

func (c *Controller) routeSingleNode(msg imessage.IMessage, node int) {
	resp := c.Elections[node].Execute(msg)
	c.Buffer.Add(resp)
}

// indexToAudID will take the human legible "Audit 1" and get the correct identity.
func (c *Controller) indexToAudID(index int) primitives.Identity {
	// TODO: Actually implement some logic if this changes
	return c.auds[index]

}

// indexToFedID will take the human legible "Leader 1" and get the correct identity
func (c *Controller) indexToFedID(index int) primitives.Identity {
	// TODO: Actually implement some logic if this changes
	return c.feds[index]

}

func (c *Controller) newVolunteer(id primitives.Identity) *messages.VolunteerMessage {
	vol := messages.NewVolunteerMessage(messages.NewEomMessage(id, c.ProcessListLocation), id)
	return &vol
}

func (c *Controller) newElection(id primitives.Identity) *election.Election {
	return election.NewElection(id, c.AuthSet.GetAuthSet(), c.ProcessListLocation)
}
