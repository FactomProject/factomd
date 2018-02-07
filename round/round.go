package round

import (
	"github.com/FactomProject/electiontesting/imessage"
	"github.com/FactomProject/electiontesting/messages"
	. "github.com/FactomProject/electiontesting/primitives"
)

const (
	// Fed States
	RoundState_FedStart int = iota
	RoundState_MajorityDecsion
	RoundState_Insistence

	// Audit States
	RoundState_AudStart
	RoundState_WaitForPublish
	RoundState_WaitForTimeout
)

type Round struct {
	Volunter          *messages.VolunteerMessage
	Votes             map[Identity]messages.VoteMessage
	MajorityDecisions []messages.MajorityDecisionMessage
	AuthSet

	// My Messages
	Self             Identity
	Vote             *messages.VoteMessage
	MajorityDecision *messages.MajorityDecisionMessage
	Insistence       *messages.InsistMessage
	IAcks            []messages.IAckMessage

	State          int
	majorityNumber int
}

func NewRound(msg imessage.IMessage, authSet AuthSet, self Identity) *Round {
	r := new(Round)

	r.AuthSet = authSet

	// Am I a fed or an audit?
	r.Self = self
	index, ok := r.IdentityMap[r.Self]
	if !ok {
		panic("I'm not a authority?")
	}

	if r.StatusArray[index] <= 0 {
		// Audit
		r.State = RoundState_AudStart
	} else {
		// Fed
		r.State = RoundState_FedStart
	}
	r.Execute(msg)

	return r
}

func (r *Round) Execute(msg imessage.IMessage) []imessage.IMessage {
	switch r.State {
	case RoundState_FedStart:
		r.fedStartExecute(msg)
	case RoundState_AudStart:
	case RoundState_MajorityDecsion:

	}

	return nil
}

func (r *Round) fedStartExecute(msg imessage.IMessage) {
	switch msg.(type) {
	case messages.VolunteerMessage:
		v := msg.(messages.VolunteerMessage)
		vote := messages.NewVoteMessage(v, r.Self)
		r.Vote = &vote

		haveMaj := r.AddVote(vote)
		var _ = haveMaj

	case messages.VoteMessage:
	case messages.MajorityDecisionMessage:
	case messages.InsistMessage:
	case messages.IAckMessage:
	}
}

func (r *Round) AddMajorityDecision() bool {
	return true
}

func (r *Round) makeMajorityDecision() {
	//m :=
}

// AddVote returns true if we have a majority of votes
func (r *Round) AddVote(vote messages.VoteMessage) bool {
	// Todo: Add warning if add twice?
	r.Votes[vote.Signer] = vote

	return len(r.Votes) > r.GetMajority()
}

func (r *Round) GetMajority() int {
	if r.majorityNumber != 0 {
		return r.majorityNumber
	}

	// Calc Majority
	for _, s := range r.StatusArray {
		if s > 0 {
			r.majorityNumber++
		}
	}
	r.majorityNumber = (r.majorityNumber / 2) + 1
	return r.majorityNumber
}
