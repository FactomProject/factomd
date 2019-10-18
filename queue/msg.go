package queue

import (
	"github.com/FactomProject/factomd/common/interfaces"

	"github.com/FactomProject/factomd/common"
	"github.com/FactomProject/factomd/telemetry"
	"github.com/FactomProject/factomd/worker"
	"time"
)

type MsgQueue struct {
	common.Name
	Package string
	Channel chan interfaces.IMsg
	Thread  *worker.Thread
}

// construct gauge w/ proper labels
func (q *MsgQueue) Metric(msg interfaces.IMsg) telemetry.Gauge {
	label := "nil"
	if msg != nil {
		label = msg.Label()
	}

	return telemetry.ChannelSize.WithLabelValues(q.Package, q.GetName(), q.Thread.Label(), label)
}

// construct counter for tracking totals
func (q *MsgQueue) TotalMetric(msg interfaces.IMsg) telemetry.Counter {
	label := "nil"
	if msg != nil {
		label = msg.Label()
	}

	return telemetry.TotalCounter.WithLabelValues(q.Package, q.GetName(), q.Thread.Label(), label)
}

// construct counter for intermittent polling of queue size
func (q *MsgQueue) PollMetric() telemetry.Gauge {
	return telemetry.ChannelSize.WithLabelValues(q.Package, q.GetName(), q.Thread.Label(), "aggregate")
}

// add metric to poll size of queue
func (q *MsgQueue) RegisterPollMetric() {
	q.Thread.RegisterMetric(func(poll *time.Ticker, exit chan bool) {
		gauge := q.PollMetric()

		for {
			select {
			case <-exit:
				return
			case <-poll.C:
				gauge.Set(float64(q.Length()))
			}
		}
	})
}

// Length of underlying channel
func (q MsgQueue) Length() int {
	return len(q.Channel)
}

// Cap of underlying channel
func (q MsgQueue) Cap() int {
	return cap(q.Channel)
}

// Enqueue adds item to channel and instruments based on type
func (q MsgQueue) Enqueue(m interfaces.IMsg) {
	q.TotalMetric(m).Inc()
	q.Metric(m).Inc()
	q.Channel <- m
}

// Dequeue removes an item from channel and instruments based on type.
// Returns nil if nothing in // queue
func (q MsgQueue) Dequeue() interfaces.IMsg {
	select {
	case v := <-q.Channel:
		q.Metric(v).Dec()
		return v
	default:
		return nil
	}
}

// BlockingDequeue will block until it retrieves from queue
func (q MsgQueue) BlockingDequeue() interfaces.IMsg {
	v := <-q.Channel
	q.Metric(v).Dec()
	return v
}
