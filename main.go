// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"github.com/FactomProject/FactomCode/factomd"
	"github.com/FactomProject/FactomCode/util"
)

func realMain() {
	util.Trace()
	factomd.Factomd_init()

	util.Trace()
	factomd.Factomd_main()

	util.Trace()
	btcd_main()

	util.Trace()
}

func main() {
	realMain()
}
