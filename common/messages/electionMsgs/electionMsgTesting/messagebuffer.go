package electionMsgTesting

import (
	"fmt"

	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/messages/electionMsgs"
)

// MessageBuffer will manage all messages in the message list
type MessageBuffer struct {
	// A list of all messages, it's an ever growing set, trimming the front
	// keeps the index the same by using the 'base' field.
	Messages []interfaces.IMsg
	Base     int

	// When looking for a message by someone, this helps searching
	MessagesMap map[[32]byte][]int
}

func NewMessageBuffer() *MessageBuffer {
	b := new(MessageBuffer)
	b.MessagesMap = make(map[[32]byte][]int)

	return b
}

func (b *MessageBuffer) Add(msg interfaces.IMsg) {
	if msg == nil {
		return
	}
	switch msg.(type) {
	case *electionMsgs.FedVoteLevelMsg:
		ll := msg.(*electionMsgs.FedVoteLevelMsg)
		b.MessagesMap[ll.Signer.Fixed()] = append(b.MessagesMap[ll.Signer.Fixed()], len(b.Messages)+b.Base)
		b.Messages = append(b.Messages, ll)
	case *electionMsgs.FedVoteProposalMsg:
		ll := msg.(*electionMsgs.FedVoteProposalMsg)
		b.MessagesMap[ll.Signer.Fixed()] = append(b.MessagesMap[ll.Signer.Fixed()], len(b.Messages)+b.Base)
		b.Messages = append(b.Messages, ll)
	case *electionMsgs.FedVoteVolunteerMsg:
		ll := msg.(*electionMsgs.FedVoteVolunteerMsg)
		b.MessagesMap[ll.ServerID.Fixed()] = append(b.MessagesMap[ll.ServerID.Fixed()], len(b.Messages)+b.Base)
		b.Messages = append(b.Messages, ll)
	default:
		fmt.Println("Message type not found")
	}

}

func (b *MessageBuffer) RetrieveLeaderLevelMessageByLevel(leader interfaces.IHash, level int) interfaces.IMsg {
	list := b.MessagesMap[leader.Fixed()]
	for _, v := range list {
		msg, _ := b.RetrieveIndex(v)
		if msg != nil {
			if ll, ok := msg.(*electionMsgs.FedVoteLevelMsg); ok {
				if int(ll.Level) == level {
					return msg
				}
			}
		}
	}
	return nil
}

// RetrieveLeaderVoteMessage takes a vol too as multiple vote 0s can be sent out
func (b *MessageBuffer) RetrieveLeaderVoteMessage(leader interfaces.IHash, vol interfaces.IHash) interfaces.IMsg {
	list := b.MessagesMap[leader.Fixed()]
	for _, v := range list {
		msg, _ := b.RetrieveIndex(v)
		if msg != nil {
			if v, ok := msg.(*electionMsgs.FedVoteProposalMsg); ok && v.Volunteer.ServerID.IsSameAs(vol) {
				return msg
			}
		}
	}
	return nil
}

func (b *MessageBuffer) RetrieveIndex(index int) (interfaces.IMsg, error) {
	i := index - b.Base
	if i == -1 {
		return nil, fmt.Errorf("index is no longer in set of messages")
	}

	if i >= len(b.Messages) {
		return nil, fmt.Errorf("index outside the set of messages")

	}
	return b.Messages[i], nil
}
