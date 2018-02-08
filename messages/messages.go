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

func NewEomMessage(vm, minute, height int) EomMessage {
	var e EomMessage
	e.Vm = vm
	e.Minute = minute
	e.Height = height

	return e
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

func NewVolunterMessage(e EomMessage, identity Identity) VolunteerMessage {
	var v VolunteerMessage
	v.Eom = e
	v.Signer = identity
	return v
}

type VoteMessage struct {
	Volunteer VolunteerMessage
	SignedMessage

	// Other votes you may have seen. Help
	// pass them along
	OtherVotes map[Identity]VoteMessage
}

func NewVoteMessage(vol VolunteerMessage, self Identity) VoteMessage {
	var vote VoteMessage
	vote.Volunteer = vol
	vote.Signer = self

	return vote
}

type MajorityDecisionMessage struct {
	MajorityVotes map[Identity]VoteMessage
	SignedMessage

	// Other MajorityDecisions you may have seen. Help
	// pass them along
	OtherMajorityDecisions map[Identity]MajorityDecisionMessage
}

func NewMajorityDecisionMessage(votes map[Identity]VoteMessage, self Identity) MajorityDecisionMessage {
	var mj MajorityDecisionMessage
	mj.MajorityVotes = votes
	mj.Signer = self

	return mj
}

type InsistMessage struct {
	MajorityMajorityDecisions map[Identity]MajorityDecisionMessage
	SignedMessage

	// Other InsistMessages you may have seen. Help
	// pass them along
	OtherInsists map[Identity]InsistMessage
}

func NewInsistenceMessage(mds map[Identity]MajorityDecisionMessage, identity Identity) InsistMessage {
	var i InsistMessage
	i.MajorityMajorityDecisions = mds
	i.Signer = identity

	return i
}

type IAckMessage struct {
	// This tells you to whom you are iacking
	Insist InsistMessage
	// IAcks can accumulate on the same message rather than broadcasting out a lot
	Signers map[Identity]bool
}

func NewIAckMessage(insist InsistMessage, identity Identity) IAckMessage {
	var iack IAckMessage
	iack.Insist = insist
	iack.Signers[identity] = true

	return iack
}

type PublishMessage struct {
	Insist               InsistMessage
	MajorityIAckMessages map[Identity]bool
	SignedMessage
}

func NewPublishMessage(insist InsistMessage, identity Identity, iackMap map[Identity]bool) PublishMessage {
	var p PublishMessage
	p.Insist = insist
	p.Signer = identity
	p.MajorityIAckMessages = iackMap

	return p
}
