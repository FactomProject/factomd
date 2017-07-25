package state

import (
	"github.com/FactomProject/factomd/common/interfaces"
)

// InMsgMSGQueue counts incoming and outgoing messages for inmsg queue
type InMsgMSGQueue chan interfaces.IMsg

func NewInMsgQueue(capacity int) InMsgMSGQueue {
	channel := make(chan interfaces.IMsg, capacity)
	return channel
}

// Length of underlying channel
func (q InMsgMSGQueue) Length() int {
	return len(chan interfaces.IMsg(q))
}

// Cap of underlying channel
func (q InMsgMSGQueue) Cap() int {
	return cap(chan interfaces.IMsg(q))
}

// Enqueue adds item to channel and instruments based on type
func (q InMsgMSGQueue) Enqueue(m interfaces.IMsg) {
	measureMessage(TotalMessageQueueInMsgGeneralVec, m, true)
	measureMessage(CurrentMessageQueueInMsgGeneralVec, m, true)
	q <- m
}

// Dequeue removes an item from channel and instruments based on type. Returns nil if nothing in
// queue
func (q InMsgMSGQueue) Dequeue() interfaces.IMsg {
	select {
	case v := <-q:
		measureMessage(CurrentMessageQueueInMsgGeneralVec, v, false)
		return v
	default:
		return nil
	}
}

// BlockingDequeue will block until it retrieves from queue
func (q InMsgMSGQueue) BlockingDequeue() interfaces.IMsg {
	v := <-q
	measureMessage(CurrentMessageQueueInMsgGeneralVec, v, false)
	return v
}
