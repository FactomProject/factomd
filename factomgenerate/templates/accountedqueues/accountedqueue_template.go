//Ͼ/*
// The FactomGenerate templates use Greek Capitol syllabary characters using "Ͼ" U+03FE, "Ͽ" U+03FF as the
// delimiters. This is done so the template can be valid go code and goimports and gofmt will work correctly on the
// code and it can be tested in unmodified form. For more information see factomgenerate/generate.go
//*/Ͽ

package accountedqueues // this is only here to make gofmt happy and is never in the generated code

//Ͼdefine "accountedqueue-imports"Ͽ

import (
	"github.com/FactomProject/factomd/common"
	"github.com/FactomProject/factomd/telemetry"
)

//ϾendϿ

// for running the test on the template, not used in the generated versions
type Ͼ_typeϿ interface{}

//Ͼdefine "accountedqueue"Ͽ
// Start accountedqueue generated go code

type Ͼ_typenameϿ struct {
	common.Name
	Channel chan Ͼ_typeϿ
}

func (q *Ͼ_typenameϿ) Init(parent common.NamedObject, name string, size int) *Ͼ_typenameϿ {
	q.Name.Init(parent, name)
	q.Channel = make(chan Ͼ_typeϿ, size)
	return q
}

// construct gauge w/ proper labels
func (q *Ͼ_typenameϿ) Metric() telemetry.Gauge {
	return telemetry.ChannelSize.WithLabelValues(q.GetPath(), "current")
}

// construct counter for tracking totals
func (q *Ͼ_typenameϿ) TotalMetric() telemetry.Counter {
	return telemetry.TotalCounter.WithLabelValues(q.GetPath(), "total")
}

// Length of underlying channel
func (q Ͼ_typenameϿ) Length() int {
	return len(q.Channel)
}

// Cap of underlying channel
func (q Ͼ_typenameϿ) Cap() int {
	return cap(q.Channel)
}

// Enqueue adds item to channel and instruments based on type
func (q Ͼ_typenameϿ) Enqueue(m Ͼ_typeϿ) {
	q.Channel <- m
	q.TotalMetric().Inc()
	q.Metric().Inc()
}

// Enqueue adds item to channel and instruments based on
// returns true if it enqueues the data
func (q Ͼ_typenameϿ) EnqueueNonBlocking(m Ͼ_typeϿ) bool {
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
func (q Ͼ_typenameϿ) Dequeue() Ͼ_typeϿ {
	select {
	case v := <-q.Channel:
		q.Metric().Dec()
		return v
	default:
		return nil
	}
}

// Dequeue removes an item from channel
func (q Ͼ_typenameϿ) BlockingDequeue() Ͼ_typeϿ {
	v := <-q.Channel
	q.Metric().Dec()
	return v
}

// End accountedqueue generated go code
// ϾendϿ
