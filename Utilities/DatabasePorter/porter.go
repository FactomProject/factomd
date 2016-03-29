// Copyright 2015 FactomProject Authors. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/hybridDB"
	"github.com/FactomProject/factomd/database/mapdb"
	"github.com/FactomProject/factomd/util"
)

func main() {
	fmt.Println("DatabasePorter")

	cfg := util.ReadConfig("")

	var dbo interfaces.DBOverlay

	switch cfg.App.DBType {
	case "Bolt":
		dbo = InitBolt(cfg)
		break
	case "LDB":
		dbo = InitLevelDB(cfg)
		break
	default:
		dbo = InitMapDB(cfg)
		break
	}

	fmt.Printf("dbo - %v", dbo)
}

func InitBolt(cfg *util.FactomdConfig) interfaces.DBOverlay {
	fmt.Println("InitBolt")
	path := cfg.App.BoltDBPath + "/"

	os.MkdirAll(path, 0777)
	dbase := hybridDB.NewBoltMapHybridDB(nil, path+"FactomBolt-Import.db")
	return databaseOverlay.NewOverlay(dbase)
}

func InitLevelDB(cfg *util.FactomdConfig) interfaces.DBOverlay {
	fmt.Println("InitLevelDB")
	path := cfg.App.LdbPath + "/" + "FactoidLevel-Import.db"

	dbase, err := hybridDB.NewLevelMapHybridDB(path, false)

	if err != nil || dbase == nil {
		dbase, err = hybridDB.NewLevelMapHybridDB(path, true)
		if err != nil {
			panic(err)
		}
	}

	return databaseOverlay.NewOverlay(dbase)
}

func InitMapDB(cfg *util.FactomdConfig) interfaces.DBOverlay {
	fmt.Println("InitMapDB")
	dbase := new(mapdb.MapDB)
	dbase.Init(nil)
	return databaseOverlay.NewOverlay(dbase)
}
