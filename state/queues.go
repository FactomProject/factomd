package state

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/queue"
)

func NewInMsgQueue(capacity int) queue.MsgQueue {
	mq := queue.MsgQueue{
		Channel: make(chan interfaces.IMsg, capacity),
	}
	mq.RegisterPollMetric()
	return mq
}

func NewInMsgQueue2(capacity int) queue.MsgQueue {
	mq := queue.MsgQueue{
		Channel: make(chan interfaces.IMsg, capacity),
	}
	mq.RegisterPollMetric()
	return mq
}

func NewElectionQueue(capacity int) queue.MsgQueue {
	mq := queue.MsgQueue{
		Channel: make(chan interfaces.IMsg, capacity),
	}
	mq.RegisterPollMetric()
	return mq
}

func NewNetOutMsgQueue(capacity int) queue.MsgQueue {
	mq := queue.MsgQueue{
		Channel: make(chan interfaces.IMsg, capacity),
	}
	mq.RegisterPollMetric()
	return mq
}

func NewAPIQueue(capacity int) queue.MsgQueue {
	mq := queue.MsgQueue{
		Channel: make(chan interfaces.IMsg, capacity),
	}
	mq.RegisterPollMetric()
	return mq
}
