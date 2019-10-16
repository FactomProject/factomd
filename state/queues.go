package state

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/queue"
	"github.com/FactomProject/factomd/worker"
)

func NewInMsgQueue(w *worker.Thread, capacity int) queue.MsgQueue {
	mq := queue.MsgQueue{
		Name:    "InMsgQueue",
		Package: "state",
		Channel: make(chan interfaces.IMsg, capacity),
		Thread:  w,
	}
	mq.RegisterPollMetric()
	return mq
}

func NewInMsgQueue2(w *worker.Thread, capacity int) queue.MsgQueue {
	mq := queue.MsgQueue{
		Name:    "InMsgQueue2",
		Package: "state",
		Channel: make(chan interfaces.IMsg, capacity),
		Thread:  w,
	}
	mq.RegisterPollMetric()
	return mq
}

func NewElectionQueue(w *worker.Thread, capacity int) queue.MsgQueue {
	mq := queue.MsgQueue{
		Name:    "ElectionQueue",
		Package: "state",
		Channel: make(chan interfaces.IMsg, capacity),
		Thread:  w,
	}
	mq.RegisterPollMetric()
	return mq
}

func NewNetOutMsgQueue(w *worker.Thread, capacity int) queue.MsgQueue {
	mq := queue.MsgQueue{
		Name:    "NetOutMsgQueue",
		Package: "state",
		Channel: make(chan interfaces.IMsg, capacity),
		Thread:  w,
	}
	mq.RegisterPollMetric()
	return mq
}

func NewAPIQueue(w *worker.Thread, capacity int) queue.MsgQueue {
	mq := queue.MsgQueue{
		Name:    "APIMSGQueue",
		Channel: make(chan interfaces.IMsg, capacity),
		Thread:  w,
	}
	mq.RegisterPollMetric()
	return mq
}
