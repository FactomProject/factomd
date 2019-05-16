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
	Acks  map[MsgHeight]interfaces.IMsg
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
	NewMsgs chan interfaces.IMsg
}

// Adds Acks to a map of the last 1000 acks
// The map of acks will be used in tandem with the message map when we get an MMR to ensure we dont ask for a message we already have.
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
	a.Acks[heights] = msg 	  // stores a message by the messages DBHeight, VMIndex and Height.
	a.MsgOrder[a.N] = msg    // keeps track of ACKMap.Acks message order
	a.N = (a.N + 1)%1000  	// increment N by 1 each time and when it reaches 1000 start at 0 again.
}

func (a * ACKMap) Get(DBHeight int, vmIndex int, height int) bool {
	heights := MsgHeight{DBHeight, vmIndex, height}
	_, exists := a.Acks[heights]
	return exists
}

// Adds messages to a map of the last 1000 messages
// The map of messages will be used in tandem with the ack map when we get an MMR to ensure we dont ask for a message we already have.
func (m * MSGMap) Add(msg interfaces.IMsg) {
	hash := msg.GetHash().Fixed()

	if m.Msgs == nil {
		m.Msgs = make(map[[32]byte]interfaces.IMsg, 0)
	}

	previous := m.MsgOrder[m.N]
	delete(m.Msgs, previous)
	m.Msgs[hash] = msg  	  // stores a message by its hash.
	m.MsgOrder[m.N] = hash   // keeps track of MSGMap.Msgs message order
	m.N = (m.N + 1)%1000  	// increment N by 1 each time and when it reaches 1000 start at 0 again.
}

func (m * MSGMap) Get(msg interfaces.IMsg) bool {
	hash := msg.GetHash().Fixed()
	_, exists := m.Msgs[hash]
	return exists
}

// Called when we receive an ask for an MMR, we check to see if we have message in out Ask AND message maps and return true or false
func (am* MissingMessageResponse) GetAckANDMsg(DBHeight int, vmIndex int, height int) bool {
	heights := MsgHeight{DBHeight, vmIndex, height}
	msg, exists := am.AcksMap.Acks[heights]
	if msg != nil && exists {
		msgExists := am.MsgsMap.Get(msg)
		return msgExists
	}
	return exists
}