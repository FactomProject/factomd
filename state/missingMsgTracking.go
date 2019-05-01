package state

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

type MsgHeight struct {
	DBHeight int
	VM int
	Height int
}

type ACKMap struct {
	Acks map[MsgHeight]interfaces.IMsg
	MsgOrder [1000][32]byte
	N int
}

type MSGMap struct {
	Msgs map[[32]byte]interfaces.IMsg
	MsgOrder [1000][32]byte
	N int
}

type MissingMessageResponse struct {
	AcksMap ACKMap
	MsgsMap MSGMap
}

func (l* ACKMap) Add(msg interfaces.IMsg) {
	ack := msg.(*messages.Ack)
	heights := MsgHeight{int(ack.DBHeight), ack.VMIndex, int(ack.Height)}

	if l.Acks == nil {
		l.Acks = make(map[MsgHeight]interfaces.IMsg, 0)
	}

	fmt.Println("ACK length: ",len(l.Acks))

	//prevous := l.MsgOrder[l.N]
	//delete(l.ACKS, prevous)
	l.Acks[heights] = msg
	l.MsgOrder[l.N] = msg.GetHash().Fixed()
	l.N = (l.N + 1)%1000
}

func (l* ACKMap) Get(DBHeight int, vmIndex int, height int) bool {
	heights := MsgHeight{DBHeight, vmIndex, height}
	msg, exists := l.Acks[heights]
	fmt.Println("ACKMap Get msg: ", msg)
	return exists
}

func (l* MSGMap) Add(msg interfaces.IMsg) {
	//ack := msg.(*messages.Ack)
	//heights := MsgHeight{ack.DBHeight, ack.VMIndex, ack.Height}
	//fmt.Println("Add heights: ", heights)

	hash := msg.GetHash().Fixed()

	if l.Msgs == nil {
		l.Msgs = make(map[[32]byte]interfaces.IMsg, 0)
	}

	//prevous := l.MsgOrder[l.N]
	//delete(l.ACKS, prevous)
	l.Msgs[hash] = msg
	l.MsgOrder[l.N] = hash
	l.N = (l.N + 1)%1000
}

func (l* MSGMap) Get(msg interfaces.IMsg) bool {
	hash := msg.GetHash().Fixed()

	_, exists := l.Msgs[hash]
	return exists
}

func (m* MissingMessageResponse) GetAckANDMsg(DBHeight int, vmIndex int, height int) bool {
	heights := MsgHeight{DBHeight, vmIndex, height}
	msg, exists := m.AcksMap.Acks[heights]
	if (msg != nil && exists) {
		fmt.Println("ACKMap Get msg: ", msg)
		msgExists := m.MsgsMap.Get(msg)
		return msgExists
	}
	return exists
}