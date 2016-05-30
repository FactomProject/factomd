// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package interfaces

import ()

// Db defines a generic interface that is used to request and insert data into db
type DBOverlay interface {
	// We let Database method calls flow through.
	IDatabase

	FetchHeadIndexByChainID(chainID IHash) (IHash, error)

	StartMultiBatch()
	PutInMultiBatch(records []Record)
	ExecuteMultiBatch() error
	GetEntryType(hash IHash) (IHash, error)

	//**********************************Entry**********************************//

	// InsertEntry inserts an entry
	InsertEntry(entry IEBEntry) (err error)

	// FetchEntry gets an entry by hash from the database.
	FetchEntryByHash(IHash) (IEBEntry, error)

	FetchAllEntriesByChainID(chainID IHash) ([]IEBEntry, error)

	FetchAllEntryIDsByChainID(chainID IHash) ([]IHash, error)

	FetchAllEntryIDs() ([]IHash, error)

	//**********************************EBlock**********************************//

	// ProcessEBlockBatche inserts the EBlock and update all it's ebentries in DB
	ProcessEBlockBatch(eblock DatabaseBlockWithEntries, checkForDuplicateEntries bool) error
	ProcessEBlockMultiBatch(eblock DatabaseBlockWithEntries, checkForDuplicateEntries bool) error

	// FetchEBlockByHash gets an entry by hash from the database.
	FetchEBlockByHash(IHash) (IEntryBlock, error)

	// FetchEBlockByKeyMR gets an entry by hash from the database.
	FetchEBlockByKeyMR(hash IHash) (IEntryBlock, error)

	// FetchEBKeyMRByHash gets an entry by hash from the database.
	FetchEBKeyMRByHash(hash IHash) (IHash, error)

	// FetchAllEBlocksByChain gets all of the blocks by chain id
	FetchAllEBlocksByChain(IHash) ([]IEntryBlock, error)

	SaveEBlockHead(block DatabaseBlockWithEntries, checkForDuplicateEntries bool) error

	FetchEBlockHead(chainID IHash) (IEntryBlock, error)

	FetchAllEBlockChainIDs() ([]IHash, error)

	//**********************************DBlock**********************************//

	// ProcessDBlockBatche inserts the EBlock and update all it's ebentries in DB
	ProcessDBlockBatch(block DatabaseBlockWithEntries) error
	ProcessDBlockMultiBatch(block DatabaseBlockWithEntries) error

	// FetchHeightRange looks up a range of blocks by the start and ending
	// heights.  Fetch is inclusive of the start height and exclusive of the
	// ending height. To fetch all hashes from the start height until no
	// more are present, use -1 as endHeight.
	FetchDBlockHeightRange(startHeight, endHeight int64) ([]IHash, error)

	// FetchBlockHeightByKeyMR returns the block height for the given hash.  This is
	// part of the database.Db interface implementation.
	FetchDBlockHeightByKeyMR(IHash) (int64, error)

	// FetchDBlock gets an entry by hash from the database.
	FetchDBlockByKeyMR(IHash) (IDirectoryBlock, error)

	// FetchDBlockByHeight gets an directory block by height from the database.
	FetchDBlockByHeight(uint32) (IDirectoryBlock, error)

	// FetchDBKeyMRByHeight gets a dBlock KeyMR from the database.
	FetchDBKeyMRByHeight(dBlockHeight uint32) (dBlockKeyMR IHash, err error)

	// FetchDBKeyMRByHash gets a DBlock KeyMR by hash.
	FetchDBKeyMRByHash(hash IHash) (dBlockHash IHash, err error)

	// FetchDBlock gets an entry by hash from the database.
	FetchDBlockByHash(dBlockHash IHash) (dBlock IDirectoryBlock, err error)

	// FetchAllFBInfo gets all of the fbInfo
	FetchAllDBlocks() ([]IDirectoryBlock, error)

	SaveDirectoryBlockHead(DatabaseBlockWithEntries) error

	FetchDirectoryBlockHead() (IDirectoryBlock, error)

	//**********************************ECBlock**********************************//

	// ProcessECBlockBatch inserts the ECBlock and update all it's ecbentries in DB
	ProcessECBlockBatch(IEntryCreditBlock, bool) (err error)
	ProcessECBlockMultiBatch(IEntryCreditBlock, bool) (err error)

	// FetchECBlockByHash gets an Entry Credit block by hash from the database.
	FetchECBlockByHash(IHash) (IEntryCreditBlock, error)

	// FetchECBlockByKeyMR gets an Entry Credit block by hash from the database.
	FetchECBlockByHeaderHash(hash IHash) (IEntryCreditBlock, error)

	// FetchAllECBlocks gets all of the entry credit blocks
	FetchAllECBlocks() ([]IEntryCreditBlock, error)

	SaveECBlockHead(IEntryCreditBlock, bool) error

	FetchECBlockHead() (IEntryCreditBlock, error)

	//**********************************ABlock**********************************//

	// ProcessABlockBatch inserts the AdminBlock
	ProcessABlockBatch(block DatabaseBatchable) error
	ProcessABlockMultiBatch(block DatabaseBatchable) error

	// FetchABlockByHash gets an admin block by hash from the database.
	FetchABlockByHash(hash IHash) (IAdminBlock, error)

	// FetchABlockByKeyMR gets an admin block by keyMR from the database.
	FetchABlockByKeyMR(hash IHash) (IAdminBlock, error)

	// FetchAllABlocks gets all of the admin blocks
	FetchAllABlocks() ([]IAdminBlock, error)

	SaveABlockHead(DatabaseBatchable) error

	FetchABlockHead() (IAdminBlock, error)

	//**********************************FBlock**********************************//

	// ProcessFBlockBatch inserts the Factoid
	ProcessFBlockBatch(DatabaseBlockWithEntries) error
	ProcessFBlockMultiBatch(DatabaseBlockWithEntries) error

	// FetchFBlockByHash gets an admin block by hash from the database.
	FetchFBlockByHash(IHash) (IFBlock, error)

	FetchFBlockByKeyMR(IHash) (IFBlock, error)

	// FetchAllFBlocks gets all of the admin blocks
	FetchAllFBlocks() ([]IFBlock, error)

	SaveFactoidBlockHead(fblock DatabaseBlockWithEntries) error

	FetchFactoidBlockHead() (IFBlock, error)

	//******************************DirBlockInfo********************************//

	// ProcessDirBlockInfoBatch inserts the dirblock info block
	ProcessDirBlockInfoBatch(block IDirBlockInfo) error

	// FetchDirBlockInfoByHash gets a dirblock info block by hash from the database.
	FetchDirBlockInfoByHash(hash IHash) (IDirBlockInfo, error)

	// FetchDirBlockInfoByKeyMR gets a dirblock info block by keyMR from the database.
	FetchDirBlockInfoByKeyMR(hash IHash) (IDirBlockInfo, error)

	// FetchAllConfirmedDirBlockInfos gets all of the confiemed dirblock info blocks
	FetchAllConfirmedDirBlockInfos() ([]IDirBlockInfo, error)

	// FetchAllUnconfirmedDirBlockInfos gets all of the unconfirmed dirblock info blocks
	FetchAllUnconfirmedDirBlockInfos() ([]IDirBlockInfo, error)

	// FetchAllDirBlockInfos gets all of the dirblock info blocks
	FetchAllDirBlockInfos() ([]IDirBlockInfo, error)

	SaveDirBlockInfo(block IDirBlockInfo) error

	//******************************IncludedIn**********************************//

	SaveIncludedIn(entry, block IHash) error
	SaveIncludedInMultiFromBlock(block DatabaseBlockWithEntries, checkForDuplicateEntries bool) error
	SaveIncludedInMulti(entries []IHash, block IHash, checkForDuplicateEntries bool) error
	FetchIncludedIn(hash IHash) (IHash, error)
	RebuildDirBlockInfo() error

	FetchPaidFor(hash IHash) (IHash, error)

	FetchFactoidTransactionByHash(hash IHash) (ITransaction, error)
	FetchECTransactionByHash(hash IHash) (IECBlockEntry, error)
}

type ISCDatabaseOverlay interface {
	DBOverlay

	FetchWalletEntryByName(addr []byte) (IWalletEntry, error)
	FetchWalletEntryByPublicKey(addr []byte) (IWalletEntry, error)
	FetchAllWalletEntriesByName() ([]IWalletEntry, error)
	FetchAllWalletEntriesByPublicKey() ([]IWalletEntry, error)
	FetchAllAddressNameKeys() ([][]byte, error)
	FetchAllAddressPublicKeys() ([][]byte, error)
	FetchTransaction(key []byte) (ITransaction, error)
	SaveTransaction(key []byte, tx ITransaction) error
	DeleteTransaction(key []byte) error
	FetchAllTransactionKeys() ([][]byte, error)
	FetchAllTransactions() ([]ITransaction, error)
	SaveRCDAddress(key []byte, we IWalletEntry) error
	SaveAddressByPublicKey(key []byte, we IWalletEntry) error
	SaveAddressByName(key []byte, we IWalletEntry) error
}
