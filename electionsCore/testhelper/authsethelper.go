package testhelper

import (
	. "github.com/PaulSnow/factom2d/electionsCore/messages"
	. "github.com/PaulSnow/factom2d/electionsCore/primitives"
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
	return NewIdentityFromInt(a.sequenceCounter)
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

func (v *VoteFactory) NextFed() Identity {
	f := v.FedList[v.index]
	v.index++
	v.index = v.index % len(v.FedList)
	return f
}

func (v *VoteFactory) Majority() int {
	return (len(v.FedList) / 2) + 1
}
