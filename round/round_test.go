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

func TestExecute(t *testing.T) {
	T = t // set ErrorHandling Test context for this test

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
	}// for all Ids ...

	// test leader state transitions
	leader_round := rounds[leaderId]
	for _, message := range test_messages {

		// copy the current set of votes
		prevVotes := map[Identity]messages.VoteMessage
		for k, v := range leader_round.Votes {
			prevVotes[k] = v
		}

		prevState := leader_round.State
		rMessages := leader_round.Execute(message) // execute the test message

		switch message.(type) {
		case messages.VolunteerMessage:
			switch prevState {
			case RoundState_FedStart:
				if len(rMessages) != 1) {
			HandleError("Expected only a vote as output")
			}
			case RoundState_MajorityDecsion, RoundState_Insistence, RoundState_Publishing:
				HandleError("Volunteer message unexpected in %v", )

			}
			if (len(leader_round.Votes) != 1) {
				HandleError("Expected vote to be one after volunteer")
			}
			for rMessage := range rMessages {
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
