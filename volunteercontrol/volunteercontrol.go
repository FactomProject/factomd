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

func (v *VolunteerControl) Execute(msg imessage.IMessage) imessage.IMessage {
	// When we get a vote, we need to add it to our map
	ll, ok := msg.(*messages.LeaderLevelMessage)
	if !ok {
		return nil
	}

	if v.Volunteer == nil {
		v.Volunteer = &ll.VolunteerMessage
	}

	for _, j := range ll.Justification {
		v.addVote(j)
	}
	v.addVote(ll)

	return v.checkVoteCount(msg)
}

// addVote just adds the vote to the vote map
func (v *VolunteerControl) addVote(msg *messages.LeaderLevelMessage) {
	// If we already have a vote from that leader for this audit, then we only replace ours if this is better
	if cur, ok := v.Votes[msg.Signer]; ok {
		if cur.Level == msg.Level {
			// Same level, same message (we have  no malicious actors)
			return
		}

		if cur.Rank > msg.Rank {
			// Greater rank is always better.
			msg.Justification = []*messages.LeaderLevelMessage{}
			v.Votes[msg.Signer] = msg
		}
	} else {
		v.Votes[msg.Signer] = msg
	}
}

// checkVoteCount will check to see if we have enough votes to issue a ranked message. We will not add
// that message to our votemap, as we may have not chosen to actually send that vote. If we decide to send that
// vote, we will get it sent back to us
// 		Returns a LeaderLevelMessage WITHOUT the level set. Don't forget to set it if you send it!
func (v *VolunteerControl) checkVoteCount(msg imessage.IMessage) imessage.IMessage {

	// No majority, no bueno. Forward the msg that we got though
	if len(v.Votes) < v.Majority() {
		return msg
	}

	var justification []*messages.LeaderLevelMessage

	// Majority votes exist, we need to find the lowest level, and issue back that level message
	level := math.MaxInt32
	for _, vote := range v.Votes {
		if vote.Level < level {
			level = vote.Level
		}
		justification = append(justification, vote)
	}

	// Now we have the lowest level, any message at that level can no longer help us.
	// We can only reuse votes at higher levels
	for k, vote := range v.Votes {
		if vote.Level == level {
			delete(v.Votes, k)
		}
	}

	llmsg := messages.NewLeaderLevelMessage(v.Self, level, -2, *v.Volunteer)
	llmsg.Justification = justification

	return &llmsg
}
