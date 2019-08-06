package eventservices_test

import (
	"fmt"
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

func TestEventMappers(t *testing.T) {
	t.Run("TestDBStateMapping", TestDBStateMapping)
	t.Run("TestCommitChainMapping", TestCommitChainMapping)
	t.Run("TestCommitEntryMapping", TestCommitEntryMapping)
	t.Run("TestRevealEntryMapping", TestRevealEntryMapping)
}

func TestDBStateMapping(t *testing.T) {
	msg := newDBStateMsg()
	data, _ := msg.MarshalBinary()
	assert.Len(t, data, 2409, msgChangedMessage("DBStateMsg"))

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

	_, err = anchorEvent.Marshal()
	assert.Nil(t, err)
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

func TestCommitChainMapping(t *testing.T) {
	msg := newCommitChainMsg()
	data, _ := msg.MarshalBinary()
	assert.Len(t, data, 201, msgChangedMessage("CommitChainMsg"))

	inputEvent := events.EventFromMessage(eventmessages.EventSource_ADD_TO_PROCESSLIST, msg)
	event, err := eventservices.MapToFactomEvent(inputEvent)
	if err != nil {
		t.Error(err)
	}

	assert.IsType(t, &eventmessages.FactomEvent_CommitChain{}, event.Value)
	commitChainEvent := event.Value.(*eventmessages.FactomEvent_CommitChain).CommitChain
	assert.NotNil(t, commitChainEvent.ChainIDHash)
	assert.NotNil(t, commitChainEvent.EcPubKey)
	assert.NotNil(t, commitChainEvent.Credits)
	assert.NotNil(t, commitChainEvent.Sig)
	assert.NotNil(t, commitChainEvent.Timestamp)
	assert.True(t, commitChainEvent.Timestamp.Nanos > 0)
	_, err = commitChainEvent.Marshal()
	assert.Nil(t, err)
}

func newCommitChainMsg() *messages.CommitChainMsg {
	msg := new(messages.CommitChainMsg)
	eBlock, _ := testHelper.CreateTestEntryBlock(nil)
	eBlock, _ = testHelper.CreateTestEntryBlock(eBlock) // Create a second entry to make sure we have a time
	msg.CommitChain = testHelper.NewCommitChain(eBlock)
	return msg
}

func TestCommitEntryMapping(t *testing.T) {
	msg := newCommitEntryMsg()
	data, _ := msg.MarshalBinary()
	assert.Len(t, data, 137, msgChangedMessage("CommitEntryMsg"))

	inputEvent := events.EventFromMessage(eventmessages.EventSource_ADD_TO_HOLDING, msg)
	event, err := eventservices.MapToFactomEvent(inputEvent)
	if err != nil {
		t.Error(err)
	}

	assert.IsType(t, &eventmessages.FactomEvent_CommitEntry{}, event.Value)
	commitEntryEvent := event.Value.(*eventmessages.FactomEvent_CommitEntry).CommitEntry
	assert.NotNil(t, commitEntryEvent.EntryHash)
	assert.NotNil(t, commitEntryEvent.EcPubKey)
	assert.NotNil(t, commitEntryEvent.Credits)
	assert.NotNil(t, commitEntryEvent.Sig)
	assert.NotNil(t, commitEntryEvent.Timestamp)
	assert.True(t, commitEntryEvent.Timestamp.Nanos > 0)

	_, err = commitEntryEvent.Marshal()
	assert.Nil(t, err)
}

func newCommitEntryMsg() *messages.CommitEntryMsg {
	msg := new(messages.CommitEntryMsg)
	eBlock, _ := testHelper.CreateTestEntryBlock(nil)
	eBlock, _ = testHelper.CreateTestEntryBlock(eBlock) // Create a second entry to make sure we have a time
	msg.CommitEntry = testHelper.NewCommitEntry(eBlock)
	return msg
}

func TestRevealEntryMapping(t *testing.T) {
	msg := newRevealEntryMsg()
	data, _ := msg.MarshalBinary()
	assert.Len(t, data, 60, msgChangedMessage("CommitEntryMsg"))

	inputEvent := events.EventFromMessage(eventmessages.EventSource_ADD_TO_PROCESSLIST, msg)
	event, err := eventservices.MapToFactomEvent(inputEvent)
	if err != nil {
		t.Error(err)
	}

	assert.IsType(t, &eventmessages.FactomEvent_RevealEntry{}, event.Value)
	revealEntryEvent := event.Value.(*eventmessages.FactomEvent_RevealEntry).RevealEntry
	assert.NotNil(t, revealEntryEvent.Entry)
	assert.NotNil(t, revealEntryEvent.Timestamp)
	assert.True(t, revealEntryEvent.Timestamp.Nanos > 0)

	_, err = revealEntryEvent.Marshal()
	assert.Nil(t, err)
}

func newRevealEntryMsg() *messages.RevealEntryMsg {
	msg := new(messages.RevealEntryMsg)
	eBlock, _ := testHelper.CreateTestEntryBlock(nil)
	eBlock, _ = testHelper.CreateTestEntryBlock(eBlock) // Create a second entry to make sure we have a time
	msg.CommitChain = newCommitChainMsg()
	msg.Entry = testHelper.CreateTestEntry(eBlock.Header.GetDBHeight())
	msg.Timestamp = msg.CommitChain.GetTimestamp()
	return msg
}

func msgChangedMessage(msgName string) string {
	return fmt.Sprintf("%s changed, please reevalate properties used by this event and adjust the expected message length.", msgName)
}
