package testHelper

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// DumpPrometheusMetric allows you to access the underlying values.
// This can be used to verify tests for example
func DumpPrometheusMetric(metric prometheus.Metric) *dto.Metric {
	dump := dto.Metric{}
	metric.Write(&dump)
	return &dump
}

func DumpPrometheusCounter(metric prometheus.Counter) float64 {
	d := DumpPrometheusMetric(metric)
	return *d.Counter.Value
}
