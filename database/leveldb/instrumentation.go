package leveldb

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	LevelDBGets = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_database_leveldb_gets",
		Help: "Counts gets from the database",
	})
	LevelDBPuts = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "factomd_database_leveldb_puts",
		Help: "Count puts to the database",
	})
	LevelDBCacheblock = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "factomd_database_leveldb_cacheblock",
		Help: "Memory used by Level DB for caching",
	})
)

var registered = false

// RegisterPrometheus registers the variables to be exposed. This can only be run once, hence the
// boolean flag to prevent panics if launched more than once. This is called in NetStart
func RegisterPrometheus() {
	if registered {
		return
	}
	registered = true

	// LevelDB
	prometheus.MustRegister(LevelDBGets)
	prometheus.MustRegister(LevelDBPuts)
	prometheus.MustRegister(LevelDBCacheblock)
}
