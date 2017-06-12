package state

import (
	"github.com/FactomProject/factomd/common/interfaces"
)

// InMsgQueueRatePrometheus is for setting the appropriate prometheus calls
type InMsgQueueRatePrometheus struct{}

func (InMsgQueueRatePrometheus) SetArrivalInstantAvg(v float64) { InMsgInstantArrivalQueueRate.Set(v) }
func (InMsgQueueRatePrometheus) SetArrivalTotalAvg(v float64)   { InMsgTotalArrivalQueueRate.Set(v) }
func (InMsgQueueRatePrometheus) SetArrivalBackup(v float64)     { InMsgQueueBackupRate.Set(v) }
func (InMsgQueueRatePrometheus) SetCompleteInstantAvg(v float64) {
	InMsgInstantCompleteQueueRate.Set(v)
}
func (InMsgQueueRatePrometheus) SetCompleteTotalAvg(v float64) { InMsgTotalCompleteQueueRate.Set(v) }
func (InMsgQueueRatePrometheus) SetMovingArrival(v float64)    { InMsgMovingArrivalQueueRate.Set(v) }
func (InMsgQueueRatePrometheus) SetMovingComplete(v float64)   { InMsgMovingCompleteQueueRate.Set(v) }

// InMsgMSGQueue counts incoming and outgoing messages for inmsg queue
type InMsgMSGQueue chan interfaces.IMsg

func NewInMsgQueue(capacity int) InMsgMSGQueue {
	channel := make(chan interfaces.IMsg, capacity)
	return channel
}

// Length of underlying channel
func (q InMsgMSGQueue) Length() int {
	return len(chan interfaces.IMsg(q))
}

// Cap of underlying channel
func (q InMsgMSGQueue) Cap() int {
	return cap(chan interfaces.IMsg(q))
}

// Enqueue adds item to channel and instruments based on type
func (q InMsgMSGQueue) Enqueue(m interfaces.IMsg) {
	//inMsgQueueRateKeeper.Arrival()
	measureMessage(q, m, true)
	q <- m
}

// Dequeue removes an item from channel and instruments based on type. Returns nil if nothing in
// queue
func (q InMsgMSGQueue) Dequeue() interfaces.IMsg {
	select {
	case v := <-q:
		measureMessage(q, v, false)
		//inMsgQueueRateKeeper.Complete()
		return v
	default:
		return nil
	}
}

// BlockingDequeue will block until it retrieves from queue
func (q InMsgMSGQueue) BlockingDequeue() interfaces.IMsg {
	v := <-q
	measureMessage(q, v, false)
	//inMsgQueueRateKeeper.Complete()
	return v
}

//
// A list of all possible messages and their prometheus incrementing/decrementing
//

func (q InMsgMSGQueue) General(increment bool) {
	if !increment {
		return
	}
	TotalMessageQueueInMsgGeneral.Inc()
}

func (q InMsgMSGQueue) EOM(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgEOM.Dec()
		return
	}
	CurrentMessageQueueInMsgEOM.Inc()
	TotalMessageQueueInMsgEOM.Inc()
}

func (q InMsgMSGQueue) ACK(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgACK.Dec()
		return
	}
	CurrentMessageQueueInMsgACK.Inc()
	TotalMessageQueueInMsgACK.Inc()
}

func (q InMsgMSGQueue) AudFault(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgAudFault.Dec()
		return
	}
	CurrentMessageQueueInMsgAudFault.Inc()
	TotalMessageQueueInMsgAudFault.Inc()
}
func (q InMsgMSGQueue) FedFault(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgFedFault.Dec()
		return
	}
	CurrentMessageQueueInMsgFedFault.Inc()
	TotalMessageQueueInMsgFedFault.Inc()
}

func (q InMsgMSGQueue) FullFault(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgFullFault.Dec()
		return
	}
	CurrentMessageQueueInMsgFullFault.Inc()
	TotalMessageQueueInMsgFullFault.Inc()
}

func (q InMsgMSGQueue) CommitChain(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgCommitChain.Dec()
		return
	}
	CurrentMessageQueueInMsgCommitChain.Inc()
	TotalMessageQueueInMsgCommitChain.Inc()
}

func (q InMsgMSGQueue) CommitEntry(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgCommitEntry.Dec()
		return
	}
	CurrentMessageQueueInMsgCommitEntry.Inc()
	TotalMessageQueueInMsgCommitEntry.Inc()
}

func (q InMsgMSGQueue) DBSig(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgDBSig.Dec()
		return
	}
	CurrentMessageQueueInMsgDBSig.Inc()
	TotalMessageQueueInMsgDBSig.Inc()
}

func (q InMsgMSGQueue) EOMTimeout(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgEOMTimeout.Dec()
		return
	}
	CurrentMessageQueueInMsgEOMTimeout.Inc()
	TotalMessageQueueInMsgEOMTimeout.Inc()
}

func (q InMsgMSGQueue) FactTx(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgFactTX.Dec()
		return
	}
	CurrentMessageQueueInMsgFactTX.Inc()
	TotalMessageQueueInMsgFactTX.Inc()
}

func (q InMsgMSGQueue) Heartbeat(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgHeartbeat.Dec()
		return
	}
	CurrentMessageQueueInMsgHeartbeat.Inc()
	TotalMessageQueueInMsgHeartbeat.Inc()
}

func (q InMsgMSGQueue) EtcdHashPickup(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgEtcdHashPickup.Dec()
		return
	}
	CurrentMessageQueueInMsgEtcdHashPickup.Inc()
	TotalMessageQueueInMsgEtcdHashPickup.Inc()
}

func (q InMsgMSGQueue) MissingMsg(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgMissingMsg.Dec()
		return
	}
	CurrentMessageQueueInMsgMissingMsg.Inc()
	TotalMessageQueueInMsgMissingMsg.Inc()
}

func (q InMsgMSGQueue) MissingMsgResp(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgMissingMsgResp.Dec()
		return
	}
	CurrentMessageQueueInMsgMissingMsgResp.Inc()
	TotalMessageQueueInMsgMissingMsgResp.Inc()
}

func (q InMsgMSGQueue) MissingData(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgMissingData.Dec()
		return
	}
	CurrentMessageQueueInMsgMissingData.Inc()
	TotalMessageQueueInMsgMissingData.Inc()
}

func (q InMsgMSGQueue) MissingDataResp(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgMissingDataResp.Dec()
		return
	}
	CurrentMessageQueueInMsgMissingDataResp.Inc()
	TotalMessageQueueInMsgMissingDataResp.Inc()
}

func (q InMsgMSGQueue) RevealEntry(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgRevealEntry.Dec()
		return
	}
	CurrentMessageQueueInMsgRevealEntry.Inc()
	TotalMessageQueueInMsgRevealEntry.Inc()
}

func (q InMsgMSGQueue) DBStateMissing(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgDbStateMissing.Dec()
		return
	}
	CurrentMessageQueueInMsgDbStateMissing.Inc()
	TotalMessageQueueInMsgDbStateMissing.Inc()
}

func (q InMsgMSGQueue) DBState(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgDbState.Dec()
		return
	}
	CurrentMessageQueueInMsgDbState.Inc()
	TotalMessageQueueInMsgDbState.Inc()
}

func (q InMsgMSGQueue) Bounce(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgBounceMsg.Inc()
		return
	}
	CurrentMessageQueueInMsgBounceMsg.Inc()
	TotalMessageQueueInMsgBounceMsg.Inc()
}

func (q InMsgMSGQueue) BounceReply(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgBounceResp.Dec()
		return
	}
	CurrentMessageQueueInMsgBounceResp.Inc()
	TotalMessageQueueInMsgBounceResp.Inc()
}

func (q InMsgMSGQueue) ReqBlock(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgReqBlock.Dec()
		return
	}
	CurrentMessageQueueInMsgReqBlock.Inc()
	TotalMessageQueueInMsgReqBlock.Inc()
}

func (q InMsgMSGQueue) Misc(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgMisc.Dec()
		return
	}
	CurrentMessageQueueInMsgMisc.Inc()
	TotalMessageQueueInMsgMisc.Inc()
}
