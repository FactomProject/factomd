package pubsub

import (
	"sync"

	"github.com/FactomProject/factomd/modules/telemetry"
)

// SubPrometheusCounter, the total number of things written to the subscriber targeting a prometheus counter
type SubPrometheusCounter struct {
	SubBase

	prometheusCounter telemetry.Counter

	sync.RWMutex
}

func NewSubPrometheusCounter(name string, help ...string) *SubPrometheusCounter {
	s := new(SubPrometheusCounter)
	if len(help) == 0 {
		help = append(help, name)
	}
	s.prometheusCounter = telemetry.NewCounter(name, help[0])
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
