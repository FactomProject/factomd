package pubsub

import (
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
)

// PubSelector only sends events to 1 subscriber on depending on the select
// function.
type PubSelector struct {
	*PubThreaded
	selector ISelector
	next     int
}

func NewPubSelector(buffer int, selector ISelector) *PubSelector {
	p := new(PubSelector)
	p.PubThreaded = NewPubThreaded(buffer)
	p.PubThreaded.ChangeWriteHandle(p.write)
	p.selector = selector

	return p
}

func (p *PubSelector) write(o interface{}) {
	for len(p.Subscribers) == 0 {
		// TODO: This isn't the best way to handle this.
		// 		Someone can unsub after we exit this for too.
		time.Sleep(100 * time.Millisecond)
	}
	index := p.selector.Select(len(p.Subscribers), o)
	if index >= 0 {
		p.Subscribers[index].write(o)
	}
}

func (p *PubSelector) Start() {
	p.PubThreaded.Start()
}

func (p *PubSelector) Publish(path string) *PubSelector {
	globalPublish(path, p)
	return p
}

type ISelector interface {
	Select(total int, o interface{}) int
}

type RoundRobinSelector struct {
	next int
}

func (r *RoundRobinSelector) Select(total int, _ interface{}) int {
	s := r.next % total
	r.next++
	return s
}

type MsgSplitSelector struct{}

type HasMsgHash interface {
	GetMsgHash() interfaces.IHash
}

func (r *MsgSplitSelector) Select(total int, o interface{}) int {
	msg, ok := o.(HasMsgHash)
	if !ok {
		return -1 // TODO: Do/say/log something?
	}

	index := int(msg.GetMsgHash().Fixed()[0]) % total
	return index
}
