package state

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/telemetry"
	"github.com/FactomProject/factomd/worker"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type MsgQueue struct {
	Name    string
	Channel chan interfaces.IMsg
	Thread  *worker.Thread
}

// use gauge w/ proper labels
func (mq *MsgQueue) Metric(msg interfaces.IMsg) telemetry.Gauge {
	label := "nil"
	if msg != nil {
		label = msg.Label()
	}

	return telemetry.ChannelSize.WithLabelValues("state", mq.Name, mq.Thread.Label(), label)
}

func (mq *MsgQueue) TotalMetric(msg interfaces.IMsg) telemetry.Counter {
	label := "nil"
	if msg != nil {
		label = msg.Label()
	}

	return telemetry.TotalCounter.WithLabelValues("state", mq.Name, mq.Thread.Label(), label)
}

func (mq *MsgQueue) PollMetric() telemetry.Gauge {
	return telemetry.ChannelSize.WithLabelValues("state", mq.Name, mq.Thread.Label(), "aggregate")
}

// add metric to poll size of queue
func (mq *MsgQueue) RegisterPollMetric() {
	mq.Thread.RegisterMetric(func(poll *time.Ticker, exit chan bool) {
		gauge := mq.PollMetric()

		for {
			select {
			case <- exit:
				return
			case <- poll.C:
				gauge.Set(float64(mq.Length()))
			}
		}
	})
}

type Counter = prometheus.Counter

// Length of underlying channel
func (mq MsgQueue) Length() int {
	return len(mq.Channel)
}

// Cap of underlying channel
func (mq MsgQueue) Cap() int {
	return cap(mq.Channel)
}

// Enqueue adds item to channel and instruments based on type
func (mq MsgQueue) Enqueue(m interfaces.IMsg) {
	mq.TotalMetric(m).Inc()
	mq.Metric(m).Inc()
	mq.Channel <- m
}

// Dequeue removes an item from channel and instruments based on type. Returns nil if nothing in
// queue
func (mq MsgQueue) Dequeue() interfaces.IMsg {
	select {
	case v := <-mq.Channel:
		mq.Metric(v).Dec()
		return v
	default:
		return nil
	}
}

// BlockingDequeue will block until it retrieves from queue
func (mq MsgQueue) BlockingDequeue() interfaces.IMsg {
	v := <-mq.Channel
	mq.Metric(v).Dec()
	return v
}
