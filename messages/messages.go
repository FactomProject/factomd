package messages

import (
	"encoding/json"
	"fmt"
	"regexp"
	. "github.com/FactomProject/electiontesting/primitives"
	. "github.com/FactomProject/electiontesting/errorhandling"

	"strings"
)

var embeddedMesssageRegEx *regexp.Regexp

func init() {
	embeddedMesssageRegEx = regexp.MustCompile("<(.+)>") // RegEx to extra a message from a string
	embeddedMesssageRegEx.Longest()                      // Make it greedy so it will handle messages with nested messages embedded
}

func jsonMarshal(r interface{}) string {
	json, err := json.Marshal(r)
	if err != nil {
		HandleErrorf("%T.String(...) failed: %v", r, err)
	}
	// get the expectedType excluding the "*messages." at the front (or what ever path is there)
	expectedType := fmt.Sprintf("%T",r)
	n := strings.LastIndex(expectedType,".")
	expectedType = expectedType[n+1:]
	return fmt.Sprintf("%s %s",expectedType, json)
}

func jsonUnmarshal(r interface{}, jsonData string) {
	var t, expectedType string

	if(jsonData[0:1] != "{") {
		//separate the type and the json data if the type is there
		n, err:= fmt.Sscanf(jsonData, "%s ", &t)
		if n!=1 || err != nil {
			HandleErrorf("%T.ReadString(\"%s\") failed: %v",r, jsonData, err)
		}
		// get the expectedType excluding the "*messages." at the front (or what ever path is there)
		expectedType = fmt.Sprintf("%T",r)
		n = strings.LastIndex(expectedType,".")
		expectedType = expectedType[n+1:]
		if t != expectedType {
			HandleErrorf("%T.ReadString(\"%s\") failed: Bad Type %s",r,jsonData, t)
		}
		jsonData = jsonData[len(t)+1:] // remove type from string
	}
	err := json.Unmarshal([]byte(jsonData), r)
	if err != nil {
		HandleErrorf("%T.ReadString(\"%s\") failed: %v",r, jsonData, err)
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

// ------------------------------------------------------------------------------------------------------------------
// Start faulting
type FaultMsg struct {
	FaultId Identity
	ProcessListLocation
	Round   int
	SignedMessage
}

func (r *FaultMsg) String() string      { return jsonMarshal(r) }
func (r *FaultMsg) ReadString(s string) { jsonUnmarshal(r, s) }

func NewFaultMessage(victim Identity, pl ProcessListLocation, r int, signer Identity) FaultMsg {
	var fault FaultMsg = FaultMsg{victim, pl, r, SignedMessage{signer}}
	return fault
}

// ------------------------------------------------------------------------------------------------------------------
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

// ------------------------------------------------------------------------------------------------------------------


type AuthChangeMessage struct {
	Id     Identity
	Status AuthorityStatus //0 < audit and >0 is leader
	SignedMessage
}

func (r *AuthChangeMessage) String() string      { return jsonMarshal(r) }
func (r *AuthChangeMessage) ReadString(s string) { jsonUnmarshal(r, s) }



// ------------------------------------------------------------------------------------------------------------------
type VolunteerMessage struct {
	Id  Identity
	Eom EomMessage
	FaultMsg
	SignedMessage
}

func (r *VolunteerMessage) String() string      { return jsonMarshal(r) }
func (r *VolunteerMessage) ReadString(s string) { jsonUnmarshal(r, s) }

func NewVolunteerMessage(e EomMessage, identity Identity) VolunteerMessage {
	var v VolunteerMessage
	v.Eom = e
	v.Signer = identity
	return v
}

// ------------------------------------------------------------------------------------------------------------------
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

func (r *LeaderLevelMessage) String() string      { return jsonMarshal(r) }
func (r *LeaderLevelMessage) ReadString(s string) { jsonUnmarshal(r, s) }

func NewLeaderLevelMessage(self Identity, rank, level int, v VolunteerMessage) LeaderLevelMessage {
	var l LeaderLevelMessage
	l.Signer = self
	l.Rank = rank
	l.Level = level
	l.VolunteerMessage = v
	return l
}

// ------------------------------------------------------------------------------------------------------------------
type VoteMessage struct {
	Volunteer VolunteerMessage
	// Other votes you may have seen. Help
	// pass them along
	OtherVotes map[Identity]SignedMessage
	SignedMessage
}

func (r *VoteMessage) String() string      { return jsonMarshal(r) }
func (r *VoteMessage) ReadString(s string) { jsonUnmarshal(r, s) }

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

func (r *MajorityDecisionMessage) String() string      { return jsonMarshal(r) }
func (r *MajorityDecisionMessage) ReadString(s string) { jsonUnmarshal(r, s) }

func NewMajorityDecisionMessage(volunteer VolunteerMessage, votes map[Identity]SignedMessage, self Identity) MajorityDecisionMessage {
	var mj MajorityDecisionMessage
	mj.Volunteer = volunteer
	mj.MajorityVotes = votes
	mj.Signer = self

	return mj
}

// ------------------------------------------------------------------------------------------------------------------
type InsistMessage struct {
	MajorityMajorityDecisions map[Identity]MajorityDecisionMessage
	SignedMessage

	// Other InsistMessages you may have seen. Help
	// pass them along
	OtherInsists map[Identity]InsistMessage
}

func (r *InsistMessage) String() string      { return jsonMarshal(r) }
func (r *InsistMessage) ReadString(s string) { jsonUnmarshal(r, s) }

func NewInsistenceMessage(mds map[Identity]MajorityDecisionMessage, identity Identity) InsistMessage {
	var i InsistMessage
	i.MajorityMajorityDecisions = mds
	i.Signer = identity

	return i
}

// ------------------------------------------------------------------------------------------------------------------
type IAckMessage struct {
	// This tells you to whom you are iacking
	Insist InsistMessage
	// IAcks can accumulate on the same message rather than broadcasting out a lot
	Signers map[Identity]bool
}

func (r *IAckMessage) String() string      { return jsonMarshal(r) }
func (r *IAckMessage) ReadString(s string) { jsonUnmarshal(r, s) }

func NewIAckMessage(insist InsistMessage, identity Identity) IAckMessage {
	var iack IAckMessage
	iack.Insist = insist
	iack.Signers = make(map[Identity]bool)
	iack.Signers[identity] = true

	return iack
}

// ------------------------------------------------------------------------------------------------------------------
type PublishMessage struct {
	Insist               InsistMessage
	MajorityIAckMessages map[Identity]bool
	SignedMessage
}

func (r *PublishMessage) String() string      { return jsonMarshal(r) }
func (r *PublishMessage) ReadString(s string) { jsonUnmarshal(r, s) }

func NewPublishMessage(insist InsistMessage, identity Identity, iackMap map[Identity]bool) PublishMessage {
	var p PublishMessage
	p.Insist = insist
	p.Signer = identity
	p.MajorityIAckMessages = iackMap

	return p
}
