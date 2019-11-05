package subscribers

import (
	"sync"
)

type Channel struct {
	Base
	updates chan interface{}

	sync.RWMutex
}

func NewChannelBasedSubscriber(buffer int) *Channel {
	s := new(Channel)
	s.updates = make(chan interface{}, buffer)

	return s
}

func (s *Channel) Write(o interface{}) {
	s.updates <- o
}

func (s *Channel) Done() {
	close(s.updates)
}

// Sub side

func (s *Channel) Channel() <-chan interface{} {
	return s.updates
}

func (s *Channel) Receive() interface{} {
	v := <-s.updates
	return v
}

func (s *Channel) ReceiveWithInfo() (interface{}, bool) {
	v, ok := <-s.updates
	return v, ok
}
