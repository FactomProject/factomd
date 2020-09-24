package eventservices

import (
	"testing"

	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/modules/events/eventmessages/generated/eventmessages"
	"github.com/stretchr/testify/assert"
)

func TestMapCommitEntryEvent(t *testing.T) {
	msg := newTestCommitEntryMsg()
	commitEntryEvent := mapCommitEntryEvent(eventmessages.EntityState_ACCEPTED, msg)

	assert.NotNil(t, commitEntryEvent)
	assert.NotNil(t, commitEntryEvent.EntryCommit)
	assert.NotNil(t, commitEntryEvent.EntryCommit.Version)
	assert.NotNil(t, commitEntryEvent.EntryCommit.Timestamp)
	assert.NotNil(t, commitEntryEvent.EntryCommit.EntryHash)
	assert.NotNil(t, commitEntryEvent.EntryCommit.Credits)
	assert.NotNil(t, commitEntryEvent.EntryCommit.EntryCreditPublicKey)
	assert.NotNil(t, commitEntryEvent.EntryCommit.Signature)
	assert.Equal(t, eventmessages.EntityState_ACCEPTED, commitEntryEvent.EntryCommit.EntityState)
}

func TestMapCommitEntryEventState(t *testing.T) {
	msg := newTestCommitEntryMsg()
	commitEntryEventState := mapCommitEntryEventState(eventmessages.EntityState_ACCEPTED, msg)

	assert.NotNil(t, commitEntryEventState)
	assert.NotNil(t, commitEntryEventState.StateChange)
	assert.NotNil(t, commitEntryEventState.StateChange.EntityHash)
	assert.NotNil(t, commitEntryEventState.StateChange.BlockHeight)
	assert.Equal(t, eventmessages.EntityState_ACCEPTED, commitEntryEventState.StateChange.EntityState)
}

func TestMapRevealEntryEvent(t *testing.T) {
	msg := newTestRevealEntryMsg()
	revealEntryEvent := mapRevealEntryEvent(eventmessages.EntityState_REQUESTED, msg)

	assert.NotNil(t, revealEntryEvent)
	assert.NotNil(t, revealEntryEvent.EntryReveal)
	assert.NotNil(t, revealEntryEvent.EntryReveal.Timestamp)
	assert.NotNil(t, revealEntryEvent.EntryReveal.Entry)
	assert.Equal(t, eventmessages.EntityState_REQUESTED, revealEntryEvent.EntryReveal.EntityState)
}

func TestMapRevealEntryEventState(t *testing.T) {
	msg := newTestRevealEntryMsg()
	revealEntryEventState := mapRevealEntryEventState(eventmessages.EntityState_REQUESTED, msg)

	assert.NotNil(t, revealEntryEventState)
	assert.NotNil(t, revealEntryEventState.StateChange)
	assert.NotNil(t, revealEntryEventState.StateChange.EntityHash)
	assert.NotNil(t, revealEntryEventState.StateChange.BlockHeight)
	assert.Equal(t, eventmessages.EntityState_REQUESTED, revealEntryEventState.StateChange.EntityState)
}

func TestMapEntryBlocks(t *testing.T) {
	blocks := newTestEntryBlocks()
	entryBlocks := mapEntryBlocks(blocks)

	assert.NotNil(t, entryBlocks)
	assert.NotNil(t, 1, len(entryBlocks))
	assert.NotNil(t, entryBlocks[0].Header)
	assert.NotNil(t, entryBlocks[0].Header.ChainID)
	assert.NotNil(t, entryBlocks[0].Header.EntryCount)
	assert.NotNil(t, entryBlocks[0].Header.BlockHeight)
	assert.NotNil(t, entryBlocks[0].Header.BlockSequence)
	assert.NotNil(t, entryBlocks[0].Header.BodyMerkleRoot)
	assert.NotNil(t, entryBlocks[0].Header.PreviousKeyMerkleRoot)
	assert.NotNil(t, entryBlocks[0].Header.PreviousFullHash)
}

func TestMapEntryBlockHashes(t *testing.T) {
	entries := []interfaces.IHash{primitives.NewZeroHash()}
	entryBlockHashes := mapEntryBlockHashes(entries)

	assert.NotNil(t, entryBlockHashes)
	assert.Equal(t, 1, len(entryBlockHashes))
	assert.NotNil(t, entryBlockHashes[0])
}

func TestMapEntryBlockHeader(t *testing.T) {
	header := newTestBlockHeader()
	entryBlockHeader := mapEntryBlockHeader(header)

	assert.NotNil(t, entryBlockHeader)
	assert.NotNil(t, entryBlockHeader.ChainID)
	assert.NotNil(t, entryBlockHeader.EntryCount)
	assert.NotNil(t, entryBlockHeader.BlockHeight)
	assert.NotNil(t, entryBlockHeader.BlockSequence)
	assert.NotNil(t, entryBlockHeader.BodyMerkleRoot)
	assert.NotNil(t, entryBlockHeader.PreviousKeyMerkleRoot)
}

func TestMapEntryBlockEntries(t *testing.T) {
	entries := []interfaces.IEBEntry{new(entryBlock.Entry)}
	entryBlockEntries := mapEntryBlockEntries(entries, false)

	assert.NotNil(t, entryBlockEntries)
	assert.Equal(t, 1, len(entryBlockEntries))

	entryBlockEntry := entryBlockEntries[0]
	assert.NotNil(t, entryBlockEntry.Version)
	assert.NotNil(t, entryBlockEntry.Hash)
	assert.Nil(t, entryBlockEntry.Content)
	assert.Nil(t, entryBlockEntry.ExternalIDs)
}

func TestMapEntryBlockEntriesWithContent(t *testing.T) {
	entry := entryBlock.RandomEntry()
	entries := []interfaces.IEBEntry{entry}
	entryBlockEntries := mapEntryBlockEntries(entries, true)

	assert.NotNil(t, entryBlockEntries)
	assert.Equal(t, 1, len(entryBlockEntries))

	entryBlockEntry := entryBlockEntries[0]
	assert.NotNil(t, entryBlockEntry.Version)
	assert.NotNil(t, entryBlockEntry.Hash)
	assert.NotNil(t, entryBlockEntry.Content)
	assert.NotNil(t, entryBlockEntry.ExternalIDs)
}

func TestMapEntryBlockEntry(t *testing.T) {
	entry := new(entryBlock.Entry)

	entryBlockEntry := mapEntryBlockEntry(entry, false)

	assert.NotNil(t, entryBlockEntry)
	assert.NotNil(t, entryBlockEntry.Version)
	assert.NotNil(t, entryBlockEntry.Hash)
	assert.Nil(t, entryBlockEntry.Content)
	assert.Nil(t, entryBlockEntry.ExternalIDs)
}

func TestMapEntryBlockEntryWithContent(t *testing.T) {
	entry := entryBlock.RandomEntry()
	entryBlockEntry := mapEntryBlockEntry(entry, true)

	assert.NotNil(t, entryBlockEntry)
	assert.NotNil(t, entryBlockEntry.Version)
	assert.NotNil(t, entryBlockEntry.Hash)
	assert.NotNil(t, entryBlockEntry.Content)
	assert.NotNil(t, entryBlockEntry.ExternalIDs)
}

func newTestCommitEntryMsg() *messages.CommitEntryMsg {
	msg := new(messages.CommitEntryMsg)
	msg.Signature = nil
	msg.CommitEntry = entryCreditBlock.NewCommitEntry()
	return msg
}

func newTestRevealEntryMsg() *messages.RevealEntryMsg {
	msg := new(messages.RevealEntryMsg)
	msg.Entry = entryBlock.NewEntry()
	msg.Timestamp = primitives.NewTimestampNow()
	return msg
}

func newTestEntryBlocks() []interfaces.IEntryBlock {
	block := new(entryBlock.EBlock)
	block.Header = newTestBlockHeader()
	blocks := []interfaces.IEntryBlock{block}
	return blocks
}

func newTestBlockHeader() interfaces.IEntryBlockHeader {
	header := new(entryBlock.EBlockHeader)
	header.ChainID = primitives.NewZeroHash()
	header.BodyMR = primitives.NewZeroHash()
	header.PrevFullHash = primitives.NewZeroHash()
	header.PrevKeyMR = primitives.NewZeroHash()
	return header
}
