// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package systemState

import (
	"sync"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

//TODO: do
//The holding queue is cleared via a timeout on each message (assuming messages are not acknowledged) or if they become or are invalid.

type MessageHoldingQueue struct {
	//Indexed by message hash
	Messages map[string]interfaces.IMsg
	//Inexed by the message hash of the message they are ACKing
	Acks map[string]interfaces.IMsg

	Semaphore sync.RWMutex
}

func (mhq *MessageHoldingQueue) Init() {
	mhq.Semaphore.Lock()
	defer mhq.Semaphore.Unlock()

	if mhq.Messages == nil {
		mhq.Messages = map[string]interfaces.IMsg{}
	}
	if mhq.Acks == nil {
		mhq.Acks = map[string]interfaces.IMsg{}
	}
}

func (mhq *MessageHoldingQueue) AddAck(ack interfaces.IMsg) {
	mhq.Init()
	mhq.Semaphore.Lock()
	defer mhq.Semaphore.Unlock()

	mhq.Acks[ack.GetHash().String()] = ack
}

func (mhq *MessageHoldingQueue) AddMessage(msg interfaces.IMsg) {
	mhq.Init()
	mhq.Semaphore.Lock()
	defer mhq.Semaphore.Unlock()

	mhq.Messages[msg.GetHash().String()] = msg
}

func (mhq *MessageHoldingQueue) IsAcked(msg interfaces.IMsg) bool {
	mhq.Init()
	mhq.Semaphore.RLock()
	defer mhq.Semaphore.RUnlock()

	ack := mhq.Acks[msg.GetHash().String()]
	if ack != nil {
		return true
	}
	return false
}

func (mhq *MessageHoldingQueue) ClearOldMessages() {
	mhq.Init()
	//TODO: do
}

func (mhq *MessageHoldingQueue) GetMessage(h interfaces.IHash) interfaces.IMsg {
	mhq.Init()
	mhq.Semaphore.RLock()
	defer mhq.Semaphore.RUnlock()

	return mhq.Messages[h.String()]
}

func (mhq *MessageHoldingQueue) GetAck(h interfaces.IHash) interfaces.IMsg {
	mhq.Init()
	mhq.Semaphore.RLock()
	defer mhq.Semaphore.RUnlock()

	return mhq.Acks[h.String()]
}
