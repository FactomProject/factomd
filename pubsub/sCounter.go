package pubsub

import (
	"sync"
)

// SubCounter the total number of things written to the subscriber
type SubCounter struct {
	SubBase

	totalCount int64

	sync.RWMutex
}

func NewCounterSubscriber() *SubCounter {
	s := new(SubCounter)

	return s
}

// Pub Side

func (s *SubCounter) Write(o interface{}) {
	s.Lock()
	s.totalCount++
	s.Unlock()
}

func (s *SubCounter) Count() int64 {
	s.RLock()
	defer s.RUnlock()

	return s.totalCount
}

func (s *SubCounter) Subscribe(path string, wrappers ...ISubscriberWrapper) *SubCounter {
	globalSubscribeWith(path, s, wrappers...)
	return s
}
