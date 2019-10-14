package state

import "github.com/FactomProject/factomd/common/interfaces"

// REVIEW: this is what we'd refactor

// InMsgMSGQueue counts incoming and outgoing messages for inmsg queue
type InMsgMSGQueue = MsgQueue

func NewInMsgQueue(capacity int) InMsgMSGQueue {
	return InMsgMSGQueue{
		make(chan interfaces.IMsg, capacity),
	}
}

// ElectionQueue counts incoming and outgoing messages for inmsg queue
type ElectionQueue = MsgQueue

func NewElectionQueue(capacity int) ElectionQueue {
	return ElectionQueue {
		make(chan interfaces.IMsg, capacity),
	}
}

// NetOutMsgQueue counts incoming and outgoing messages for netout queue
type NetOutMsgQueue = MsgQueue

func NewNetOutMsgQueue(capacity int) NetOutMsgQueue {
	return NetOutMsgQueue {
		make(chan interfaces.IMsg, capacity),
	}
}

// APIMSGQueue counts incoming and outgoing messages for API queue
type APIMSGQueue = MsgQueue

func NewAPIQueue(capacity int) APIMSGQueue {
	return APIMSGQueue {
		make(chan interfaces.IMsg, capacity),
	}
}
