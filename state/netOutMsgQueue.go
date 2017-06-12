package state

import (
	"github.com/FactomProject/factomd/common/interfaces"
)

// NetOutQueueRatePrometheus is for setting the appropriate prometheus calls
type NetOutQueueRatePrometheus struct{}

func (NetOutQueueRatePrometheus) SetArrivalInstantAvg(v float64) {
	NetOutInstantArrivalQueueRate.Set(v)
}
func (NetOutQueueRatePrometheus) SetArrivalTotalAvg(v float64) { NetOutTotalArrivalQueueRate.Set(v) }
func (NetOutQueueRatePrometheus) SetArrivalBackup(v float64)   { NetOutQueueBackupRate.Set(v) }
func (NetOutQueueRatePrometheus) SetCompleteInstantAvg(v float64) {
	NetOutInstantCompleteQueueRate.Set(v)
}
func (NetOutQueueRatePrometheus) SetCompleteTotalAvg(v float64) { NetOutTotalCompleteQueueRate.Set(v) }
func (NetOutQueueRatePrometheus) SetMovingArrival(v float64)    { NetOutMovingArrivalQueueRate.Set(v) }
func (NetOutQueueRatePrometheus) SetMovingComplete(v float64)   { NetOutMovingCompleteQueueRate.Set(v) }

// NetOutMsgQueue counts incoming and outgoing messages for netout queue
type NetOutMsgQueue chan interfaces.IMsg

func NewNetOutMsgQueue(capacity int) NetOutMsgQueue {
	channel := make(chan interfaces.IMsg, capacity)
	return channel
}

// Length of underlying channel
func (q NetOutMsgQueue) Length() int {
	return len(chan interfaces.IMsg(q))
}

// Cap of underlying channel
func (q NetOutMsgQueue) Cap() int {
	return cap(chan interfaces.IMsg(q))
}

// Enqueue adds item to channel and instruments based on type
func (q NetOutMsgQueue) Enqueue(m interfaces.IMsg) {
	//NetOutMsgQueueRateKeeper.Arrival()
	measureMessage(q, m, true)
	q <- m
}

// Dequeue removes an item from channel and instruments based on type. Returns nil if nothing in
// queue
func (q NetOutMsgQueue) Dequeue() interfaces.IMsg {
	select {
	case v := <-q:
		//NetOutMsgQueueRateKeeper.Complete()
		return v
	default:
		return nil
	}
}

// BlockingDequeue will block until it retrieves from queue
func (q NetOutMsgQueue) BlockingDequeue() interfaces.IMsg {
	v := <-q
	//NetOutMsgQueueRateKeeper.Complete()
	return v
}

//
// A list of all possible messages and their prometheus incrementing/decrementing
//

func (q NetOutMsgQueue) General(increment bool) {
	TotalMessageQueueNetOutMsgGeneral.Inc()
}

func (q NetOutMsgQueue) EOM(increment bool) {
	TotalMessageQueueNetOutMsgEOM.Inc()
}

func (q NetOutMsgQueue) ACK(increment bool) {
	TotalMessageQueueNetOutMsgACK.Inc()
}

func (q NetOutMsgQueue) AudFault(increment bool) {
	TotalMessageQueueNetOutMsgAudFault.Inc()
}
func (q NetOutMsgQueue) FedFault(increment bool) {
	TotalMessageQueueNetOutMsgFedFault.Inc()
}

func (q NetOutMsgQueue) FullFault(increment bool) {
	TotalMessageQueueNetOutMsgFullFault.Inc()
}

func (q NetOutMsgQueue) CommitChain(increment bool) {
	TotalMessageQueueNetOutMsgCommitChain.Inc()
}

func (q NetOutMsgQueue) CommitEntry(increment bool) {
	TotalMessageQueueNetOutMsgCommitEntry.Inc()
}

func (q NetOutMsgQueue) DBSig(increment bool) {
	TotalMessageQueueNetOutMsgDBSig.Inc()
}

func (q NetOutMsgQueue) EOMTimeout(increment bool) {
	TotalMessageQueueNetOutMsgEOMTimeout.Inc()
}

func (q NetOutMsgQueue) FactTx(increment bool) {
	TotalMessageQueueNetOutMsgFactTX.Inc()
}

func (q NetOutMsgQueue) Heartbeat(increment bool) {
	TotalMessageQueueNetOutMsgHeartbeat.Inc()
}

func (q NetOutMsgQueue) EtcdHashPickup(increment bool) {
	TotalMessageQueueNetOutMsgEtcdHashPickup.Inc()
}

func (q NetOutMsgQueue) MissingMsg(increment bool) {
	TotalMessageQueueNetOutMsgMissingMsg.Inc()
}

func (q NetOutMsgQueue) MissingMsgResp(increment bool) {
	TotalMessageQueueNetOutMsgMissingMsgResp.Inc()
}

func (q NetOutMsgQueue) MissingData(increment bool) {
	TotalMessageQueueNetOutMsgMissingData.Inc()
}

func (q NetOutMsgQueue) MissingDataResp(increment bool) {
	TotalMessageQueueNetOutMsgMissingDataResp.Inc()
}

func (q NetOutMsgQueue) RevealEntry(increment bool) {
	TotalMessageQueueNetOutMsgRevealEntry.Inc()
}

func (q NetOutMsgQueue) DBStateMissing(increment bool) {
	TotalMessageQueueNetOutMsgDbStateMissing.Inc()
}

func (q NetOutMsgQueue) DBState(increment bool) {
	TotalMessageQueueNetOutMsgDbState.Inc()
}

func (q NetOutMsgQueue) Bounce(increment bool) {
	TotalMessageQueueNetOutMsgBounceMsg.Inc()
}

func (q NetOutMsgQueue) BounceReply(increment bool) {
	TotalMessageQueueNetOutMsgBounceResp.Inc()
}

func (q NetOutMsgQueue) ReqBlock(increment bool) {
	TotalMessageQueueNetOutMsgReqBlock.Inc()
}

func (q NetOutMsgQueue) Misc(increment bool) {
	TotalMessageQueueNetOutMsgMisc.Inc()
}
