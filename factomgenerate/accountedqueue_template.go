//+build ignore

var r map[string]string // dummy variable to make the template delimiters work for gofmt...

r["/*
This looks syntatically off because it is a template used to generate go code. Inorder to make the template be
gofmt able the parse delimiters are set to 'r["'  and '"]' so r[" .typename "] will be replaced by the typename
from the //FactomGenerate command
*/"]

//r[" define "accountedqueue" "]
// Start accountedqueue generated go code

type r[" .typename "] struct {
	common.Name
	Channel chan r[" .type "]
}

func  (q *r[" .typename "]) Init(parent common.NamedObject, name string, size int) *r[" .typename "]{
    q.Name.Init(parent, name)
    q.Channel = make(chan r[" .type "], size)
    return q
}

// construct gauge w/ proper labels
func (q *r[" .typename "]) Metric() telemetry.Gauge {
	return telemetry.ChannelSize.WithLabelValues("state", q.GetPath(), "thread", "current")
}

// construct counter for tracking totals
func (q *r[" .typename "]) TotalMetric() telemetry.Counter {
	return telemetry.TotalCounter.WithLabelValues("state", q.GetPath(), "thread", "total")
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
	q.TotalMetric().Inc()
	q.Metric().Inc()
	q.Channel <- m
}

// Dequeue removes an item from channel and instruments based on type.
// Returns nil if nothing in // queue
func (q r[" .typename "]) DequeueNonBlocking() r[" .type "] {
	select {
	case v := <-q.Channel:
		q.Metric().Dec()
		return v
	default:
		return nil
	}
}

// Dequeue removes an item from channel and instruments based on type.
// Returns nil if nothing in // queue
func (q r[" .typename "]) Dequeue() r[" .type "] {
        v := <-q.Channel
		q.Metric().Dec()
		return v
}

// BlockingDequeue will block until it retrieves from queue
func (q r[" .typename "]) BlockingDequeue() r[" .type "] {
	v := <- q.Channel
	q.Metric().Dec()
	return v
}
// End accountedqueue generated go code
r[" end "]"
