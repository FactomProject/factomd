package messages

import (
	. "github.com/FactomProject/electiontesting/primitives"
	. "github.com/FactomProject/electiontesting/errorhandling"
	"fmt"
	"regexp"
	"strconv"
	"github.com/FactomProject/electiontesting/imessage"
)

var embeddedMesssageRegEx *regexp.Regexp

func init() {
	embeddedMesssageRegEx = regexp.MustCompile("<(.+)>") // RegEx to extra a message from a string
	embeddedMesssageRegEx.Longest()                      // Make it greedy so it will handle messages with nested messages embedded
}

type SignedMessage struct {
	Signer Identity
}

var dummySignedMessage SignedMessage

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
	return fmt.Sprintf("EOM %v %v", m.ProcessListLocation.String(), m.SignedMessage.String())
}

func (m *EomMessage) ReadString(s string) {
	var (
		pl string
		sm string
	)
	n, err := fmt.Sscanf(s, "EOM %s %s", &pl, &sm)
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
	Round   int
	SignedMessage
}

func (m *FaultMsg) String() string {
	return fmt.Sprintf("FAULT %v %v %v %v", m.FaultId.String(), m.ProcessListLocation.String(), m.Round, m.SignedMessage.String())
}

func (m *FaultMsg) ReadString(s string) {
	var (
		id string
		pl string
		r  int
		sm string
	)
	n, err := fmt.Sscanf(s, "FAULT %v %v %v %v", &id, &pl, &r, &sm)
	if err != nil || n != 4 {
		HandleErrorf("EomMessage.ReadString(%v) failed: %d %v", s, n, err)
	}
	m.FaultId.ReadString(id)
	m.ProcessListLocation.ReadString(pl)
	m.Round = r
	m.SignedMessage.ReadString(sm)
}

type DbsigMessage struct {
	Prev   Hash
	Height int
	Eom    EomMessage
	SignedMessage
}

func (m *DbsigMessage) String() string {
	return fmt.Sprintf("DBSIG %v %d <%v> %v", m.Prev.String(), m.Height, m.Eom.String(), m.SignedMessage.String())
}

func (m *DbsigMessage) ReadString(s string) {

	// todo: Move all the regex's to init
	//	mTypeRegex := "([A-Z]+) ?"
	hashRegEx := "(-[0-9a-fA-F]+-) ?"
	numberRegex := "([0-9]+) ?"
	// may regret not including the <..> in the message itself and instead only using it when nesting
	messageRegex := "<(.*)> ?" // must be greedy for messages that contain messages
	idRegex := "(ID-[0-9a-fA-F]+) ?"

	DbsigMessageRegEx := regexp.MustCompile("DBSIG " + hashRegEx + numberRegex + messageRegex + idRegex) // RegEx split a DbsigMessage from a string

	parts := DbsigMessageRegEx.FindStringSubmatch(s) // Split the message

	if (parts == nil || len(parts) != 5) {
		HandleErrorf("DbsigMessage.ReadString(%v) failed: found %d parts", s, len(parts))
		return
	}

	m.Prev.ReadString(parts[1])
	m.Height, _ = strconv.Atoi(parts[2])
	m.Eom.ReadString(parts[3])
	m.SignedMessage.ReadString(parts[4])
}

func NewDBSigMessage(identity Identity, eom EomMessage, prev Hash) DbsigMessage {
	var dbs DbsigMessage
	dbs.Prev = prev
	dbs.Eom = eom
	dbs.Signer = identity
	return dbs
}

type AuthChangeMessage struct {
	Id     Identity
	Status int //0 < audit and >0 is leader
	SignedMessage
}

func (m *AuthChangeMessage) StatusString() string {
	switch m.Status {
	case 0:
		return "AUDIT"
	case 1:
		return "LEADER"
	default:
		return fmt.Sprintf("UNKNOWN%d", m.Status)
	}
}

func (m *AuthChangeMessage) String() string {
	return fmt.Sprintf("AUTH %v %v %v", m.Id.String(), m.StatusString(), m.SignedMessage.String())
}

func (m *AuthChangeMessage) ReadString(s string) {
	var (
		id     string
		status string
		sm     string
	)
	n, err := fmt.Sscanf(s, "AUTH %v %v %v", &id, &status, &sm)
	if err != nil || n != 3 {
		HandleErrorf("AuthChangeMessage.ReadString(%v) failed: %d %v", s, n, err)
	}
	m.Id.ReadString(id)
	switch status {
	case "AUDIT":
		m.Status = 0
	case "LEADER":
		m.Status = 1
	default:
		HandleErrorf("AuthChangeMessage.ReadString(%v) bad status %v", s, status)
		m.Status = -1
	}
	m.SignedMessage.ReadString(sm)
}

type VolunteerMessage struct {
	Id  Identity
	Eom EomMessage
	FaultMsg
	SignedMessage
}

func (m *VolunteerMessage) String() string {
	return fmt.Sprintf("VOLUNTEER %v <%v> <%v> %v", m.Id.String(), m.Eom.String(), m.FaultMsg.String(), m.SignedMessage.String())
}

func (m *VolunteerMessage) ReadString(s string) {

	// todo: Move all the regex's to init
	//	mTypeRegex := "([A-Z]+) ?"
	//	hashRegEx := "(-[0-9a-fA-F]+-) ?"
	//	numberRegex := "([0-9]+) ?"
	// may regret not including the <..> in the message itself and instead only using it when nesting
	messageRegex := "<(.*)> ?" // must be greedy for messages that contain messages
	idRegex := "(ID-[0-9a-fA-F]+) ?"

	VolunteerRegex := "VOLUNTEER " + idRegex + messageRegex + messageRegex + idRegex
	VolunteerMessageRegEx :=
		regexp.MustCompile(VolunteerRegex) // RegEx split a VolunteerMessage from a string

	parts := VolunteerMessageRegEx.FindStringSubmatch(s) // Split the message

	if (parts == nil || len(parts) != 5) {
		HandleErrorf("VolunteerMessage.ReadString(%v) failed: found %d parts", s, len(parts))
		return
	}
	m.Id.ReadString(parts[1])
	m.Eom.ReadString(parts[2])
	m.FaultMsg.ReadString(parts[3])
	m.SignedMessage.ReadString(parts[4])
}

func NewVolunteerMessage(e EomMessage, identity Identity) VolunteerMessage {
	var v VolunteerMessage
	v.Eom = e
	v.Signer = identity
	return v
}

// ------------------------------------------------------------------------------------------------------------------
type VoteMessage struct {
	Volunteer VolunteerMessage
	// Other votes you may have seen. Help
	// pass them along
	OtherVotes map[Identity]SignedMessage
	SignedMessage
}

func mapString(msgMap map[Identity]SignedMessage) (r string) {
	for id, m := range msgMap {
		r += fmt.Sprintf("(%s : %s) ", id.String(), m.String())
	}
	return r
}

func voteMapReadString(s string) (msgMap map[Identity]SignedMessage) {
	messageRegex := "<(.*)> ?"                            // must be greedy for messages that contain messages
	messageRegexRegEx := regexp.MustCompile(messageRegex) // RegEx split a msgMap from a string

	votes := messageRegexRegEx.FindAllString(s,-1)

	if len(votes) == 0 {
		HandleErrorf("VoteMessage.ReadString(%v) failed: no votes", s)
		return nil
	}

	msgMap = make(map[Identity]SignedMessage, len(votes))
	for _, pair := range votes {
		var idString, sigString string
		fmt.Sscanf(pair,"(%s : %s)", &idString, &sigString)
		var id Identity
		var sig SignedMessage
		id.ReadString(idString)
		sig.ReadString(sigString)
		msgMap[id] = sig
	}

	return msgMap
}

func (m *VoteMessage) String() string {
	mapString := mapString(m.OtherVotes)
	return fmt.Sprintf("VOTE <%v> {%v} %v", m.Volunteer.String(), mapString, m.SignedMessage.String())
}

func (m *VoteMessage) ReadString(s string) {

	// todo: Move all the regex's to init
	//	mTypeRegex := "([A-Z]+) ?"
	//	hashRegEx := "(-[0-9a-fA-F]+-) ?"
	//	numberRegex := "([0-9]+) ?"
	// may regret not including the <..> in the message itself and instead only using it when nesting
	messageRegex := "<(.*)> ?" // must be greedy for messages that contain messages
	idRegex := "(ID-[0-9a-fA-F]+) ?"
	messageMapRegex := "{(.*)} ?" // must be greedy for messages that contain messages

	VolunteerRegex := "VOTE " + messageRegex + messageMapRegex + idRegex
	VolunteerMessageRegEx :=
		regexp.MustCompile(VolunteerRegex) // RegEx split a VolunteerMessage from a string

	parts := VolunteerMessageRegEx.FindStringSubmatch(s) // Split the message

	if (parts == nil || len(parts) != 4) {
		HandleErrorf("VoteMessage.ReadString(%v) failed: found %d parts", s, len(parts))
		return
	}
	m.Volunteer.ReadString(parts[1])

	m.Signer.ReadString(parts[3])
}

func NewVoteMessage(vol VolunteerMessage, self Identity) VoteMessage {
	var vote VoteMessage
	vote.Volunteer = vol
	vote.Signer = self

	return vote
}

// ------------------------------------------------------------------------------------------------------------------

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
