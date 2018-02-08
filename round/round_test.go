package round_test

import (
	"github.com/FactomProject/electiontesting/imessage"
	"github.com/FactomProject/electiontesting/messages"
	. "github.com/FactomProject/electiontesting/primitives"
	. "github.com/FactomProject/electiontesting/round"
	"math/rand"
	"testing"
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

func TestExecute(t *testing.T) {

	leaderId := authSet.IdentityList[3]
	auditId := authSet.IdentityList[0]

	sm := messages.SignedMessage{Signer: leaderId}

	loc := ProcessListLocation{0, 0, 0}
	eom := messages.NewEomMessage(sm.Signer, loc)
	volunteerMessage := messages.VolunteerMessage{Eom: eom, SignedMessage: sm}
	voteMessage := messages.VoteMessage{Volunteer: volunteerMessage, SignedMessage: sm}
	majorityDecisionMessage := messages.NewMajorityDecisionMessage(map[Identity]messages.VoteMessage{voteMessage.Signer: voteMessage}, voteMessage.Signer)

	test_messages := []imessage.IMessage{volunteerMessage, voteMessage, majorityDecisionMessage}
	ids := []Identity{leaderId, auditId}

	for _, id := range ids {
		r := NewRound(authSet, id, volunteerMessage, loc)
		if authSet.IsLeader(id) {
			if r.State != RoundState_FedStart {
				t.Errorf("Expect to start in RoundState_FedStart, found %s", RoundStateString(r.State))
			}
		} else {
			if r.State != RoundState_AudStart {
				t.Errorf("Expect to start in RoundState_AudStart, found %s", RoundStateString(r.State))
			}

		}
		for message := range test_messages {
			rMessage := r.Execute(message)
			_ = rMessage
		}

	}

	// m := r.Execute()
}

// Test jumping to MJ if receive one
func TestAcceptMajorityDecision(t *testing.T) {
	// mj := messages.NewMajorityDecisionMessage()
}
