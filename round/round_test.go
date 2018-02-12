package round_test

import (
	"testing"
	"math/rand"
	. "github.com/FactomProject/electiontesting/errorhandling"
	"github.com/FactomProject/electiontesting/imessage"
	"github.com/FactomProject/electiontesting/messages"
	. "github.com/FactomProject/electiontesting/primitives"
	. "github.com/FactomProject/electiontesting/round"
)

var authSet AuthSet

func init() {
	authSet.New()
	// add three leaders and two audit servers
	for i := 0; i < 5; i++ {
		id := rand.Int()
		authSet.Add(Identity(id), i%2)
	}

}

/*
type Round struct {
	// The audit server that we are trying to get majority to pass
	Volunteer         *messages.VolunteerMessage
	Votes             map[Identity]messages.VoteMessage
	MajorityDecisions map[Identity]messages.MajorityDecisionMessage
	Insistences       map[Identity]messages.InsistMessage
	AuthSet

	// My Messages
	Self             Identity
	Vote             *messages.VoteMessage
	MajorityDecision *messages.MajorityDecisionMessage
	Insistence       *messages.InsistMessage
	Publish          *messages.PublishMessage
	IAcks            map[Identity]bool

	State          RoundState
	majorityNumber int

	// EOM Info
	ProcessListLocation
}

 */

func TestRoundString(t *testing.T) {
	var (
		r         Round
		volunteer messages.VolunteerMessage
		vote      messages.VoteMessage
		id        Identity
	)
	volunteer.ReadString("VOLUNTEER ID-76543210 <EOM 1/2/3 ID-89abcdef> <FAULT ID-01234567 1/2/3 99 ID-89abcdef> ID-89abcdef")

	r.Volunteer = &volunteer
	r.Votes = make(map[Identity]messages.SignedMessage)
	//create a vote
	id.ReadString("ID-76543210")
	vote.ReadString("VOTE <VOLUNTEER ID-76543210 <EOM 1/2/3 ID-89abcdef> <FAULT ID-01234567 1/2/3 99 ID-89abcdef> ID-89abcdef> {" +
		"(ID-76543210 ID-76543210) " +
		"(ID-76543211 ID-76543211)" +
		" } " + id.String())
	r.Votes[id] = vote.SignedMessage
	// change the ID and add another vote
	id.ReadString("ID-FEDCBA89")
	vote.Signer = id
	r.Votes[id] = vote.SignedMessage

	r.Vote = &vote

	r.MajorityDecisions = make(map[Identity]messages.MajorityDecisionMessage, 0)
	//r.MajorityDecision[] leave this empty for now

	r.Insistences = make(map[Identity]messages.InsistMessage, 0)
	r.AuthSet.ReadString(`{"IdentityList":[1985229328,19088743],"StatusArray":[1,0],"IdentityMap":{"19088743":1,"1985229328":0}}`)

	r.Self.ReadString("ID-00000001")

	s:= r.String()
	_= s;


}

func TestExecute(t *testing.T) {
	T = t // set ErrorHandling Test context for this test

	leaderId := authSet.IdentityList[3]
	auditId := authSet.IdentityList[0]

	sm := messages.SignedMessage{Signer: leaderId}

	loc := ProcessListLocation{0, MinuteLocation{0, 0}}
	eom := messages.NewEomMessage(sm.Signer, loc)
	volunteerMessage := messages.VolunteerMessage{Eom: eom, SignedMessage: sm}
	voteMessage := messages.VoteMessage{Volunteer: volunteerMessage, SignedMessage: sm}
	majorityDecisionMessage := messages.NewMajorityDecisionMessage(volunteerMessage, map[Identity]messages.SignedMessage{voteMessage.Signer: voteMessage.SignedMessage}, voteMessage.Signer)

	test_messages := []imessage.IMessage{volunteerMessage, voteMessage, majorityDecisionMessage}
	ids := []Identity{leaderId, auditId}

	var rounds map[Identity]*Round = make(map[Identity]*Round)

	for _, id := range ids {
		rounds[id] = NewRound(authSet, id, volunteerMessage, loc)
		if authSet.IsLeader(id) {
			if rounds[id].State != RoundState_FedStart {
				HandleErrorf("Expect to start in RoundState_FedStart, found %s", rounds[id].State.String())
			}
		} else {
			if rounds[id].State != RoundState_AudStart {
				HandleErrorf("Expect to start in RoundState_AudStart, found %s", rounds[id].State.String())
			}

		}
	} // for all Ids ...

	// test leader state transitions
	leader_round := rounds[leaderId]
	for _, message := range test_messages {

		// copy the current set of votes
		prevVotes := make(map[Identity]messages.SignedMessage, 0)
		for k, v := range leader_round.Votes {
			prevVotes[k] = v
		}

		prevState := leader_round.State
		rMessages := leader_round.Execute(message) // execute the test message

		switch message.(type) {
		case messages.VolunteerMessage:
			switch prevState {
			case RoundState_FedStart:
				if len(rMessages) != 1 {
					HandleError("Expected only a vote as output")
				}
			case RoundState_MajorityDecsion, RoundState_Insistence, RoundState_Publishing:
				HandleError("Volunteer message unexpected in %v", )

			}
			if (len(leader_round.Votes) != 1) {
				HandleError("Expected vote to be one after volunteer")
			}
			for _, rMessage := range rMessages {
				switch rMessage.(type) {
				case messages.VolunteerMessage:
					/* expected */
				default:
					HandleError("Don't double count votes")
				}
			} // for all return messages
		case messages.VoteMessage:
			v := message.(messages.VoteMessage)
			_, ok := prevVotes[v.Signer]

			if ok {
				if (len(leader_round.Votes) != len(prevVotes)) {
					HandleError("Don't double count votes")
				}

			} else {
				if (len(leader_round.Votes) != len(prevVotes)+1) {
					HandleError("Don't double count votes")
				}
			}

		case messages.MajorityDecisionMessage:
			mj := message.(messages.MajorityDecisionMessage)
			for k, v := range mj.MajorityVotes {
				prevVotes[k] = v // copy the majority votes to my list
			}
			if (len(leader_round.Votes) != len(prevVotes)) {
				HandleError("votes")
			}
		}

	} // for all message
} //TestExecute() {...}

// Test jumping to MJ if receive one
func TestAcceptMajorityDecision(t *testing.T) {
	// mj := messages.NewMajorityDecisionMessage()
}
