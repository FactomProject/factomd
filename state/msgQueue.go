package state

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/telemetry"
)

type MsgQueue struct {
	q chan interfaces.IMsg
}

// Length of underlying channel
func (mq MsgQueue) Length() int {
	return len(mq.q)
}

// Cap of underlying channel
func (mq MsgQueue) Cap() int {
	return cap(mq.q)
}

// Enqueue adds item to channel and instruments based on type
func (mq MsgQueue) Enqueue(m interfaces.IMsg) {
	//measureMessage(TotalMessageQueueInMsgGeneralVec, m, true)
	//measureMessage(CurrentMessageQueueInMsgGeneralVec, m, true)
	mq.q <- m
}

// Dequeue removes an item from channel and instruments based on type. Returns nil if nothing in
// queue
func (mq MsgQueue) Dequeue() interfaces.IMsg {
	select {
	case v := <-mq.q:
		//measureMessage(CurrentMessageQueueInMsgGeneralVec, v, false)
		return v
	default:
		return nil
	}
}

// BlockingDequeue will block until it retrieves from queue
func (mq MsgQueue) BlockingDequeue() interfaces.IMsg {
	v := <-mq.q
	//measureMessage(CurrentMessageQueueInMsgGeneralVec, v, false)
	return v
}

// measureMessage will increment/decrement prometheus based on type
func measureMessage(counter *telemetry.GaugeVec, msg interfaces.IMsg, increment bool) {
	if msg == nil {
		return
	}
	amt := float64(1)
	if !increment {
		amt = -1
	}

	if counter == nil {
		panic("nil counter")
	}

	counter.WithLabelValues(msg.Label()).Add(amt)
}
