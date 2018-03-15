package controller

import (
	"fmt"

	"github.com/FactomProject/factomd/electionsCore/imessage"
	"github.com/FactomProject/factomd/electionsCore/messages"
	"github.com/FactomProject/factomd/electionsCore/primitives"
)

// MessageBuffer will manage all messages in the message list
type MessageBuffer struct {
	// A list of all messages, it's an ever growing set, trimming the front
	// keeps the index the same by using the 'base' field.
	Messages []imessage.IMessage
	Base     int

	// When looking for a message by someone, this helps searching
	MessagesMap map[primitives.Identity][]int
}

func NewMessageBuffer() *MessageBuffer {
	b := new(MessageBuffer)
	b.MessagesMap = make(map[primitives.Identity][]int)

	return b
}

func (b *MessageBuffer) Add(msg imessage.IMessage) {
	if msg == nil {
		return
	}
	switch msg.(type) {
	case *messages.LeaderLevelMessage:
		ll := msg.(*messages.LeaderLevelMessage)
		b.MessagesMap[ll.Signer] = append(b.MessagesMap[ll.Signer], len(b.Messages)+b.Base)
		b.Messages = append(b.Messages, ll)
	case *messages.VoteMessage:
		ll := msg.(*messages.VoteMessage)
		b.MessagesMap[ll.Signer] = append(b.MessagesMap[ll.Signer], len(b.Messages)+b.Base)
		b.Messages = append(b.Messages, ll)
	case *messages.VolunteerMessage:
		ll := msg.(*messages.VolunteerMessage)
		b.MessagesMap[ll.Signer] = append(b.MessagesMap[ll.Signer], len(b.Messages)+b.Base)
		b.Messages = append(b.Messages, ll)
	default:
		fmt.Println("Message type not found")
	}

}

func (b *MessageBuffer) RetrieveLeaderLevelMessageByLevel(leader primitives.Identity, level int) imessage.IMessage {
	list := b.MessagesMap[leader]
	for _, v := range list {
		msg, _ := b.RetrieveIndex(v)
		if msg != nil {
			if ll, ok := msg.(*messages.LeaderLevelMessage); ok {
				if ll.Level == level {
					return msg
				}
			}
		}
	}
	return nil
}

// RetrieveLeaderVoteMessage takes a vol too as multiple vote 0s can be sent out
func (b *MessageBuffer) RetrieveLeaderVoteMessage(leader primitives.Identity, vol primitives.Identity) imessage.IMessage {
	list := b.MessagesMap[leader]
	for _, v := range list {
		msg, _ := b.RetrieveIndex(v)
		if msg != nil {
			if v, ok := msg.(*messages.VoteMessage); ok && v.Volunteer.Signer == vol {
				return msg
			}
		}
	}
	return nil
}

func (b *MessageBuffer) RetrieveIndex(index int) (imessage.IMessage, error) {
	i := index - b.Base
	if i == -1 {
		return nil, fmt.Errorf("index is no longer in set of messages")
	}

	if i >= len(b.Messages) {
		return nil, fmt.Errorf("index outside the set of messages")

	}
	return b.Messages[i], nil
}
