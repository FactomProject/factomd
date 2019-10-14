package leveldb

import (
	"github.com/FactomProject/factomd/telemetry"
)

var (
	LevelDBGets = telemetry.NewCounter(
		"factomd_database_leveldb_gets",
		"Counts gets from the database",
	)
	LevelDBPuts = telemetry.NewCounter(
		"factomd_database_leveldb_puts",
		"Count puts to the database",
	)
	LevelDBCacheblock = telemetry.NewGauge(
		"factomd_database_leveldb_cacheblock",
		"Memory used by Level DB for caching",
	)
)
