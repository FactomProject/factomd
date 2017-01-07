// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
)

// StartProfiler runs the go pprof tool
// `go tool pprof http://localhost:6060/debug/pprof/profile`
// https://golang.org/pkg/net/http/pprof/
func StartProfiler() {

	log.Println(http.ListenAndServe(fmt.Sprintf("localhost:%s", logPort), nil))
}
