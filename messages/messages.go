package messages

import (
	. "github.com/FactomProject/electiontesting/primitives"
	. "github.com/FactomProject/electiontesting/errorhandling"
	"fmt"
)

type SignedMessage struct {
	Signer Identity
}

func (m *SignedMessage) String() string {
	return fmt.Sprintf("%s", m.Signer.String())
}

func (m *SignedMessage) ReadString(s string) {
	m.Signer.ReadString(s)
}

type EomMessage struct {
	ProcessListLocation
	SignedMessage
}

func (m *EomMessage) String() string {
	return fmt.Sprintf("EOM %s %s", m.ProcessListLocation.String(), m.SignedMessage.String())
}

func (m *EomMessage) ReadString(s string) string {
	var (
		pl string
		sm string
	)
	n, err := fmt.Scanf(s, "EOM %s %s", &pl, &sm)
	if err != nil || n != 2 {
		HandleErrorf("EomMessage.ReadString(%v) failed: %d %v", s, n, err)
	}
	m.ProcessListLocation.ReadString(pl)
	m.SignedMessage.ReadString(sm)
}

func NewEomMessage(identity Identity, loc ProcessListLocation) EomMessage {
	var e EomMessage
	e.Signer = identity
	e.ProcessListLocation = loc
	return e
}

// Start faulting
type FaultMsg struct {
	FaultId Identity
	ProcessListLocation
	SignedMessage
}

func (m *FaultMsg) String() string {
	return fmt.Sprintf("FAULT %s %s %s", m.FaultId.String(), m.ProcessListLocation.String(), m.SignedMessage.String())
}

func (m *FaultMsg) ReadString(s string) string {
	var (
		id string
		pl string
		sm string
	)
	n, err := fmt.Scanf(s, "FAULT %s %s %s", &id, &pl, &sm)
	if err != nil || n != 3 {
		HandleErrorf("EomMessage.ReadString(%v) failed: %d %v", s, n, err)
	}
	m.FaultId.ReadString(id)
	m.ProcessListLocation.ReadString(pl)
	m.SignedMessage.ReadString(sm)
}

func NewFault(loc ProcessListLocation) FaultMsg {
	return FaultMsg{loc}
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

func NewVolunteerMessage(e EomMessage, identity Identity) VolunteerMessage {
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
	iack.Signers = make(map[Identity]bool)
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
