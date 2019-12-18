package pubsub

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/DiSiqueira/GoTree"
)

var globalReg *Registry

func init() {
	Reset()
}

func Reset() {
	globalReg = NewRegistry()
}

func ResetGlobalRegistry() {
	globalReg = NewRegistry()
}

func GlobalRegistry() *Registry {
	return globalReg
}

type Registry struct {
	Publishers map[string]IPublisher
	// pubLock guards the map access
	publock sync.RWMutex

	// useLock guards the registry. Some publishers need
	// further coordination ontop of just a safe map
	useLock sync.RWMutex

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
	// TODO: Create a logger for the publisher file

	r.publock.Lock()
	defer r.publock.Unlock()

	_, ok := r.Publishers[path]
	if ok {
		return fmt.Errorf("publisher already exists at that path (%s)", path)
	}

	pub.setPath(path)
	r.Publishers[path] = pub
	r.AddPath(path)
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
	sub.setUnsubscribe(func() {
		pub.Unsubscribe(sub)
	})
	sub.setPublisher(pub)

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

func globalPublishWith(path string, p IPublisher, wrappers ...IPublisherWrapper) IPublisher {
	if len(wrappers) > 0 {
		// Check the first
		if _, ok := wrappers[0].(*PubMultiWrapper); ok && (0) != len(wrappers)-1 {
			panic("A mutli must always be the last wrapper")
		}

		newpub := wrappers[0].Wrap(p)
		for i, wrap := range wrappers[1:] {
			if _, ok := wrap.(*PubMultiWrapper); ok && (i+1) != len(wrappers)-1 {
				panic("A mutli must always be the last wrapper")
			}
			newpub = wrap.Wrap(newpub)
		}
		return newpub.Publish(path)
	}

	// No wrappers
	err := globalReg.Register(path, p)
	if err != nil {
		tree := globalReg.PrintTree()
		fmt.Printf("Publish Tree\n%s\n", tree)
		panic(fmt.Sprintf("failed to publish: %s", err.Error()))
	}
	return p
}

func globalPublish(path string, p IPublisher) IPublisher {
	fmt.Printf("globalPublish: %v\n", path)
	err := globalReg.Register(path, p)
	if err != nil {
		tree := globalReg.PrintTree()
		fmt.Printf("Publish Tree\n%s\n", tree)
		panic(fmt.Sprintf("failed to publish: %s %s", path, err.Error()))
	}
	return p
}

func globalSubscribe(path string, sub IPubSubscriber) IPubSubscriber {
	fmt.Printf("globalSubscribe: %v\n", path)

	err := globalReg.SubscribeTo(path, sub)
	if err != nil {
		panic(fmt.Sprintf("failed to subscribe: %s %s", path, err.Error()))
	}
	return sub
}

func GetPath(dirs ...string) string {
	return filepath.Join(dirs...)
}
