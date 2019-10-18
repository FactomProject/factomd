//+build ignore

var r map[string]string // dummy variable to make teh template delimiters work for gofmt...

r["/*
This looks syntatically off because it is a template used to generate go code. Inorder to make the template be
gofmt able the parse delimiters are set to 'r["'  and '"]' so r[" .typename "] will be replaced by the typename
from the //FactomGenerate command
*/"]

//r[" define "accountedqueue" "]
// Start accountedqueue generated go code

type r[" .typename "] struct {
	common.Name
	Package string
	Channel chan r[" .type "]
	Thread  *worker.Thread
}
// construct gauge w/ proper labels
func (q *r[" .typename "]) Metric(msg r[" .type "]) telemetry.Gauge {
	label := "nil"
	if msg != nil {
		label = msg.Label()
	}

	return telemetry.ChannelSize.WithLabelValues(q.Package, q.GetName(), q.Thread.Label(), label)
}

// construct counter for tracking totals
func (q *r[" .typename "]) TotalMetric(msg r[" .type "]) telemetry.Counter {
	label := "nil"
	if msg != nil {
		label = msg.Label()
	}

	return telemetry.TotalCounter.WithLabelValues(q.Package, q.GetName(), q.Thread.Label(), label)
}

// construct counter for intermittent polling of queue size
func (q *r[" .typename "]) PollMetric() telemetry.Gauge {
	return telemetry.ChannelSize.WithLabelValues(q.Package, q.GetName(), q.Thread.Label(), "aggregate")
}

// add metric to poll size of queue
func (q *r[" .typename "]) RegisterPollMetric() {
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
func (q r[" .typename "]) Length() int {
	return len(q.Channel)
}

// Cap of underlying channel
func (q r[" .typename "]) Cap() int {
	return cap(q.Channel)
}

// Enqueue adds item to channel and instruments based on type
func (q r[" .typename "]) Enqueue(m r[" .type "]) {
	q.TotalMetric(m).Inc()
	q.Metric(m).Inc()
	q.Channel <- m
}

// Dequeue removes an item from channel and instruments based on type.
// Returns nil if nothing in // queue
func (q r[" .typename "]) Dequeue() r[" .type "] {
	select {
	case v := <-q.Channel:
		q.Metric(v).Dec()
		return v
	default:
		return nil
	}
}

// BlockingDequeue will block until it retrieves from queue
func (q r[" .typename "]) BlockingDequeue() r[" .type "] {
	v := <- q.Channel
	q.Metric(v).Dec()
	return v
}
// End accountedqueue generated go code
r[" end "]"
