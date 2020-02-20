// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package blockExtractor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/util"
)

type BlockExtractor struct {
	DataStorePath string
}

func FileNotExists(name string) bool {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return true
	}
	return err != nil
}

func (be *BlockExtractor) SaveBinary(block interfaces.DatabaseBatchable) error {
	data, err := block.MarshalBinary()
	if err != nil {
		return err
	}

	strChainID := fmt.Sprintf("%x", block.GetChainID().Bytes())
	dir := be.DataStorePath + strChainID
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

func (be *BlockExtractor) SaveEntryBinary(entry interfaces.DatabaseBatchable, blockHeight uint32) error {
	data, err := entry.MarshalBinary()
	if err != nil {
		return err
	}

	strChainID := fmt.Sprintf("%x", entry.GetChainID().Bytes())
	dir := be.DataStorePath + strChainID + "/entries"
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

func (be *BlockExtractor) SaveJSON(block interfaces.DatabaseBatchable) error {
	data, err := block.(interfaces.Printable).JSONByte()
	if err != nil {
		return err
	}

	var out bytes.Buffer
	json.Indent(&out, data, "", "\t")
	data = out.Next(out.Len())

	strChainID := fmt.Sprintf("%x", block.GetChainID().Bytes())
	dir := be.DataStorePath + strChainID
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

func (be *BlockExtractor) SaveEntryJSON(entry interfaces.DatabaseBatchable, blockHeight uint32) error {
	data, err := entry.(interfaces.Printable).JSONByte()
	if err != nil {
		return err
	}

	var out bytes.Buffer
	json.Indent(&out, data, "", "\t")
	data = out.Next(out.Len())

	strChainID := fmt.Sprintf("%x", entry.GetChainID().Bytes())
	dir := be.DataStorePath + strChainID + "/entries"
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

func (be *BlockExtractor) ExportEChain(chainID string, db interfaces.DBOverlay) error {
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
	sort.Sort(util.ByEBlockIDAscending(eBlocks))

	for _, block := range eBlocks {
		be.SaveBinary(block.(interfaces.DatabaseBatchable))
		be.SaveJSON(block.(interfaces.DatabaseBatchable))
		height := block.GetDatabaseHeight()
		entryHashes := block.GetBody().GetEBEntries()
		for _, hash := range entryHashes {
			if hash.IsMinuteMarker() {
				continue
			}
			entry, err := db.FetchEntry(hash)
			if err != nil {
				return err
			}
			err = be.ExportEntry(entry.(interfaces.DatabaseBatchable), height)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (be *BlockExtractor) ExportDChain(db interfaces.DBOverlay) error {
	fmt.Printf("ExportDChain\n")
	// get all ecBlocks from db
	dBlocks, err := db.FetchAllDBlocks()
	if err != nil {
		return err
	}
	sort.Sort(util.ByDBlockIDAscending(dBlocks))

	for _, block := range dBlocks {
		//Making sure Hash and KeyMR are set for the JSON export
		block.GetFullHash()
		block.GetKeyMR()
		err = be.ExportBlock(block.(interfaces.DatabaseBatchable))
		if err != nil {
			return err
		}
	}
	return nil
}

func (be *BlockExtractor) ExportECChain(db interfaces.DBOverlay) error {
	fmt.Printf("ExportECChain\n")
	// get all ecBlocks from db
	ecBlocks, err := db.FetchAllECBlocks()
	if err != nil {
		return err
	}
	sort.Sort(util.ByECBlockIDAscending(ecBlocks))

	for _, block := range ecBlocks {
		err = be.ExportBlock(block.(interfaces.DatabaseBatchable))
		if err != nil {
			return err
		}
	}
	return nil
}

func (be *BlockExtractor) ExportAChain(db interfaces.DBOverlay) error {
	fmt.Printf("ExportAChain\n")
	// get all aBlocks from db
	aBlocks, err := db.FetchAllABlocks()
	if err != nil {
		return err
	}
	sort.Sort(util.ByABlockIDAscending(aBlocks))

	for _, block := range aBlocks {
		err = be.ExportBlock(block.(interfaces.DatabaseBatchable))
		if err != nil {
			return err
		}
	}
	return nil
}

func (be *BlockExtractor) ExportFctChain(db interfaces.DBOverlay) error {
	fmt.Printf("ExportFctChain\n")
	// get all aBlocks from db
	fBlocks, err := db.FetchAllFBlocks()
	if err != nil {
		return err
	}
	sort.Sort(util.ByFBlockIDAscending(fBlocks))

	for _, block := range fBlocks {
		err = be.ExportBlock(block.(interfaces.DatabaseBatchable))
		if err != nil {
			return err
		}
	}
	return nil
}

func (be *BlockExtractor) ExportDirBlockInfo(db interfaces.DBOverlay) error {
	fmt.Printf("ExportDirBlockInfo\n")
	// get all aBlocks from db
	dbi, err := db.FetchAllDirBlockInfos()
	if err != nil {
		return err
	}
	fmt.Printf("Fetched %v blocks\n", len(dbi))
	sort.Sort(util.ByDirBlockInfoIDAscending(dbi))

	for _, block := range dbi {
		err = be.ExportBlock(block)
		if err != nil {
			return err
		}
	}
	return nil
}

func (be *BlockExtractor) ExportBlock(block interfaces.DatabaseBatchable) error {
	err := be.SaveBinary(block)
	if err != nil {
		return err
	}
	err = be.SaveJSON(block)
	if err != nil {
		return err
	}
	return nil
}

func (be *BlockExtractor) ExportEntry(entry interfaces.DatabaseBatchable, height uint32) error {
	err := be.SaveEntryBinary(entry, height)
	if err != nil {
		return err
	}
	err = be.SaveEntryJSON(entry, height)
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
	if FileNotExists(be.DataStorePath + strChainID) {
		err := os.MkdirAll(be.DataStorePath+strChainID, 0777)
		if err == nil {
			fmt.Println("Created directory " + be.DataStorePath + strChainID)
		} else {
			fmt.Println(err)
		}
	}
	err = ioutil.WriteFile(fmt.Sprintf(be.DataStorePath+strChainID+"/store.%09d.block", block.Header.DBHeight), data, 0777)
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
	if FileNotExists(be.DataStorePath + strChainID) {
		err := os.MkdirAll(be.DataStorePath+strChainID, 0777)
		if err == nil {
			fmt.Println("Created directory " + be.DataStorePath + strChainID)
		} else {
			fmt.Println(err)
		}
	}

	err = ioutil.WriteFile(fmt.Sprintf(be.DataStorePath+strChainID+"/store.%09d.%09d.block", block.Header.EBSequence, block.Header.DBHeight), data, 0777)
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
	if FileNotExists(be.DataStorePath + strChainID) {
		err := os.MkdirAll(be.DataStorePath+strChainID, 0777)
		if err == nil {
			fmt.Println("Created directory " + be.DataStorePath + strChainID)
		} else {
			fmt.Println(err)
		}
	}
	err = ioutil.WriteFile(fmt.Sprintf(be.DataStorePath+strChainID+"/store.%09d.block", block.Header.DBHeight), data, 0777)
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
	if FileNotExists(be.DataStorePath + strChainID) {
		err := os.MkdirAll(be.DataStorePath+strChainID, 0777)
		if err == nil {
			fmt.Println("Created directory " + be.DataStorePath + strChainID)
		} else {
			fmt.Println(err)
		}
	}
	err = ioutil.WriteFile(fmt.Sprintf(be.DataStorePath+strChainID+"/store.%09d.block", block.Header.DBHeight), data, 0777)
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
	if FileNotExists(be.DataStorePath + strChainID) {
		err := os.MkdirAll(be.DataStorePath+strChainID, 0777)
		if err == nil {
			fmt.Println("Created directory " + be.DataStorePath + strChainID)
		} else {
			fmt.Println(err)
		}
	}
	err = ioutil.WriteFile(fmt.Sprintf(be.DataStorePath+strChainID+"/store.%09d.block", block.GetDBHeight()), data, 0777)
	if err != nil {
		return err
	}

}
*/
