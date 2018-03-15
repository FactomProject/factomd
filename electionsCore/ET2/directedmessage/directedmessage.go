package directedmessage

import 	"github.com/FactomProject/electiontesting/imessage"

type DirectedMessage struct {
	LeaderIdx int
	Msg       imessage.IMessage

}
