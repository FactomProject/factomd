// Copyright 2015 FactomProject Authors. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package blockExtractor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/util"
	"io/ioutil"
	"os"
	"sort"
)

var DataStorePath string = ""

func FileNotExists(name string) bool {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return true
	}
	return err != nil
}

func SaveBinary(block interfaces.DatabaseBatchable) error {
	data, err := block.MarshalBinary()
	if err != nil {
		return err
	}

	strChainID := fmt.Sprintf("%x", block.GetChainID().Bytes())
	dir := DataStorePath + strChainID
	if FileNotExists(dir) {
		err := os.MkdirAll(dir, 0777)
		if err == nil {
			fmt.Println("Created directory " + dir)
		} else {
			return err
		}
	}

	err = ioutil.WriteFile(fmt.Sprintf(dir+"/store.%09d.block", block.GetDatabaseHeight()), data, 0777)
	if err != nil {
		return err
	}
	return nil
}

func SaveEntryBinary(entry interfaces.DatabaseBatchable, blockHeight uint32) error {
	data, err := entry.MarshalBinary()
	if err != nil {
		return err
	}

	strChainID := fmt.Sprintf("%x", entry.GetChainID().Bytes())
	dir := DataStorePath + strChainID + "/entries"
	if FileNotExists(dir) {
		err := os.MkdirAll(dir, 0777)
		if err == nil {
			fmt.Println("Created directory " + dir)
		} else {
			return err
		}
	}

	err = ioutil.WriteFile(fmt.Sprintf(dir+"/store.%09d.%v.entry", blockHeight, entry.DatabasePrimaryIndex().String()), data, 0777)
	if err != nil {
		return err
	}
	return nil
}

func SaveJSON(block interfaces.DatabaseBatchable) error {
	data, err := block.(interfaces.Printable).JSONByte()
	if err != nil {
		return err
	}

	var out bytes.Buffer
	json.Indent(&out, data, "", "\t")
	data = out.Bytes()

	strChainID := fmt.Sprintf("%x", block.GetChainID().Bytes())
	dir := DataStorePath + strChainID
	if FileNotExists(dir) {
		err := os.MkdirAll(dir, 0777)
		if err == nil {
			fmt.Println("Created directory " + dir)
		} else {
			return err
		}
	}

	err = ioutil.WriteFile(fmt.Sprintf(dir+"/storeJSON.%09d.block", block.GetDatabaseHeight()), data, 0777)
	if err != nil {
		return err
	}
	return nil
}

func SaveEntryJSON(entry interfaces.DatabaseBatchable, blockHeight uint32) error {
	data, err := entry.(interfaces.Printable).JSONByte()
	if err != nil {
		return err
	}

	var out bytes.Buffer
	json.Indent(&out, data, "", "\t")
	data = out.Bytes()

	strChainID := fmt.Sprintf("%x", entry.GetChainID().Bytes())
	dir := DataStorePath + strChainID + "/entries"
	if FileNotExists(dir) {
		err := os.MkdirAll(dir, 0777)
		if err == nil {
			fmt.Println("Created directory " + dir)
		} else {
			return err
		}
	}

	err = ioutil.WriteFile(fmt.Sprintf(dir+"/storeJSON.%09d.%v.entry", blockHeight, entry.DatabasePrimaryIndex().String()), data, 0777)
	if err != nil {
		return err
	}
	return nil
}

func ExportEChain(chainID string, db interfaces.DBOverlay) error {
	fmt.Printf("ExportEChain %v\n", chainID)
	id, err := primitives.NewShaHashFromStr(chainID)
	if err != nil {
		return err
	}
	eBlocks, err := db.FetchAllEBlocksByChain(id)
	if err != nil {
		return err
	}
	fmt.Printf("Fetched %v blocks\n", len(eBlocks))
	sort.Sort(util.ByEBlockIDAccending(eBlocks))

	for _, block := range eBlocks {
		SaveBinary(block.(interfaces.DatabaseBatchable))
		SaveJSON(block.(interfaces.DatabaseBatchable))
		height := block.GetDatabaseHeight()
		entryHashes := block.GetBody().GetEBEntries()
		for _, hash := range entryHashes {
			entry, err := db.FetchEntryByHash(hash)
			if err != nil {
				return err
			}
			err = ExportEntry(entry.(interfaces.DatabaseBatchable), height)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func ExportDChain(db interfaces.DBOverlay) error {
	fmt.Printf("ExportDChain\n")
	// get all ecBlocks from db
	dBlocks, err := db.FetchAllDBlocks()
	if err != nil {
		return err
	}
	sort.Sort(util.ByDBlockIDAccending(dBlocks))

	for _, block := range dBlocks {
		//Making sure Hash and KeyMR are set for the JSON export
		block.GetHash()
		block.GetKeyMR()
		err = ExportBlock(block.(interfaces.DatabaseBatchable))
		if err != nil {
			return err
		}
	}
	return nil
}

func ExportECChain(db interfaces.DBOverlay) error {
	fmt.Printf("ExportECChain\n")
	// get all ecBlocks from db
	ecBlocks, err := db.FetchAllECBlocks()
	if err != nil {
		return err
	}
	sort.Sort(util.ByECBlockIDAccending(ecBlocks))

	for _, block := range ecBlocks {
		err = ExportBlock(block.(interfaces.DatabaseBatchable))
		if err != nil {
			return err
		}
	}
	return nil
}

func ExportAChain(db interfaces.DBOverlay) error {
	fmt.Printf("ExportAChain\n")
	// get all aBlocks from db
	aBlocks, err := db.FetchAllABlocks()
	if err != nil {
		return err
	}
	sort.Sort(util.ByABlockIDAccending(aBlocks))

	for _, block := range aBlocks {
		err = ExportBlock(block.(interfaces.DatabaseBatchable))
		if err != nil {
			return err
		}
	}
	return nil
}

func ExportFctChain(db interfaces.DBOverlay) error {
	fmt.Printf("ExportFctChain\n")
	// get all aBlocks from db
	fBlocks, err := db.FetchAllFBlocks()
	if err != nil {
		return err
	}
	sort.Sort(util.ByFBlockIDAccending(fBlocks))

	for _, block := range fBlocks {
		err = ExportBlock(block.(interfaces.DatabaseBatchable))
		if err != nil {
			return err
		}
	}
	return nil
}

func ExportDirBlockInfo(db interfaces.DBOverlay) error {
	fmt.Printf("ExportDirBlockInfo\n")
	// get all aBlocks from db
	dbi, err := db.FetchAllDirBlockInfos()
	if err != nil {
		return err
	}
	fmt.Printf("Fetched %v blocks\n", len(dbi))
	sort.Sort(util.ByDirBlockInfoIDAccending(dbi))

	for _, block := range dbi {
		err = ExportBlock(block)
		if err != nil {
			return err
		}
	}
	return nil
}

func ExportBlock(block interfaces.DatabaseBatchable) error {
	err := SaveBinary(block)
	if err != nil {
		return err
	}
	err = SaveJSON(block)
	if err != nil {
		return err
	}
	return nil
}

func ExportEntry(entry interfaces.DatabaseBatchable, height uint32) error {
	err := SaveEntryBinary(entry, height)
	if err != nil {
		return err
	}
	err = SaveEntryJSON(entry, height)
	if err != nil {
		return err
	}
	return nil
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
		return err
	}

	strChainID := dchain.ChainID.String()
	if FileNotExists(DataStorePath + strChainID) {
		err := os.MkdirAll(DataStorePath+strChainID, 0777)
		if err == nil {
			fmt.Println("Created directory " + DataStorePath + strChainID)
		} else {
			fmt.Println(err)
		}
	}
	err = ioutil.WriteFile(fmt.Sprintf(DataStorePath+strChainID+"/store.%09d.block", block.Header.DBHeight), data, 0777)
	if err != nil {
		return err
	}

}

func exportEBlock(block *EBlock, db interfaces.DBOverlay) {
	if block == nil {
		return
	}

	data, err := block.MarshalBinary()
	if err != nil {
		return err
	}

	strChainID := block.Header.ChainID.String()
	if FileNotExists(DataStorePath + strChainID) {
		err := os.MkdirAll(DataStorePath+strChainID, 0777)
		if err == nil {
			fmt.Println("Created directory " + DataStorePath + strChainID)
		} else {
			fmt.Println(err)
		}
	}

	err = ioutil.WriteFile(fmt.Sprintf(DataStorePath+strChainID+"/store.%09d.%09d.block", block.Header.EBSequence, block.Header.DBHeight), data, 0777)
	if err != nil {
		return err
	}

}

func exportECBlock(block *ECBlock, db interfaces.DBOverlay) {
	if block == nil {
		return
	}

	data, err := block.MarshalBinary()
	if err != nil {
		return err
	}

	strChainID := block.Header.ECChainID.String()
	if FileNotExists(DataStorePath + strChainID) {
		err := os.MkdirAll(DataStorePath+strChainID, 0777)
		if err == nil {
			fmt.Println("Created directory " + DataStorePath + strChainID)
		} else {
			fmt.Println(err)
		}
	}
	err = ioutil.WriteFile(fmt.Sprintf(DataStorePath+strChainID+"/store.%09d.block", block.Header.DBHeight), data, 0777)
	if err != nil {
		return err
	}

}

func exportABlock(block *AdminBlock, db interfaces.DBOverlay) {
	if block == nil {
		return
	}

	data, err := block.MarshalBinary()
	if err != nil {
		return err
	}

	strChainID := block.Header.AdminChainID.String()
	if FileNotExists(DataStorePath + strChainID) {
		err := os.MkdirAll(DataStorePath+strChainID, 0777)
		if err == nil {
			fmt.Println("Created directory " + DataStorePath + strChainID)
		} else {
			fmt.Println(err)
		}
	}
	err = ioutil.WriteFile(fmt.Sprintf(DataStorePath+strChainID+"/store.%09d.block", block.Header.DBHeight), data, 0777)
	if err != nil {
		return err
	}

}

func exportFctBlock(block interfaces.IFBlock, db interfaces.DBOverlay) {
	if block == nil {
		return
	}

	data, err := block.MarshalBinary()
	if err != nil {
		return err
	}

	strChainID := block.GetChainID().String()
	if FileNotExists(DataStorePath + strChainID) {
		err := os.MkdirAll(DataStorePath+strChainID, 0777)
		if err == nil {
			fmt.Println("Created directory " + DataStorePath + strChainID)
		} else {
			fmt.Println(err)
		}
	}
	err = ioutil.WriteFile(fmt.Sprintf(DataStorePath+strChainID+"/store.%09d.block", block.GetDBHeight()), data, 0777)
	if err != nil {
		return err
	}

}
*/
