package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

var registeredMetrics = make(map[string]prometheus.Collector)

// Don't let other packages reference prometheus directly
type Counter = prometheus.Counter
type Gauge = prometheus.Gauge
type GaugeVec = prometheus.GaugeVec
type Summary = prometheus.Summary

type MetricHandler interface {
	Counter(name string, help string) Counter
	Gauge(name string, help string) Gauge
	GaugeVec(name string, help string, labels []string) *GaugeVec
	Summary(name string, help string) Summary
}

type metric struct {
	MetricHandler
}

var RegisterMetric = metric{}

func (metric) Counter(name string, help string) prometheus.Counter {
	c := prometheus.NewCounter(prometheus.CounterOpts{
		Name: name,
		Help: help,
	})
	registeredMetrics[name] = c
	return c
}

func (metric) Gauge(name string, help string) prometheus.Gauge {
	g := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	})
	registeredMetrics[name] = g
	return g
}

func (metric) GaugeVec(name string, help string, labels []string) *prometheus.GaugeVec {
	v := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	}, labels)

	registeredMetrics[name] = v
	return v
}

func (metric) Summary(name string, help string) prometheus.Summary {
	s := prometheus.NewSummary(prometheus.SummaryOpts{
		Name: name,
		Help: help,
	})

	registeredMetrics[name] = s
	return s
}

var registered sync.Once

func RegisterPrometheus() {
	registered.Do(func() {
		for _, m := range registeredMetrics {
			prometheus.MustRegister(m)
		}
	})
}
