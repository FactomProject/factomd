// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package interfaces

import ()

// Db defines a generic interface that is used to request and insert data into db
type DBOverlay interface {
	// We let Database method calls flow through.
	IDatabase

	// InsertEntry inserts an entry
	InsertEntry(entry DatabaseBatchable) (err error)

	// FetchEntry gets an entry by hash from the database.
	FetchEntryByHash(entrySha IHash, dst DatabaseBatchable) (entry DatabaseBatchable, err error)

	// ProcessEBlockBatche inserts the EBlock and update all it's ebentries in DB
	ProcessEBlockBatch(eblock DatabaseBatchable) error

	// FetchEntryBlock gets an entry by hash from the database.
	FetchEBlockByHash(IHash) (DatabaseBatchable, error)

	// FetchAllEBlocksByChain gets all of the blocks by chain id
	FetchAllEBlocksByChain(IHash) ([]IEntryBlock, error)
	
	// FetchDBlock gets an entry by hash from the database.
	FetchDBlockByHash(dBlockHash IHash, dst DatabaseBatchable) (dBlock DatabaseBatchable, err error)

	// FetchDBHashByMR gets a DBHash by MR from the database.
	FetchDBHashByMR(dBMR IHash) (dBlockHash IHash, err error)

	// ProcessDBlockBatche inserts the EBlock and update all it's ebentries in DB
	ProcessDBlockBatch(block DatabaseBatchable) error

	// FetchHeightRange looks up a range of blocks by the start and ending
	// heights.  Fetch is inclusive of the start height and exclusive of the
	// ending height. To fetch all hashes from the start height until no
	// more are present, use -1 as endHeight.
	FetchHeightRange(startHeight, endHeight int64) (rshalist []IHash, err error)

	// FetchAllECBlocks gets all of the entry credit blocks
	FetchAllECBlocks() ([]IEntryCreditBlock, error)
	
	// FetchAllFBInfo gets all of the fbInfo
	FetchAllDBlocks() ([]IDirectoryBlock, error)	

	// FetchDBHashByHeight gets a dBlockHash from the database.
	FetchDBHashByHeight(dBlockHeight uint32) (dBlockHash IHash, err error)

	// FetchDBlockByHeight gets an directory block by height from the database.
	FetchDBlockByHeight(uint32) (DatabaseBatchable, error)

	// ProcessECBlockBatche inserts the ECBlock and update all it's ecbentries in DB
	ProcessECBlockBatch(block DatabaseBatchable) (err error)

	// FetchECBlockByHash gets an Entry Credit block by hash from the database.
	FetchECBlockByHash(IHash) (DatabaseBatchable, error)

	// ProcessABlockBatch inserts the AdminBlock
	ProcessABlockBatch(block DatabaseBatchable) error

	// FetchABlockByHash gets an admin block by hash from the database.
	FetchABlockByHash(hash IHash) (IAdminBlock, error)

	// FetchAllABlocks gets all of the admin blocks
	FetchAllABlocks() ([]IAdminBlock, error)
	
	// ProcessFBlockBatch inserts the Factoid
	ProcessFBlockBatch(DatabaseBatchable) error

	// FetchFBlockByHash gets an admin block by hash from the database.
	FetchFBlockByHash(IHash) (DatabaseBatchable, error)

	// FetchAllFBlocks gets all of the admin blocks
	FetchAllFBlocks() ([]IFBlock, error)

	SaveFactoidBlockHead(fblock DatabaseBatchable) error

}
