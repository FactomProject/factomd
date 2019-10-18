package state

import (
	"github.com/FactomProject/factomd/queue"
	"github.com/FactomProject/factomd/worker"
)

// Now really sure the thread shodul be the parent but for now ...
func NewInMsgQueue(w *worker.Thread, capacity int) *queue.MsgQueue {
	return new(queue.MsgQueue).Init(w, "state", w, "InMsgQueue", capacity)
}

func NewInMsgQueue2(w *worker.Thread, capacity int) *queue.MsgQueue {
	return new(queue.MsgQueue).Init(w, "state", w, "InMsgQueue2", capacity)
}

func NewElectionQueue(w *worker.Thread, capacity int) *queue.MsgQueue {
	return new(queue.MsgQueue).Init(w, "state", w, "ElectionQueue", capacity)
}

func NewNetOutMsgQueue(w *worker.Thread, capacity int) *queue.MsgQueue {
	return new(queue.MsgQueue).Init(w, "state", w, "NetworkOutputQueue", capacity)
}

func NewAPIQueue(w *worker.Thread, capacity int) *queue.MsgQueue {
	return new(queue.MsgQueue).Init(w, "state", w, "APInQueue", capacity)
}
