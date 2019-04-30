package state

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

type MsgHeight struct {
	DBHeight uint32
	VM int
	Height uint32
}

type ACKMap struct {
	ACKS map[MsgHeight]interfaces.IMsg
	MsgOrder [1000][32]byte
	N int
}

type MSGMap struct {
	Msgs map[[32]byte]interfaces.IMsg
	MsgOrder [1000][32]byte
	N int
}

type MissingMessageResponse struct {
	ACKSMap ACKMap
	Msgs MSGMap
}

func (l* ACKMap) Add(msg interfaces.IMsg) {
	ack := msg.(*messages.Ack)
	heights := MsgHeight{ack.DBHeight, ack.VMIndex, ack.Height}
	fmt.Println("Add heights: ", heights)

	if l.ACKS == nil {
		l.ACKS = make(map[MsgHeight]interfaces.IMsg, 0)
	}

	//prevous := l.MsgOrder[l.N]
	//delete(l.ACKS, prevous)
	l.ACKS[heights] = msg
	l.MsgOrder[l.N] = msg.GetHash().Fixed()
	l.N = (l.N + 1)%1000
}

func (l* ACKMap) Get(msg interfaces.IMsg) bool {
	ack := msg.(*messages.Ack)
	heights := MsgHeight{ack.DBHeight, ack.VMIndex, ack.Height}
	fmt.Println("Get heights: ", heights)
	_, exists := l.ACKS[heights]
	return exists
}