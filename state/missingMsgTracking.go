package state

import (
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

type MsgHeight struct {
	DBHeight int
	VM       int
	Height   int
}

// map of the last N Acks indexed by height
type ACKMap struct {
	Acks     map[MsgHeight]interfaces.IMsg
	MsgOrder [1000]MsgHeight
	N        int
}

// map of the last N inMessages indexed by hash
type MSGMap struct {
	Msgs     map[[32]byte]interfaces.IMsg
	MsgOrder [1000][32]byte
	N        int
}

// Keep the history last N acks and Ackable inMessages
type RecentMessage struct {
	AcksMap ACKMap
	MsgsMap MSGMap
	NewMsgs chan interfaces.IMsg
}

// Adds Acks to a map of the last 1000 acks
// The map of acks will be used in tandem with the message map when we get an MMR to ensure we don't ask for a message we already have.
func (a *ACKMap) Add(msg interfaces.IMsg) {
	if a.Acks == nil {
		a.Acks = make(map[MsgHeight]interfaces.IMsg, 1000)
	}
	// delete the oldest record from the map
	prevH := a.MsgOrder[a.N]
	delete(a.Acks, prevH)

	// add the new ack
	ack := msg.(*messages.Ack)
	height := MsgHeight{int(ack.DBHeight), ack.VMIndex, int(ack.Height)}
	a.Acks[height] = msg     // stores a message by the inMessages DBHeight, VMIndex and Height.
	a.MsgOrder[a.N] = height // keeps track of ACKMap.Acks message order
	a.N = (a.N + 1) % 1000   // increment N by 1 each time and when it reaches 1000 start at 0 again.
}

func (a *ACKMap) Get(DBHeight int, vmIndex int, height int) interfaces.IMsg {
	heights := MsgHeight{DBHeight, vmIndex, height}
	ack := a.Acks[heights]
	return ack
}

// If a message is rejected we need to delete it from the recent message history so we will ask a neighbor
func (m *RecentMessage) HandleRejection(msg interfaces.IMsg, iAck interfaces.IMsg) {
	ack, ok := iAck.(*messages.Ack)
	if ok {
		delete(m.AcksMap.Acks, MsgHeight{int(ack.DBHeight), ack.VMIndex, int(ack.Height)})
		delete(m.MsgsMap.Msgs, ack.GetMsgHash().Fixed())
	} else {
		panic("expected ack")
	}
}

// Adds inMessages to a map
// The map of inMessages will be used in tandem with the ack map when we get an MMR to ensure we don't ask for a message we already have.
func (m *RecentMessage) Add(msg interfaces.IMsg) {
	if msg.Type() == constants.ACK_MSG {
		m.AcksMap.Add(msg) // adds Acks to a Ack map for MMR
	} else {
		if constants.NeedsAck(msg.Type()) {
			if m.MsgsMap.Msgs == nil {
				m.MsgsMap.Msgs = make(map[[32]byte]interfaces.IMsg, 1000)
			}
			// delete the oldest message
			oldHash := m.MsgsMap.MsgOrder[m.MsgsMap.N]
			delete(m.MsgsMap.Msgs, oldHash)

			// Add the new message
			hash := msg.GetMsgHash().Fixed()
			m.MsgsMap.Msgs[hash] = msg             // stores a message by its hash.
			m.MsgsMap.MsgOrder[m.MsgsMap.N] = hash // keeps track of MSGMap.Msgs message order
			m.MsgsMap.N = (m.MsgsMap.N + 1) % 1000 // increment N by 1 each time and when it reaches 1000 start at 0 again.
		}
	}
}

// Called when we receive an ask for an MMR, we check to see if we have an ack and message in out Ask amd Message maps
func (am *RecentMessage) GetAckAndMsg(DBHeight int, vmIndex int, height int, s interfaces.IState) (ack interfaces.IMsg, msg interfaces.IMsg) {
	heights := MsgHeight{DBHeight, vmIndex, height}
	ack = am.AcksMap.Acks[heights]
	if ack != nil {
		msg = am.MsgsMap.Msgs[ack.GetHash().Fixed()]
		return ack, msg
	}
	return nil, nil
}
