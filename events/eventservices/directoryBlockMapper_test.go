package eventservices

import (
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMapDirectoryBlock(t *testing.T) {
	block := new(directoryBlock.DirectoryBlock)
	directoryBlock := mapDirectoryBlock(block)

	assert.NotNil(t, directoryBlock.Header)
	assert.NotNil(t, directoryBlock.Entries)
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
	assert.NotNil(t, directoryBlockHeader.PreviousFullHash)
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
