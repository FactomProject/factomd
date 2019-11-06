package pubsub

import (
	"runtime"
	"sync"
)

// AtomicValue handles a single atomic value of the last write.
type AtomicValue struct {
	SubBase

	value interface{}

	sync.RWMutex
}

func NewAtomicValueSubscriber() *AtomicValue {
	s := new(AtomicValue)

	return s
}

// Pub Side

func (s *AtomicValue) Write(o interface{}) {
	s.Lock()
	s.value = o
	s.Unlock()
}

func (s *AtomicValue) Value() interface{} {
	runtime.Gosched()
	s.RLock()
	defer s.RUnlock()
	return s.value
}

func (s *AtomicValue) Subscribe(path string) *AtomicValue {
	globalSubscribe(path, s)
	return s
}
