package state

import (
	"github.com/FactomProject/factomd/common/interfaces"
)

// NetOutMsgQueue counts incoming and outgoing messages for netout queue
type NetOutMsgQueue chan interfaces.IMsg

func NewNetOutMsgQueue(capacity int) NetOutMsgQueue {
	channel := make(chan interfaces.IMsg, capacity)
	return channel
}

// Length of underlying channel
func (q NetOutMsgQueue) Length() int {
	return len(chan interfaces.IMsg(q))
}

// Cap of underlying channel
func (q NetOutMsgQueue) Cap() int {
	return cap(chan interfaces.IMsg(q))
}

// Enqueue adds item to channel and instruments based on type
func (q NetOutMsgQueue) Enqueue(m interfaces.IMsg) {
	measureMessage(TotalMessageQueueNetOutMsgGeneralVec, m, true)

	q <- m
}

// Dequeue removes an item from channel and instruments based on type. Returns nil if nothing in
// queue
func (q NetOutMsgQueue) Dequeue() interfaces.IMsg {
	select {
	case v := <-q:
		//NetOutMsgQueueRateKeeper.Complete()
		return v
	default:
		return nil
	}
}

// BlockingDequeue will block until it retrieves from queue
func (q NetOutMsgQueue) BlockingDequeue() interfaces.IMsg {
	v := <-q
	//NetOutMsgQueueRateKeeper.Complete()
	return v
}
