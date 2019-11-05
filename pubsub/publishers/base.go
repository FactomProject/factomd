package publishers

import (
	"sync"

	"github.com/Emyrk/pubsub/pubregistry"
)

// Base publisher has the basic necessary function implementations.
type Base struct {
	Subscribers []pubregistry.IPubSubscriber
	sync.RWMutex

	// path is set by registry
	path string
}

func (p *Base) SetPath(path string) { p.path = path }
func (p Base) Path() string         { return p.path }

func (p *Base) Close() error {
	p.Lock()
	for i := range p.Subscribers {
		p.Subscribers[i].Done()
	}
	p.Unlock()
	return nil
}

func (p *Base) NumberOfSubscribers() int {
	return len(p.Subscribers)
}

func (p *Base) Unsubscribe(subscriber pubregistry.IPubSubscriber) bool {
	p.Lock()
	defer p.Unlock()

	for i := range p.Subscribers {
		if p.Subscribers[i] == subscriber {
			newSlice := make([]pubregistry.IPubSubscriber, len(p.Subscribers)-1)
			copy(newSlice, p.Subscribers[:i])
			copy(newSlice[i:], p.Subscribers[i+1:])
			return true
		}
	}
	return false
}

func (p *Base) Subscribe(subscriber pubregistry.IPubSubscriber) bool {
	p.Lock()
	p.Subscribers = append(p.Subscribers, subscriber)
	p.Unlock()
	return true
}

func (p *Base) Write(o interface{}) {
	p.RLock()
	for i := range p.Subscribers {
		p.Subscribers[i].Write(o)
	}
	p.RUnlock()
}
