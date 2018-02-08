package round

import (
	"testing"
	. "github.com/FactomProject/electiontesting/primitives"
	"github.com/FactomProject/electiontesting/messages"
	"github.com/FactomProject/electiontesting/imessage"
	"math/rand"
)

var authSet AuthSet

func init {
	authSet.New()
	// add three leaders nad two audit servers
	for i := 0; i < 5; i++ {
		id := rand.Int()
		authSet.Add(Identity(id), i/2)
	}

}

func TestExecute(t *testing.T) {

	leaderId := authSet.IdentityList[3]
	auditId := authSet.IdentityList[0]

	sm := messages.SignedMessage{leaderId}

	eom := messages.EomMessage{Vm: 0, Minute: 0, Height: 0, SignedMessage: sm}
	volunteerMessage := messages.VolunteerMessage{Eom: eom, SignedMessage: sm}
	voteMessage := messages.VoteMessage{Volunteer: volunteerMessage, SignedMessage: sm}
	majorityDecisionMessage := messages.MajorityDecisionMessage{MajorityVotes: []messages.VoteMessage{voteMessage}, SignedMessage: sm, ...}

	test_messages := []imessage.IMessage{volunteerMessage, voteMessage, majorityDecisionMessage}
	ids := []Identity{leaderId, auditId}

	for _, id := range ids {
		r := NewRound(authSet, id, volunteerMessage, 0,0,0,)
		if authSet.IsLeader(id) {
			if (r.State != RoundState_FedStart) {
				t.Error("Expect to start in RoundState_FedStart")
			}
		} else {
			if (r.State != RoundState_AudStart) {
				t.Error("Expect to start in RoundState_FedStart")
			}

		}
		for message := range test_messages {
			rMessage := r.Execute(message)
			_ = rMessage
		}

	}

	m := r.Execute()

}
