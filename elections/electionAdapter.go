package elections

import (
	"github.com/FactomProject/electiontesting/election"
	"github.com/FactomProject/electiontesting/imessage"
	"github.com/FactomProject/electiontesting/primitives"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages/electionMsgs"
	//"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/electiontesting/messages"
	"github.com/FactomProject/factomd/common/messages/msgbase"
)

// ElectionAdapter is used to drive the election package, abstracting away factomd
// logic and messages
type ElectionAdapter struct {
	Election *Elections

	// All messages we adapt so we can expand them
	tagedMessages map[[32]byte]interfaces.IMsg

	simulatedElection *election.Election
}

func NewElectionAdapter(e Elections) *ElectionAdapter {
	ea := new(ElectionAdapter)
	ea.tagedMessages = make(map[[32]byte]interfaces.IMsg)

	ea.Election = e
	// Build the authset
	// TODO: Check the order!
	authset := primitives.NewAuthSet()
	for _, f := range ea.Election.Federated {
		authset.AddHash(f.GetChainID(), 1)
	}

	for _, f := range ea.Election.Audit {
		authset.AddHash(f.GetChainID(), 0)
	}

	ea.simulatedElection = election.NewElection(primitives.Identity(ea.Election.FedID.Fixed()), *authset)

	// Set who I am

	return ea
}

// Execute will:
// 	take in a message
// 	convert it to the adapted message
//	convert returned message to imsg
//	return
func (ea *ElectionAdapter) Execute(msg interfaces.IMsg) interfaces.IMsg {
	simmessage := ea.adaptMessage(msg)
	if simmessage == nil {
		// TODO: Handle error case
		return nil
	}

	return nil
}

/***
 *
 * Expanding a message goes from simulation --> factomd
 *
 */

func (ea *ElectionAdapter) expandMessage(msg imessage.IMessage) interfaces.IMsg {
	switch msg.(type) {
	case *messages.VolunteerMessage:
		return ea.expandGeneral(msg)
	case *messages.VoteMessage:
		return ea.expandGeneral(msg)
	case *messages.LeaderLevelMessage:
		ea.expandLevelMessage(msg.(*messages.LeaderLevelMessage), false)
	}

	return nil
}

func (ea *ElectionAdapter) expandGeneral(msg imessage.IMessage) interfaces.IMsg {
	tagable, ok := msg.(imessage.Taggable)
	if !ok {
		return nil
	}
	expandedGeneral, ok := ea.tagedMessages[tagable.Tag()]
	if !ok {
		return nil
	}

	return expandedGeneral
}

func (ea *ElectionAdapter) expandLevelMessage(msg *messages.LeaderLevelMessage, single bool) *electionMsgs.FedVoteLevelMsg {
	expandedGeneral, ok := ea.tagedMessages[msg.Tag()]
	if !ok {
		return nil
	}

	expanded, ok := expandedGeneral.(*electionMsgs.FedVoteLevelMsg)
	if !ok {
		// TODO: Handle error case
		return nil
	}

	if !single {
		for _, j := range msg.Justification {
			je := ea.expandLevelMessage(j, true)
			if je != nil {
				expanded.Justification = append(expanded.Justification, *je)
			}
		}
	}

	return expanded
}

/***
 *
 * Adapting a message goes from factomd --> simulation
 *
 */

func (ea *ElectionAdapter) adaptMessage(msg interfaces.IMsg) imessage.IMessage {
	switch msg.(type) {
	case *electionMsgs.FedVoteVolunteerMsg:
		return ea.adaptVolunteerMessage(msg.(*electionMsgs.FedVoteVolunteerMsg))
	case *electionMsgs.FedVoteMsg:
		return ea.adaptVoteMessage(msg.(*electionMsgs.FedVoteLevelMsg))
	case *electionMsgs.FedVoteLevelMsg:
		return ea.adaptLevelMessage(msg.(*electionMsgs.FedVoteLevelMsg), false)
	}

	return nil
}

func (ea *ElectionAdapter) adaptVolunteerMessage(msg *electionMsgs.FedVoteVolunteerMsg) *messages.VolunteerMessage {
	ea.tagMessage(msg)

	vol := msg.ServerID.Fixed()
	volid := primitives.Identity(vol)
	volmsg := messages.NewVolunteerMessageWithoutEOM(volid)
	volmsg.TagMessage(msg.MsgHash.Fixed())
	return &volmsg
}

func (ea *ElectionAdapter) adaptVoteMessage(msg *electionMsgs.FedVoteLevelMsg) *messages.VoteMessage {
	ea.tagMessage(msg)

	vol := msg.ServerID.Fixed()
	volid := primitives.Identity(vol)
	volmsg := messages.NewVolunteerMessageWithoutEOM(volid)
	vote := messages.NewVoteMessage(volmsg, primitives.Identity(msg.Signer.Fixed()))
	vote.TagMessage(msg.MsgHash.Fixed())
	return &vote
}

// adaptLevelMessage
// To stop possible infinite recursive behavior, only adapt the first level of justifications
func (ea *ElectionAdapter) adaptLevelMessage(msg *electionMsgs.FedVoteLevelMsg, single bool) *messages.LeaderLevelMessage {
	ea.tagMessage(msg)

	vol := msg.ServerID.Fixed()
	volid := primitives.Identity(vol)
	volmsg := messages.NewVolunteerMessageWithoutEOM(volid)
	ll := messages.NewLeaderLevelMessage(primitives.Identity(msg.Signer.Fixed()), int(msg.Rank), int(msg.Level), volmsg)
	ll.TagMessage(msg.MsgHash.Fixed())

	if !single {
		for _, m := range msg.Justification {
			ll.Justification = append(ll.Justification, ea.adaptLevelMessage(&m, true))
		}
	}

	return &ll
}

/*************/

func (ea *ElectionAdapter) tagMessage(msg interfaces.IMsg) {
	ea.tagedMessages[msg.GetHash().Fixed()] = msg
}
