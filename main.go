// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"github.com/FactomProject/FactomCode/factomd"
	"github.com/FactomProject/FactomCode/util"
)

func RealMain() {
	util.Trace()
	factomd.Factomd_main()
	util.Trace()
	btcd_main()
	util.Trace()
}

func main() {
	RealMain()
}
