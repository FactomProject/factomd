// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package interfaces

//A simplified DBOverlay to make sure we are not calling functions that could cause problems
type DBOverlaySimple interface {
	Close() error
	DoesKeyExist(bucket, key []byte) (bool, error)
	ExecuteMultiBatch() error
	FetchABlock(IHash) (IAdminBlock, error)
	FetchABlockByHeight(blockHeight uint32) (IAdminBlock, error)
	FetchDBKeyMRByHeight(dBlockHeight uint32) (dBlockKeyMR IHash, err error)
	FetchDBlock(IHash) (IDirectoryBlock, error)
	FetchDBlockByHeight(uint32) (IDirectoryBlock, error)
	FetchDBlockHead() (IDirectoryBlock, error)
	FetchEBlock(IHash) (IEntryBlock, error)
	FetchEBlockHead(chainID IHash) (IEntryBlock, error)
	FetchECBlock(IHash) (IEntryCreditBlock, error)
	FetchECBlockByHeight(blockHeight uint32) (IEntryCreditBlock, error)
	FetchECTransaction(hash IHash) (IECBlockEntry, error)
	FetchEntry(IHash) (IEBEntry, error)
	FetchFBlock(IHash) (IFBlock, error)
	FetchFBlockByHeight(blockHeight uint32) (IFBlock, error)
	FetchFactoidTransaction(hash IHash) (ITransaction, error)
	FetchHeadIndexByChainID(chainID IHash) (IHash, error)
	FetchIncludedIn(hash IHash) (IHash, error)
	FetchPaidFor(hash IHash) (IHash, error)
	FetchAllEBlocksByChain(IHash) ([]IEntryBlock, error)
	InsertEntryMultiBatch(entry IEBEntry) error
	InsertEntry(entry IEBEntry) error
	ProcessABlockMultiBatch(block DatabaseBatchable) error
	ProcessDBlockMultiBatch(block DatabaseBlockWithEntries) error
	ProcessEBlockBatch(eblock DatabaseBlockWithEntries, checkForDuplicateEntries bool) error
	ProcessEBlockMultiBatch(eblock DatabaseBlockWithEntries, checkForDuplicateEntries bool) error
	ProcessEBlockMultiBatchWithoutHead(eblock DatabaseBlockWithEntries, checkForDuplicateEntries bool) error
	ProcessECBlockMultiBatch(IEntryCreditBlock, bool) (err error)
	ProcessFBlockMultiBatch(DatabaseBlockWithEntries) error
	FetchDirBlockInfoByKeyMR(hash IHash) (IDirBlockInfo, error)
	SetExportData(path string)
	StartMultiBatch()
	Trim()
	FetchAllEntriesByChainID(chainID IHash) ([]IEBEntry, error)
	SaveKeyValueStore(kvs BinaryMarshallable, key []byte) error
	FetchKeyValueStore(key []byte, dst BinaryMarshallable) (BinaryMarshallable, error)
	SaveDatabaseEntryHeight(height uint32) error
	FetchDatabaseEntryHeight() (uint32, error)
}

// Db defines a generic interface that is used to request and insert data into db
type DBOverlay interface {
	// We let Database method calls flow through.
	IDatabase

	FetchHeadIndexByChainID(chainID IHash) (IHash, error)
	SetExportData(path string)

	StartMultiBatch()
	PutInMultiBatch(records []Record)
	ExecuteMultiBatch() error
	GetEntryType(hash IHash) (IHash, error)

	//**********************************Entry**********************************//

	// InsertEntry inserts an entry
	InsertEntry(entry IEBEntry) (err error)
	InsertEntryMultiBatch(entry IEBEntry) error

	// FetchEntry gets an entry by hash from the database.
	FetchEntry(IHash) (IEBEntry, error)

	FetchAllEntriesByChainID(chainID IHash) ([]IEBEntry, error)

	FetchAllEntryIDsByChainID(chainID IHash) ([]IHash, error)

	FetchAllEntryIDs() ([]IHash, error)

	//**********************************EBlock**********************************//

	// ProcessEBlockBatche inserts the EBlock and update all it's ebentries in DB
	ProcessEBlockBatch(eblock DatabaseBlockWithEntries, checkForDuplicateEntries bool) error
	ProcessEBlockBatchWithoutHead(eblock DatabaseBlockWithEntries, checkForDuplicateEntries bool) error
	ProcessEBlockMultiBatchWithoutHead(eblock DatabaseBlockWithEntries, checkForDuplicateEntries bool) error
	ProcessEBlockMultiBatch(eblock DatabaseBlockWithEntries, checkForDuplicateEntries bool) error

	FetchEBlock(IHash) (IEntryBlock, error)

	// FetchEBlockByHash gets an entry by hash from the database.
	FetchEBlockByPrimary(IHash) (IEntryBlock, error)

	// FetchEBlockByKeyMR gets an entry by hash from the database.
	FetchEBlockBySecondary(hash IHash) (IEntryBlock, error)

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
	ProcessDBlockBatchWithoutHead(block DatabaseBlockWithEntries) error
	ProcessDBlockMultiBatch(block DatabaseBlockWithEntries) error

	// FetchHeightRange looks up a range of blocks by the start and ending
	// heights.  Fetch is inclusive of the start height and exclusive of the
	// ending height. To fetch all hashes from the start height until no
	// more are present, use -1 as endHeight.
	FetchDBlockHeightRange(startHeight, endHeight int64) ([]IHash, error)

	// FetchBlockHeightByKeyMR returns the block height for the given hash.  This is
	// part of the database.Db interface implementation.
	FetchDBlockHeightByKeyMR(IHash) (int64, error)

	FetchDBlock(IHash) (IDirectoryBlock, error)

	// FetchDBlock gets an entry by hash from the database.
	FetchDBlockByPrimary(IHash) (IDirectoryBlock, error)

	// FetchDBlock gets an entry by hash from the database.
	FetchDBlockBySecondary(IHash) (IDirectoryBlock, error)

	// FetchDBlockByHeight gets an directory block by height from the database.
	FetchDBlockByHeight(uint32) (IDirectoryBlock, error)

	FetchDBlockHead() (IDirectoryBlock, error)

	// FetchDBKeyMRByHeight gets a dBlock KeyMR from the database.
	FetchDBKeyMRByHeight(dBlockHeight uint32) (dBlockKeyMR IHash, err error)

	// FetchDBKeyMRByHash gets a DBlock KeyMR by hash.
	FetchDBKeyMRByHash(hash IHash) (dBlockHash IHash, err error)

	// FetchAllFBInfo gets all of the fbInfo
	FetchAllDBlocks() ([]IDirectoryBlock, error)
	FetchAllDBlockKeys() ([]IHash, error)

	SaveDirectoryBlockHead(DatabaseBlockWithEntries) error

	FetchDirectoryBlockHead() (IDirectoryBlock, error)

	//**********************************ECBlock**********************************//

	// ProcessECBlockBatch inserts the ECBlock and update all it's ecbentries in DB
	ProcessECBlockBatch(IEntryCreditBlock, bool) (err error)
	ProcessECBlockBatchWithoutHead(IEntryCreditBlock, bool) (err error)
	ProcessECBlockMultiBatch(IEntryCreditBlock, bool) (err error)

	FetchECBlock(IHash) (IEntryCreditBlock, error)

	// FetchECBlockByHash gets an Entry Credit block by hash from the database.
	FetchECBlockByPrimary(IHash) (IEntryCreditBlock, error)

	// FetchECBlockByKeyMR gets an Entry Credit block by hash from the database.
	FetchECBlockBySecondary(hash IHash) (IEntryCreditBlock, error)
	FetchECBlockByHeight(blockHeight uint32) (IEntryCreditBlock, error)

	// FetchAllECBlocks gets all of the entry credit blocks
	FetchAllECBlocks() ([]IEntryCreditBlock, error)
	FetchAllECBlockKeys() ([]IHash, error)

	SaveECBlockHead(IEntryCreditBlock, bool) error

	FetchECBlockHead() (IEntryCreditBlock, error)

	//**********************************ABlock**********************************//

	// ProcessABlockBatch inserts the AdminBlock
	ProcessABlockBatch(block DatabaseBatchable) error
	ProcessABlockBatchWithoutHead(block DatabaseBatchable) error
	ProcessABlockMultiBatch(block DatabaseBatchable) error

	FetchABlock(IHash) (IAdminBlock, error)

	// FetchABlockByHash gets an admin block by hash from the database.
	FetchABlockByPrimary(hash IHash) (IAdminBlock, error)

	// FetchABlockByKeyMR gets an admin block by keyMR from the database.
	FetchABlockBySecondary(hash IHash) (IAdminBlock, error)
	FetchABlockByHeight(blockHeight uint32) (IAdminBlock, error)

	// FetchAllABlocks gets all of the admin blocks
	FetchAllABlocks() ([]IAdminBlock, error)
	FetchAllABlockKeys() ([]IHash, error)

	SaveABlockHead(DatabaseBatchable) error

	FetchABlockHead() (IAdminBlock, error)

	//**********************************FBlock**********************************//

	// ProcessFBlockBatch inserts the Factoid
	ProcessFBlockBatch(DatabaseBlockWithEntries) error
	ProcessFBlockBatchWithoutHead(DatabaseBlockWithEntries) error
	ProcessFBlockMultiBatch(DatabaseBlockWithEntries) error

	FetchFBlock(IHash) (IFBlock, error)

	// FetchFBlockByHash gets a factoid block by hash from the database.
	FetchFBlockByPrimary(IHash) (IFBlock, error)
	FetchFBlockBySecondary(IHash) (IFBlock, error)
	FetchFBlockByHeight(blockHeight uint32) (IFBlock, error)

	// FetchAllFBlocks gets all of the factoid blocks
	FetchAllFBlocks() ([]IFBlock, error)
	FetchAllFBlockKeys() ([]IHash, error)

	SaveFactoidBlockHead(fblock DatabaseBlockWithEntries) error

	FetchFactoidBlockHead() (IFBlock, error)
	FetchFBlockHead() (IFBlock, error)

	//******************************DirBlockInfo********************************//

	// ProcessDirBlockInfoBatch inserts the dirblock info block
	ProcessDirBlockInfoBatch(block IDirBlockInfo) error

	// FetchDirBlockInfoByHash gets a dirblock info block by hash from the database.
	FetchDirBlockInfoByHash(hash IHash) (IDirBlockInfo, error)

	// FetchDirBlockInfoByKeyMR gets a dirblock info block by keyMR from the database.
	FetchDirBlockInfoByKeyMR(hash IHash) (IDirBlockInfo, error)

	// FetchAllConfirmedDirBlockInfos gets all of the confirmed dirblock info blocks
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

	ReparseAnchorChains() error
	SetBitcoinAnchorRecordPublicKeysFromHex([]string) error
	SetEthereumAnchorRecordPublicKeysFromHex([]string) error

	FetchPaidFor(hash IHash) (IHash, error)

	FetchFactoidTransaction(hash IHash) (ITransaction, error)
	FetchECTransaction(hash IHash) (IECBlockEntry, error)

	//******************************KeyValueStore**********************************//
	SaveKeyValueStore(kvs BinaryMarshallable, key []byte) error
	FetchKeyValueStore(key []byte, dst BinaryMarshallable) (BinaryMarshallable, error)
	SaveDatabaseEntryHeight(height uint32) error
	FetchDatabaseEntryHeight() (uint32, error)
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
