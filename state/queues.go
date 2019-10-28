package state

import (
	"github.com/FactomProject/factomd/queue"
	"github.com/FactomProject/factomd/worker"
)

// Now really sure the thread should be the parent but for now ...
func NewInMsgQueue(w *worker.Thread, capacity int) *queue.MsgQueue {
	return new(queue.MsgQueue).Init(w,"InMsgQueue", capacity)
}

func NewInMsgQueue2(w *worker.Thread, capacity int) *queue.MsgQueue {
	return new(queue.MsgQueue).Init(w, "InMsgQueue2", capacity)
}

func NewElectionQueue(w *worker.Thread, capacity int) *queue.MsgQueue {
	return new(queue.MsgQueue).Init(w, "ElectionQueue", capacity)
}

func NewNetOutMsgQueue(w *worker.Thread, capacity int) *queue.MsgQueue {
	return new(queue.MsgQueue).Init(w, "NetworkOutputQueue", capacity)
}

func NewAPIQueue(w *worker.Thread, capacity int) *queue.MsgQueue {
	return new(queue.MsgQueue).Init(w, "APInQueue", capacity)
}
