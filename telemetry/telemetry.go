package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Gauge = prometheus.Gauge
type GaugeVec = prometheus.GaugeVec
type Counter = prometheus.Counter
type Summary = prometheus.Summary


func NewCounter(name string, help string) Counter {
	c := prometheus.NewCounter(prometheus.CounterOpts{
		Name: name,
		Help: help,
	})
	prometheus.MustRegister(c)
	return c
}

func NewGauge(name string, help string) Gauge {
	g := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	})
	prometheus.MustRegister(g)
	return g
}

func NewGaugeVec(name string, help string, labels []string) *GaugeVec {
	v := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	}, labels)
	prometheus.MustRegister(v)
	return v
}

func NewSummary(name string, help string) Summary {
	s := prometheus.NewSummary(prometheus.SummaryOpts{
		Name: name,
		Help: help,
	})
	prometheus.MustRegister(s)
	return s
}
