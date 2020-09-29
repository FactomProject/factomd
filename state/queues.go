package state

import (
	"github.com/FactomProject/factomd/common"
	"github.com/FactomProject/factomd/queue"
)

// Now really sure the thread should be the parent but for now ...
func NewInMsgQueue(o common.NamedObject, capacity int) *queue.MsgQueue {
	return new(queue.MsgQueue).Init(o, "InMsgQueue", capacity)
}

func NewInMsgQueue2(o common.NamedObject, capacity int) *queue.MsgQueue {
	return new(queue.MsgQueue).Init(o, "InMsgQueue2", capacity)
}

func NewElectionQueue(o common.NamedObject, capacity int) *queue.MsgQueue {
	return new(queue.MsgQueue).Init(o, "ElectionQueue", capacity)
}

func NewNetOutMsgQueue(o common.NamedObject, capacity int) *queue.MsgQueue {
	return new(queue.MsgQueue).Init(o, "NetworkOutputQueue", capacity)
}

func NewAPIQueue(o common.NamedObject, capacity int) *queue.MsgQueue {
	return new(queue.MsgQueue).Init(o, "APInQueue", capacity)
}
