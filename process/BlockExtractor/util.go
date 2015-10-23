// Copyright 2015 FactomProject Authors. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package BlockExtractor

import (
	"fmt"
	"github.com/FactomProject/factomd/util"
	"github.com/davecgh/go-spew/spew"
	"io/ioutil"
	"os"
	"sort"

	. "github.com/FactomProject/factomd/common"
	. "github.com/FactomProject/factomd/common/adminBlock"
	. "github.com/FactomProject/factomd/common/directoryBlock"
	. "github.com/FactomProject/factomd/common/entryBlock"
	. "github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
)

var _ = util.Trace
var _ = spew.Sdump

var dataStorePath string
var dchain *DChain

func exportDChain(chain *DChain, db interfaces.DBOverlay) {
	// get all ecBlocks from db
	dBlocks, _ := db.FetchAllDBlocks()
	sort.Sort(util.ByDBlockIDAccending(dBlocks))

	for _, block := range dBlocks {

		data, err := block.MarshalBinary()
		if err != nil {
			panic(err)
		}

		strChainID := chain.ChainID.String()
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
}

func exportEChain(chain *EChain, db interfaces.DBOverlay) {
	eBlocks, _ := db.FetchAllEBlocksByChain(chain.ChainID)
	sort.Sort(util.ByEBlockIDAccending(eBlocks))

	for _, block := range eBlocks {

		data, err := block.MarshalBinary()
		if err != nil {
			panic(err)
		}

		strChainID := chain.ChainID.String()
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
}

func exportECChain(chain *ECChain, db interfaces.DBOverlay) {
	// get all ecBlocks from db
	ecBlocks, _ := db.FetchAllECBlocks()
	sort.Sort(util.ByECBlockIDAccending(ecBlocks))

	for _, block := range ecBlocks {
		data, err := block.MarshalBinary()
		if err != nil {
			panic(err)
		}

		strChainID := chain.ChainID.String()
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
}

func exportAChain(chain *AdminChain, db interfaces.DBOverlay) {
	// get all aBlocks from db
	aBlocks, _ := db.FetchAllABlocks()
	sort.Sort(util.ByABlockIDAccending(aBlocks))

	for _, block := range aBlocks {

		data, err := block.MarshalBinary()
		if err != nil {
			panic(err)
		}

		strChainID := chain.ChainID.String()
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
}

func exportFctChain(chain *FctChain, db interfaces.DBOverlay) {
	// get all aBlocks from db
	FBlocks, _ := db.FetchAllFBlocks()
	sort.Sort(util.ByFBlockIDAccending(FBlocks))

	for _, block := range FBlocks {

		data, err := block.MarshalBinary()
		if err != nil {
			panic(err)
		}

		strChainID := chain.ChainID.String()
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
}

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

func fileNotExists(name string) bool {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return true
	}
	return err != nil
}
