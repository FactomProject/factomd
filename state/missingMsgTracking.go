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
	//MsgOrder [1000][32]byte
	MsgOrder [100]interfaces.IMsg
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

	fmt.Println("len(l.Acks): ", len(l.Acks))
	if len(l.Acks) == 100 {
		holder := l.MsgOrder[(l.N - 100)%100]
		ack2 := holder.(*messages.Ack)
		oldHeights := MsgHeight{int(ack2.DBHeight), ack2.VMIndex, int(ack2.Height)}
		fmt.Println("height of old: ", oldHeights)

		var oldestHeight MsgHeight
		for i, j := range l.Acks {
			fmt.Println("i: ", i, " j: ", j)
			fmt.Println("oldestHeight.Height: ", oldestHeight.Height)
			fmt.Println("i.Height: ", i.Height)
			if oldestHeight.DBHeight == 0 {
				oldestHeight = i
			}
			if oldestHeight.Height > i.Height  {

				oldestHeight = i
			}
		}
		fmt.Println("oldest Height: ", oldestHeight)

		fmt.Println("holder: ", holder)
		//fmt.Println("Delete this???: ", l.Acks[holder])
	}

	//prevous := l.MsgOrder[l.N]
	//delete(l.ACKS, prevous)
	l.Acks[heights] = msg
	//l.MsgOrder[l.N] = msg.GetHash().Fixed()
	l.MsgOrder[l.N] = msg
	l.N = (l.N + 1)%100
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
	fmt.Println("adding Message: ", msg.GetMsgHash())

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