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
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"

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

func MsgTypeToLabel(msg interfaces.IMsg) string{
	switch msg.Type() {
	case constants.EOM_MSG: // 1
		return "eom"
	case constants.ACK_MSG: // 2
		return "ack"
	case constants.FULL_SERVER_FAULT_MSG: // 5
		return "fault"
	case constants.COMMIT_CHAIN_MSG: // 6
		return "commitchain"
	case constants.COMMIT_ENTRY_MSG: // 7
		return "commitentry"
	case constants.DIRECTORY_BLOCK_SIGNATURE_MSG: // 8
		return "dbsig"
	case constants.FACTOID_TRANSACTION_MSG: // 10
		return "factoid"
	case constants.HEARTBEAT_MSG: // 11
		return "heartbeat"
	case constants.MISSING_MSG: // 13
		return "missingmsg"
	case constants.MISSING_MSG_RESPONSE: // 14
		return "missingmsgresp"
	case constants.MISSING_DATA: // 15
		return "missingdata"
	case constants.DATA_RESPONSE: // 16
		return "dataresp"
	case constants.REVEAL_ENTRY_MSG: // 17
		return "revealentry"
	case constants.REQUEST_BLOCK_MSG: // 18
		return "requestblock"
	case constants.DBSTATE_MISSING_MSG: // 19
		return "dbstatmissing"
	case constants.DBSTATE_MSG: // 20
		return "dbstate"
	default: // 23
		return "misc"
	}
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
		counter.WithLabelValues(MsgTypeToLabel(msg)).Add(amt)
	}
}
