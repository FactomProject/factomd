package state

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

//
// Can use custom structs to do instrumenting
//

type IPrometheusChannel interface {
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
	InvalidDBlock(increment bool)
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

// InMsgMSGQueue counts incoming and outgoing messages for inmsg queue
type InMsgMSGQueue struct {
	GeneralMSGQueue
}

func NewInMsgQueue(capacity int) *InMsgMSGQueue {
	i := new(InMsgMSGQueue)
	channel := make(chan interfaces.IMsg, capacity)
	i.GeneralMSGQueue = channel
	return i
}

// Enqueue adds item to channel and instruments based on type
func (q InMsgMSGQueue) Enqueue(m interfaces.IMsg) {
	measureMessage(q, m, true)
	q.GeneralMSGQueue.Enqueue(m)
}

// Dequeue removes an item from channel and instruments based on type
func (q InMsgMSGQueue) Dequeue() interfaces.IMsg {
	v := q.GeneralMSGQueue.Dequeue()
	measureMessage(q, v, false)
	return v
}

func (q InMsgMSGQueue) EOM(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgQueueEOM.Dec()
		return
	}
	CurrentMessageQueueInMsgQueueEOM.Inc()
	TotalMessageQueueInMsgQueueEOM.Inc()
}

func (q InMsgMSGQueue) ACK(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgQueueACK.Dec()
		return
	}
	CurrentMessageQueueInMsgQueueACK.Inc()
	TotalMessageQueueInMsgQueueACK.Inc()
}

func (q InMsgMSGQueue) AudFault(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgQueueAudFault.Dec()
		return
	}
	CurrentMessageQueueInMsgQueueAudFault.Inc()
	TotalMessageQueueInMsgQueueAudFault.Inc()
}
func (q InMsgMSGQueue) FedFault(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgQueueFedFault.Dec()
		return
	}
	CurrentMessageQueueInMsgQueueFedFault.Inc()
	TotalMessageQueueInMsgQueueFedFault.Inc()
}

func (q InMsgMSGQueue) FullFault(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgQueueFullFault.Dec()
		return
	}
	CurrentMessageQueueInMsgQueueFullFault.Inc()
	TotalMessageQueueInMsgQueueFullFault.Inc()
}

func (q InMsgMSGQueue) CommitChain(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgQueueCommitChain.Dec()
		return
	}
	CurrentMessageQueueInMsgQueueCommitChain.Inc()
	TotalMessageQueueInMsgQueueCommitChain.Inc()
}

func (q InMsgMSGQueue) CommitEntry(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgQueueCommitEntry.Dec()
		return
	}
	CurrentMessageQueueInMsgQueueCommitEntry.Inc()
	TotalMessageQueueInMsgQueueCommitEntry.Inc()
}

func (q InMsgMSGQueue) DBSig(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgQueueDBSig.Dec()
		return
	}
	CurrentMessageQueueInMsgQueueDBSig.Inc()
	TotalMessageQueueInMsgQueueDBSig.Inc()
}

func (q InMsgMSGQueue) EOMTimeout(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgQueueEOMTimeout.Dec()
		return
	}
	CurrentMessageQueueInMsgQueueEOMTimeout.Inc()
	TotalMessageQueueInMsgQueueEOMTimeout.Inc()
}

func (q InMsgMSGQueue) FactTx(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgQueueFactTX.Dec()
		return
	}
	CurrentMessageQueueInMsgQueueFactTX.Inc()
	TotalMessageQueueInMsgQueueFactTX.Inc()
}

func (q InMsgMSGQueue) Heartbeat(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgQueueHeartbeat.Dec()
		return
	}
	CurrentMessageQueueInMsgQueueHeartbeat.Inc()
	TotalMessageQueueInMsgQueueHeartbeat.Inc()
}

func (q InMsgMSGQueue) InvalidDBlock(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgQueueInvalidDB.Dec()
		return
	}
	CurrentMessageQueueInMsgQueueInvalidDB.Inc()
	TotalMessageQueueInMsgQueueInvalidDB.Inc()
}

func (q InMsgMSGQueue) MissingMsg(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgQueueMissingMsg.Dec()
		return
	}
	CurrentMessageQueueInMsgQueueMissingMsg.Inc()
	TotalMessageQueueInMsgQueueMissingMsg.Inc()
}

func (q InMsgMSGQueue) MissingMsgResp(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgQueueMissingMsgResp.Dec()
		return
	}
	CurrentMessageQueueInMsgQueueMissingMsgResp.Inc()
	TotalMessageQueueInMsgQueueMissingMsgResp.Inc()
}

func (q InMsgMSGQueue) MissingData(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgQueueMissingData.Dec()
		return
	}
	CurrentMessageQueueInMsgQueueMissingData.Inc()
	TotalMessageQueueInMsgQueueMissingData.Inc()
}

func (q InMsgMSGQueue) MissingDataResp(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgQueueMissingDataResp.Dec()
		return
	}
	CurrentMessageQueueInMsgQueueMissingDataResp.Inc()
	TotalMessageQueueInMsgQueueMissingDataResp.Inc()
}

func (q InMsgMSGQueue) RevealEntry(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgQueueRevealEntry.Dec()
		return
	}
	CurrentMessageQueueInMsgQueueRevealEntry.Inc()
	TotalMessageQueueInMsgQueueRevealEntry.Inc()
}

func (q InMsgMSGQueue) DBStateMissing(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgQueueDbStateMissing.Dec()
		return
	}
	CurrentMessageQueueInMsgQueueDbStateMissing.Inc()
	TotalMessageQueueInMsgQueueDbStateMissing.Inc()
}

func (q InMsgMSGQueue) DBState(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgQueueDbState.Dec()
		return
	}
	CurrentMessageQueueInMsgQueueDbState.Inc()
	TotalMessageQueueInMsgQueueDbState.Inc()
}

func (q InMsgMSGQueue) Bounce(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgQueueBounceMsg.Inc()
		return
	}
	CurrentMessageQueueInMsgQueueBounceMsg.Inc()
	TotalMessageQueueInMsgQueueBounceMsg.Inc()
}

func (q InMsgMSGQueue) BounceReply(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgQueueBounceResp.Dec()
		return
	}
	CurrentMessageQueueInMsgQueueBounceResp.Inc()
	TotalMessageQueueInMsgQueueBounceResp.Inc()
}

func (q InMsgMSGQueue) ReqBlock(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgQueueReqBlock.Dec()
		return
	}
	CurrentMessageQueueInMsgQueueReqBlock.Inc()
	TotalMessageQueueInMsgQueueReqBlock.Inc()
}

func (q InMsgMSGQueue) Misc(increment bool) {
	if !increment {
		CurrentMessageQueueInMsgQueueMisc.Dec()
		return
	}
	CurrentMessageQueueInMsgQueueMisc.Inc()
	TotalMessageQueueInMsgQueueMisc.Inc()
}

// measureMessage will increment/decrement prometheus based on type
func measureMessage(channel IPrometheusChannel, msg interfaces.IMsg, increment bool) {
	if msg == nil {
		return
	}
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
		channel.InvalidDBlock(increment)
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
