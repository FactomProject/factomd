// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
)

func StartProfiler() {
	log.Println("DEBUG: Starting profiler")
	log.Println(http.ListenAndServe("localhost:6060", nil))
}
