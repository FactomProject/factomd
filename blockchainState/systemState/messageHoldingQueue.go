// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package systemState

import (
	"sync"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
)

//TODO: do
//The holding queue is cleared via a timeout on each message (assuming messages are not acknowledged) or if they become or are invalid.
const ExpirationTime int64 = 5 * 60 * 1000
const ResendTime int64 = 20000

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

	if ack.SetExpireTime().GetTimeMilli() == 0 {
		ack.SetExpiretime(primitives.NewTimestampNow())
	}
	if ack.GetResendTime().GetTimeMilli() == 0 {
		ack.SetResendtime(primitives.NewTimestampNow())
	}

	mhq.Acks[ack.GetHash().String()] = ack
}

func (mhq *MessageHoldingQueue) AddMessage(msg interfaces.IMsg) {
	mhq.Init()
	mhq.Semaphore.Lock()
	defer mhq.Semaphore.Unlock()

	if msg.SetExpireTime().GetTimeMilli() == 0 {
		msg.SetExpiretime(primitives.NewTimestampNow())
	}
	if msg.GetResendTime().GetTimeMilli() == 0 {
		msg.SetResendtime(primitives.NewTimestampNow())
	}

	mhq.Messages[msg.GetHash().String()] = msg
}

func (mhq *MessageHoldingQueue) IsAcked(h interfaces.IHash) bool {
	mhq.Init()
	mhq.Semaphore.RLock()
	defer mhq.Semaphore.RUnlock()

	ack := mhq.Acks[h.String()]
	if ack != nil {
		return true
	}
	return false
}

func (mhq *MessageHoldingQueue) ClearOldMessages() {
	mhq.Init()

	now := primitives.NewTimestampNow().GetTimeMilli()
	for k, v := range mhq.Acks {
		exp := v.GetExpireTime().GetTimeMilli()
		if now-exp > ExpirationTime {
			delete(mhq.Acks, k)
		}
	}
	for k, v := range mhq.Messages {
		exp := v.GetExpireTime().GetTimeMilli()
		if now-exp > ExpirationTime {
			delete(mhq.Messages, k)
		}
	}
}

func (mhq *MessageHoldingQueue) GetMessagesForResend() []interfaces.IMsg {
	mhq.Init()
	answer := []interfaces.IMsg{}

	nowT := primitives.NewTimestampNow()
	now := nowT.GetTimeMilli()
	for k, v := range mhq.Acks {
		res := v.GetResendTime().GetTimeMilli()
		if now-res > ResendTime {
			v.SetResendTime(nowT)
			answer = append(answer, v)
		}
	}
	for k, v := range mhq.Messages {
		res := v.GetResendTime().GetTimeMilli()
		if now-res > ResendTime {
			v.SetResendTime(nowT)
			answer = append(answer, v)
		}
	}

	return answer
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
