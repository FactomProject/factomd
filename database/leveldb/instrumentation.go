package leveldb

import (
	"github.com/FactomProject/factomd/telemetry"
)

var (
	LevelDBGets = telemetry.Counter(
		"factomd_database_leveldb_gets",
		"Counts gets from the database",
	)
	LevelDBPuts = telemetry.Counter(
		"factomd_database_leveldb_puts",
		"Count puts to the database",
	)
	LevelDBCacheblock = telemetry.Gauge(
		"factomd_database_leveldb_cacheblock",
		"Memory used by Level DB for caching",
	)
)
