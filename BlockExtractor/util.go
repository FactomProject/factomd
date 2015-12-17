// Copyright 2015 FactomProject Authors. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/util"
	"io/ioutil"
	"os"
	"sort"
	/*
		. "github.com/FactomProject/factomd/common"
		. "github.com/FactomProject/factomd/common/adminBlock"
		. "github.com/FactomProject/factomd/common/directoryBlock"
		. "github.com/FactomProject/factomd/common/entryBlock"
		. "github.com/FactomProject/factomd/common/entryCreditBlock"*/)

const level string = "level"
const bolt string = "bolt"

var dataStorePath string = ""

func main() {
	fmt.Println("Usage:")
	fmt.Println("BlockExtractor level/bolt [ChainID-To-Extract]")
	fmt.Println("Leave out the last one to export basic chains (A, D, EC, F)")
	if len(os.Args) < 1 {
		fmt.Println("\nNot enough arguments passed")
		os.Exit(1)
	}
	if len(os.Args) > 2 {
		fmt.Println("\nToo many arguments passed")
		os.Exit(1)
	}

	levelBolt := os.Args[1]

	if levelBolt != level && levelBolt != bolt {
		fmt.Println("\nFirst argument should be `level` or `bolt`")
		os.Exit(1)
	}

	chainID := ""
	if len(os.Args) == 3 {
		chainID = os.Args[2]
	}

	state := new(state.State)
	state.Cfg = util.ReadConfig("")
	if levelBolt == level {
		err := state.InitLevelDB()
		if err != nil {
			panic(err)
		}
	}
	if levelBolt == bolt {
		err := state.InitBoltDB()
		if err != nil {
			panic(err)
		}
	}
	dbo := state.GetDB()

	if chainID != "" {
		exportEChain(chainID, dbo)
	} else {
		exportDChain(dbo)
		exportECChain(dbo)
		exportAChain(dbo)
		exportFctChain(dbo)
	}
}

func fileNotExists(name string) bool {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return true
	}
	return err != nil
}

func SaveBinary(block interfaces.DatabaseBatchable) {
	data, err := block.MarshalBinary()
	if err != nil {
		panic(err)
	}

	strChainID := fmt.Sprintf("%x", block.GetChainID())
	if fileNotExists(dataStorePath + strChainID) {
		err := os.MkdirAll(dataStorePath+strChainID, 0777)
		if err == nil {
			fmt.Println("Created directory " + dataStorePath + strChainID)
		} else {
			panic(err)
		}
	}

	err = ioutil.WriteFile(fmt.Sprintf(dataStorePath+strChainID+"/store.%09d.block", block.GetDatabaseHeight()), data, 0777)
	if err != nil {
		panic(err)
	}
}

func SaveJSON(block interfaces.DatabaseBatchable) {
	data, err := block.(interfaces.Printable).JSONByte()
	if err != nil {
		panic(err)
	}

	var out bytes.Buffer
	json.Indent(&out, data, "", "\t")
	data = out.Bytes()

	strChainID := fmt.Sprintf("%x", block.GetChainID())
	if fileNotExists(dataStorePath + strChainID) {
		err := os.MkdirAll(dataStorePath+strChainID, 0777)
		if err == nil {
			fmt.Println("Created directory " + dataStorePath + strChainID)
		} else {
			panic(err)
		}
	}

	err = ioutil.WriteFile(fmt.Sprintf(dataStorePath+strChainID+"/storeJSON.%09d.block", block.GetDatabaseHeight()), data, 0777)
	if err != nil {
		panic(err)
	}
}

func exportEChain(chainID string, db interfaces.DBOverlay) {
	fmt.Printf("exportEChain %v\n", chainID)
	id, err := primitives.NewShaHashFromStr(chainID)
	if err != nil {
		panic(err)
	}
	eBlocks, err := db.FetchAllEBlocksByChain(id)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Fetched %v blocks\n", len(eBlocks))
	sort.Sort(util.ByEBlockIDAccending(eBlocks))

	for _, block := range eBlocks {
		SaveBinary(block.(interfaces.DatabaseBatchable))
		SaveJSON(block.(interfaces.DatabaseBatchable))
	}
}

func exportDChain(db interfaces.DBOverlay) {
	// get all ecBlocks from db
	dBlocks, err := db.FetchAllDBlocks()
	if err != nil {
		panic(err)
	}
	sort.Sort(util.ByDBlockIDAccending(dBlocks))

	for _, block := range dBlocks {
		SaveBinary(block.(interfaces.DatabaseBatchable))
		SaveJSON(block.(interfaces.DatabaseBatchable))
	}
}

func exportECChain(db interfaces.DBOverlay) {
	// get all ecBlocks from db
	ecBlocks, err := db.FetchAllECBlocks()
	if err != nil {
		panic(err)
	}
	sort.Sort(util.ByECBlockIDAccending(ecBlocks))

	for _, block := range ecBlocks {
		SaveBinary(block.(interfaces.DatabaseBatchable))
		SaveJSON(block.(interfaces.DatabaseBatchable))
	}
}

func exportAChain(db interfaces.DBOverlay) {
	// get all aBlocks from db
	aBlocks, err := db.FetchAllABlocks()
	if err != nil {
		panic(err)
	}
	sort.Sort(util.ByABlockIDAccending(aBlocks))

	for _, block := range aBlocks {
		SaveBinary(block.(interfaces.DatabaseBatchable))
		SaveJSON(block.(interfaces.DatabaseBatchable))
	}
}

func exportFctChain(db interfaces.DBOverlay) {
	// get all aBlocks from db
	fBlocks, err := db.FetchAllFBlocks()
	if err != nil {
		panic(err)
	}
	sort.Sort(util.ByFBlockIDAccending(fBlocks))

	for _, block := range fBlocks {
		SaveBinary(block.(interfaces.DatabaseBatchable))
		SaveJSON(block.(interfaces.DatabaseBatchable))
	}
}

/*

var dchain *DChain


// to export individual block once at a time - for debugging ------------------------
func exportDBlock(block *DirectoryBlock, db interfaces.DBOverlay) {
	if block == nil {
		//log.Println("no blocks to save for chain: " + string (*chain.ChainID))
		return
	}

	data, err := block.MarshalBinary()
	if err != nil {
		panic(err)
	}

	strChainID := dchain.ChainID.String()
	if fileNotExists(dataStorePath + strChainID) {
		err := os.MkdirAll(dataStorePath+strChainID, 0777)
		if err == nil {
			fmt.Println("Created directory " + dataStorePath + strChainID)
		} else {
			fmt.Println(err)
		}
	}
	err = ioutil.WriteFile(fmt.Sprintf(dataStorePath+strChainID+"/store.%09d.block", block.Header.DBHeight), data, 0777)
	if err != nil {
		panic(err)
	}

}

func exportEBlock(block *EBlock, db interfaces.DBOverlay) {
	if block == nil {
		return
	}

	data, err := block.MarshalBinary()
	if err != nil {
		panic(err)
	}

	strChainID := block.Header.ChainID.String()
	if fileNotExists(dataStorePath + strChainID) {
		err := os.MkdirAll(dataStorePath+strChainID, 0777)
		if err == nil {
			fmt.Println("Created directory " + dataStorePath + strChainID)
		} else {
			fmt.Println(err)
		}
	}

	err = ioutil.WriteFile(fmt.Sprintf(dataStorePath+strChainID+"/store.%09d.%09d.block", block.Header.EBSequence, block.Header.DBHeight), data, 0777)
	if err != nil {
		panic(err)
	}

}

func exportECBlock(block *ECBlock, db interfaces.DBOverlay) {
	if block == nil {
		return
	}

	data, err := block.MarshalBinary()
	if err != nil {
		panic(err)
	}

	strChainID := block.Header.ECChainID.String()
	if fileNotExists(dataStorePath + strChainID) {
		err := os.MkdirAll(dataStorePath+strChainID, 0777)
		if err == nil {
			fmt.Println("Created directory " + dataStorePath + strChainID)
		} else {
			fmt.Println(err)
		}
	}
	err = ioutil.WriteFile(fmt.Sprintf(dataStorePath+strChainID+"/store.%09d.block", block.Header.DBHeight), data, 0777)
	if err != nil {
		panic(err)
	}

}

func exportABlock(block *AdminBlock, db interfaces.DBOverlay) {
	if block == nil {
		return
	}

	data, err := block.MarshalBinary()
	if err != nil {
		panic(err)
	}

	strChainID := block.Header.AdminChainID.String()
	if fileNotExists(dataStorePath + strChainID) {
		err := os.MkdirAll(dataStorePath+strChainID, 0777)
		if err == nil {
			fmt.Println("Created directory " + dataStorePath + strChainID)
		} else {
			fmt.Println(err)
		}
	}
	err = ioutil.WriteFile(fmt.Sprintf(dataStorePath+strChainID+"/store.%09d.block", block.Header.DBHeight), data, 0777)
	if err != nil {
		panic(err)
	}

}

func exportFctBlock(block interfaces.IFBlock, db interfaces.DBOverlay) {
	if block == nil {
		return
	}

	data, err := block.MarshalBinary()
	if err != nil {
		panic(err)
	}

	strChainID := block.GetChainID().String()
	if fileNotExists(dataStorePath + strChainID) {
		err := os.MkdirAll(dataStorePath+strChainID, 0777)
		if err == nil {
			fmt.Println("Created directory " + dataStorePath + strChainID)
		} else {
			fmt.Println(err)
		}
	}
	err = ioutil.WriteFile(fmt.Sprintf(dataStorePath+strChainID+"/store.%09d.block", block.GetDBHeight()), data, 0777)
	if err != nil {
		panic(err)
	}

}
*/
