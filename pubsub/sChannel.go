package pubsub

import (
	"sync"
)

type Channel struct {
	SubBase
	Updates chan interface{}

	sync.RWMutex
}

func NewChannelBasedSubscriber(buffer int) *Channel {
	s := new(Channel)
	s.Updates = make(chan interface{}, buffer)

	return s
}

func (s *Channel) Write(o interface{}) {
	s.Updates <- o
}

func (s *Channel) Done() {
	close(s.Updates)
}

// Sub side

func (s *Channel) Channel() <-chan interface{} {
	return s.Updates
}

func (s *Channel) Receive() interface{} {
	v := <-s.Updates
	return v
}

func (s *Channel) ReceiveWithInfo() (interface{}, bool) {
	v, ok := <-s.Updates
	return v, ok
}

func (s *Channel) Subscribe(path string) *Channel {
	globalSubscribe(path, s)
	return s
}
