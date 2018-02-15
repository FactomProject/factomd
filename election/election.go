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
	// If I have committed to an answer and found enough to finish Election
	Committed bool
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

func (e *Election) updateCurrentVote(new messages.LeaderLevelMessage) {
	e.CurrentVote = new
}

func (e *Election) Execute(msg imessage.IMessage) (imessage.IMessage, bool) {
	e.executeDisplay(msg)

	// We are done, never use anything else
	if e.Committed {
		return &e.CurrentVote, false
	}

	switch msg.(type) {
	case *messages.LeaderLevelMessage:
		return e.executeLeaderLevelMessage(msg.(*messages.LeaderLevelMessage))
	case *messages.VolunteerMessage:
		vol := msg.(*messages.VolunteerMessage)
		vote := messages.NewVoteMessage(*vol, e.Self)
		msg, _ := e.Execute(&vote)
		if ll, ok := msg.(*messages.LeaderLevelMessage); ok {
			// TODO: Add votes to vote array in leader level message
			var _ = ll
		}
		return &vote, true
	case *messages.VoteMessage:
		// Colecting these allows us to issue out 0.#
		vote := msg.(*messages.VoteMessage)
		vol := vote.Volunteer.Signer

		if e.VolunteerVotes[vol] == nil {
			e.VolunteerVotes[vol] = make(map[Identity]*messages.VoteMessage)
		}
		e.VolunteerVotes[vol][vote.Signer] = vote
		e.executeDisplay(vote)
		if len(e.VolunteerVotes[vote.Volunteer.Signer]) >= e.Majority() {
			// We have a majority of level 0 votes and can issue a rank 0 LeaderLevel Message

			// No current vote, so send!
			if e.CurrentVote.Rank == -1 {
				goto SendRank0
			}

			// If we have a rank 1+, we will not issue a rank 0
			if e.CurrentVote.Rank > 0 {
				return &e.CurrentVote, false // forward our answer again
			}

			// If we have a rank 0, and higher priority volunteer (or same, don't vote again)
			if e.CurrentVote.VolunteerPriority >= e.getVolunteerPriority(vol) {
				return &e.CurrentVote, false // forward our answer again
			}

		SendRank0:
			ll := messages.NewLeaderLevelMessage(e.Self, 0, e.CurrentLevel, vote.Volunteer)
			ll.VolunteerPriority = e.getVolunteerPriority(vote.Volunteer.Signer)
			for _, v := range e.VolunteerVotes[vote.Volunteer.Signer] {
				ll.VoteMessage = append(ll.VoteMessage, v)
			}

			e.CurrentLevel++
			e.updateCurrentVote(ll)

			//e.executeDisplay(&ll)
			ret, _ := e.Execute(&ll)
			// The response could be a better vote. If it is, send that out instead
			if ret != nil {
				if l, ok := ret.(*messages.LeaderLevelMessage); ok {
					if (&ll).Less(l) {
						return l, true
					}
				}
			}

			return &ll, true
		}
	}
	return nil, false
}

func (e *Election) getVolunteerPriority(vol Identity) int {
	return e.GetVolunteerPriority(vol, e.ProcessListLocation)
}

func (e *Election) executeDisplay(msg imessage.IMessage) {
	if e.Display != nil {
		e.Display.Execute(msg)
	}
}

func (e *Election) executeLeaderLevelMessage(msg *messages.LeaderLevelMessage) (imessage.IMessage, bool) {
	if e.VolunteerControls[msg.VolunteerMessage.Signer] == nil {
		e.VolunteerControls[msg.VolunteerMessage.Signer] = volunteercontrol.NewVolunteerControl(e.Self, e.AuthSet)
	}

	if msg.Level <= 0 {
		panic("level <= 0 should never happen")
	}

	// Did a vote change happen?
	voteChange := false
	// Votes exist, so we can add these to our vote map
	if len(msg.VoteMessage) > 0 {
		// If we get a new current vote from this, we will get a vote change. In which case we will
		// send out our current vote if nothing is better
		for _, v := range msg.VoteMessage {
			_, v := e.Execute(v)
			voteChange = v || voteChange
		}
	}

	// If commit is true, then we are done. Return the EOM
	commit := e.CommitmentIndicator.ShouldICommit(msg)
	if commit {
		// TODO: Add the justification for others to also agree
		e.Committed = true
		// Need to make our last leaderlevel message to go to commitment
		lvl := e.CurrentLevel
		if msg.Rank >= lvl {
			lvl = msg.Rank + 1
			e.CurrentLevel = msg.Rank + 1
		} else {
			e.CurrentLevel++
		}
		ll := messages.NewLeaderLevelMessage(e.Self, msg.Rank+1, lvl, msg.VolunteerMessage)
		e.updateCurrentVote(ll)
		e.Execute(&ll)
		return &ll, true
	}

	res, change := e.VolunteerControls[msg.VolunteerMessage.Signer].Execute(msg)
	if res != nil {
		// If it is a vote from us, then we need to decide if we should send it
		// If it already has a volunteer priority, then we decided to already send it out.
		ll, ok := res.(*messages.LeaderLevelMessage)
		if ok && ll.Signer == e.Self && ll.Level < 0 {
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

				e.updateCurrentVote(*ll)
				e.executeDisplay(ll)

				// This vote may change our state, so call ourselves again
				resp, _ := e.Execute(ll)
				if resp != nil {
					return resp, true
				}
				return ll, true
			} else {
				// This message was from us, and we decided not to sent it
				if ll.Level < 0 {
					return nil, change
				}
			}
		}
	}

	// No response, send our current vote because we had a change above
	if res == nil {
		if voteChange {
			return &e.CurrentVote, voteChange
		}
	}

	// return the result even if we didn't change our current vote
	return res, change
}

func (e *Election) VolunteerControlString() string {
	str := "VolunteerControls\n"
	if e.Display == nil {
		return "No display\n"
	}

	vcs := make([]string, 0)

	for i, v := range e.VolunteerControls {
		line := fmt.Sprintf("(%d) ", e.getVolunteerPriority(i))
		if e.VolunteerControls[i] == nil {
			line += "nil"
		} else {
			votes := ""
			sep := ""
			for _, vo := range v.Votes {
				votes += sep + e.Display.FormatMessage(vo)
				sep = ","
			}
			line += votes
		}

		vcs = append(vcs, line)
	}

	for _, l := range vcs {
		str += l + "\n"
	}

	return str
}
