package queue

import (
	"bytes"
	"text/template"
)

// TODO: refactor to use an interface where we can add comments

var sourceFileFormat string = `package queue

import (
	"{{ .Import }}"
	"github.com/FactomProject/factomd/telemetry"
	"github.com/FactomProject/factomd/worker"
	"time"
)

type {{ .Name }} struct {
	Name    string
	Package string
	Channel chan {{ .Type }}
	Thread  *worker.Thread
}

// construct gauge w/ proper labels
func (q *{{ .Name }}) Metric(msg {{ .Type }}) telemetry.Gauge {
	label := "nil"
	if msg != nil {
		label = msg.Label()
	}

	return telemetry.ChannelSize.WithLabelValues(q.Package, q.Name, q.Thread.Label(), label)
}

// construct counter for tracking totals
func (q *{{ .Name }}) TotalMetric(msg {{ .Type }}) telemetry.Counter {
	label := "nil"
	if msg != nil {
		label = msg.Label()
	}

	return telemetry.TotalCounter.WithLabelValues(q.Package, q.Name, q.Thread.Label(), label)
}

// construct counter for intermittent polling of queue size
func (q *{{ .Name }}) PollMetric() telemetry.Gauge {
	return telemetry.ChannelSize.WithLabelValues(q.Package, q.Name, q.Thread.Label(), "aggregate")
}

// add metric to poll size of queue
func (q *{{ .Name }}) RegisterPollMetric() {
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
func (q {{ .Name }}) Length() int {
	return len(q.Channel)
}

// Cap of underlying channel
func (q {{ .Name }}) Cap() int {
	return cap(q.Channel)
}

// Enqueue adds item to channel and instruments based on type
func (q {{ .Name }}) Enqueue(m {{ .Type }}) {
	q.TotalMetric(m).Inc()
	q.Metric(m).Inc()
	q.Channel <- m
}

// Dequeue removes an item from channel and instruments based on type.
// Returns nil if nothing in // queue
func (q {{ .Name }}) Dequeue() {{ .Type }} {
	select {
	case v := <-q.Channel:
		q.Metric(v).Dec()
		return v
	default:
		return nil
	}
}

// BlockingDequeue will block until it retrieves from queue
func (q {{ .Name }}) BlockingDequeue() {{ .Type }} {
	v := <- q.Channel
	q.Metric(v).Dec()
	return v
}
`

var sourceTemplate *template.Template = template.Must(
	template.New("").Parse(sourceFileFormat),
)

type SourceFile struct {
	File string
	Name string
    Type string
	Import string
}

func (f SourceFile) Generate() *bytes.Buffer {
	b := &bytes.Buffer{}
	err := sourceTemplate.Execute(b, f)
	if nil != err {
		panic(err)
	}
	return b
}
