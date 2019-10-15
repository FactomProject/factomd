package telemetry

import (
	"github.com/FactomProject/factomd/fnode"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type Gauge = prometheus.Gauge
type GaugeVec = prometheus.GaugeVec
type Counter = prometheus.Counter
type CounterVec = prometheus.CounterVec
type Summary = prometheus.Summary

type Handle func(*time.Ticker, chan bool)
type MetricHandler func(Handle)

var exit = make(chan bool)
var metricTicker = time.NewTicker(500 * time.Millisecond)

func init() {
	fnode.AddInterruptHandler(Exit) // trigger exit behavior in the case of SIGINT
}

// cause all polling metrics exit
func Exit() {
	close(exit)
}

// add a metric reporting goroutine
func RegisterMetric(handler Handle) {
	go handler(metricTicker, exit)
}

func NewCounter(name string, help string) Counter {
	c := prometheus.NewCounter(prometheus.CounterOpts{
		Name: name,
		Help: help,
	})
	prometheus.MustRegister(c)
	return c
}

func NewCounterVec(name string, help string, labels []string) *CounterVec {
	c := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: name,
		Help: help,
	}, labels)
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
