package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
)

type MetricHandler interface {
	Counter(name string, help string) prometheus.Counter
	Gauge(name string, help string) prometheus.Gauge
	GaugeVec(name string, help string, labels []string) *prometheus.GaugeVec
	Summary(name string, help string) prometheus.Summary
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
	prometheus.MustRegister(c)
	return c
}

func (metric) Gauge(name string, help string) prometheus.Gauge {
	g := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	})
	prometheus.MustRegister(g)
	return g
}

func (metric) GaugeVec(name string, help string, labels []string) *prometheus.GaugeVec {
	v := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	}, labels)

	prometheus.MustRegister(v)
	return v
}

func (metric) Summary(name string, help string) prometheus.Summary {
	s := prometheus.NewSummary(prometheus.SummaryOpts{
		Name: name,
		Help: help,
	})

	prometheus.MustRegister(s)
	return s
}
