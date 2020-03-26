package pubsub

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

// SubPrometheusCounter, the total number of things written to the subscriber targeting a prometheus counter
type SubPrometheusCounter struct {
	SubBase

	prometheusCounter prometheus.Counter

	sync.RWMutex
}

func NewSubPrometheusCounter(name string, help ...string) *SubPrometheusCounter {
	s := new(SubPrometheusCounter)
	opts := prometheus.CounterOpts{Name: name}
	if len(help) > 0 {
		opts.Help = help[0]
	}
	s.prometheusCounter = prometheus.NewCounter(opts)
	return s
}

// Pub Side

func (s *SubPrometheusCounter) write(o interface{}) {
	s.Lock()
	s.prometheusCounter.Inc()
	s.Unlock()
}

func (s *SubPrometheusCounter) Subscribe(path string, wrappers ...ISubscriberWrapper) *SubPrometheusCounter {
	globalSubscribeWith(path, s, wrappers...)
	return s
}
