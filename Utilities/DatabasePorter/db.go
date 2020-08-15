package main

import (
	//"fmt"
	"os"

	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/database/databaseOverlay"
	"github.com/PaulSnow/factom2d/database/hybridDB"
	"github.com/PaulSnow/factom2d/database/mapdb"
	"github.com/PaulSnow/factom2d/util"
)

//DBInit

func InitBolt(cfg *util.FactomdConfig) interfaces.DBOverlay {
	//fmt.Println("InitBolt")
	path := cfg.App.BoltDBPath + "/"

	os.MkdirAll(path, 0777)
	dbase := hybridDB.NewBoltMapHybridDB(nil, path+"FactomBolt-Import.db")
	return databaseOverlay.NewOverlay(dbase)
}

func InitLevelDB(cfg *util.FactomdConfig) interfaces.DBOverlay {
	//fmt.Println("InitLevelDB")
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
	//fmt.Println("InitMapDB")
	dbase := new(mapdb.MapDB)
	dbase.Init(nil)
	return databaseOverlay.NewOverlay(dbase)
}
