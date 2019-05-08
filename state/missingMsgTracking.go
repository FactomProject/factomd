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

func (a * ACKMap) Add(msg interfaces.IMsg) {
	ack := msg.(*messages.Ack)
	heights := MsgHeight{int(ack.DBHeight), ack.VMIndex, int(ack.Height)}
	if a.Acks == nil {
		a.Acks = make(map[MsgHeight]interfaces.IMsg, 0)
	}

	if len(a.Acks) > 0 {
		previous := a.MsgOrder[a.N]
		if previous != nil {
			prevAck := previous.(*messages.Ack)
			prevHeights := MsgHeight{int(prevAck.DBHeight), prevAck.VMIndex, int(prevAck.Height)}
			delete(a.Acks, prevHeights)
		}
	}
	a.Acks[heights] = msg
	a.MsgOrder[a.N] = msg
	a.N = (a.N + 1)%1000
}

func (a * ACKMap) Get(DBHeight int, vmIndex int, height int) bool {
	heights := MsgHeight{DBHeight, vmIndex, height}
	_, exists := a.Acks[heights]
	return exists
}

func (m * MSGMap) Add(msg interfaces.IMsg) {
	hash := msg.GetHash().Fixed()

	if m.Msgs == nil {
		m.Msgs = make(map[[32]byte]interfaces.IMsg, 0)
	}

	previous := m.MsgOrder[m.N]
	delete(m.Msgs, previous)
	m.Msgs[hash] = msg
	m.MsgOrder[m.N] = hash
	m.N = (m.N + 1)%1000
}

func (m * MSGMap) Get(msg interfaces.IMsg) bool {
	hash := msg.GetHash().Fixed()
	_, exists := m.Msgs[hash]
	return exists
}

func (am* MissingMessageResponse) GetAckANDMsg(DBHeight int, vmIndex int, height int) bool {
	heights := MsgHeight{DBHeight, vmIndex, height}
	msg, exists := am.AcksMap.Acks[heights]
	if msg != nil && exists {
		msgExists := am.MsgsMap.Get(msg)
		return msgExists
	}
	return exists
}