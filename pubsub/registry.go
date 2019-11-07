package pubsub

import (
	"fmt"
	"sync"

	"github.com/DiSiqueira/GoTree"
)

var globalReg *Registry

func init() {
	globalReg = NewRegistry()
}

func GlobalRegistry() *Registry {
	return globalReg
}

type Registry struct {
	Publishers map[string]IPublisher
	publock    sync.RWMutex

	// Add indexing
	// TODO: Should we keep the tree ok.
	tree gotree.Tree
}

func NewRegistry() *Registry {
	p := new(Registry)
	p.Publishers = make(map[string]IPublisher)
	p.tree = gotree.New("registry")

	return p
}

func (r *Registry) FindPublisher(path string) IPublisher {
	r.publock.RLock()
	defer r.publock.RUnlock()

	return r.Publishers[path]
}

func (r *Registry) Register(path string, pub IPublisher) error {
	r.publock.Lock()
	defer r.publock.Unlock()

	_, ok := r.Publishers[path]
	if ok {
		return fmt.Errorf("publisher already exists at that path")
	}

	pub.setPath(path)
	r.Publishers[path] = pub
	return nil
}

func (r *Registry) Remove(path string) {
	r.publock.Lock()
	defer r.publock.Unlock()

	delete(r.Publishers, path)
}

// SubscribeTo subscribes a subscriber to a specific publisher
func (r *Registry) SubscribeTo(path string, sub IPubSubscriber) error {
	r.publock.RLock()
	defer r.publock.RUnlock()

	pub, ok := r.Publishers[path]
	if !ok {
		return fmt.Errorf("path does not exist")
	}

	pub.Subscribe(sub)
	sub.SetUnsubscribe(func() {
		pub.Unsubscribe(sub)
	})

	return nil
}

func globalSubscribeWith(path string, sub IPubSubscriber, wrappers ...ISubscriberWrapper) IPubSubscriber {
	newsub := sub
	for _, wrap := range wrappers {
		newsub = wrap.Wrap(newsub)
	}

	globalSubscribe(path, newsub)
	return newsub
}

func globalPublish(path string, p IPublisher) IPublisher {
	err := globalReg.Register(path, p)
	if err != nil {
		panic(fmt.Sprintf("failed to publish: %s", err.Error()))
	}
	return p
}

func globalSubscribe(path string, sub IPubSubscriber) IPubSubscriber {
	err := globalReg.SubscribeTo(path, sub)
	if err != nil {
		panic(fmt.Sprintf("failed to subscribe: %s", err.Error()))
	}
	return sub
}
