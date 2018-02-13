package messages

import (
	"encoding/json"
	"fmt"
	. "github.com/FactomProject/electiontesting/errorhandling"
	. "github.com/FactomProject/electiontesting/primitives"
	"regexp"
)

var embeddedMesssageRegEx *regexp.Regexp

func init() {
	embeddedMesssageRegEx = regexp.MustCompile("<(.+)>") // RegEx to extra a message from a string
	embeddedMesssageRegEx.Longest()                      // Make it greedy so it will handle messages with nested messages embedded
}

func jsonMarshal(r interface{}) string {
	rval, err := json.Marshal(r)
	if err != nil {
		fmt.Printf("%T.String(...) failed: %v", r, err)
	}
	return string(rval)
}

func jsonUnmarshal(r interface{}, s string) {
	err := json.Unmarshal([]byte(s), r)
	if err != nil {
		fmt.Printf("%T.ReadString(%s) failed: %v", r, s, err)
	}
}

type SignedMessage struct {
	Signer Identity
}

func (r *SignedMessage) String() string      { return jsonMarshal(r) }
func (r *SignedMessage) ReadString(s string) { jsonUnmarshal(r, s) }

type EomMessage struct {
	ProcessListLocation
	SignedMessage
}

func (r *EomMessage) String() string      { return jsonMarshal(r) }
func (r *EomMessage) ReadString(s string) { jsonUnmarshal(r, s) }

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
	Round int
	SignedMessage
}

func (r *FaultMsg) String() string      { return jsonMarshal(r) }
func (r *FaultMsg) ReadString(s string) { jsonUnmarshal(r, s) }

type DbsigMessage struct {
	Prev   Hash
	Height int
	Eom    EomMessage
	SignedMessage
}

func (r *DbsigMessage) String() string      { return jsonMarshal(r) }
func (r *DbsigMessage) ReadString(s string) { jsonUnmarshal(r, s) }

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

func (m VolunteerMessage) String() string {
	return fmt.Sprintf("VOLUNTEER %v <%v> <%v> %v", m.Id.String(), m.Eom.String(), m.FaultMsg.String(), m.SignedMessage.String())
}

func (m VolunteerMessage) ReadString(s string) {

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

	if parts == nil || len(parts) != 5 {
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

type LeaderLevelMessage struct {
	// Usually to prove your rank you have to explicitly show
	// the votes you used to obtain that rank, however we don't have
	// to here
	Rank int
	// Leaders must never have 2 messages of the same level
	Level int
	VolunteerMessage
	SignedMessage

	// messages used to justify
	Justification []LeaderLevelMessage
}

func NewLeaderLevelMessage(self Identity, rank, level int, v VolunteerMessage) LeaderLevelMessage {
	var l LeaderLevelMessage
	l.Signer = self
	l.Rank = rank
	l.Level = level
	l.VolunteerMessage = v
	return l
}

func (r LeaderLevelMessage) String() string      { return jsonMarshal(r) }
func (r LeaderLevelMessage) ReadString(s string) { jsonUnmarshal(r, s) }

// ------------------------------------------------------------------------------------------------------------------
type VoteMessage struct {
	Volunteer VolunteerMessage
	// Other votes you may have seen. Help
	// pass them along
	OtherVotes map[Identity]SignedMessage
	SignedMessage
}

func voteMapString(msgMap map[Identity]SignedMessage) (r string) {
	for id, m := range msgMap {
		r += fmt.Sprintf("(%s %s) ", id.String(), m.String())
	}
	return r
}

func voteMapReadString(s string) (msgMap map[Identity]SignedMessage) {
	messageRegex := "[(]([^)]+ [^)]+)[)] ?"               // must be greedy for messages that contain messages
	messageRegexRegEx := regexp.MustCompile(messageRegex) // RegEx split a msgMap from a string

	votes := messageRegexRegEx.FindAllStringSubmatch(s, -1)

	if len(votes) == 0 {
		HandleErrorf("VoteMessage.ReadString(%v) failed: no votes", s)
		return nil
	}

	msgMap = make(map[Identity]SignedMessage, len(votes))
	for _, pair := range votes {
		var idString, sigString string
		fmt.Sscanf(pair[1], "%s %s", &idString, &sigString)
		var id Identity
		var sig SignedMessage
		id.ReadString(idString)
		sig.ReadString(sigString)
		msgMap[id] = sig
	}

	return msgMap
}

func (m VoteMessage) String() string {
	mapString := voteMapString(m.OtherVotes)
	return fmt.Sprintf("VOTE <%v> {%v} %v", m.Volunteer.String(), mapString, m.SignedMessage.String())
}

func (m VoteMessage) ReadString(s string) {

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
	VolunteerMessageRegEx.Longest()                      // Embedded message has embedded messages so be greedy
	parts := VolunteerMessageRegEx.FindStringSubmatch(s) // Split the message

	if parts == nil || len(parts) != 4 {
		HandleErrorf("VoteMessage.ReadString(%v) failed: found %d parts", s, len(parts))
		return
	}
	m.Volunteer.ReadString(parts[1])
	m.OtherVotes = voteMapReadString(parts[2])
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
	Volunteer     VolunteerMessage
	MajorityVotes map[Identity]SignedMessage
	SignedMessage

	// Other MajorityDecisions you may have seen. Help
	// pass them along
	OtherMajorityDecisions map[Identity]MajorityDecisionMessage
}

func (m MajorityDecisionMessage) String() string {
	panic("")
	return ""
}

func (m MajorityDecisionMessage) ReadString(s string) {
	panic("")
}

func NewMajorityDecisionMessage(volunteer VolunteerMessage, votes map[Identity]SignedMessage, self Identity) MajorityDecisionMessage {
	var mj MajorityDecisionMessage
	mj.Volunteer = volunteer
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

func (m InsistMessage) String() string {
	panic("")
	return ""
}

func (m InsistMessage) ReadString(s string) {
	panic("")
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

func (m IAckMessage) String() string {
	panic("")
	return ""
}

func (m IAckMessage) ReadString(s string) {
	panic("")
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

func (m PublishMessage) String() string {
	panic("")
	return ""
}

func (m PublishMessage) ReadString(s string) {
	panic("")
}

func NewPublishMessage(insist InsistMessage, identity Identity, iackMap map[Identity]bool) PublishMessage {
	var p PublishMessage
	p.Insist = insist
	p.Signer = identity
	p.MajorityIAckMessages = iackMap

	return p
}
