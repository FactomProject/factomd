package pubsub

var _ IPublisher = (*PubThreaded)(nil)

// PubThreaded handles all writes on a separate go routine.
type PubThreaded struct {
	PubBase

	handleWrite func(o interface{})
	inputs      chan interface{}
	subChanges  chan subAction
}

type action = int

const (
	COUNT action = iota + 1
	SUBSCRIBE
	UNSUBSCRIBE
)

type subAction struct {
	Subscriber IPubSubscriber
	Action     action
	Reply      chan interface{}
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
	return p.UnsubscribeSync(subscriber)
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
	return p.SubscribeSync(subscriber)
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
			switch action.Action {
			case UNSUBSCRIBE:
				p.unsubscribe(action.Subscriber)
			case SUBSCRIBE:
				p.subscribe(action.Subscriber)
			case COUNT:
			default:
				panic("UnknownAction")
			}
			if action.Reply != nil {
				action.Reply <- len(p.Subscribers)
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

func (p *PubThreaded) CountSubscriberSync(replyChannel ...chan interface{}) {
	var reply chan interface{}
	if len(replyChannel) == 1 {
		reply = replyChannel[0]
	} else {
		reply = make(chan interface{})
	}

	p.subChanges <- subAction{
		Subscriber: nil,
		Action:     COUNT,
		Reply:      reply,
	}
}

// optional reply channl allow waiting on async calls to unsubscribe
func (p *PubThreaded) UnsubscribeSync(subscriber IPubSubscriber, replyChannel ...chan interface{}) bool {

	var reply chan interface{}

	if len(replyChannel) == 1 {
		reply = replyChannel[0]
	}

	// Send the command to unsub
	p.subChanges <- subAction{
		Subscriber: subscriber,
		Action:     UNSUBSCRIBE,
		Reply:      reply,
	}
	return true
}

// optional reply channl allow waiting on async calls to subscribe
func (p *PubThreaded) SubscribeSync(subscriber IPubSubscriber, replyChannel ...chan interface{}) bool {
	var reply chan interface{}

	if len(replyChannel) == 1 {
		reply = replyChannel[0]
	}

	// Send the command to sub
	p.subChanges <- subAction{
		Subscriber: subscriber,
		Action:     SUBSCRIBE,
		Reply:      reply,
	}
	return true
}
