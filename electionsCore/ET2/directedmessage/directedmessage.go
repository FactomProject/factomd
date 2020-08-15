package directedmessage

import "github.com/PaulSnow/factom2d/electionsCore/imessage"

type DirectedMessage struct {
	LeaderIdx int
	Msg       imessage.IMessage
}
