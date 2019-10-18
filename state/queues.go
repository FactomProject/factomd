package state

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/queue"
	"github.com/FactomProject/factomd/worker"
)

func NewInMsgQueue(w *worker.Thread, capacity int) queue.MsgQueue {
	mq := queue.MsgQueue{
		Package: "state",
		Channel: make(chan interfaces.IMsg, capacity),
		Thread:  w,
	}
	mq.RegisterPollMetric()
	return mq
}

func NewInMsgQueue2(w *worker.Thread, capacity int) queue.MsgQueue {
	mq := queue.MsgQueue{
		Package: "state",
		Channel: make(chan interfaces.IMsg, capacity),
		Thread:  w,
	}
	mq.RegisterPollMetric()
	return mq
}

func NewElectionQueue(w *worker.Thread, capacity int) queue.MsgQueue {
	mq := queue.MsgQueue{
		Package: "state",
		Channel: make(chan interfaces.IMsg, capacity),
		Thread:  w,
	}
	mq.RegisterPollMetric()
	return mq
}

func NewNetOutMsgQueue(w *worker.Thread, capacity int) queue.MsgQueue {
	mq := queue.MsgQueue{
		Package: "state",
		Channel: make(chan interfaces.IMsg, capacity),
		Thread:  w,
	}
	mq.RegisterPollMetric()
	return mq
}

func NewAPIQueue(w *worker.Thread, capacity int) queue.MsgQueue {
	mq := queue.MsgQueue{
		Channel: make(chan interfaces.IMsg, capacity),
		Thread:  w,
	}
	mq.RegisterPollMetric()
	return mq
}
