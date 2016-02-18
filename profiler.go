// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
)

// StartProfiler runs pprof. get pprof data by running the go tool pprof
// commands.
// go tool pprof -pdf http://localhost:6060/debug/pprof/heap >/tmp/heap.pdf
// go tool pprof -pdf http://localhost:6060/debug/pprof/profile >/tmp/profile.pdf
// go tool pprof -pdf http://localhost:6060/debug/pprof/block >/tmp/block.pdf
func StartProfiler() {
	log.Println("DEBUG: Starting profiler")
	log.Println(http.ListenAndServe("localhost:6060", nil))
}
