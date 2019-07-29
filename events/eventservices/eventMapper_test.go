package eventservices_test

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/events"
	"github.com/FactomProject/factomd/events/eventmessages"
	"github.com/FactomProject/factomd/events/eventservices"
	"github.com/FactomProject/factomd/testHelper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDBStateMapping(t *testing.T) {
	msg := newDBStateMsg()
	msg.String()

	inputEvent := events.EventFromMessage(eventmessages.EventSource_COMMIT_DIRECTORY_BLOCK, msg)
	event, err := eventservices.MapToFactomEvent(inputEvent)
	if err != nil {
		t.Error(err)
	}

	assert.IsType(t, &eventmessages.FactomEvent_AnchorEvent{}, event.Value)
	anchorEvent := event.Value.(*eventmessages.FactomEvent_AnchorEvent).AnchorEvent
	assert.NotNil(t, anchorEvent.DirectoryBlock)
	assertFactoidBlock(t, anchorEvent.FactoidBlock)
	assertEntryBlocks(t, anchorEvent.EntryBlocks)
	assertEntryBlockEntries(t, anchorEvent.EntryBlockEntries)
}

func assertFactoidBlock(t *testing.T, factoidBlock *eventmessages.FactoidBlock) {
	assert.NotNil(t, factoidBlock)
	assert.NotNil(t, factoidBlock.BodyMerkleRoot)
	assert.NotNil(t, factoidBlock.PreviousKeyMerkleRoot)
	assert.NotNil(t, factoidBlock.PreviousLedgerKeyMerkleRoot)
	assert.NotNil(t, factoidBlock.ExchRate)
	assert.NotNil(t, factoidBlock.BlockHeight)
	assertTransactions(t, factoidBlock.Transactions)
}

func assertTransactions(t *testing.T, transactions []*eventmessages.Transaction) {
	assert.NotNil(t, transactions)
	for _, transaction := range transactions {
		assert.NotNil(t, transaction)
		assert.NotNil(t, transaction.TransactionId)
		assert.NotNil(t, transaction.BlockHeight)
		assert.NotNil(t, transaction.Outputs)
		for _, output := range transaction.Outputs {
			assert.True(t, output.Amount > 0)
			assert.NotNil(t, output.Address)
		}
	}
}

func assertEntryBlocks(t *testing.T, entryBlocks []*eventmessages.EntryBlock) {
	assert.NotNil(t, entryBlocks)
	for _, entryBlock := range entryBlocks {
		assert.NotNil(t, entryBlock)
		assert.NotNil(t, entryBlock.EntryBlockHeader)
		assertHashes(t, entryBlock.EntryHashes)
	}
}

func assertEntryBlockEntries(t *testing.T, blockEntries []*eventmessages.EntryBlockEntry) {
	assert.NotNil(t, blockEntries)
	for _, blockEntry := range blockEntries {
		assert.NotNil(t, blockEntry)
		assert.NotNil(t, blockEntry.Content)
		assert.NotNil(t, blockEntry.Hash)
		assertExtIds(t, blockEntry.ExternalIDs)
	}
}

func assertHashes(t *testing.T, hashes []*eventmessages.Hash) {
	assert.NotNil(t, hashes)
	for _, hash := range hashes {
		assert.NotNil(t, hash)
		assert.NotNil(t, hash.HashValue)
	}
}

func assertExtIds(t *testing.T, extIds []*eventmessages.ExternalId) {
	assert.NotNil(t, extIds)
	for _, extId := range extIds {
		assert.NotNil(t, extId)
		assert.NotNil(t, extId.BinaryValue)
	}
}

func newDBStateMsg() *messages.DBStateMsg {
	msg := new(messages.DBStateMsg)
	msg.Timestamp = primitives.NewTimestampNow()

	set := testHelper.CreateTestBlockSet(nil)
	set = testHelper.CreateTestBlockSet(set)

	msg.DirectoryBlock = set.DBlock
	msg.AdminBlock = set.ABlock
	msg.FactoidBlock = set.FBlock
	msg.EntryCreditBlock = set.ECBlock
	msg.EBlocks = []interfaces.IEntryBlock{set.EBlock, set.AnchorEBlock}
	for _, e := range set.Entries {
		msg.Entries = append(msg.Entries, e)
	}

	return msg
}
