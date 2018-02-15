package volunteercontrol

import (
	"math"

	"fmt"
	"github.com/FactomProject/electiontesting/imessage"
	"github.com/FactomProject/electiontesting/messages"
	. "github.com/FactomProject/electiontesting/primitives"
)

var _ = fmt.Println

// VolunteerControl will keep a record of all votes
// for a given volunteer and produce the best vote
// possible.
type VolunteerControl struct {
	AuthSet
	Self      Identity
	Volunteer *messages.VolunteerMessage

	Votes map[Identity]*messages.LeaderLevelMessage
}

func NewVolunteerControl(self Identity, authset AuthSet) *VolunteerControl {
	v := new(VolunteerControl)
	v.Votes = make(map[Identity]*messages.LeaderLevelMessage)
	v.Self = self
	v.AuthSet = authset

	return v
}

func (a *VolunteerControl) Copy() *VolunteerControl {
	b := NewVolunteerControl(a.Self, a.AuthSet.Copy())
	if a.Volunteer != nil {
		v := *a.Volunteer
		b.Volunteer = &v
	}

	for k, v := range a.Votes {
		b.Votes[k] = v.Copy()
	}

	return b
}

func (v *VolunteerControl) Execute(msg imessage.IMessage) (imessage.IMessage, bool) {
	// When we get a vote, we need to add it to our map
	ll, ok := msg.(*messages.LeaderLevelMessage)
	if !ok {
		return nil, false
	}

	if v.Volunteer == nil {
		v.Volunteer = &ll.VolunteerMessage
	}

	statechange := false

	if ll.Justification != nil {
		for _, j := range ll.Justification {
			change := v.AddVote(j)
			if change {
				statechange = change
			}
		}
	}

	change := v.AddVote(ll)
	if change {
		statechange = change
	}

	resp, checkChange := v.checkVoteCount(msg)
	return resp, checkChange || statechange
}

// addVote just adds the vote to the vote map
func (v *VolunteerControl) AddVote(msg *messages.LeaderLevelMessage) bool {
	// If we already have a vote from that leader for this audit, then we only replace ours if this is better
	if cur, ok := v.Votes[msg.Signer]; ok {
		if cur.Level == msg.Level {
			// Same level, same message (we have  no malicious actors)
			return false
		}

		if msg.Rank > cur.Rank {
			// Greater rank is always better.
			msg.Justification = nil
			v.Votes[msg.Signer] = msg
			return true
		}
	}
	v.Votes[msg.Signer] = msg
	return true
}

// checkVoteCount will check to see if we have enough votes to issue a ranked message. We will not add
// that message to our votemap, as we may have not chosen to actually send that vote. If we decide to send that
// vote, we will get it sent back to us
// 		Returns a LeaderLevelMessage WITHOUT the level set. Don't forget to set it if you send it!
func (v *VolunteerControl) checkVoteCount(msg imessage.IMessage) (imessage.IMessage, bool) {
	// No majority, no bueno. Forward the msg that we got though
	if len(v.Votes) < v.Majority() {
		return msg, false
	}

	m := v.Majority()
	l := len(v.Votes)

	var _, _ = m, l

	var justification []*messages.LeaderLevelMessage

	// Majority votes exist, we need to find the lowest level, and issue back that level message
	rank := math.MaxInt32
	for _, vote := range v.Votes {
		if vote.Level < rank {
			rank = vote.Level
		}
		justification = append(justification, vote)
	}

	// Now we have the lowest level, any message at that level can no longer help us.
	// We can only reuse votes at higher levels
	for k, vote := range v.Votes {
		if vote.Level == rank {
			delete(v.Votes, k)
		}
	}

	llmsg := messages.NewLeaderLevelMessage(v.Self, rank, -2, *v.Volunteer)
	llmsg.Justification = justification

	return &llmsg, true
}
