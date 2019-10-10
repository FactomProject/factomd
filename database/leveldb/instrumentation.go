package leveldb

import (
	"github.com/FactomProject/factomd/telemetry"
)

var counter = telemetry.RegisterMetric.Counter
var guage = telemetry.RegisterMetric.Gauge

var (
	LevelDBGets = counter(
		"factomd_database_leveldb_gets",
		"Counts gets from the database",
	)
	LevelDBPuts = counter(
		"factomd_database_leveldb_puts",
		"Count puts to the database",
	)
	LevelDBCacheblock = guage(
		"factomd_database_leveldb_cacheblock",
		"Memory used by Level DB for caching",
	)
)
