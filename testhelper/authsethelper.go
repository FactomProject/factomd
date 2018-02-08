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

func (a *AuthSetHelper) GetFeds() []Identity {
	var feds []Identity
	for i, id := range a.IdentityList {
		if a.StatusArray[i] > 0 {
			feds = append(feds, id)
		}
	}
	return feds
}

func (a *AuthSetHelper) GetAuds() []Identity {
	var auds []Identity
	for i, id := range a.IdentityList {
		if a.StatusArray[i] <= 0 {
			auds = append(auds, id)
		}
	}
	return auds
}

type VoteFactory struct {
	index     int
	FedList   []Identity
	Volunteer VolunteerMessage
}

func (v *VoteFactory) NextVote() VoteMessage {
	vote := NewVoteMessage(v.Volunteer, v.FedList[v.index])
	v.index++
	v.index = v.index % len(v.FedList)
	return vote
}
