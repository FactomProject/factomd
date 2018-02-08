package election

import (
	"github.com/FactomProject/electiontesting/imessage"
	"github.com/FactomProject/electiontesting/messages"
	. "github.com/FactomProject/electiontesting/primitives"
	"github.com/FactomProject/electiontesting/round"
)

const (
	_ int = iota
	ElectionState_Working
	ElectionState_Publishing
)

// TODO: Add Dbsig kill code

type Election struct {
	Rounds []*round.Round
	// The key is the volunter for the election msg. The index is the round integer
	PublishingRound int
	PublishMsg      *messages.PublishMessage

	// In/Out chan

	// Authority Information
	AuthSet

	// vm, min, height
	ProcessListLocation

	Self  Identity
	State int
}

func NewElection(a AuthSet, loc ProcessListLocation) *Election {
	e := new(Election)
	e.AuthSet = a
	e.ProcessListLocation = loc
	e.Rounds = make([]*round.Round, len(a.StatusArray))
	e.State = ElectionState_Working

	return e
}

func (e *Election) ExecuteMsg(msg imessage.IMessage) []imessage.IMessage {
	r := e.GetRoundFromMsg(msg)

	// No matter the state, we check the publishing
	if p, ok := msg.(messages.PublishMessage); ok {
		// Set state if this is new
		if e.PublishMsg == nil {
			e.State = ElectionState_Publishing
		}

		// Is this publish better than what we have?
		if e.PublishMsg == nil || (e.PublishMsg != nil && e.PublishingRound > r) {
			// No competing, set our publishing round
			e.PublishingRound = r
			e.PublishMsg = &p
			return imessage.MakeMessageArray(p)
		}

		// We have something better (should not be nil)
		return imessage.MakeMessageArray(*e.PublishMsg)
	}

	// We should filter all messages if we are publishing
	if e.PublishMsg != nil {
		if r >= e.PublishingRound {
			return imessage.MakeMessageArray(*e.PublishMsg)
		}
	}

	var response []imessage.IMessage
	// Guarented any messaged here is lower than any we have publishing.
	switch e.State {
	case ElectionState_Working:
		// Default to execute and look for publish
	case ElectionState_Publishing:
		// Lower round, let it through but add our publish message to the response
		response = append(response, *e.PublishMsg)
	default:
		panic("Election does not have a valid state")
	}

	response = imessage.MakeMessageArrayFromArray(response, e.executeWorking(msg, r))
	if pub := ContainsPublish(response); pub != nil {
		e.setPublishing(*pub, r)
	}
	return response
}

func (e *Election) setPublishing(msg messages.PublishMessage, r int) {
	e.PublishingRound = r
	e.PublishMsg = &msg
	e.State = ElectionState_Publishing
}

func (e *Election) executeWorking(msg imessage.IMessage, r int) []imessage.IMessage {
	switch msg.(type) {
	case messages.FaultMsg:
		// Should have a volunteer message
	default:
		// This means it is an election msg.
		vol := messages.GetVolunteerMsg(msg)
		if vol == nil {
			panic("All messages should have a volunteer msg in them")
		}

		r := e.GetRound(vol.Signer)
		// Ensure round exists
		if r > len(e.Rounds)-1 {
			panic("This should never happen. The round is outside our round possibilities")
		}

		if e.Rounds[r] == nil {
			e.Rounds[r] = round.NewRound(e.AuthSet, e.Self, *vol, e.ProcessListLocation)
		}

		return e.executeInRound(msg, r)
	}

	panic("Should not reach this in executeWorking")
	return nil
}

// executeInRound is guarenteed the election round exists
func (e *Election) executeInRound(msg imessage.IMessage, r int) []imessage.IMessage {
	return e.Rounds[r].Execute(msg)
}

func (e *Election) GetRound(vol Identity) int {
	// TODO: Make a better round determinate
	// Currently just their ID mod length of authority set
	i := int(vol)
	round := i % len(e.StatusArray)
	return round
}

func (e *Election) GetRoundFromMsg(msg imessage.IMessage) int {
	switch msg.(type) {
	case messages.FaultMsg:
		return 1
	default:
		// This means it is an election msg.
		vol := messages.GetVolunteerMsg(msg)
		if vol == nil {
			panic("All messages should have a volunteer msg in them")
		}

		r := e.GetRound(vol.Signer)
		return r
	}
}
