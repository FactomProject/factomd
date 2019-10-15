package state

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/worker"
)

// InMsgMSGQueue counts incoming and outgoing messages for inmsg queue
type InMsgMSGQueue = MsgQueue

func NewInMsgQueue(w *worker.Thread, capacity int) InMsgMSGQueue {
	mq := InMsgMSGQueue{
		Name:    "InMsgQueue",
		Channel: make(chan interfaces.IMsg, capacity),
		Thread:  w,
	}
	mq.RegisterPollMetric()
	return mq
}

func NewInMsgQueue2(w *worker.Thread, capacity int) InMsgMSGQueue {
	mq := InMsgMSGQueue{
		Name:    "InMsgQueue2",
		Channel: make(chan interfaces.IMsg, capacity),
		Thread:  w,
	}
	mq.RegisterPollMetric()
	return mq
}

// ElectionQueue counts incoming and outgoing messages for inmsg queue
type ElectionQueue = MsgQueue

func NewElectionQueue(w *worker.Thread, capacity int) ElectionQueue {
	mq := ElectionQueue{
		Name:    "ElectionQueue",
		Channel: make(chan interfaces.IMsg, capacity),
		Thread:  w,
	}
	mq.RegisterPollMetric()
	return mq
}

// NetOutMsgQueue counts incoming and outgoing messages for netout queue
type NetOutMsgQueue = MsgQueue

func NewNetOutMsgQueue(w *worker.Thread, capacity int) NetOutMsgQueue {
	mq := NetOutMsgQueue{
		Name:    "NetOutMsgQueue",
		Channel: make(chan interfaces.IMsg, capacity),
		Thread:  w,
	}
	mq.RegisterPollMetric()
	return mq
}

// APIMSGQueue counts incoming and outgoing messages for API queue
type APIMSGQueue = MsgQueue

func NewAPIQueue(w *worker.Thread, capacity int) APIMSGQueue {
	mq := APIMSGQueue{
		Name:    "APIMSGQueue",
		Channel: make(chan interfaces.IMsg, capacity),
		Thread:  w,
	}
	mq.RegisterPollMetric()
	return mq
}
