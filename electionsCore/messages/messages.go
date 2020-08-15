package messages

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	. "github.com/PaulSnow/factom2d/electionsCore/errorhandling"
	"github.com/PaulSnow/factom2d/electionsCore/imessage"
	. "github.com/PaulSnow/factom2d/electionsCore/primitives"
)

var Nothing NoMessage

type NoMessage struct{}

func (r *NoMessage) String() string      { return jsonMarshal(r) }
func (r *NoMessage) ReadString(s string) { jsonUnmarshal(r, s) }

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
	expectedType := fmt.Sprintf("%T", r)
	n := strings.LastIndex(expectedType, ".")
	expectedType = expectedType[n+1:]
	return fmt.Sprintf("%s %s", expectedType, json)
}

func jsonUnmarshal(r interface{}, jsonData string) {
	var t, expectedType string

	if jsonData[0:1] != "{" {
		//separate the type and the json data if the type is there
		n, err := fmt.Sscanf(jsonData, "%s ", &t)
		if n != 1 || err != nil {
			HandleErrorf("%T.ReadString(\"%s\") failed: %v", r, jsonData, err)
		}
		// get the expectedType excluding the "*messages." at the front (or what ever path is there)
		expectedType = fmt.Sprintf("%T", r)
		n = strings.LastIndex(expectedType, ".")
		expectedType = expectedType[n+1:]
		if t != expectedType {
			HandleErrorf("%T.ReadString(\"%s\") failed: Bad Type %s", r, jsonData, t)
		}
		jsonData = jsonData[len(t)+1:] // remove type from string
	}
	err := json.Unmarshal([]byte(jsonData), r)
	if err != nil {
		HandleErrorf("%T.ReadString(\"%s\") failed: %v", r, jsonData, err)
	}
}

// A tagged message has a tag to recall it later. Optional
type TaggedMessage struct {
	tag [32]byte
}

func (r *TaggedMessage) Tag() [32]byte {
	return r.tag
}
func (r *TaggedMessage) String() string      { return jsonMarshal(r) }
func (r *TaggedMessage) ReadString(s string) { jsonUnmarshal(r, s) }
func (r *TaggedMessage) TagMessage(tag [32]byte) {
	r.tag = tag
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

func (r EomMessage) String() string      { return jsonMarshal(r) }
func (r EomMessage) ReadString(s string) { jsonUnmarshal(r, s) }

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
	Round int
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
	TaggedMessage
}

var _ imessage.IMessage = (*VolunteerMessage)(nil)

func (r *VolunteerMessage) String() string      { return jsonMarshal(r) }
func (r *VolunteerMessage) ReadString(s string) { jsonUnmarshal(r, s) }

func NewVolunteerMessageWithoutEOM(identity Identity) VolunteerMessage {
	var v VolunteerMessage
	v.Signer = identity
	return v
}

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
	Level             int
	VolunteerPriority int

	VolunteerMessage
	SignedMessage
	TaggedMessage

	// Every vote also includes their previous
	PreviousVote *LeaderLevelMessage

	// For the rank 0 case
	VoteMessages []*VoteMessage

	// messages used to justify
	Justification []LeaderLevelMessage
	Committed     bool

	// 		Used internally
	// If you skip to EOM, set this so we know who you skipped from
	EOMFrom Identity
}

func (a *LeaderLevelMessage) Less(b *LeaderLevelMessage) bool {
	// Committed is trump. People won't even issue after that
	//if a.Committed {
	//	return true
	//}
	//if b.Committed {
	//	return false
	//}

	if a.Rank == b.Rank {
		return a.VolunteerPriority < b.VolunteerPriority
	}
	return a.Rank < b.Rank
}

func (a *LeaderLevelMessage) Copy() *LeaderLevelMessage {
	b := NewLeaderLevelMessage(a.Signer, a.Rank, a.Level, a.VolunteerMessage)
	b.Justification = make([]LeaderLevelMessage, len(a.Justification))
	for i, v := range a.Justification {
		b.Justification[i] = *v.Copy()
	}

	if a.PreviousVote != nil {
		b.PreviousVote = a.PreviousVote.Copy()
	}

	b.VoteMessages = make([]*VoteMessage, len(a.VoteMessages))
	for i, v := range a.VoteMessages {
		b.VoteMessages[i] = v.Copy()
	}
	b.Committed = a.Committed
	b.VolunteerPriority = a.VolunteerPriority
	b.EOMFrom = a.EOMFrom
	return &b
}

func (r *LeaderLevelMessage) String() string      { return jsonMarshal(r) }
func (r *LeaderLevelMessage) ReadString(s string) { jsonUnmarshal(r, s) }
func NewLeaderLevelMessage(self Identity, rank, level int, v VolunteerMessage) LeaderLevelMessage {
	var l LeaderLevelMessage
	l.Signer = self
	l.Rank = rank
	l.Level = level
	l.VolunteerMessage = v
	// l.Justification = make([]*LeaderLevelMessage, 0)
	return l
}

// ------------------------------------------------------------------------------------------------------------------
type VoteMessage struct {
	Volunteer VolunteerMessage
	// Other votes you may have seen. Help
	// pass them along
	// OtherVotes map[Identity]SignedMessage
	SignedMessage
	TaggedMessage
}

func (a *VoteMessage) Copy() *VoteMessage {
	b := new(VoteMessage)
	b.Volunteer = a.Volunteer
	// b.OtherVotes = make(map[Identity]SignedMessage)
	// for k, v := range a.OtherVotes {
	// 	b.OtherVotes[k] = v
	// }
	b.SignedMessage = a.SignedMessage

	return b
}
func (r *VoteMessage) String() string      { return jsonMarshal(r) }
func (r *VoteMessage) ReadString(s string) { jsonUnmarshal(r, s) }

func NewVoteMessage(vol VolunteerMessage, self Identity) VoteMessage {
	var vote VoteMessage
	vote.Volunteer = vol
	vote.Signer = self
	// vote.OtherVotes = make(map[Identity]SignedMessage)

	return vote
}
