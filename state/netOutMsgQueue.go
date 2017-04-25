package state

import (
	"github.com/FactomProject/factomd/common/interfaces"
)

// NetOutMsgQueue counts incoming and outgoing messages for inmsg queue
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
	measureMessage(q, m, true)
	q <- m
}

// Dequeue removes an item from channel and instruments based on type. Returns nil if nothing in
// queue
func (q NetOutMsgQueue) Dequeue() interfaces.IMsg {
	select {
	case v := <-q:
		return v
	default:
		return nil
	}
}

// BlockingDequeue will block until it retrieves from queue
func (q NetOutMsgQueue) BlockingDequeue() interfaces.IMsg {
	v := <-q
	return v
}

//
// A list of all possible messages and their prometheus incrementing/decrementing
//

func (q NetOutMsgQueue) EOM(increment bool) {
	if !increment {
		return
	}
	TotalMessageQueueNetOutMsgEOM.Inc()
}

func (q NetOutMsgQueue) ACK(increment bool) {
	if !increment {
		return
	}
	TotalMessageQueueNetOutMsgACK.Inc()
}

func (q NetOutMsgQueue) AudFault(increment bool) {
	if !increment {
		return
	}
	TotalMessageQueueNetOutMsgAudFault.Inc()
}
func (q NetOutMsgQueue) FedFault(increment bool) {
	if !increment {
		return
	}
	TotalMessageQueueNetOutMsgFedFault.Inc()
}

func (q NetOutMsgQueue) FullFault(increment bool) {
	if !increment {
		return
	}
	TotalMessageQueueNetOutMsgFullFault.Inc()
}

func (q NetOutMsgQueue) CommitChain(increment bool) {
	if !increment {
		return
	}
	TotalMessageQueueNetOutMsgCommitChain.Inc()
}

func (q NetOutMsgQueue) CommitEntry(increment bool) {
	if !increment {
		return
	}
	TotalMessageQueueNetOutMsgCommitEntry.Inc()
}

func (q NetOutMsgQueue) DBSig(increment bool) {
	if !increment {
		return
	}
	TotalMessageQueueNetOutMsgDBSig.Inc()
}

func (q NetOutMsgQueue) EOMTimeout(increment bool) {
	if !increment {
		return
	}
	TotalMessageQueueNetOutMsgEOMTimeout.Inc()
}

func (q NetOutMsgQueue) FactTx(increment bool) {
	if !increment {
		return
	}
	TotalMessageQueueNetOutMsgFactTX.Inc()
}

func (q NetOutMsgQueue) Heartbeat(increment bool) {
	if !increment {
		return
	}
	TotalMessageQueueNetOutMsgHeartbeat.Inc()
}

func (q NetOutMsgQueue) InvalidDBlock(increment bool) {
	if !increment {
		return
	}
	TotalMessageQueueNetOutMsgInvalidDB.Inc()
}

func (q NetOutMsgQueue) MissingMsg(increment bool) {
	if !increment {
		return
	}
	TotalMessageQueueNetOutMsgMissingMsg.Inc()
}

func (q NetOutMsgQueue) MissingMsgResp(increment bool) {
	if !increment {
		return
	}
	TotalMessageQueueNetOutMsgMissingMsgResp.Inc()
}

func (q NetOutMsgQueue) MissingData(increment bool) {
	if !increment {
		return
	}
	TotalMessageQueueNetOutMsgMissingData.Inc()
}

func (q NetOutMsgQueue) MissingDataResp(increment bool) {
	if !increment {
		return
	}
	TotalMessageQueueNetOutMsgMissingDataResp.Inc()
}

func (q NetOutMsgQueue) RevealEntry(increment bool) {
	if !increment {
		return
	}
	TotalMessageQueueNetOutMsgRevealEntry.Inc()
}

func (q NetOutMsgQueue) DBStateMissing(increment bool) {
	if !increment {
		return
	}
	TotalMessageQueueNetOutMsgDbStateMissing.Inc()
}

func (q NetOutMsgQueue) DBState(increment bool) {
	if !increment {
		return
	}
	TotalMessageQueueNetOutMsgDbState.Inc()
}

func (q NetOutMsgQueue) Bounce(increment bool) {
	if !increment {
		return
	}
	TotalMessageQueueNetOutMsgBounceMsg.Inc()
}

func (q NetOutMsgQueue) BounceReply(increment bool) {
	if !increment {
		return
	}
	TotalMessageQueueNetOutMsgBounceResp.Inc()
}

func (q NetOutMsgQueue) ReqBlock(increment bool) {
	if !increment {
		return
	}
	TotalMessageQueueNetOutMsgReqBlock.Inc()
}

func (q NetOutMsgQueue) Misc(increment bool) {
	if !increment {
		return
	}
	TotalMessageQueueNetOutMsgMisc.Inc()
}
