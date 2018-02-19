package election

import (
	"github.com/FactomProject/electiontesting/imessage"
	"github.com/FactomProject/electiontesting/messages"
)

// RoutingElection is just an election that returns msgs for broadcasting
type RoutingElection struct {
	*Election
}

func NewRoutingElection(e *Election) *RoutingElection {
	r := new(RoutingElection)
	r.Election = e

	return r
}

func (r *RoutingElection) Execute(msg imessage.IMessage) (imessage.IMessage, bool) {
	resp, ch := r.Election.execute(msg)
	if resp == nil && ch {
		return msg, ch
	}

	if ll, ok := msg.(*messages.LeaderLevelMessage); ok {
		if r.Election.CurrentVote.Level > 0 {
			if r.Election.CurrentVote.VolunteerPriority != ll.VolunteerPriority {
				return ll, ch
			}
		} else {
			return ll, ch
		}
	}

	if resp == nil && r.Election.CurrentVote.Level > 0 {
		return &r.Election.CurrentVote, false
	}

	return resp, ch
}
