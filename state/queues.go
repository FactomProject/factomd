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
	"github.com/PaulSnow/factom2d/common/constants"
	"github.com/PaulSnow/factom2d/common/interfaces"

	"github.com/prometheus/client_golang/prometheus"
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
func measureMessage(counter *prometheus.GaugeVec, msg interfaces.IMsg, increment bool) {
	if msg == nil {
		return
	}
	amt := float64(1)
	if !increment {
		amt = -1
	}

	if counter != nil {
		switch msg.Type() {
		case constants.EOM_MSG: // 1
			counter.WithLabelValues("eom").Add(amt)
		case constants.ACK_MSG: // 2
			counter.WithLabelValues("ack").Add(amt)
		case constants.FULL_SERVER_FAULT_MSG: // 5
			counter.WithLabelValues("fault").Add(amt)
		case constants.COMMIT_CHAIN_MSG: // 6
			counter.WithLabelValues("commitchain").Add(amt)
		case constants.COMMIT_ENTRY_MSG: // 7
			counter.WithLabelValues("commitentry").Add(amt)
		case constants.DIRECTORY_BLOCK_SIGNATURE_MSG: // 8
			counter.WithLabelValues("dbsig").Add(amt)
		case constants.FACTOID_TRANSACTION_MSG: // 10
			counter.WithLabelValues("factoid").Add(amt)
		case constants.HEARTBEAT_MSG: // 11
			counter.WithLabelValues("heartbeat").Add(amt)
		case constants.MISSING_MSG: // 13
			counter.WithLabelValues("missingmsg").Add(amt)
		case constants.MISSING_MSG_RESPONSE: // 14
			counter.WithLabelValues("missingmsgresp").Add(amt)
		case constants.MISSING_DATA: // 15
			counter.WithLabelValues("missingdata").Add(amt)
		case constants.DATA_RESPONSE: // 16
			counter.WithLabelValues("dataresp").Add(amt)
		case constants.REVEAL_ENTRY_MSG: // 17
			counter.WithLabelValues("revealentry").Add(amt)
		case constants.REQUEST_BLOCK_MSG: // 18
			counter.WithLabelValues("requestblock").Add(amt)
		case constants.DBSTATE_MISSING_MSG: // 19
			counter.WithLabelValues("dbstatmissing").Add(amt)
		case constants.DBSTATE_MSG: // 20
			counter.WithLabelValues("dbstate").Add(amt)
		default: // 23
			counter.WithLabelValues("misc").Add(amt)
		}
	}
}
