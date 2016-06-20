// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package databaseOverlay

import (
	"encoding/binary"
	"sync"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/blockExtractor"
)

// the "table" prefix
var (
	// Directory Block
	DIRECTORYBLOCK                = []byte("DirectoryBlock")
	DIRECTORYBLOCK_NUMBER         = []byte("DirectoryBlockNumber")
	DIRECTORYBLOCK_SECONDARYINDEX = []byte("DirectoryBlockSecondaryIndex")

	// Admin Block
	ADMINBLOCK                = []byte("AdminBlock")
	ADMINBLOCK_NUMBER         = []byte("AdminBlockNumber")
	ADMINBLOCK_SECONDARYINDEX = []byte("AdminBlockSecondaryIndex")

	//Factoid Block
	FACTOIDBLOCK                = []byte("FactoidBlock")
	FACTOIDBLOCK_NUMBER         = []byte("FactoidBlockNumber")
	FACTOIDBLOCK_SECONDARYINDEX = []byte("FactoidBlockSecondaryIndex")

	// Entry Credit Block
	ENTRYCREDITBLOCK                = []byte("EntryCreditBlock")
	ENTRYCREDITBLOCK_NUMBER         = []byte("EntryCreditBlockNumber")
	ENTRYCREDITBLOCK_SECONDARYINDEX = []byte("EntryCreditBlockSecondaryIndex")

	// Entry Chain
	//ENTRYCHAIN //11

	// The latest Block MR for chains including special chains
	CHAIN_HEAD = []byte("ChainHead")

	// Entry Block
	ENTRYBLOCK                = []byte("EntryBlock")
	ENTRYBLOCK_CHAIN_NUMBER   = []byte("EntryBlockNumber")
	ENTRYBLOCK_SECONDARYINDEX = []byte("EntryBlockSecondaryIndex")

	//Entry
	ENTRY = []byte("Entry")

	//Directory Block Info
	DIRBLOCKINFO                = []byte("DirBlockInfo")
	DIRBLOCKINFO_UNCONFIRMED    = []byte("DirBlockInfoUnconfirmed")
	DIRBLOCKINFO_NUMBER         = []byte("DirBlockInfoNumber")
	DIRBLOCKINFO_SECONDARYINDEX = []byte("DirBlockInfoSecondaryIndex")

	//IncludedIn
	INCLUDED_IN = []byte("IncludedIn")

	//Which EC transaction paid for this Entry
	PAID_FOR = []byte("PaidFor")
)

var ConstantNamesMap map[string]string

func init() {
	ConstantNamesMap = map[string]string{}
	ConstantNamesMap[string(DIRECTORYBLOCK)] = "DirectoryBlock"
	ConstantNamesMap[string(DIRECTORYBLOCK_NUMBER)] = "DirectoryBlockNumber"
	ConstantNamesMap[string(DIRECTORYBLOCK_SECONDARYINDEX)] = "DirectoryBlockSecondaryIndex"

	ConstantNamesMap[string(ADMINBLOCK)] = "AdminBlock"
	ConstantNamesMap[string(ADMINBLOCK_NUMBER)] = "AdminBlockNumber"
	ConstantNamesMap[string(ADMINBLOCK_SECONDARYINDEX)] = "AdminBlockSecondaryIndex"

	ConstantNamesMap[string(FACTOIDBLOCK)] = "FactoidBlock"
	ConstantNamesMap[string(FACTOIDBLOCK_NUMBER)] = "FactoidBlockNumber"
	ConstantNamesMap[string(FACTOIDBLOCK_SECONDARYINDEX)] = "FactoidBlockSecondaryIndex"

	ConstantNamesMap[string(ENTRYCREDITBLOCK)] = "EntryCreditBlock"
	ConstantNamesMap[string(ENTRYCREDITBLOCK_NUMBER)] = "EntryCreditBlockNumber"
	ConstantNamesMap[string(ENTRYCREDITBLOCK_SECONDARYINDEX)] = "EntryCreditBlockSecondaryIndex"

	ConstantNamesMap[string(CHAIN_HEAD)] = "ChainHead"

	ConstantNamesMap[string(ENTRYBLOCK)] = "EntryBlock"
	ConstantNamesMap[string(ENTRYBLOCK_CHAIN_NUMBER)] = "EntryBlockChainNumber"
	ConstantNamesMap[string(ENTRYBLOCK_SECONDARYINDEX)] = "EntryBlockSecondaryIndex"

	ConstantNamesMap[string(ENTRY)] = "Entry"

	ConstantNamesMap[string(DIRBLOCKINFO)] = "DirBlockInfo"
	ConstantNamesMap[string(DIRBLOCKINFO_UNCONFIRMED)] = "DirBlockInfoUnconfirmed"
	ConstantNamesMap[string(DIRBLOCKINFO_NUMBER)] = "DirBlockInfoNumber"
	ConstantNamesMap[string(DIRBLOCKINFO_SECONDARYINDEX)] = "DirBlockInfoSecondaryIndex"

	ConstantNamesMap[string(INCLUDED_IN)] = "IncludedIn"

	ConstantNamesMap[string(PAID_FOR)] = "PaidFor"
}

type Overlay struct {
	DB interfaces.IDatabase

	ExportData     bool
	ExportDataPath string

	BatchSemaphore sync.Mutex
	MultiBatch     []interfaces.Record
	BlockExtractor blockExtractor.BlockExtractor
}

func (db *Overlay) ListAllBuckets() ([][]byte, error) {
	return db.DB.ListAllBuckets()
}

func (db *Overlay) SetExportData(path string) {
	db.ExportData = true
	db.ExportDataPath = path
	db.BlockExtractor.DataStorePath = path
}

func (db *Overlay) StartMultiBatch() {
	db.BatchSemaphore.Lock()
	db.MultiBatch = make([]interfaces.Record, 0, 128)
}

func (db *Overlay) PutInMultiBatch(records []interfaces.Record) {
	db.MultiBatch = append(db.MultiBatch, records...)
}

func (db *Overlay) ExecuteMultiBatch() error {
	defer func() {
		db.MultiBatch = nil
		db.BatchSemaphore.Unlock()
	}()
	return db.PutInBatch(db.MultiBatch)
}

func (db *Overlay) PutInBatch(records []interfaces.Record) error {
	return db.DB.PutInBatch(records)
}

func (db *Overlay) Put(bucket, key []byte, data interfaces.BinaryMarshallable) error {
	return db.DB.Put(bucket, key, data)
}

func (db *Overlay) ListAllKeys(bucket []byte) ([][]byte, error) {
	return db.DB.ListAllKeys(bucket)
}

func (db *Overlay) GetAll(bucket []byte, sample interfaces.BinaryMarshallableAndCopyable) ([]interfaces.BinaryMarshallableAndCopyable, [][]byte, error) {
	return db.DB.GetAll(bucket, sample)
}

func (db *Overlay) Get(bucket, key []byte, destination interfaces.BinaryMarshallable) (interfaces.BinaryMarshallable, error) {
	return db.DB.Get(bucket, key, destination)
}

func (db *Overlay) Clear(bucket []byte) error {
	return db.DB.Clear(bucket)
}

func (db *Overlay) Close() (err error) {
	return db.DB.Close()
}

func (db *Overlay) Delete(bucket, key []byte) error {
	return db.DB.Delete(bucket, key)
}

func NewOverlay(db interfaces.IDatabase) *Overlay {
	answer := new(Overlay)
	answer.DB = db
	return answer
}

func (db *Overlay) FetchBlockByHeight(heightBucket []byte, blockBucket []byte, blockHeight uint32, dst interfaces.DatabaseBatchable) (interfaces.DatabaseBatchable, error) {
	index, err := db.FetchBlockIndexByHeight(heightBucket, blockHeight)
	if err != nil {
		return nil, err
	}
	if index == nil {
		return nil, nil
	}
	return db.FetchBlock(blockBucket, index, dst)
}

func (db *Overlay) FetchBlockIndexByHeight(bucket []byte, blockHeight uint32) (interfaces.IHash, error) {
	key := make([]byte, 4)
	binary.BigEndian.PutUint32(key, blockHeight)

	block, err := db.DB.Get(bucket, key, new(primitives.Hash))
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(interfaces.IHash), nil
}

func (db *Overlay) FetchPrimaryIndexBySecondaryIndex(secondaryIndexBucket []byte, key interfaces.IHash) (interfaces.IHash, error) {
	block, err := db.DB.Get(secondaryIndexBucket, key.Bytes(), new(primitives.Hash))
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(interfaces.IHash), nil
}

func (db *Overlay) FetchBlockBySecondaryIndex(secondaryIndexBucket, blockBucket []byte, index interfaces.IHash, dst interfaces.DatabaseBatchable) (interfaces.DatabaseBatchable, error) {
	hash, err := db.FetchPrimaryIndexBySecondaryIndex(secondaryIndexBucket, index)
	if err != nil {
		return nil, err
	}
	if hash == nil {
		return nil, nil
	}
	return db.FetchBlock(blockBucket, hash, dst)
}

func (db *Overlay) FetchBlock(bucket []byte, key interfaces.IHash, dst interfaces.DatabaseBatchable) (interfaces.DatabaseBatchable, error) {
	block, err := db.DB.Get(bucket, key.Bytes(), dst)
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(interfaces.DatabaseBatchable), nil
}

func (db *Overlay) FetchAllBlocksFromBucket(bucket []byte, sample interfaces.BinaryMarshallableAndCopyable) ([]interfaces.BinaryMarshallableAndCopyable, error) {
	answer, _, err := db.DB.GetAll(bucket, sample)
	if err != nil {
		return nil, err
	}
	return answer, nil
}

func (db *Overlay) FetchAllBlockKeysFromBucket(bucket []byte) ([]interfaces.IHash, error) {
	entries, err := db.ListAllKeys(bucket)
	if err != nil {
		return nil, err
	}
	answer := make([]interfaces.IHash, len(entries))
	for i := range entries {
		answer[i], err = primitives.NewShaHash(entries[i])
		if err != nil {
			return nil, err
		}
	}
	return answer, nil
}

func (db *Overlay) Insert(bucket []byte, entry interfaces.DatabaseBatchable) error {
	err := db.DB.Put(bucket, entry.DatabasePrimaryIndex().Bytes(), entry)
	if err != nil {
		return err
	}
	return nil
}

func (db *Overlay) ProcessBlockMultiBatch(blockBucket, numberBucket, secondaryIndexBucket []byte, block interfaces.DatabaseBatchable) error {
	if block == nil {
		return nil
	}

	batch := []interfaces.Record{}

	batch = append(batch, interfaces.Record{blockBucket, block.DatabasePrimaryIndex().Bytes(), block})

	if numberBucket != nil {
		bytes := make([]byte, 4)
		binary.BigEndian.PutUint32(bytes, block.GetDatabaseHeight())
		batch = append(batch, interfaces.Record{numberBucket, bytes, block.DatabasePrimaryIndex()})
	}

	if secondaryIndexBucket != nil {
		batch = append(batch, interfaces.Record{secondaryIndexBucket, block.DatabaseSecondaryIndex().Bytes(), block.DatabasePrimaryIndex()})
	}

	batch = append(batch, interfaces.Record{CHAIN_HEAD, block.GetChainID().Bytes(), block.DatabasePrimaryIndex()})

	db.PutInMultiBatch(batch)

	if db.ExportData {
		err := db.BlockExtractor.ExportBlock(block)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *Overlay) ProcessBlockBatch(blockBucket, numberBucket, secondaryIndexBucket []byte, block interfaces.DatabaseBatchable) error {
	if block == nil {
		return nil
	}

	batch := []interfaces.Record{}

	batch = append(batch, interfaces.Record{blockBucket, block.DatabasePrimaryIndex().Bytes(), block})

	if numberBucket != nil {
		bytes := make([]byte, 4)
		binary.BigEndian.PutUint32(bytes, block.GetDatabaseHeight())
		batch = append(batch, interfaces.Record{numberBucket, bytes, block.DatabasePrimaryIndex()})
	}

	if secondaryIndexBucket != nil {
		batch = append(batch, interfaces.Record{secondaryIndexBucket, block.DatabaseSecondaryIndex().Bytes(), block.DatabasePrimaryIndex()})
	}

	batch = append(batch, interfaces.Record{CHAIN_HEAD, block.GetChainID().Bytes(), block.DatabasePrimaryIndex()})

	err := db.DB.PutInBatch(batch)
	if err != nil {
		return err
	}

	if db.ExportData {
		err = db.BlockExtractor.ExportBlock(block)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *Overlay) ProcessBlockBatchWithoutHead(blockBucket, numberBucket, secondaryIndexBucket []byte, block interfaces.DatabaseBatchable) error {
	if block == nil {
		return nil
	}

	batch := []interfaces.Record{}

	batch = append(batch, interfaces.Record{blockBucket, block.DatabasePrimaryIndex().Bytes(), block})

	if numberBucket != nil {
		bytes := make([]byte, 4)
		binary.BigEndian.PutUint32(bytes, block.GetDatabaseHeight())
		batch = append(batch, interfaces.Record{numberBucket, bytes, block.DatabasePrimaryIndex()})
	}

	if secondaryIndexBucket != nil {
		batch = append(batch, interfaces.Record{secondaryIndexBucket, block.DatabaseSecondaryIndex().Bytes(), block.DatabasePrimaryIndex()})
	}

	err := db.DB.PutInBatch(batch)
	if err != nil {
		return err
	}

	return nil
}

// FetchHeadMRByChainID gets an index of the highest block from the database.
func (db *Overlay) FetchHeadIndexByChainID(chainID interfaces.IHash) (interfaces.IHash, error) {
	if chainID == nil {
		return nil, nil
	}

	bucket := CHAIN_HEAD
	key := chainID.Bytes()

	block, err := db.DB.Get(bucket, key, new(primitives.Hash))
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}

	return block.(interfaces.IHash), nil
}

func (db *Overlay) FetchChainHeadByChainID(bucket []byte, chainID interfaces.IHash, dst interfaces.DatabaseBatchable) (interfaces.DatabaseBatchable, error) {
	blockHash, err := db.FetchHeadIndexByChainID(chainID)
	if err != nil {
		return nil, err
	}
	if blockHash == nil {
		return nil, nil
	}
	return db.FetchBlock(bucket, blockHash, dst)
}

//Use endHeight of -1 (or other negative numbers) to fetch all / as many entries as possibe
func (db *Overlay) FetchBlockIndexesInHeightRange(numberBucket []byte, startHeight, endHeight int64) ([]interfaces.IHash, error) {
	var endidx int64
	if endHeight < 0 {
		endidx = startHeight + constants.MaxBlocksPerMsg
	} else {
		endidx = endHeight
	}

	shalist := make([]interfaces.IHash, 0, endidx-startHeight)
	for height := startHeight; height < endidx; height++ {
		dbhash, err := db.FetchBlockIndexByHeight(numberBucket, uint32(height))
		if err != nil {
			return nil, err
		}
		if dbhash == nil {
			break
		}

		shalist = append(shalist, dbhash)
	}

	return shalist, nil
}

func (db *Overlay) GetEntryType(hash interfaces.IHash) (interfaces.IHash, error) {
	if hash == nil {
		return nil, nil
	}

	in, err := db.FetchIncludedIn(hash)
	if err != nil {
		return nil, err
	}

	if in == nil {
		//Entry not included anywhere, it might be still a dBlock (or the special block not included in dblocks)
		dBlock, err := db.FetchDBlock(hash)
		if err != nil {
			return nil, err
		}
		if dBlock == nil {
			//TODO: search for the free-floating block here

			//Entry is nowhere to be found
			return nil, nil
		}
		return dBlock.GetChainID(), nil
	}

	eBlock, err := db.FetchEBlock(in)
	if err != nil {
		return nil, err
	}
	if eBlock != nil {
		return eBlock.GetChainID(), nil
	}

	ecBlock, err := db.FetchECBlock(in)
	if err != nil {
		return nil, err
	}
	if ecBlock != nil {
		return ecBlock.GetChainID(), nil
	}

	fblock, err := db.FetchFBlock(in)
	if err != nil {
		return nil, err
	}
	if fblock != nil {
		return fblock.GetChainID(), nil
	}

	ablock, err := db.FetchABlock(in)
	if err != nil {
		return nil, err
	}
	if ablock != nil {
		return ablock.GetChainID(), nil
	}

	dBlock, err := db.FetchDBlock(in)
	if err != nil {
		return nil, err
	}
	if dBlock != nil {
		dbEntries := dBlock.GetDBEntries()
		for _, dbEntry := range dbEntries {
			if dbEntry.GetKeyMR().IsSameAs(hash) {
				return dbEntry.GetChainID(), nil
			}
		}
		if dBlock.GetKeyMR().IsSameAs(hash) == true {
			return dBlock.GetChainID(), nil
		}
	}

	return nil, nil
}
