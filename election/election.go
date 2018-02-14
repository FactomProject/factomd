package election

import (
	"fmt"

	"github.com/FactomProject/electiontesting/imessage"
	"github.com/FactomProject/electiontesting/messages"
	. "github.com/FactomProject/electiontesting/primitives"
	"github.com/FactomProject/electiontesting/volunteercontrol"
)

var _ = fmt.Println

type Election struct {
	// Level 0 volunteer votes map[vol]map[leader]msg
	VolunteerVotes map[Identity]map[Identity]*messages.VoteMessage

	// Indexed by volunteer
	VolunteerControls map[Identity]*volunteercontrol.VolunteerControl

	CommitmentIndicator *DiamondShop

	CurrentLevel int
	CurrentVote  messages.LeaderLevelMessage
	Self         Identity
	AuthSet
	ProcessListLocation

	Display *Display
}

func NewElection(self Identity, authset AuthSet, loc ProcessListLocation) *Election {
	e := new(Election)
	e.VolunteerVotes = make(map[Identity]map[Identity]*messages.VoteMessage)
	e.VolunteerControls = make(map[Identity]*volunteercontrol.VolunteerControl)
	e.Self = self
	e.AuthSet = authset
	// Majority level starts at 1
	e.CurrentLevel = 1
	e.CurrentVote.Rank = -1
	e.CurrentVote.VolunteerPriority = -1

	// Used to determine volunteer priority
	e.ProcessListLocation = loc

	e.CommitmentIndicator = NewDiamondShop(e.AuthSet)
	return e
}

// AddDisplay takes a global tracker. Send nil if you don't care about
// a global state
func (e *Election) AddDisplay(global *Display) *Display {
	e.Display = NewDisplay(e, global)
	return e.Display
}

func (e *Election) Execute(msg imessage.IMessage) imessage.IMessage {
	e.executeDisplay(msg)

	switch msg.(type) {
	case *messages.LeaderLevelMessage:
		return e.executeLeaderLevelMessage(msg.(*messages.LeaderLevelMessage))
	case *messages.VolunteerMessage:
		vol := msg.(*messages.VolunteerMessage)
		vote := messages.NewVoteMessage(*vol, e.Self)
		return &vote
	case *messages.VoteMessage:
		// Colecting these allows us to issue out 0.#
		vote := msg.(*messages.VoteMessage)
		vol := vote.Volunteer.Signer

		if e.VolunteerVotes[vol] == nil {
			e.VolunteerVotes[vol] = make(map[Identity]*messages.VoteMessage)
		}
		e.VolunteerVotes[vol][vote.Signer] = vote
		if len(e.VolunteerVotes[vote.Volunteer.Signer]) > e.Majority() {
			// We have a majority of level 0 votes and can issue a rank 0 LeaderLevel Message

			// No current vote, so send!
			if e.CurrentVote.Rank == -1 {
				goto SendRank0
			}

			// If we have a rank 1+, we will not issue a rank 0
			if e.CurrentVote.Rank > 0 {
				return &e.CurrentVote // forward our answer again
			}

			// If we have a rank 0, and higher priority volunteer (or same, don't vote again)
			if e.CurrentVote.VolunteerPriority >= e.getVolunteerPriority(vol) {
				return &e.CurrentVote // forward our answer again
			}

		SendRank0:
			ll := messages.NewLeaderLevelMessage(e.Self, 0, e.CurrentLevel, vote.Volunteer)
			e.CurrentLevel++
			e.CurrentVote = ll
			e.executeDisplay(&ll)
			return e.Execute(&ll)
		}
	}
	return nil
}

func (e *Election) getVolunteerPriority(vol Identity) int {
	return e.GetVolunteerPriority(vol, e.ProcessListLocation)
}

func (e *Election) executeDisplay(msg imessage.IMessage) {
	if e.Display != nil {
		e.Display.Execute(msg)
	}
}

func (e *Election) executeLeaderLevelMessage(msg *messages.LeaderLevelMessage) imessage.IMessage {
	if e.VolunteerControls[msg.VolunteerMessage.Signer] == nil {
		e.VolunteerControls[msg.VolunteerMessage.Signer] = volunteercontrol.NewVolunteerControl(e.Self, e.AuthSet)
	}

	if msg.Level <= 0 {
		panic("Whuy")
	}

	// If commit is true, then we are done. Return the EOM
	commit := e.CommitmentIndicator.ShouldICommit(msg)
	if commit {
		return msg.VolunteerMessage.Eom
	}

	res := e.VolunteerControls[msg.VolunteerMessage.Signer].Execute(msg)
	if res != nil {
		// If it is a vote from us, then we need to decide if we should send it
		if ll, ok := res.(*messages.LeaderLevelMessage); ok && ll.Signer == e.Self {
			// We need to set the volunteer priority for comparing
			ll.VolunteerPriority = e.getVolunteerPriority(ll.VolunteerMessage.Signer)

			// We have a new vote we might be able to cast
			if e.CurrentVote.Less(ll) {
				// This vote is better than our current, let's pass it out.
				if ll.Rank >= e.CurrentLevel {
					// We cannot issue a rank 2 on level 2. It has to be on level 3
					ll.Level = ll.Rank + 1
					e.CurrentLevel = ll.Rank + 2
				} else {
					// Set level to our level, increment
					ll.Level = e.CurrentLevel
					e.CurrentLevel++
				}
				e.CurrentVote = *ll
				e.executeDisplay(ll)

				// This vote may change our state, so call ourselves again
				return e.Execute(ll)
			} else {
				// This message was from us, and we decided not to sent it
				if ll.Level <= 0 {
					return nil
				}
			}
		}
	}

	// return the result even if we didn't change our current vote
	return res
}
