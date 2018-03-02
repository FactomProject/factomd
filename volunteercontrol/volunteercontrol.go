package volunteercontrol

import (
	"math"

	"fmt"

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

	Votes map[Identity]messages.LeaderLevelMessage
}

func NewVolunteerControl(self Identity, authset AuthSet) *VolunteerControl {
	v := new(VolunteerControl)
	v.Votes = make(map[Identity]messages.LeaderLevelMessage)
	v.Self = self
	v.AuthSet = authset

	return v
}

// addVote just adds the vote to the vote map, and will not act upon it
func (v *VolunteerControl) AddVote(msg messages.LeaderLevelMessage) bool {
	if v.Volunteer == nil {
		v.Volunteer = &msg.VolunteerMessage
	}

	// If we already have a vote from that leader for this audit, then we only replace ours if this is better
	if cur, ok := v.Votes[msg.Signer]; ok {
		if cur.Level == msg.Level {
			// Same level, same message (we have  no malicious actors)
			return false
		}

		if msg.Rank > cur.Rank {
			// Greater rank is always better. Replace their current with the new
			msg.Justification = nil
			v.Votes[msg.Signer] = msg
			return true
		} else {
			return false
		}
	}

	v.Votes[msg.Signer] = msg

	// New Vote, if we have more than a majority, delete the lowest vote
	// to keep the majority the best majority possible
	if len(v.Votes) > v.Majority() {
		// Delete the lowest one, we don't need it
		lowest := math.MaxInt32
		var lowestvote messages.LeaderLevelMessage
		lowestvote.Rank = math.MaxInt32
		remove := NewIdentityFromInt(-1)
		for k, v := range v.Votes {
			if v.Level < lowest || (v.Level == lowest && !v.Less(&lowestvote)) {
				// If level is lower OR equal and less
				lowest = v.Level
				remove = k
				lowestvote = v
			}
		}

		delete(v.Votes, remove)
	}

	return true
}

// checkVoteCount will check to see if we have enough votes to issue a ranked message. We will not add
// that message to our votemap, as we may have not chosen to actually send that vote. If we decide to send that
// vote, we will get it sent back to us
// 		Returns a LeaderLevelMessage with the level set, however it may need adjusting! (Can only adjust it up)
func (v *VolunteerControl) CheckVoteCount() *messages.LeaderLevelMessage {
	// No majority, no bueno.
	if len(v.Votes) < v.Majority() {
		return nil
	}

	var justification []messages.LeaderLevelMessage

	// Majority votes exist, we need to find the lowest level, and use it for our rank.
	rank := math.MaxInt32
	highestlevel := 0
	for _, vote := range v.Votes {
		// If vote level is less than current rank, bring down our rank
		if vote.Level < rank {
			rank = vote.Level
		}
		if vote.Level > highestlevel {
			highestlevel = vote.Level
		}
		justification = append(justification, vote)
	}

	// Now we have the lowest level, any message at that level can no longer help us.
	// We can only reuse votes at higher levels
	for k, vote := range v.Votes {
		if vote.Level <= rank {
			delete(v.Votes, k)
		}
	}

	llmsg := messages.NewLeaderLevelMessage(v.Self, rank, highestlevel, *v.Volunteer)
	llmsg.Justification = justification

	return &llmsg
}

func (a *VolunteerControl) Copy() *VolunteerControl {
	b := NewVolunteerControl(a.Self, a.AuthSet.Copy())
	if a.Volunteer != nil {
		v := *a.Volunteer
		b.Volunteer = &v
	}

	for k, v := range a.Votes {
		b.Votes[k] = *v.Copy()
	}

	return b
}
