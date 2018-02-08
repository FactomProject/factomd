package round

import (
	"fmt"

	"github.com/FactomProject/electiontesting/imessage"
	"github.com/FactomProject/electiontesting/messages"
	. "github.com/FactomProject/electiontesting/primitives"
)

var _ = fmt.Println

const (
	_ int = iota
	// Fed States
	RoundState_FedStart
	RoundState_MajorityDecsion
	RoundState_Insistence

	// Audit States
	RoundState_AudStart
	RoundState_WaitForPublish
	RoundState_WaitForTimeout

	RoundState_Publishing
)

type Round struct {
	Volunteer         *messages.VolunteerMessage
	Votes             map[Identity]messages.VoteMessage
	MajorityDecisions map[Identity]messages.MajorityDecisionMessage
	Insistences       map[Identity]messages.InsistMessage
	AuthSet

	// My Messages
	Self             Identity
	Vote             *messages.VoteMessage
	MajorityDecision *messages.MajorityDecisionMessage
	Insistence       *messages.InsistMessage
	Publish          *messages.PublishMessage
	IAcks            map[Identity]bool

	State          int
	majorityNumber int

	// EOM Info
	Vm     int
	Minute int
	Height int
}

func NewRound(authSet AuthSet, self Identity, volunteer messages.VolunteerMessage, minute int, vm int, height int) *Round {
	r := new(Round)

	r.AuthSet = authSet
	r.Minute = minute
	r.Vm = vm
	r.Height = height
	r.Volunteer = &volunteer

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

	r.Votes = make(map[Identity]messages.VoteMessage)
	r.MajorityDecisions = make(map[Identity]messages.MajorityDecisionMessage)
	r.Insistences = make(map[Identity]messages.InsistMessage)
	r.IAcks = make(map[Identity]bool)

	return r
}

func (r *Round) Execute(msg imessage.IMessage) []imessage.IMessage {
	switch r.State {
	case RoundState_FedStart:
		return r.fedStartExecute(msg)
	case RoundState_AudStart:
		// This means we are an audit. Let's broadcast our volunteer message
		return imessage.MakeMessageArray(msg, *r.Volunteer)
	case RoundState_MajorityDecsion:
		return r.majorityDecisionExecute(msg)
	case RoundState_Insistence:
		return r.insistExecute(msg)
	case RoundState_WaitForPublish:
		// If we get a publish, this round is over. just broadcast whatever we get.
		// TODO: Add whatever we saw to the msg, help aggregate
		return imessage.MakeMessageArray(msg, *r.Volunteer)
	case RoundState_WaitForTimeout:
		// We are waiting for timeout to start the next round. Just broadcast
		// TODO: Add whatever we saw to the msg, help aggregate
		return imessage.MakeMessageArray(msg, *r.Volunteer)
	case RoundState_Publishing:
		return imessage.MakeMessageArray(*r.Publish)
	}

	return nil
}

func (r *Round) fedStartExecute(msg imessage.IMessage) []imessage.IMessage {
	switch msg.(type) {
	case messages.VolunteerMessage:
		v := msg.(messages.VolunteerMessage)
		vote := r.makeVote(v)
		return r.fedStartExecute(vote)
	case messages.VoteMessage:
		vote := msg.(messages.VoteMessage)
		haveMaj := r.AddVote(vote)
		if haveMaj {
			md := r.makeMajorityDecision()
			r.State = RoundState_MajorityDecsion
			return r.majorityDecisionExecute(md)
		}
		// If we don't have a majority broadcast vote
		return imessage.MakeMessageArray(vote)
	case messages.MajorityDecisionMessage:
		r.State = RoundState_MajorityDecsion
		md := msg.(messages.MajorityDecisionMessage)
		// Take the votes and MDs they have
		r.CopyMajorityDecision(md)

		// Make our own MajDec
		myMD := r.makeMajorityDecision()
		return r.majorityDecisionExecute(myMD)
	case messages.IAckMessage:
		// IAck cannot be for us or we would be in the insistence state.
		// Lets call this function to transition us, and broadcast out the
		// iack with whatever else we want to send out
		iack := msg.(messages.IAckMessage)
		insistence := iack.Insist
		return imessage.MakeMessageArrayFromArray(r.fedStartExecute(insistence), iack)
	case messages.InsistMessage:
		r.State = RoundState_Insistence
		insist := msg.(messages.InsistMessage)
		// Get MajDecisions and Votes. So now we can insist
		r.CopyInsist(insist)
		myI := r.makeInsist()
		return imessage.MakeMessageArrayFromArray(r.insistExecute(msg), myI)
	}

	return nil
}

func (r *Round) majorityDecisionExecute(msg imessage.IMessage) []imessage.IMessage {
	switch msg.(type) {
	case messages.VolunteerMessage:
		// We have already signed the volunteer and voted.
		return imessage.MakeMessageArray(r.makeMajorityDecision())
	case messages.VoteMessage:
		// We have already signed the volunteer and voted.
		vote := msg.(messages.VoteMessage)
		r.AddVote(vote)
		return imessage.MakeMessageArray(r.makeMajorityDecision())
	case messages.MajorityDecisionMessage:
		md := msg.(messages.MajorityDecisionMessage)
		haveMaj := r.AddMajorityDecision(md)
		if haveMaj {
			// Move to insistence and send out ours
			r.State = RoundState_Insistence
			insist := r.makeInsist()
			return imessage.MakeMessageArrayFromArray(r.insistExecute(insist), insist)
		}
		return imessage.MakeMessageArray(md)
	case messages.IAckMessage:
		// Forward IAck and take the Insistence
		iack := msg.(messages.IAckMessage)
		insistence := iack.Insist
		return imessage.MakeMessageArrayFromArray(r.majorityDecisionExecute(insistence), iack)
	case messages.InsistMessage:
		r.State = RoundState_Insistence
		insist := msg.(messages.InsistMessage)
		// Get MajDecisions and Votes. So now we can insist.
		r.CopyInsist(insist)
		myI := r.makeInsist()
		return imessage.MakeMessageArrayFromArray(r.insistExecute(msg), myI)
	}

	return nil
}

func (r *Round) insistExecute(msg imessage.IMessage) []imessage.IMessage {
	switch msg.(type) {
	case messages.VolunteerMessage:
		// We have already signed the volunteer and voted.
		return imessage.MakeMessageArray(r.makeInsist())
	case messages.VoteMessage:
		// We have already signed the volunteer and voted.
		vote := msg.(messages.VoteMessage)
		r.AddVote(vote)
		return imessage.MakeMessageArray(r.makeInsist())
	case messages.MajorityDecisionMessage:
		// We already have a majority of majority decisions
		md := msg.(messages.MajorityDecisionMessage)
		r.AddMajorityDecision(md)
		return imessage.MakeMessageArray(r.makeInsist())
	case messages.IAckMessage:
		iack := msg.(messages.IAckMessage)
		// Steal there votes/mds
		r.CopyInsist(iack.Insist)

		// Check if it is to us
		if iack.Insist.Signer == r.Self {
			if r.AddIAck(iack) {
				// Publishing is the last state
				r.State = RoundState_Publishing
				r.makePublish()
				return imessage.MakeMessageArray(r.makePublish())
			}
			return imessage.MakeMessageArray(r.makeInsist())
		}
		// IAck is not to us, we should add ourselves and broadcast.
		iack.Signers[r.Self] = true
		// Let's also rebroadcast our insist
		return imessage.MakeMessageArray(iack, r.makeInsist())
	case messages.InsistMessage:
		insist := msg.(messages.InsistMessage)
		iack := messages.NewIAckMessage(insist, r.Self)
		// Take what they have as well
		r.CopyInsist(insist)
		return imessage.MakeMessageArray(iack, r.makeInsist())
	}
	return nil
}

func (r *Round) CopyVotes(votes map[Identity]messages.VoteMessage) {
	for k, vote := range votes {
		r.Votes[k] = vote
	}
}

func (r *Round) AddIAck(iack messages.IAckMessage) bool {
	r.CopyIAck(iack)
	return len(r.IAcks) > r.GetMajority()
}

// CopyIAck will take an IAck to us and add all the Identities that IAcked us
func (r *Round) CopyIAck(iack messages.IAckMessage) {
	for k, v := range iack.Signers {
		r.IAcks[k] = v
	}
}

func (r *Round) CopyMajorityDecision(md messages.MajorityDecisionMessage) {
	r.MajorityDecisions[md.Signer] = md
	r.CopyVotes(md.MajorityVotes)
	for k, md := range md.OtherMajorityDecisions {
		r.MajorityDecisions[k] = md
		r.CopyVotes(md.MajorityVotes)
	}
}

func (r *Round) CopyInsist(i messages.InsistMessage) {
	for _, md := range i.MajorityMajorityDecisions {
		r.CopyMajorityDecision(md)
	}

}

func (r *Round) AddMajorityDecision(md messages.MajorityDecisionMessage) bool {
	r.MajorityDecisions[md.Signer] = md

	return len(r.MajorityDecisions) > r.GetMajority()
}

func (r *Round) makePublish() messages.PublishMessage {
	if r.Publish != nil {
		return *r.Publish
	}

	publish := messages.NewPublishMessage(r.makeInsist(), r.Self, r.IAcks)
	r.Publish = &publish
	return *r.Publish
}

func (r *Round) makeInsist() messages.InsistMessage {
	if r.Insistence != nil {
		r.Insistence.OtherInsists = r.Insistences
		return *r.Insistence
	}

	i := messages.NewInsistenceMessage(r.MajorityDecisions, r.Self)
	r.Insistence = &i
	r.Insistences[r.Self] = i

	return i
}

// makeVote should only be called once. Once we make our vote, we should never do it again
func (r *Round) makeVote(vol messages.VolunteerMessage) messages.VoteMessage {
	if r.Vote != nil {
		r.Vote.OtherVotes = r.Votes
		return *r.Vote
	}

	vote := messages.NewVoteMessage(vol, r.Self)
	vote.OtherVotes = r.Votes
	r.Votes[r.Self] = vote
	r.Vote = &vote
	return vote
}

func (r *Round) makeMajorityDecision() messages.MajorityDecisionMessage {
	if r.MajorityDecision != nil {
		r.MajorityDecision.OtherMajorityDecisions = r.MajorityDecisions
		return *r.MajorityDecision
	}

	m := messages.NewMajorityDecisionMessage(r.Votes, r.Self)
	m.OtherMajorityDecisions = r.MajorityDecisions
	r.MajorityDecision = &m
	r.MajorityDecisions[r.Self] = m
	return m
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
