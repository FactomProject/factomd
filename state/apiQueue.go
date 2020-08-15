package state

import (
	"github.com/PaulSnow/factom2d/common/interfaces"
)

// APIMSGQueue counts incoming and outgoing messages for API queue
type APIMSGQueue chan interfaces.IMsg

func NewAPIQueue(capacity int) APIMSGQueue {
	channel := make(chan interfaces.IMsg, capacity)
	return channel
}

// Length of underlying channel
func (q APIMSGQueue) Length() int {
	return len(chan interfaces.IMsg(q))
}

// Cap of underlying channel
func (q APIMSGQueue) Cap() int {
	return cap(chan interfaces.IMsg(q))
}

// Enqueue adds item to channel and instruments based on type
func (q APIMSGQueue) Enqueue(m interfaces.IMsg) {
	measureMessage(TotalMessageQueueApiGeneralVec, m, true)
	measureMessage(CurrentMessageQueueApiGeneralVec, m, true)
	q <- m
}

// Dequeue removes an item from channel and instruments based on type. Returns nil if nothing in
// queue
func (q APIMSGQueue) Dequeue() interfaces.IMsg {
	select {
	case v := <-q:
		measureMessage(CurrentMessageQueueApiGeneralVec, v, false)
		return v
	default:
		return nil
	}
}

// BlockingDequeue will block until it retrieves from queue
func (q APIMSGQueue) BlockingDequeue() interfaces.IMsg {
	v := <-q
	measureMessage(CurrentMessageQueueApiGeneralVec, v, false)
	return v
}
