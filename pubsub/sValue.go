package pubsub

import (
	"sync"
)

// SubValue handles a single atomic value of the last write.
type SubValue struct {
	SubBase

	value interface{}

	sync.RWMutex
}

func NewSubValue() *SubValue {
	s := new(SubValue)

	return s
}

// Pub Side

func (s *SubValue) write(o interface{}) {
	s.Lock()
	s.value = o
	s.Unlock()
}

func (s *SubValue) Read() interface{} {
	s.RLock()
	defer s.RUnlock()
	return s.value
}

func (s *SubValue) Subscribe(path string, wrappers ...ISubscriberWrapper) *SubValue {
	globalSubscribeWith(path, s, wrappers...)
	return s
}

// UnsafeSubValue handles a single value of the last write.
type UnsafeSubValue struct {
	SubBase
	value interface{}
}

func NewUnsafeSubValue() *UnsafeSubValue {
	s := new(UnsafeSubValue)

	return s
}

// Pub Side

func (s *UnsafeSubValue) write(o interface{}) {
	s.value = o
}

func (s *UnsafeSubValue) Read() interface{} {
	return s.value
}

func (s *UnsafeSubValue) Subscribe(path string, wrappers ...ISubscriberWrapper) *UnsafeSubValue {
	globalSubscribeWith(path, s, wrappers...)
	return s
}
