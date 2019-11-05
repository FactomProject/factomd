package subscribers

import (
	"sync"
)

// Counts the total number of things written to the subscriber
type Counter struct {
	Base

	totalCount int64

	sync.RWMutex
}

func NewCounterSubscriber() *Counter {
	s := new(Counter)

	return s
}

// Pub Side

func (s *Counter) Write(o interface{}) {
	s.Lock()
	s.totalCount++
	s.Unlock()
}

func (s *Counter) Count() int64 {
	s.RLock()
	defer s.RUnlock()

	return s.totalCount
}
