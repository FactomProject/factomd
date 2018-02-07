package messages

import (
	. "github.com/FactomProject/electiontesting/primitives"
)

type EomMessage struct {
	Vm     int
	Minute int
	Height int
	Id     Identity
}

type DbsigMessage struct {
	Prev Hash
	Eom  EomMessage
	Id   Identity
}

type AuthChangeMessage struct {
	Id     Identity
	Status int //0 < audit and >0 is leader
}

type VolunteerMessage struct {
	Eom EomMessage
	Id  Identity
}

type VoteMessage struct {
	Vote VolunteerMessage
	Id   Identity

	// Other votes you may have seen. Help
	// pass them along
	OtherVotes []VoteMessage
}

type MajorityDecisionMessage struct {
	MajorityVotes []VoteMessage
	Signer        Identity

	// Other MajorityDecisions you may have seen. Help
	// pass them along
	OtherMajorityDecisions []MajorityDecisionMessage
}

type InsistMessage struct {
	MajorityMajorityDecision []MajorityDecisionMessage
	Signer                   Identity

	// Other InsistMessages you may have seen. Help
	// pass them along
	OtherInsists []InsistMessage
}

type IAckMessage struct {
	// This tells you to whom you are iacking
	Insist InsistMessage
	Signer Identity

	// If you see other IAck's to the same person, we can help accumulate them
	OtherIAckMessages []IAckMessage
}

type PublishMessage struct {
	Insist               InsistMessage
	MajorityIAckMessages []IAckMessage
}
