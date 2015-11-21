// Copyright 2015 FactomProject Authors. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// logger is based on github.com/alexcesaro/log and
// github.com/alexcesaro/log/golog (MIT License)

package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/FactomProject/factomd/log"

	"github.com/FactomProject/factomd/anchor"
	"github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/directoryBlock"
	. "github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	ldb "github.com/FactomProject/factomd/database/leveldb"
	"github.com/FactomProject/factomd/util"
	"github.com/btcsuite/btcd/wire"
	"github.com/davecgh/go-spew/spew"
)

var (
	_   = log.Print
	cfg *util.FactomdConfig
	db  database.Db
)

func main() {
	cfg = util.ReadConfig()
	ldbpath := cfg.App.LdbPath
	initDB(ldbpath)

	anchorChainID, _ := HexToHash(cfg.Anchor.AnchorChainID)
	//log.Println("anchorChainID: ", cfg.Anchor.AnchorChainID)

	processAnchorChain(anchorChainID)

	//initDB("/home/bw/.factom/ldb.prd")
	//dirBlockInfoMap, _ := db.FetchAllDirBlockInfo() // map[string]*DirBlockInfo
	//for _, dirBlockInfo := range dirBlockInfoMap {
	//log.Printf("dirBlockInfo: %s\n", spew.Sdump(dirBlockInfo))
	//}
}

func processAnchorChain(anchorChainID interfaces.IHash) {
	eblocks, _ := db.FetchAllEBlocksByChain(anchorChainID)
	//log.Println("anchorChain length: ", len(*eblocks))
	for _, eblock := range *eblocks {
		//log.Printf("anchor chain block=%s\n", spew.Sdump(eblock))
		if eblock.Header.EBSequence == 0 {
			continue
		}
		for _, ebEntry := range eblock.Body.EBEntries {
			entry, _ := db.FetchEntryByHash(ebEntry)
			if entry != nil {
				//log.Printf("entry=%s\n", spew.Sdump(entry))
				aRecord, err := entryToAnchorRecord(entry)
				if err != nil {
					log.Println(err)
				}
				dirBlockInfo, _ := dirBlockInfoToAnchorChain(aRecord)
				err = db.InsertDirBlockInfo(dirBlockInfo)
				if err != nil {
					log.Printf("InsertDirBlockInfo error: %s, DirBlockInfo=%s\n", err, spew.Sdump(dirBlockInfo))
				}
			}
		}
	}
}

func dirBlockInfoToAnchorChain(aRecord *anchor.AnchorRecord) (*DirBlockInfo, error) {
	dirBlockInfo := new(DirBlockInfo)
	dirBlockInfo.DBHeight = aRecord.DBHeight
	dirBlockInfo.BTCTxOffset = aRecord.Bitcoin.Offset
	dirBlockInfo.BTCBlockHeight = aRecord.Bitcoin.BlockHeight
	mrBytes, _ := hex.DecodeString(aRecord.KeyMR)
	dirBlockInfo.DBMerkleRoot, _ = NewShaHash(mrBytes)
	dirBlockInfo.BTCConfirmed = true

	txSha, _ := wire.NewShaHashFromStr(aRecord.Bitcoin.TXID)
	dirBlockInfo.BTCTxHash = toHash(txSha)
	blkSha, _ := wire.NewShaHashFromStr(aRecord.Bitcoin.BlockHash)
	dirBlockInfo.BTCBlockHash = toHash(blkSha)

	dblock, err := db.FetchDBlockByHeight(aRecord.DBHeight)
	if err != nil {
		log.Printf("err in FetchDBlockByHeight: %d\n", aRecord.DBHeight)
		dirBlockInfo.DBHash = new(primitives.Hash)
	} else {
		dirBlockInfo.Timestamp = int64(dblock.Header.Timestamp)
		dirBlockInfo.DBHash = dblock.DBHash
	}
	log.Printf("dirBlockInfo: %s\n", spew.Sdump(dirBlockInfo))
	return dirBlockInfo, nil
}

func entryToAnchorRecord(entry interfaces.IEBEntry) (*anchor.AnchorRecord, error) {
	content := entry.GetContent()
	jsonARecord := content[:(len(content) - 128)]
	jsonSigBytes := content[(len(content) - 128):]
	jsonSig, err := hex.DecodeString(string(jsonSigBytes))
	if err != nil {
		log.Printf("*** hex.Decode jsonSigBytes error: %s\n", err.Error())
	}

	//log.Println("bytes decoded: ", hex.DecodedLen(len(jsonSigBytes)))
	//log.Printf("jsonARecord: %s\n", string(jsonARecord))
	//log.Printf("    jsonSig: %s\n", string(jsonSigBytes))

	pubKeySlice := make([]byte, 32, 32)
	pubKey := PubKeyFromString(SERVER_PUB_KEY)
	copy(pubKeySlice, pubKey.Key[:])
	verified := VerifySlice(pubKeySlice, jsonARecord, jsonSig)

	if !verified {
		log.Printf("*** anchor chain signature does NOT match:\n")
	} else {
		log.Printf("&&& anchor chain signature does MATCH:\n")
	}

	aRecord := new(anchor.AnchorRecord)
	err = json.Unmarshal(jsonARecord, aRecord)
	if err != nil {
		return nil, fmt.Errorf("json.UnMarshall error: %s", err)
	}
	log.Printf("entryToAnchorRecord: %s", spew.Sdump(aRecord))

	return aRecord, nil
}

func initDB(ldbpath string) {
	var err error
	db, err = ldb.OpenLevelDB(ldbpath, false)
	if err != nil {
		fmt.Errorf("err opening db: %v\n", err)
	}

	if db == nil {
		log.Println("Creating new db ...")
		db, err = ldb.OpenLevelDB(ldbpath, true)
		if err != nil {
			panic(err)
		}
	}
	log.Println("Database started from: " + ldbpath)
}

func toHash(txHash *wire.ShaHash) *Hash {
	h := new(primitives.Hash)
	h.SetBytes(txHash.Bytes())
	return h
}
