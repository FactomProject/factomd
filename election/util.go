package election

import (
	"github.com/FactomProject/electiontesting/imessage"
	"github.com/FactomProject/electiontesting/messages"
)

func ContainsPublish(msgs []imessage.IMessage) *messages.PublishMessage {
	for _, m := range msgs {
		if publish, ok := m.(messages.PublishMessage); ok {
			return &publish
		}
	}
	return nil
}
