// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package database

import (
	. "github.com/FactomProject/factomd/common/DirectoryBlock"
	. "github.com/FactomProject/factomd/common/adminBlock"
	. "github.com/FactomProject/factomd/common/entryBlock"
	. "github.com/FactomProject/factomd/common/entryCreditBlock"
	. "github.com/FactomProject/factomd/common/interfaces"
)

// AllShas is a special value that can be used as the final sha when requesting
// a range of shas by height to request them all.
const AllShas = int64(^uint64(0) >> 1)

// Db defines a generic interface that is used to request and insert data into db
type DBOverlay interface {
	// Close cleanly shuts down the database and syncs all data.
	Close() (err error)

	// RollbackClose discards the recent database changes to the previously
	// saved data at last Sync and closes the database.
	// RollbackClose() (err error)

	// Sync verifies that the database is coherent on disk and no
	// outstanding transactions are in flight.
	// Sync() (err error)

	// InsertEntry inserts an entry
	InsertEntry(entry IEntry) (err error)

	// FetchEntry gets an entry by hash from the database.
	FetchEntryByHash(entrySha IHash) (entry IEntry, err error)

	// FetchEBEntriesFromQueue gets all of the ebentries that have not been processed
	//FetchEBEntriesFromQueue(chainID *[]byte, startTime *[]byte) (ebentries []*EBEntry, err error)

	// ProcessEBlockBatche inserts the EBlock and update all it's ebentries in DB
	ProcessEBlockBatch(eblock *EBlock) error

	// FetchDBEntriesFromQueue gets all of the dbentries that have not been processed
	//FetchDBEntriesFromQueue(startTime *[]byte) (dbentries []*DBEntry, err error)

	// InsertChain inserts the newly created chain into db
	InsertChain(chain *EChain) (err error)

	// FetchChainByHash gets a chain by chainID
	// FetchChainByHash(chainID IHash) (chain *EChain, err error)

	//FetchAllChains gets all of the chains
	FetchAllChains() (chains []*EChain, err error)

	// FetchEntryInfoBranchByHash gets an EntryInfo obj
	//FetchEntryInfoByHash(entryHash IHash) (entryInfo *EntryInfo, err error)

	// FetchEntryInfoBranchByHash gets an EntryInfoBranch obj
	// FetchEntryInfoBranchByHash(entryHash IHash) (entryInfoBranch *EntryInfoBranch, err error)

	// FetchEntryBlock gets an entry by hash from the database.
	FetchEBlockByHash(eBlockHash IHash) (eBlock *EBlock, err error)

	// FetchEBlockByMR gets an entry block by merkle root from the database.
	FetchEBlockByMR(eBMR IHash) (eBlock *EBlock, err error)

	// FetchEBlockByHeight gets an entry block by height from the database.
	//FetchEBlockByHeight(chainID * Hash, eBlockHeight uint32) (eBlock *EBlock, err error)

	// FetchEBHashByMR gets an entry by hash from the database.
	FetchEBHashByMR(eBMR IHash) (eBlockHash IHash, err error)

	// FetchAllEBlocksByChain gets all of the blocks by chain id
	FetchAllEBlocksByChain(chainID IHash) (eBlocks []*EBlock, err error)

	// FetchDBlock gets an entry by hash from the database.
	FetchDBlockByHash(dBlockHash IHash) (dBlock *DirectoryBlock, err error)

	// FetchDBlockByMR gets a directory block by merkle root from the database.
	FetchDBlockByMR(dBMR IHash) (dBlock *DirectoryBlock, err error)

	// FetchDBHashByMR gets a DBHash by MR from the database.
	FetchDBHashByMR(dBMR IHash) (dBlockHash IHash, err error)

	// FetchDBBatchByHash gets an FBBatch obj
	// FetchDirBlockInfoByHash(dbHash IHash) (dirBlockInfo *DirBlockInfo, err error)

	// Insert the Directory Block meta data into db
	InsertDirBlockInfo(dirBlockInfo *DirBlockInfo) (err error)

	// FetchAllDirBlockInfo gets all of the dirBlockInfo
	// FetchAllDirBlockInfo() (ddirBlockInfoMap map[string]*DirBlockInfo, err error)

	// FetchAllUnconfirmedDirBlockInfo gets all of the dirBlockInfos that have BTC Anchor confirmation
	//FetchAllUnconfirmedDirBlockInfo() (dBInfoSlice []DirBlockInfo, err error)
	FetchAllUnconfirmedDirBlockInfo() (dirBlockInfoMap map[string]*DirBlockInfo, err error)

	// ProcessDBlockBatche inserts the EBlock and update all it's ebentries in DB
	ProcessDBlockBatch(block *DirectoryBlock) error

	// FetchHeightRange looks up a range of blocks by the start and ending
	// heights.  Fetch is inclusive of the start height and exclusive of the
	// ending height. To fetch all hashes from the start height until no
	// more are present, use the special id `AllShas'.
	FetchHeightRange(startHeight, endHeight int64) (rshalist []IHash, err error)

	// FetchBlockHeightBySha returns the block height for the given hash.  This is
	// part of the database.Db interface implementation.
	FetchBlockHeightBySha(sha IHash) (int64, error)

	// FetchAllECBlocks gets all of the entry credit blocks
	FetchAllECBlocks() (cBlocks []*ECBlock, err error)

	// FetchAllFBInfo gets all of the fbInfo
	FetchAllDBlocks() (fBlocks []*DirectoryBlock, err error)

	// FetchDBHashByHeight gets a dBlockHash from the database.
	FetchDBHashByHeight(dBlockHeight uint32) (dBlockHash IHash, err error)

	// FetchDBlockByHeight gets an directory block by height from the database.
	FetchDBlockByHeight(dBlockHeight uint32) (dBlock *DirectoryBlock, err error)

	// ProcessECBlockBatche inserts the ECBlock and update all it's ecbentries in DB
	ProcessECBlockBatch(block *ECBlock) (err error)

	// FetchECBlockByHash gets an Entry Credit block by hash from the database.
	FetchECBlockByHash(cBlockHash IHash) (ecBlock *ECBlock, err error)

	// Initialize External ID map for explorer search
	// InitializeExternalIDMap() (extIDMap map[string]bool, err error)

	// ProcessABlockBatch inserts the AdminBlock
	ProcessABlockBatch(block *AdminBlock) error

	// FetchABlockByHash gets an admin block by hash from the database.
	FetchABlockByHash(aBlockHash IHash) (aBlock *AdminBlock, err error)

	// FetchAllABlocks gets all of the admin blocks
	FetchAllABlocks() (aBlocks []*AdminBlock, err error)

	// ProcessABlockBatch inserts the AdminBlock
	ProcessFBlockBatch(IFBlock) error

	// FetchABlockByHash gets an admin block by hash from the database.
	FetchFBlockByHash(IHash) (IFBlock, error)

	// FetchAllABlocks gets all of the admin blocks
	FetchAllFBlocks() ([]IFBlock, error)

	// UpdateBlockHeightCache updates the dir block height cache in db
	UpdateBlockHeightCache(dirBlkHeigh uint32, dirBlkHash IHash) error

	// FetchBlockHeightCache returns the hash and block height of the most recent dir block
	FetchBlockHeightCache() (sha IHash, height int64, err error)

	// UpdateNextBlockHeightCache updates the next dir block height cache (from server) in db
	// UpdateNextBlockHeightCache(dirBlkHeigh uint32) error

	// FetchNextBlockHeightCache returns the next block height from server
	//FetchNextBlockHeightCache() (height int64)

	// FtchHeadMRByChainID gets a MR of the highest block from the database.
	FetchHeadMRByChainID(chainID IHash) (blkMR IHash, err error)
}
