package eventservices

import (
	"encoding/json"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMapEntryCreditBlock(t *testing.T) {
	serverIndexNumber := entryCreditBlock.NewServerIndexNumber()
	serverIndexNumber.ServerIndexNumber = 2
	minuteNumber := entryCreditBlock.NewMinuteNumber(2)
	commitChain := entryCreditBlock.NewCommitChain()
	commitEntry := entryCreditBlock.NewCommitEntry()

	block := entryCreditBlock.NewECBlock()
	block.GetHeader().SetDBHeight(2)
	block.GetHeader().SetBodySize(2)
	block.GetHeader().SetObjectCount(2)

	block.GetBody().AddEntry(serverIndexNumber)
	block.GetBody().AddEntry(minuteNumber)
	block.GetBody().AddEntry(commitChain)
	block.GetBody().AddEntry(commitEntry)

	mappedBlock := mapEntryCreditBlock(block)

	j, _ := json.MarshalIndent(mappedBlock, "", "  ")
	t.Logf("%s", j)
	t.Log(mappedBlock)

	assert.NotNil(t, mappedBlock)
	if assert.NotNil(t, mappedBlock.Header) {
		assert.Equal(t, block.GetHeader().GetDBHeight(), mappedBlock.Header.BlockHeight)
		assert.Equal(t, block.GetHeader().GetBodyHash().Bytes(), mappedBlock.Header.BodyHash)
		assert.Equal(t, block.GetHeader().GetPrevFullHash().Bytes(), mappedBlock.Header.PreviousFullHash)
		assert.Equal(t, block.GetHeader().GetPrevHeaderHash().Bytes(), mappedBlock.Header.PreviousHeaderHash)
		assert.Equal(t, block.GetHeader().GetObjectCount(), mappedBlock.Header.ObjectCount)
	}
	assert.NotNil(t, mappedBlock.Entries)
	if assert.Equal(t, 4, len(mappedBlock.Entries)) {
		if assert.NotNil(t, mappedBlock.Entries[0].EntryCreditBlockEntry) {
			assert.Equal(t, uint32(2), mappedBlock.Entries[0].GetServerIndexNumber().ServerIndexNumber)
		}
		if assert.NotNil(t, mappedBlock.Entries[1].GetMinuteNumber()) {
			assert.Equal(t, uint32(2), mappedBlock.Entries[1].GetMinuteNumber().MinuteNumber)
		}
		if assert.NotNil(t, mappedBlock.Entries[2].GetChainCommit()) {
			assert.NotNil(t, mappedBlock.Entries[2].GetChainCommit().ChainIDHash)
			assert.NotNil(t, mappedBlock.Entries[2].GetChainCommit().Version)
			assert.NotNil(t, mappedBlock.Entries[2].GetChainCommit().Timestamp)
			assert.NotNil(t, mappedBlock.Entries[2].GetChainCommit().EntityState)
			assert.NotNil(t, mappedBlock.Entries[2].GetChainCommit().EntryHash)
			assert.NotNil(t, mappedBlock.Entries[2].GetChainCommit().Credits)
			assert.NotNil(t, mappedBlock.Entries[2].GetChainCommit().Weld)
			assert.NotNil(t, mappedBlock.Entries[2].GetChainCommit().Signature)
			assert.NotNil(t, mappedBlock.Entries[2].GetChainCommit().EntryCreditPublicKey)
		}
		if assert.NotNil(t, mappedBlock.Entries[3].GetEntryCommit()) {
			assert.NotNil(t, mappedBlock.Entries[3].GetEntryCommit().Version)
			assert.NotNil(t, mappedBlock.Entries[3].GetEntryCommit().Timestamp)
			assert.NotNil(t, mappedBlock.Entries[3].GetEntryCommit().EntityState)
			assert.NotNil(t, mappedBlock.Entries[3].GetEntryCommit().EntryHash)
			assert.NotNil(t, mappedBlock.Entries[3].GetEntryCommit().Credits)
			assert.NotNil(t, mappedBlock.Entries[3].GetEntryCommit().Signature)
			assert.NotNil(t, mappedBlock.Entries[3].GetEntryCommit().EntryCreditPublicKey)
		}
	}
}
