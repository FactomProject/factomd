// Copyright 2015 Factom Foundation

package main

import (
	"github.com/FactomProject/FactomCode/factomd"
	"github.com/FactomProject/FactomCode/util"
)

func RealMain() {
	util.Trace()
	factomd.Factomd_init()
	util.Trace()
	factomd.Factomd_main()
	util.Trace()
	btcd_main()
	util.Trace()
}

func main() {
	RealMain()
}
