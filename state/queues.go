package state

//
// Addressing Performance
// 	IQueues replace channels and monitor enqueues and dequeues
// 	with prometheus instrumentation. By tripping a prometheus call,
// 	performance is lost, but compared to the insight gained, is worth it.
// 	The performance does not affect our queue management.
//
// Benchmarks :: `go test -bench=. queues_test.go `
// 	BenchmarkChannels-4            	20000000	        94.7 ns/op
// 	BenchmarkQueues-4              	10000000	       153 ns/op
// 	BenchmarkConcurentChannels-4   	10000000	       138 ns/op
// 	BenchmarkConcurrentQueues-4    	 5000000	       251 ns/op
// 	BenchmarkCompetingChannels-4   	 3000000	       360 ns/op
// 	BenchmarkCompetingQueues-4     	 1000000	      1302 ns/op

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/telemetry"
)

// Returning this is non-instrumented way
type GeneralMSGQueue chan interfaces.IMsg

// Length of underlying channel
func (q GeneralMSGQueue) Length() int {
	return len(chan interfaces.IMsg(q))
}

// Cap of underlying channel
func (q GeneralMSGQueue) Cap() int {
	return cap(chan interfaces.IMsg(q))
}

// Enqueue adds item to channel
func (q GeneralMSGQueue) Enqueue(t interfaces.IMsg) {
	q <- t
}

// Dequeue returns the channel dequeue
func (q GeneralMSGQueue) Dequeue() interfaces.IMsg {
	select {
	case v := <-q:
		return v
	default:
		return nil
	}
}

func (q GeneralMSGQueue) BlockingDequeue() interfaces.IMsg {
	return <-q
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
