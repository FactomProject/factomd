package directedmessage

import 	"github.com/FactomProject/factomd/electionsCore/imessage"

type DirectedMessage struct {
	LeaderIdx int
	Msg       imessage.IMessage

}
