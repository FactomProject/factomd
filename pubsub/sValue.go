package pubsub

import (
	"runtime"
	"sync"
)

// SubValue handles a single atomic value of the last write.
type SubValue struct {
	SubBase

	value interface{}

	sync.RWMutex
}

func NewAtomicValueSubscriber() *SubValue {
	s := new(SubValue)

	return s
}

// Pub Side

func (s *SubValue) Write(o interface{}) {
	s.Lock()
	s.value = o
	s.Unlock()
}

func (s *SubValue) Value() interface{} {
	runtime.Gosched()
	s.RLock()
	defer s.RUnlock()
	return s.value
}

func (s *SubValue) Subscribe(path string, wrappers ...ISubscriberWrapper) *SubValue {
	globalSubscribeWith(path, s, wrappers...)
	return s
}
