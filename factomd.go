// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"github.com/FactomProject/factomd/engine"
)

func main() {
	// uncomment StartProfiler() to run the pprof tool (for testing)
	engine.Factomd()
}
