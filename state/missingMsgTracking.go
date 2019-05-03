package state

import (
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
	MsgOrder [1000]interfaces.IMsg
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

	if len(l.Acks) > 0 {
		previous := l.MsgOrder[l.N]
		if previous != nil {
			prevAck := previous.(*messages.Ack)
			prevHeights := MsgHeight{int(prevAck.DBHeight), prevAck.VMIndex, int(prevAck.Height)}
			delete(l.Acks, prevHeights)
		}
	}
	l.Acks[heights] = msg
	l.MsgOrder[l.N] = msg
	l.N = (l.N + 1)%1000
}

func (l* ACKMap) Get(DBHeight int, vmIndex int, height int) bool {
	heights := MsgHeight{DBHeight, vmIndex, height}
	_, exists := l.Acks[heights]
	return exists
}

func (l* MSGMap) Add(msg interfaces.IMsg) {
	hash := msg.GetHash().Fixed()

	if l.Msgs == nil {
		l.Msgs = make(map[[32]byte]interfaces.IMsg, 0)
	}

	prevous := l.MsgOrder[l.N]
	delete(l.Msgs, prevous)
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
		msgExists := m.MsgsMap.Get(msg)
		return msgExists
	}
	return exists
}