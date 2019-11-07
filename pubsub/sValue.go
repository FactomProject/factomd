package pubsub

import (
	"sync"
)

// Read handles a single atomic value of the last write.
type Value struct {
	SubBase
	value interface{}
	sync.RWMutex
}

func NewValueSubscriber() *Value {
	s := new(Value)
	return s
}

// Pub Side

func (s *Value) Write(o interface{}) {
	s.Lock()
	s.value = o
	s.Unlock()
}

func (s *Value) Read() interface{} {
	s.RLock()
	defer s.RUnlock()
	return s.value
}

func (s *Value) Subscribe(path string) *Value {
	globalSubscribe(path, s)
	return s
}
