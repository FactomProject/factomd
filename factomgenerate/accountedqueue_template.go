//+build ignore

//ᐸ/*
//This looks syntatically off because it is a template used to generate go code. In order to make the template be
//gofmt able the parse delimiters are set to 'ᐸ'  and ' ᐳ' so ᐸ.typename ᐳ will be replaced by the typename
//from the //FactomGenerate command
//*/ᐳ

//ᐸif false  ᐳ
package Dummy // this is only here to make gofmt happy and is never in the generated code
//ᐸend ᐳ

import (
	"github.com/FactomProject/factomd/telemetry"
)

//ᐸdefine "accountedqueue" ᐳ
// Start accountedqueue generated go code

type ᐸ.typename ᐳ struct {
common.Name
Channel chan ᐸ.type ᐳ
}

func (q *ᐸ.typename ᐳ) Init(parent common.NamedObject, name string, size int) *ᐸ.typename ᐳ {
	q.Name.Init(parent, name)
	q.Channel = make(chan ᐸ.type ᐳ, size)
	return q
}

// construct gauge w/ proper labels
func (q *ᐸ.typename ᐳ) Metric() telemetry.Gauge {
	return telemetry.ChannelSize.WithLabelValues("state", q.GetPath(), "thread", "current")
}

// construct counter for tracking totals
func (q *ᐸ.typename ᐳ) TotalMetric() telemetry.Counter {
	return telemetry.TotalCounter.WithLabelValues("state", q.GetPath(), "thread", "total")
}

// Length of underlying channel
func (q ᐸ.typename ᐳ) Length() int {
	return len(q.Channel)
}

// Cap of underlying channel
func (q ᐸ.typename ᐳ) Cap() int {
	return cap(q.Channel)
}

// Enqueue adds item to channel and instruments based on type
func (q ᐸ.typename ᐳ) Enqueue(m ᐸ.type ᐳ) {
	q.Channel <- m
	q.TotalMetric().Inc()
	q.Metric().Inc()
}

// Enqueue adds item to channel and instruments based on
// returns true it it enqueues the data
func (q ᐸ.typename ᐳ) EnqueueNonBlocking(m ᐸ.type ᐳ) bool {
	select {
	case q.Channel <- m:
		q.TotalMetric().Inc()
		q.Metric().Inc()
		return true
	default:
		return false
	}
}

// Dequeue removes an item from channel
// Returns nil if nothing in // queue
func (q ᐸ.typename ᐳ) Dequeue() ᐸ.type ᐳ {
	select {
	case v := <-q.Channel:
		q.Metric().Dec()
		return v
	default:
		return nil
	}
}

// Dequeue removes an item from channel
func (q ᐸ.typename ᐳ) BlockingDequeue() ᐸ.type ᐳ {
	v := <-q.Channel
	q.Metric().Dec()
	return v
}

// End accountedqueue generated go code
// ᐸend  ᐳ
