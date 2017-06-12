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

//
// Can use custom structs to do instrumenting
//

type IPrometheusChannel interface {
	General(increment bool)

	EOM(increment bool)
	ACK(increment bool)
	AudFault(increment bool)
	FedFault(increment bool)
	FullFault(increment bool)
	CommitChain(increment bool)
	CommitEntry(increment bool)
	DBSig(increment bool)
	EOMTimeout(increment bool)
	FactTx(increment bool)
	Heartbeat(increment bool)
	EtcdHashPickup(increment bool)
	MissingMsg(increment bool)
	MissingMsgResp(increment bool)
	MissingData(increment bool)
	MissingDataResp(increment bool)
	RevealEntry(increment bool)
	DBStateMissing(increment bool)
	DBState(increment bool)
	Bounce(increment bool)
	BounceReply(increment bool)
	ReqBlock(increment bool)
	Misc(increment bool)
}

// measureMessage will increment/decrement prometheus based on type
func measureMessage(channel IPrometheusChannel, msg interfaces.IMsg, increment bool) {
	if msg == nil {
		return
	}
	channel.General(increment)
	switch msg.Type() {
	case constants.EOM_MSG: // 1
		channel.EOM(increment)
	case constants.ACK_MSG: // 2
		channel.ACK(increment)
	case constants.AUDIT_SERVER_FAULT_MSG: // 3
		channel.AudFault(increment)
	case constants.FED_SERVER_FAULT_MSG: // 4
		channel.FedFault(increment)
	case constants.FULL_SERVER_FAULT_MSG: // 5
		channel.FullFault(increment)
	case constants.COMMIT_CHAIN_MSG: // 6
		channel.CommitChain(increment)
	case constants.COMMIT_ENTRY_MSG: // 7
		channel.CommitEntry(increment)
	case constants.DIRECTORY_BLOCK_SIGNATURE_MSG: // 8
		channel.DBSig(increment)
	case constants.EOM_TIMEOUT_MSG: // 9
		channel.EOMTimeout(increment)
	case constants.FACTOID_TRANSACTION_MSG: // 10
		channel.FactTx(increment)
	case constants.HEARTBEAT_MSG: // 11
		channel.Heartbeat(increment)
	case constants.INVALID_DIRECTORY_BLOCK_MSG: // 12
		channel.EtcdHashPickup(increment)
	case constants.MISSING_MSG: // 13
		channel.MissingMsg(increment)
	case constants.MISSING_MSG_RESPONSE: // 14
		channel.MissingMsgResp(increment)
	case constants.MISSING_DATA: // 15
		channel.MissingData(increment)
	case constants.DATA_RESPONSE: // 16
		channel.MissingDataResp(increment)
	case constants.REVEAL_ENTRY_MSG: // 17
		channel.RevealEntry(increment)
	case constants.REQUEST_BLOCK_MSG: // 18
		channel.ReqBlock(increment)
	case constants.DBSTATE_MISSING_MSG: // 19
		channel.DBStateMissing(increment)
	case constants.DBSTATE_MSG: // 20
		channel.DBState(increment)
	case constants.BOUNCE_MSG: // 21
		channel.Bounce(increment)
	case constants.BOUNCEREPLY_MSG: // 22
		channel.BounceReply(increment)
	default: // 23
		channel.Misc(increment)
	}
}
