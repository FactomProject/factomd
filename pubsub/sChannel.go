package pubsub

import (
	"sync"
)

type SubChannel struct {
	SubBase
	Updates chan interface{}

	sync.RWMutex
}

func NewSubChannel(buffer int) *SubChannel {
	s := new(SubChannel)
	s.Updates = make(chan interface{}, buffer)

	return s
}

func (s *SubChannel) write(o interface{}) {
	s.Updates <- o
}

func (s *SubChannel) done() {
	close(s.Updates)
}

// Sub side

func (s *SubChannel) Channel() <-chan interface{} {
	return s.Updates
}

func (s *SubChannel) Read() interface{} {
	v := <-s.Updates
	return v
}

func (s *SubChannel) ReadWithInfo() (interface{}, bool) {
	v, ok := <-s.Updates
	return v, ok
}

func (s *SubChannel) Subscribe(path string, wrappers ...ISubscriberWrapper) *SubChannel {
	globalSubscribeWith(path, s, wrappers...)
	return s
}
