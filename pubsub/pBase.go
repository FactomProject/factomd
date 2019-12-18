package pubsub

import (
	"sync"

	"github.com/FactomProject/factomd/common/interfaces"
)

var _ IPublisher = (*PubBase)(nil)

// PubBase publisher has the basic necessary function implementations.
type PubBase struct {
	Subscribers []IPubSubscriber
	sync.RWMutex

	// path is set by registry
	path string
	Log  interfaces.Log
}

func (p *PubBase) Publish(path string, wrappers ...IPublisherWrapper) IPublisher {
	return globalPublishWith(path, p, wrappers...)
}

func (p *PubBase) setPath(path string)          { p.path = path }
func (p PubBase) Path() string                  { return p.path }
func (p *PubBase) SetLogger(log interfaces.Log) { p.Log = log }
func (p PubBase) Logger() interfaces.Log        { return p.Log }

func (p *PubBase) Close() error {
	p.Lock()
	for i := range p.Subscribers {
		p.Subscribers[i].done()
	}
	p.Unlock()
	return nil
}

func (p *PubBase) NumberOfSubscribers() int {
	return len(p.Subscribers)
}

func (p *PubBase) Unsubscribe(subscriber IPubSubscriber) bool {
	p.Lock()
	defer p.Unlock()

	for i := range p.Subscribers {
		if p.Subscribers[i] == subscriber {
			newSlice := make([]IPubSubscriber, len(p.Subscribers)-1)
			copy(newSlice, p.Subscribers[:i])
			copy(newSlice[i:], p.Subscribers[i+1:])
			return true
		}
	}
	return false
}

func (p *PubBase) Subscribe(subscriber IPubSubscriber) bool {
	p.Lock()
	p.Subscribers = append(p.Subscribers, subscriber)
	p.Unlock()
	return true
}

func (p *PubBase) Write(o interface{}) {
	p.RLock()
	for i := range p.Subscribers {
		p.Subscribers[i].write(o)
	}
	p.RUnlock()
}

func (PubBase) Start() {
}
