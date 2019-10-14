package state

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/worker"
)

// InMsgMSGQueue counts incoming and outgoing messages for inmsg queue
type InMsgMSGQueue = MsgQueue

func NewInMsgQueue(w *worker.Thread, capacity int) InMsgMSGQueue {
	return InMsgMSGQueue{
		name: "InMsgQueue",
		q:    make(chan interfaces.IMsg, capacity),
		w:    w,
	}
}

func NewInMsgQueue2(w *worker.Thread, capacity int) InMsgMSGQueue {
	return InMsgMSGQueue{
		name: "InMsgQueue2",
		q:    make(chan interfaces.IMsg, capacity),
		w:    w,
	}
}

// ElectionQueue counts incoming and outgoing messages for inmsg queue
type ElectionQueue = MsgQueue

func NewElectionQueue(w *worker.Thread, capacity int) ElectionQueue {
	return ElectionQueue{
		name: "InMsgQueue",
		q:    make(chan interfaces.IMsg, capacity),
		w:    w,
	}
}

// NetOutMsgQueue counts incoming and outgoing messages for netout queue
type NetOutMsgQueue = MsgQueue

func NewNetOutMsgQueue(w *worker.Thread, capacity int) NetOutMsgQueue {
	return NetOutMsgQueue{
		name: "InMsgQueue",
		q:    make(chan interfaces.IMsg, capacity),
		w:    w,
	}
}

// APIMSGQueue counts incoming and outgoing messages for API queue
type APIMSGQueue = MsgQueue

func NewAPIQueue(w *worker.Thread, capacity int) APIMSGQueue {
	return APIMSGQueue{
		name: "InMsgQueue",
		q:    make(chan interfaces.IMsg, capacity),
		w:    w,
	}
}
