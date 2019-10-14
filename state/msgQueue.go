package state

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/telemetry"
	"github.com/FactomProject/factomd/worker"
)

type MsgQueue struct {
	name string
	q    chan interfaces.IMsg
	w    *worker.Thread
}

// use gauge w/ proper labels
func (mq *MsgQueue) Metric(msg interfaces.IMsg) telemetry.Gauge {
	return telemetry.ChannelSize.WithLabelValues("state", mq.name, mq.w.Label(), msg.Label())
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
	// REVIEW: do we want to record totals as prometheus metrics?
	//measureMessage(TotalMessageQueueInMsgGeneralVec, m, true)
	mq.Metric(m).Inc()
	mq.q <- m
}

// Dequeue removes an item from channel and instruments based on type. Returns nil if nothing in
// queue
func (mq MsgQueue) Dequeue() interfaces.IMsg {
	select {
	case v := <-mq.q:
		mq.Metric(v).Dec()
		return v
	default:
		return nil
	}
}

// BlockingDequeue will block until it retrieves from queue
func (mq MsgQueue) BlockingDequeue() interfaces.IMsg {
	v := <-mq.q
	mq.Metric(v).Dec()
	return v
}
