package pubsub

var _ IPublisher = (*PubThreaded)(nil)

// PubThreaded handles all writes on a separate go routine.
type PubThreaded struct {
	PubBase

	handleWrite func(o interface{})
	inputs      chan interface{}
	subChanges  chan subAction
}

type subAction struct {
	Subscriber IPubSubscriber
	Subscribe  bool
}

func NewPubThreaded(buffer int) *PubThreaded {
	p := new(PubThreaded)
	p.inputs = make(chan interface{}, buffer)
	p.subChanges = make(chan subAction, 10)
	p.handleWrite = func(o interface{}) {
		p.write(o)
	}

	return p
}

func (p *PubThreaded) Close() error {
	close(p.inputs)
	return nil
}

func (p *PubThreaded) Write(o interface{}) {
	p.inputs <- o
}

func (p *PubThreaded) Unsubscribe(subscriber IPubSubscriber) bool {
	// Send the command to unsub
	p.subChanges <- subAction{
		Subscriber: subscriber,
		Subscribe:  false,
	}
	return true
}

func (p *PubThreaded) unsubscribe(subscriber IPubSubscriber) (ok bool) {
	for i := range p.Subscribers {
		if p.Subscribers[i] == subscriber {
			newSlice := make([]IPubSubscriber, len(p.Subscribers)-1)
			copy(newSlice, p.Subscribers[:i])
			copy(newSlice[i:], p.Subscribers[i+1:])
			p.Subscribers = newSlice
			return true
		}
	}
	return false
}

func (p *PubThreaded) Subscribe(subscriber IPubSubscriber) bool {
	// Send the command to sub
	p.subChanges <- subAction{
		Subscriber: subscriber,
		Subscribe:  true,
	}
	return true
}

func (p *PubThreaded) subscribe(subscriber IPubSubscriber) bool {
	p.Subscribers = append(p.Subscribers, subscriber)
	return true
}

func (p *PubThreaded) ChangeWriteHandle(handle func(o interface{})) {
	p.handleWrite = handle
}

// Run handles all changes to the threaded state and writes. Since Threaded
// has a single thread to handle all state changes, no mutexs are used.
// All subscribe/unsubscribe changes are handled alongside writes to ensure
// the writes can be handles in a threadsafe context with all publisher state
// access.
func (p *PubThreaded) Start() {
ThreadedRunLoop:
	for {
		select {
		case in, open := <-p.inputs:
			if !open {
				break ThreadedRunLoop
			}
			p.handleWrite(in)
		case action := <-p.subChanges:
			if action.Subscribe {
				p.subscribe(action.Subscriber)
			} else {
				p.unsubscribe(action.Subscriber)
			}
		}
	}

	// Close when out of things to write and channel is closed
	_ = p.PubBase.Close()
}

func (p *PubThreaded) write(o interface{}) {
	p.PubBase.Write(o)
}

func (p *PubThreaded) Publish(path string, wrappers ...IPublisherWrapper) IPublisher {
	return globalPublishWith(path, p, wrappers...)
}
