package messages

import (
	. "github.com/FactomProject/electiontesting/primitives"
)

type SignedMessage struct {
	Signer Identity
}

type EomMessage struct {
	Vm     int
	Minute int
	Height int
	SignedMessage
}

type DbsigMessage struct {
	Prev Hash
	Eom  EomMessage
	SignedMessage
}

type AuthChangeMessage struct {
	Id     Identity
	Status int //0 < audit and >0 is leader
	SignedMessage
}

type VolunteerMessage struct {
	Eom EomMessage
	SignedMessage
}

type VoteMessage struct {
	Volunteer VolunteerMessage
	SignedMessage

	// Other votes you may have seen. Help
	// pass them along
	OtherVotes []VoteMessage
}

func NewVoteMessage(vol VolunteerMessage, self Identity) VoteMessage {
	var vote VoteMessage
	vote.Volunteer = vol
	vote.Signer = self

	return vote
}

type MajorityDecisionMessage struct {
	MajorityVotes []VoteMessage
	SignedMessage

	// Other MajorityDecisions you may have seen. Help
	// pass them along
	OtherMajorityDecisions []MajorityDecisionMessage
}

func NewMajorityDecisionMessage(votes []VoteMessage, self Identity) MajorityDecisionMessage {
	var mj MajorityDecisionMessage
	mj.MajorityVotes = votes
	mj.Signer = self

	return mj
}

type InsistMessage struct {
	MajorityMajorityDecision []MajorityDecisionMessage
	SignedMessage

	// Other InsistMessages you may have seen. Help
	// pass them along
	OtherInsists []InsistMessage
}

type IAckMessage struct {
	// This tells you to whom you are iacking
	Insist InsistMessage
	SignedMessage

	// If you see other IAck's to the same person, we can help accumulate them
	OtherIAckMessages []IAckMessage
}

type PublishMessage struct {
	Insist               InsistMessage
	MajorityIAckMessages []IAckMessage
	SignedMessage
}
