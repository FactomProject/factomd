package testhelper

import (
	. "github.com/FactomProject/electiontesting/messages"
	. "github.com/FactomProject/electiontesting/primitives"
)

// AuthSetHelper will help testing authority sets. Add whatever functions you need to help you
// manage the authsets in tests
type AuthSetHelper struct {
	feds int
	auds int
	AuthSet

	sequenceCounter int
}

func NewAuthSetHelper(feds, auds int) *AuthSetHelper {
	a := new(AuthSetHelper)
	a.AuthSet = *NewAuthSet()

	for i := 0; i < feds; i++ {
		a.AddFed()
	}
	for i := 0; i < auds; i++ {
		a.AddAudit()
	}

	return a
}

func (a *AuthSetHelper) NextIdentity() Identity {
	a.sequenceCounter++
	return Identity(a.sequenceCounter)
}

func (a *AuthSetHelper) AddFed() Identity {
	id := a.NextIdentity()
	a.Add(id, 1)
	return id
}

func (a *AuthSetHelper) AddAudit() Identity {
	id := a.NextIdentity()
	a.Add(id, 0)

	return id
}

func (a *AuthSetHelper) GetAuthSet() AuthSet {
	return a.AuthSet
}

func (a *AuthSetHelper) NewVoteFactory(vol VolunteerMessage) *VoteFactory {
	return NewVoteFactory(a.GetFeds(), vol)
}

func (a *AuthSetHelper) Majority() int {
	return (len(a.GetFeds()) / 2) + 1
}

type VoteFactory struct {
	index     int
	FedList   []Identity
	Volunteer VolunteerMessage
}

func NewVoteFactory(feds []Identity, vol VolunteerMessage) *VoteFactory {
	vf := new(VoteFactory)
	vf.FedList = feds
	vf.Volunteer = vol
	return vf
}

func (v *VoteFactory) NextVote() VoteMessage {
	vote := NewVoteMessage(v.Volunteer, v.NextFed())
	v.index = v.index % len(v.FedList)
	return vote
}

func (v *VoteFactory) VotesMapWithMajority() map[Identity]VoteMessage {
	votes := make(map[Identity]VoteMessage)
	for i := 0; i < v.Majority(); i++ {
		vote := v.NextVote()
		votes[vote.Signer] = vote
	}
	return votes
}

func (v *VoteFactory) VotesListWithMajority() []VoteMessage {
	// List to map
	var votes []VoteMessage
	m := v.VotesMapWithMajority()
	for _, v := range m {
		votes = append(votes, v)
	}
	return votes
}

func (v *VoteFactory) NextMajorityDecision() MajorityDecisionMessage {
	return NewMajorityDecisionMessage(v.VotesMapWithMajority(), v.FedList[v.index])
}

func (v *VoteFactory) MajorityDecisionMapWithMajority() map[Identity]MajorityDecisionMessage {
	votes := make(map[Identity]MajorityDecisionMessage)
	for i := 0; i < v.Majority(); i++ {
		vote := v.NextMajorityDecision()
		votes[vote.Signer] = vote
	}
	return votes
}

func (v *VoteFactory) MajorityDecisionListWithMajority() []MajorityDecisionMessage {
	// List to map
	var votes []MajorityDecisionMessage
	m := v.MajorityDecisionMapWithMajority()
	for _, v := range m {
		votes = append(votes, v)
	}
	return votes
}

func (v *VoteFactory) NextInsistence() InsistMessage {
	return NewInsistenceMessage(v.MajorityDecisionMapWithMajority(), v.FedList[v.index])
}

func (v *VoteFactory) InsistenceMapWithMajority() map[Identity]InsistMessage {
	votes := make(map[Identity]InsistMessage)
	for i := 0; i < v.Majority(); i++ {
		vote := v.NextInsistence()
		votes[vote.Signer] = vote
	}
	return votes
}

func (v *VoteFactory) InsistenceListWithMajority() []InsistMessage {
	// List to map
	var votes []InsistMessage
	m := v.InsistenceMapWithMajority()
	for _, v := range m {
		votes = append(votes, v)
	}
	return votes
}

func (v *VoteFactory) NextFed() Identity {
	f := v.FedList[v.index]
	v.index++
	v.index = v.index % len(v.FedList)
	return f
}

func (v *VoteFactory) MajorityIAck(i InsistMessage) IAckMessage {
	iack := NewIAckMessage(i, v.NextFed())
	for i := 0; i < v.Majority(); i++ {
		f := v.NextFed()
		iack.Signers[f] = true
		if len(iack.Signers) == v.Majority() {
			// Since we sign the iack ourselves too, we might have
			// maj+1 in this loop. Check for that
			break
		}
	}

	return iack
}

func (v *VoteFactory) NextIAck(i InsistMessage) IAckMessage {
	return NewIAckMessage(i, v.NextFed())
}

func (v *VoteFactory) NextPublish() PublishMessage {
	insist := v.NextInsistence()
	return NewPublishMessage(insist, insist.Signer, v.MajorityIAck(insist).Signers)
}

func (v *VoteFactory) Majority() int {
	return (len(v.FedList) / 2) + 1
}
