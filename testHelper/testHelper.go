package testHelper

//A package for functions used multiple times in tests that aren't useful in production code.

import (
	"encoding/hex"
	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
	factoidBlock "github.com/FactomProject/factomd/common/factoid/block"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/mapdb"
	//"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/state"
)

var BlockCount int = 10
var DefaultCoinbaseAmount uint64 = 100000000

func CreateAndPopulateTestState() *state.State {
	s := new(state.State)
	s.DB = CreateAndPopulateTestDatabaseOverlay()
	s.Init("")
	err := s.RecalculateBalances()
	if err != nil {
		panic(err)
	}
	return s
}

func CreateAndPopulateTestDatabaseOverlay() *databaseOverlay.Overlay {
	dbo := CreateEmptyTestDatabaseOverlay()

	aBlocks := []*adminBlock.AdminBlock{}
	var prevABlock *adminBlock.AdminBlock = nil

	eBlocks := []*entryBlock.EBlock{}
	var prevEBlock *entryBlock.EBlock = nil

	ecBlocks := []interfaces.IEntryCreditBlock{}
	var prevECBlock interfaces.IEntryCreditBlock = nil

	fBlocks := []interfaces.IFBlock{}
	var prevFBlock interfaces.IFBlock = nil

	dBlocks := []*directoryBlock.DirectoryBlock{}
	var prevDBlock *directoryBlock.DirectoryBlock = nil

	var de *directoryBlock.DBEntry
	dbEntries := []interfaces.IDBEntry{}

	var err error

	for i := 0; i < BlockCount; i++ {
		dbEntries = []interfaces.IDBEntry{}
		prevABlock = CreateTestAdminBlock(prevABlock)
		aBlocks = append(aBlocks, prevABlock)
		err = dbo.SaveABlockHead(prevABlock)
		if err != nil {
			panic(err)
		}

		de = new(directoryBlock.DBEntry)
		de.ChainID, err = primitives.NewShaHash(prevABlock.GetChainID())
		if err != nil {
			panic(err)
		}
		de.KeyMR, err = prevABlock.GetKeyMR()
		if err != nil {
			panic(err)
		}

		dbEntries = append(dbEntries, de)

		prevEBlock = CreateTestEntryBlock(prevEBlock)
		eBlocks = append(eBlocks, prevEBlock)
		err = dbo.SaveEBlockHead(prevEBlock)
		if err != nil {
			panic(err)
		}

		de = new(directoryBlock.DBEntry)
		de.ChainID, err = primitives.NewShaHash(prevEBlock.GetChainID())
		if err != nil {
			panic(err)
		}
		de.KeyMR, err = prevEBlock.KeyMR()
		if err != nil {
			panic(err)
		}

		dbEntries = append(dbEntries, de)

		prevECBlock = CreateTestEntryCreditBlock(prevECBlock)
		ecBlocks = append(ecBlocks, prevECBlock)
		err = dbo.SaveECBlockHead(prevECBlock)
		if err != nil {
			panic(err)
		}

		de = new(directoryBlock.DBEntry)
		de.ChainID, err = primitives.NewShaHash(prevECBlock.GetChainID())
		if err != nil {
			panic(err)
		}
		de.KeyMR, err = prevECBlock.HeaderHash()
		if err != nil {
			panic(err)
		}

		dbEntries = append(dbEntries, de)

		prevFBlock = CreateTestFactoidBlock(prevFBlock)
		fBlocks = append(fBlocks, prevFBlock)
		err = dbo.SaveFactoidBlockHead(prevFBlock)
		if err != nil {
			panic(err)
		}

		de = new(directoryBlock.DBEntry)
		de.ChainID, err = primitives.NewShaHash(prevFBlock.GetChainID())
		if err != nil {
			panic(err)
		}
		de.KeyMR = prevFBlock.GetKeyMR()

		dbEntries = append(dbEntries, de)

		prevDBlock = CreateTestDirectoryBlock(prevDBlock)
		prevDBlock.SetDBEntries(dbEntries)
		dBlocks = append(dBlocks)

		err := dbo.SaveDirectoryBlockHead(prevDBlock)
		if err != nil {
			panic(err)
		}
	}

	return dbo
}

func CreateEmptyTestDatabaseOverlay() *databaseOverlay.Overlay {
	return databaseOverlay.NewOverlay(new(mapdb.MapDB))
}

func CreateTestAdminBlock(prev *adminBlock.AdminBlock) *adminBlock.AdminBlock {
	block := new(adminBlock.AdminBlock)
	block.SetHeader(CreateTestAdminHeader(prev))
	block.GetHeader().SetMessageCount(uint32(len(block.GetABEntries())))
	return block
}

func CreateTestAdminHeader(prev *adminBlock.AdminBlock) *adminBlock.ABlockHeader {
	header := new(adminBlock.ABlockHeader)

	if prev == nil {
		header.PrevLedgerKeyMR = primitives.NewZeroHash()
		header.DBHeight = 0
	} else {
		keyMR, err := prev.GetKeyMR()
		if err != nil {
			panic(err)
		}
		header.PrevLedgerKeyMR = keyMR
		header.DBHeight = prev.Header.GetDBHeight() + 1
	}

	header.HeaderExpansionSize = 5
	header.HeaderExpansionArea = []byte{0x00, 0x01, 0x02, 0x03, 0x04}
	header.MessageCount = 0
	header.BodySize = 0

	return header
}

func CreateTestDirectoryBlock(prevBlock *directoryBlock.DirectoryBlock) *directoryBlock.DirectoryBlock {
	dblock := new(directoryBlock.DirectoryBlock)

	dblock.SetHeader(CreateTestDirectoryBlockHeader(prevBlock))

	dblock.SetDBEntries(make([]interfaces.IDBEntry, 0, 5))

	de := new(directoryBlock.DBEntry)
	de.ChainID = primitives.NewZeroHash()
	de.KeyMR = primitives.NewZeroHash()

	dblock.SetDBEntries(append(dblock.GetDBEntries(), de))
	//dblock.GetHeader().SetBlockCount(uint32(len(dblock.GetDBEntries())))

	return dblock
}

func CreateTestDirectoryBlockHeader(prevBlock *directoryBlock.DirectoryBlock) *directoryBlock.DBlockHeader {
	header := new(directoryBlock.DBlockHeader)

	header.SetBodyMR(primitives.Sha(primitives.NewZeroHash().Bytes()))
	header.SetBlockCount(0)
	header.SetNetworkID(0xffff)

	if prevBlock == nil {
		header.SetDBHeight(0)
		header.SetPrevLedgerKeyMR(primitives.NewZeroHash())
		header.SetPrevKeyMR(primitives.NewZeroHash())
		header.SetTimestamp(1234)
	} else {
		header.SetDBHeight(prevBlock.Header.GetDBHeight() + 1)
		header.SetPrevLedgerKeyMR(prevBlock.GetHash())
		keyMR, err := prevBlock.BuildKeyMerkleRoot()
		if err != nil {
			panic(err)
		}
		header.SetPrevKeyMR(keyMR)
		header.SetTimestamp(prevBlock.Header.GetTimestamp() + 1)
	}

	header.SetVersion(1)

	return header
}

func CreateTestEntryBlock(prev *entryBlock.EBlock) *entryBlock.EBlock {
	e := entryBlock.NewEBlock()
	entryStr := "4bf71c177e71504032ab84023d8afc16e302de970e6be110dac20adbf9a1974625f25d9375533b44505964af993212ef7c13314736b2c76a37c73571d89d8b21c6180f7430677d46d93a3e17b68e6a25dc89ecc092cee1459101578859f7f6969d171a092a1d04f067d55628b461c6a106b76b4bc860445f87b0052cdc5f2bfd000002d800001b080000000272d72e71fdee4984ecb30eedcc89cb171d1f5f02bf9a8f10a8b2cfbaf03efe1c0000000000000000000000000000000000000000000000000000000000000001"
	h, err := hex.DecodeString(entryStr)
	if err != nil {
		panic(err)
	}
	err = e.UnmarshalBinary(h)
	if err != nil {
		panic(err)
	}

	if prev != nil {
		keyMR, err := prev.KeyMR()
		if err != nil {
			panic(err)
		}

		e.Header.SetPrevKeyMR(keyMR)
		hash, err := prev.Hash()
		if err != nil {
			panic(err)
		}
		e.Header.SetPrevLedgerKeyMR(hash)
		e.Header.SetDBHeight(prev.Header.GetDBHeight() + 1)

		e.Header.SetChainID(prev.Header.GetChainID())
	} else {
		e.Header.SetPrevKeyMR(primitives.NewZeroHash())
		e.Header.SetDBHeight(0)
	}

	return e
}

func CreateTestEntryCreditBlock(prev interfaces.IEntryCreditBlock) interfaces.IEntryCreditBlock {
	block, err := entryCreditBlock.NextECBlock(prev)
	if err != nil {
		panic(err)
	}
	return block
}

func CreateTestFactoidBlock(prev interfaces.IFBlock) interfaces.IFBlock {
	return CreateTestFactoidBlockWithCoinbase(prev, NewFactoidAddress(0), DefaultCoinbaseAmount)
}

func CreateTestFactoidBlockWithCoinbase(prev interfaces.IFBlock, address interfaces.IAddress, amount uint64) interfaces.IFBlock {
	block := factoidBlock.NewFBlockFromPreviousBlock(1, prev)
	tx := new(factoid.Transaction)
	tx.AddOutput(address, amount)
	err := block.AddCoinbase(tx)
	if err != nil {
		panic(err)
	}
	return block
}
