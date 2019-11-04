//Ͼ/*
// The FactomGenerate templates use Canadian Aboriginal syllabary characters using "Ͼ" U+1438, "ᐳ" U+1433 as the
// delimiters. This is done so the template can be valid go code and goimports and gofmt will work correctly on the
// code and it can be tested in unmodified form. For more information see factomgenerate/generate.go
//*/ᐳ

package templates // this is only here to make gofmt happy and is never in the generated code

//Ͼdefine "accountedqueue-imports"ᐳ

import (
	"github.com/FactomProject/factomd/common"
	"github.com/FactomProject/factomd/telemetry"
)

//Ͼendᐳ

//Ͼdefine "accountedqueue"ᐳ
// Start accountedqueue generated go code

type Ͼ_typenameᐳ struct {
	common.Name
	Channel chan Ͼ_typeᐳ
}

func (q *Ͼ_typenameᐳ) Init(parent common.NamedObject, name string, size int) *Ͼ_typenameᐳ {
	q.Name.Init(parent, name)
	q.Channel = make(chan Ͼ_typeᐳ, size)
	return q
}

// construct gauge w/ proper labels
func (q *Ͼ_typenameᐳ) Metric() telemetry.Gauge {
	return telemetry.ChannelSize.WithLabelValues(q.GetPath(), "current")
}

// construct counter for tracking totals
func (q *Ͼ_typenameᐳ) TotalMetric() telemetry.Counter {
	return telemetry.TotalCounter.WithLabelValues(q.GetPath(), "total")
}

// Length of underlying channel
func (q Ͼ_typenameᐳ) Length() int {
	return len(q.Channel)
}

// Cap of underlying channel
func (q Ͼ_typenameᐳ) Cap() int {
	return cap(q.Channel)
}

// Enqueue adds item to channel and instruments based on type
func (q Ͼ_typenameᐳ) Enqueue(m Ͼ_typeᐳ) {
	q.Channel <- m
	q.TotalMetric().Inc()
	q.Metric().Inc()
}

// Enqueue adds item to channel and instruments based on
// returns true it it enqueues the data
func (q Ͼ_typenameᐳ) EnqueueNonBlocking(m Ͼ_typeᐳ) bool {
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
func (q Ͼ_typenameᐳ) Dequeue() Ͼ_typeᐳ {
	select {
	case v := <-q.Channel:
		q.Metric().Dec()
		return v
	default:
		return nil
	}
}

// Dequeue removes an item from channel
func (q Ͼ_typenameᐳ) BlockingDequeue() Ͼ_typeᐳ {
	v := <-q.Channel
	q.Metric().Dec()
	return v
}

// End accountedqueue generated go code
// Ͼendᐳ
