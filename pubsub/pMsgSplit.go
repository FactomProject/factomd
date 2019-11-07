package pubsub

import (
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
)

type HasMsgHash interface {
	GetMsgHash() interfaces.IHash
}

// PubMsgSplit sends a msg to a subscriber based on it's hash value
type PubMsgSplit struct {
	*PubThreaded
}

func NewPubMsgSplit(buffer int) *PubMsgSplit {
	p := new(PubMsgSplit)
	p.PubThreaded = NewPubThreaded(buffer)
	p.PubThreaded.ChangeWriteHandle(p.write)

	return p
}

func (p *PubMsgSplit) write(o interface{}) {
	for len(p.Subscribers) == 0 {
		// TODO: This isn't the best way to handle this.
		// 		Someone can unsub after we exit this for too.
		time.Sleep(100 * time.Millisecond)
	}

	msg, ok := o.(HasMsgHash)
	if !ok {
		return // TODO: Do/say/log something?
	}

	// Adding a mutex to protect this from adding subscribers adds a
	// constant overhead per msg.
	index := int(msg.GetMsgHash().Fixed()[0]) % len(p.Subscribers)
	p.Subscribers[index].write(o)
}

func (p *PubMsgSplit) Run() {
	p.PubThreaded.Run()
}

func (p *PubMsgSplit) Publish(path string) *PubRoundRobin {
	globalPublish(path, p)
	return p
}
