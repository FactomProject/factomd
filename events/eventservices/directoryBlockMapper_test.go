package eventservices

import (
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/directoryBlock/dbInfo"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMapDirectoryBlock(t *testing.T) {
	block := newDirectoryBlock()
	directoryBlock := mapDirectoryBlock(block)

	assert.NotNil(t, directoryBlock.Header)
	assert.NotNil(t, directoryBlock.Entries)
	assert.NotNil(t, directoryBlock.ChainID)
	assert.NotNil(t, directoryBlock.KeyMerkleRoot)
	assert.NotNil(t, directoryBlock.Hash)
}

func newDirectoryBlock() *directoryBlock.DirectoryBlock {
	block := new(directoryBlock.DirectoryBlock)
	dbEntries := make([]interfaces.IDBEntry, 1)
	dbEntries[0] = &directoryBlock.DBEntry{
		ChainID: primitives.NewZeroHash(),
		KeyMR:   primitives.NewZeroHash()}
	block.DBEntries = dbEntries
	return block
}

func TestMapDirHeader(t *testing.T) {
	header := newTestDBlockHeader()
	directoryBlockHeader := mapDirectoryBlockHeader(header)

	assert.NotNil(t, directoryBlockHeader.Version)
	assert.NotNil(t, directoryBlockHeader.NetworkID)
	assert.NotNil(t, directoryBlockHeader.Timestamp)
	assert.NotNil(t, directoryBlockHeader.BlockHeight)
	assert.NotNil(t, directoryBlockHeader.BlockCount)
	assert.NotNil(t, directoryBlockHeader.BodyMerkleRoot)
	assert.NotNil(t, directoryBlockHeader.PreviousKeyMerkleRoot)
}

func TestMapDirectoryBlockEntries(t *testing.T) {
	entries := []interfaces.IDBEntry{newTestEntry()}
	directoryBlockEntries := mapDirectoryBlockEntries(entries)

	assert.NotNil(t, directoryBlockEntries)
	assert.Equal(t, 1, len(directoryBlockEntries))
	directoryBlockEntry := directoryBlockEntries[0]
	assert.NotNil(t, directoryBlockEntry.ChainID)
	assert.NotNil(t, directoryBlockEntry.KeyMerkleRoot)
}

func TestMapDirectoryBlockEntry(t *testing.T) {
	entry := newTestEntry()
	directoryBlockEntry := mapDirectoryBlockEntry(entry)

	assert.NotNil(t, directoryBlockEntry)
	assert.NotNil(t, directoryBlockEntry.ChainID)
	assert.NotNil(t, directoryBlockEntry.KeyMerkleRoot)
}

func newTestDBlockHeader() *directoryBlock.DBlockHeader {
	header := new(directoryBlock.DBlockHeader)
	header.BodyMR = primitives.NewZeroHash()
	header.PrevKeyMR = primitives.NewZeroHash()
	header.PrevFullHash = primitives.NewZeroHash()

	return header
}

func newTestEntry() interfaces.IDBEntry {
	entry := new(directoryBlock.DBEntry)
	entry.SetChainID(primitives.NewZeroHash())
	entry.SetKeyMR(primitives.NewZeroHash())
	return entry
}

func TestMapDirectoryBlockInfo(t *testing.T) {
	block := newDirectoryblockInfo()
	directoryBlockInfo := mapDirectoryBlockInfo(block)
	dirBlockAnchor := directoryBlockInfo.DirectoryBlockAnchor
	assert.NotNil(t, dirBlockAnchor)
	assert.NotNil(t, dirBlockAnchor.DirectoryBlockHash)
	assert.NotNil(t, dirBlockAnchor.DirectoryBlockMerkleRoot)
	assert.NotNil(t, dirBlockAnchor.BlockHeight)
	assert.NotNil(t, dirBlockAnchor.Timestamp)
	assert.NotNil(t, dirBlockAnchor.BtcTxHash)
	assert.NotNil(t, dirBlockAnchor.BtcTxOffset)
	assert.NotNil(t, dirBlockAnchor.BtcBlockHash)
	assert.NotNil(t, dirBlockAnchor.BtcBlockHeight)
	assert.NotNil(t, dirBlockAnchor.BtcConfirmed)
	assert.NotNil(t, dirBlockAnchor.EthereumAnchorRecordEntryHash)
	assert.NotNil(t, dirBlockAnchor.EthereumConfirmed)
}

func newDirectoryblockInfo() *dbInfo.DirBlockInfo {
	info := new(dbInfo.DirBlockInfo)
	info.DBHeight = 123
	info.DBHash = primitives.NewZeroHash()
	info.Timestamp = new(primitives.Timestamp).GetTimeMilli()
	info.BTCTxHash = primitives.NewZeroHash()
	info.BTCTxOffset = 1000
	info.BTCBlockHash = primitives.NewZeroHash()
	info.BTCBlockHeight = 1001
	info.BTCConfirmed = true
	info.EthereumAnchorRecordEntryHash = primitives.NewZeroHash()
	info.EthereumConfirmed = true
	return info
}
