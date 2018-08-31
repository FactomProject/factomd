package main

import (
	"fmt"
	"net/http"
	"runtime"

	log "github.com/sirupsen/logrus"
)

// StartProfiler runs the go pprof tool
// `go tool pprof http://localhost:6060/debug/pprof/profile`
// https://golang.org/pkg/net/http/pprof/
func StartProfiler(mpr int, expose bool) {
	_ = log.Print
	runtime.MemProfileRate = mpr
	pre := "localhost"
	if expose {
		pre = ""
	}
	url := fmt.Sprintf("%s:%s", pre, "6060")
	log.Infof("Profiling on %s", url)
	log.Println(http.ListenAndServe(url, nil))
}
