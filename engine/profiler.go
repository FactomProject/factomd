// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
)

// StartProfiler runs the go pprof tool
// `go tool pprof http://localhost:6060/debug/pprof/profile`
// https://golang.org/pkg/net/http/pprof/
func StartProfiler(mpr int) {
	_ = log.Print
	runtime.MemProfileRate = mpr
	log.Println(http.ListenAndServe(fmt.Sprintf("localhost:%s", logPort), nil))
	//runtime.SetBlockProfileRate(100000)
}

func launchPrometheus(port int) {
	http.Handle("/metrics", prometheus.Handler())
	go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
