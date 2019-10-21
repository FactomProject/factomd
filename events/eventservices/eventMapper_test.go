package eventservices_test

import (
	"fmt"
	"github.com/FactomProject/factomd/common/constants/runstate"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/engine"
	"github.com/FactomProject/factomd/events"
	"github.com/FactomProject/factomd/events/eventmessages/generated/eventmessages"
	"github.com/FactomProject/factomd/events/eventservices"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
	"github.com/stretchr/testify/assert"
	"testing"
)

var testState = &state.State{
	IdentityChainID: primitives.NewZeroHash(),
	RunState:        runstate.Running,
}

func TestEventMappers(t *testing.T) {
	t.Run("Run eventmapper tests", func(t *testing.T) {
		t.Run("TestDBStateMapping", testDBStateMapping)
		t.Run("TestCommitChainMapping", testCommitChainMapping)
		t.Run("TestCommitEntryMapping", testCommitEntryMapping)
		t.Run("TestStateChangeMapping", testStateChangeMapping)
		t.Run("TestEntryRevealMapping", testEntryRevealMapping)
	})
}

func testDBStateMapping(t *testing.T) {
	msg := newDBStateMsg()
	data, _ := msg.MarshalBinary()
	assert.Len(t, data, 2409, msgChangedMessage("DBStateMsg"))

	inputEvent := events.NewStateChangeEventFromMsg(eventmessages.EventSource_LIVE, eventmessages.EntityState_COMMITTED_TO_DIRECTORY_BLOCK, msg)
	event, err := eventservices.MapToFactomEvent(inputEvent, eventservices.BroadcastAlways, true)
	if err != nil {
		t.Error(err)
	}

	assert.IsType(t, &eventmessages.FactomEvent_DirectoryBlockCommit{}, event.Value)
	anchorEvent := event.Value.(*eventmessages.FactomEvent_DirectoryBlockCommit).DirectoryBlockCommit
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
	assert.NotNil(t, factoidBlock.ExchangeRate)
	assert.NotNil(t, factoidBlock.BlockHeight)
	assertTransactions(t, factoidBlock.Transactions)
}

func assertTransactions(t *testing.T, transactions []*eventmessages.Transaction) {
	assert.NotNil(t, transactions)
	for _, transaction := range transactions {
		assert.NotNil(t, transaction)
		assert.NotNil(t, transaction.TransactionID)
		assert.NotNil(t, transaction.BlockHeight)
		assert.NotNil(t, transaction.FactoidOutputs)
		for _, output := range transaction.FactoidOutputs {
			assert.True(t, output.Amount > 0)
			assert.NotNil(t, output.Address)
		}
	}
}

func assertEntryBlocks(t *testing.T, entryBlocks []*eventmessages.EntryBlock) {
	assert.NotNil(t, entryBlocks)
	for _, entryBlock := range entryBlocks {
		assert.NotNil(t, entryBlock)
		assert.NotNil(t, entryBlock.Header)
		assertHashes(t, entryBlock.EntryHashes)
	}
}

func assertEntryBlockEntries(t *testing.T, blockEntries []*eventmessages.EntryBlockEntry) {
	assert.NotNil(t, blockEntries)
	for _, blockEntry := range blockEntries {
		assert.NotNil(t, blockEntry)
		assert.NotNil(t, blockEntry.Content)
		assert.NotNil(t, blockEntry.Hash)
		assertHashes(t, blockEntry.ExternalIDs)
	}
}

func assertHashes(t *testing.T, hashes [][]byte) {
	assert.NotNil(t, hashes)
	for _, hash := range hashes {
		assert.NotNil(t, hash)
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

func testCommitChainMapping(t *testing.T) {
	msg := newCommitChainMsg()
	data, _ := msg.MarshalBinary()
	assert.Len(t, data, 201, msgChangedMessage("CommitChainMsg"))

	inputEvent := events.NewRegistrationEvent(eventmessages.EventSource_LIVE, msg)
	event, err := eventservices.MapToFactomEvent(inputEvent, eventservices.BroadcastAlways, true)
	if err != nil {
		t.Error(err)
	}

	assert.IsType(t, &eventmessages.FactomEvent_ChainCommit{}, event.Value)
	commitChainEvent := event.Value.(*eventmessages.FactomEvent_ChainCommit).ChainCommit
	assert.NotNil(t, commitChainEvent.ChainIDHash)
	assert.NotNil(t, commitChainEvent.EntryCreditPublicKey)
	assert.NotNil(t, commitChainEvent.Credits)
	assert.NotNil(t, commitChainEvent.Signature)
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

func testCommitEntryMapping(t *testing.T) {
	msg := newCommitEntryMsg()
	data, _ := msg.MarshalBinary()
	assert.Len(t, data, 137, msgChangedMessage("CommitEntryMsg"))

	inputEvent := events.NewRegistrationEvent(eventmessages.EventSource_LIVE, msg)
	event, err := eventservices.MapToFactomEvent(inputEvent, eventservices.BroadcastAlways, true)
	if err != nil {
		t.Error(err)
	}

	assert.IsType(t, &eventmessages.FactomEvent_EntryCommit{}, event.Value)
	commitEntryEvent := event.Value.(*eventmessages.FactomEvent_EntryCommit).EntryCommit
	assert.NotNil(t, commitEntryEvent.EntryHash)
	assert.NotNil(t, commitEntryEvent.EntryCreditPublicKey)
	assert.NotNil(t, commitEntryEvent.Credits)
	assert.NotNil(t, commitEntryEvent.Signature)
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

func testEntryRevealMapping(t *testing.T) {
	msg := newEntryRevealMsg()
	data, _ := msg.MarshalBinary()
	assert.Len(t, data, 60, msgChangedMessage("RevealEntryMsg"))

	inputEvent := events.NewRegistrationEvent(eventmessages.EventSource_LIVE, msg)
	event, err := eventservices.MapToFactomEvent(inputEvent, eventservices.BroadcastAlways, true)
	if err != nil {
		t.Error(err)
	}

	assert.IsType(t, &eventmessages.FactomEvent_EntryReveal{}, event.Value)
	entryCommit := event.Value.(*eventmessages.FactomEvent_EntryReveal).EntryReveal
	assert.NotNil(t, entryCommit.Entry)
	assert.NotNil(t, entryCommit.Timestamp)
	assert.True(t, entryCommit.Timestamp.Nanos > 0)
	assert.NotNil(t, entryCommit.Entry.Version)
	assert.NotNil(t, entryCommit.Entry.Hash)
	assert.NotNil(t, entryCommit.Entry.ChainID)
	assert.NotNil(t, entryCommit.Entry.Content)
	assert.NotNil(t, entryCommit.Entry.ExternalIDs)

	_, err = entryCommit.Marshal()
	assert.Nil(t, err)
}

func newEntryRevealMsg() *messages.RevealEntryMsg {
	msg := new(messages.RevealEntryMsg)
	eBlock, _ := testHelper.CreateTestEntryBlock(nil)
	eBlock, _ = testHelper.CreateTestEntryBlock(eBlock) // Create a second entry to make sure we have a time
	msg.CommitChain = newCommitChainMsg()
	msg.Entry = testHelper.CreateTestEntry(eBlock.Header.GetDBHeight())
	msg.Timestamp = msg.CommitChain.GetTimestamp()
	return msg
}

func testStateChangeMapping(t *testing.T) {
	msg := newCommitEntryMsg()
	data, _ := msg.MarshalBinary()
	assert.Len(t, data, 137, msgChangedMessage("CommitEntryMsg"))

	inputEvent := events.NewStateChangeEventFromMsg(eventmessages.EventSource_LIVE, eventmessages.EntityState_ACCEPTED, msg)
	event, err := eventservices.MapToFactomEvent(inputEvent, eventservices.BroadcastAlways, true)
	if err != nil {
		t.Error(err)
	}

	assert.IsType(t, &eventmessages.FactomEvent_StateChange{}, event.Value)
	stateChangedEvent := event.Value.(*eventmessages.FactomEvent_StateChange).StateChange
	assert.NotNil(t, stateChangedEvent)
	assert.EqualValues(t, eventmessages.EntityState_ACCEPTED, stateChangedEvent.EntityState)
	entityHash := msg.CommitEntry.EntryHash.Bytes()
	assert.EqualValues(t, entityHash, stateChangedEvent.EntityHash)
}

func msgChangedMessage(msgName string) string {
	return fmt.Sprintf("%s changed, please reevalate properties used by this event and adjust the expected message length.", msgName)
}

func TestMapToFactomEvent(t *testing.T) {
	testCases := map[string]struct {
		Input                    events.EventInput
		BroadcastContent         eventservices.BroadcastContent
		EventReplayDuringStartup bool
		Assertion                func(t *testing.T, event *eventmessages.FactomEvent)
	}{
		"RegistrationEvent": {
			Input:            events.NewRegistrationEvent(eventmessages.EventSource_REPLAY_BOOT, newTestCommitChain()),
			BroadcastContent: eventservices.BroadcastAlways,
			Assertion: func(t *testing.T, event *eventmessages.FactomEvent) {
				assert.Equal(t, eventmessages.EventSource_REPLAY_BOOT, event.EventSource)
				if assert.NotNil(t, event.GetChainCommit()) {
					assert.NotNil(t, event.GetChainCommit().Version)
					assert.NotNil(t, event.GetChainCommit().Timestamp)
					assert.NotNil(t, event.GetChainCommit().ChainIDHash)
					assert.NotNil(t, event.GetChainCommit().Signature)
					assert.NotNil(t, event.GetChainCommit().EntryHash)
					assert.NotNil(t, event.GetChainCommit().EntryCreditPublicKey)
					assert.NotNil(t, event.GetChainCommit().Credits)
					assert.NotNil(t, event.GetChainCommit().Weld)
					assert.Equal(t, eventmessages.EntityState_REQUESTED, event.GetChainCommit().EntityState)
				}
			},
		},
		"RegistrationEventWithRevealContent": {
			Input:            events.NewRegistrationEvent(eventmessages.EventSource_REPLAY_BOOT, newTestEntryReveal()),
			BroadcastContent: eventservices.BroadcastAlways,
			Assertion: func(t *testing.T, event *eventmessages.FactomEvent) {
				assert.Equal(t, eventmessages.EventSource_REPLAY_BOOT, event.EventSource)
				if assert.NotNil(t, event.GetEntryReveal()) {
					assert.NotNil(t, event.GetEntryReveal().Timestamp)
					assert.NotNil(t, event.GetEntryReveal().Entry)
					assert.Equal(t, eventmessages.EntityState_REQUESTED, event.GetEntryReveal().EntityState)
				}
			},
		},
		"StateChangeEventFromDirectoryBlock": {
			Input:                    events.NewStateChangeEventFromMsg(eventmessages.EventSource_REPLAY_BOOT, eventmessages.EntityState_ACCEPTED, newDBStateMsg()),
			BroadcastContent:         eventservices.BroadcastOnce,
			EventReplayDuringStartup: true,
			Assertion: func(t *testing.T, event *eventmessages.FactomEvent) {
				assert.Equal(t, eventmessages.EventSource_REPLAY_BOOT, event.EventSource)
				if assert.NotNil(t, event.GetDirectoryBlockCommit()) {
					assert.NotNil(t, event.GetDirectoryBlockCommit().GetDirectoryBlock())
					assert.NotNil(t, event.GetDirectoryBlockCommit().GetEntryBlockEntries())
					assert.NotNil(t, event.GetDirectoryBlockCommit().GetEntryBlocks())
					assert.NotNil(t, event.GetDirectoryBlockCommit().GetFactoidBlock())
				}
			},
		},
		"StateChangeEventFromCommitChain": {
			Input:                    events.NewStateChangeEventFromMsg(eventmessages.EventSource_REPLAY_BOOT, eventmessages.EntityState_REJECTED, newTestCommitChain()),
			BroadcastContent:         eventservices.BroadcastOnce,
			EventReplayDuringStartup: true,
			Assertion: func(t *testing.T, event *eventmessages.FactomEvent) {
				assert.Equal(t, eventmessages.EventSource_REPLAY_BOOT, event.EventSource)
				if assert.NotNil(t, event.GetStateChange()) {
					assert.NotNil(t, event.GetStateChange().GetBlockHeight())
					assert.NotNil(t, event.GetStateChange().GetEntityHash())
					assert.Equal(t, eventmessages.EntityState_REJECTED, event.GetStateChange().GetEntityState())
				}
			},
		},
		"StateChangeEventFromCommitEntry": {
			Input:                    events.NewStateChangeEventFromMsg(eventmessages.EventSource_REPLAY_BOOT, eventmessages.EntityState_REJECTED, newTestCommitEntry()),
			BroadcastContent:         eventservices.BroadcastOnce,
			EventReplayDuringStartup: true,
			Assertion: func(t *testing.T, event *eventmessages.FactomEvent) {
				assert.Equal(t, eventmessages.EventSource_REPLAY_BOOT, event.EventSource)
				if assert.NotNil(t, event.GetStateChange()) {
					assert.NotNil(t, event.GetStateChange().GetBlockHeight())
					assert.NotNil(t, event.GetStateChange().GetEntityHash())
					assert.NotNil(t, eventmessages.EntityState_REJECTED, event.GetStateChange().GetEntityState())
				}
			},
		},
		"StateChangeEventEntryReveal": {
			Input:                    events.NewStateChangeEventFromMsg(eventmessages.EventSource_REPLAY_BOOT, eventmessages.EntityState_REJECTED, newTestEntryReveal()),
			BroadcastContent:         eventservices.BroadcastOnce,
			EventReplayDuringStartup: true,
			Assertion: func(t *testing.T, event *eventmessages.FactomEvent) {
				assert.Equal(t, eventmessages.EventSource_REPLAY_BOOT, event.EventSource)
				if assert.NotNil(t, event.GetStateChange()) {
					assert.NotNil(t, event.GetStateChange().GetBlockHeight())
					assert.NotNil(t, event.GetStateChange().GetEntityHash())
					assert.NotNil(t, eventmessages.EntityState_REJECTED, event.GetStateChange().GetEntityState())
				}
			},
		},
		"StateChangeEventFromCommitChainResend": {
			BroadcastContent:         eventservices.BroadcastAlways,
			EventReplayDuringStartup: false,
			Input:                    events.NewStateChangeEventFromMsg(eventmessages.EventSource_REPLAY_BOOT, eventmessages.EntityState_REJECTED, newTestCommitChain()),
			Assertion: func(t *testing.T, event *eventmessages.FactomEvent) {
				assert.Equal(t, eventmessages.EventSource_REPLAY_BOOT, event.EventSource)
				if assert.NotNil(t, event.GetChainCommit()) {
					assert.NotNil(t, event.GetChainCommit().Version)
					assert.NotNil(t, event.GetChainCommit().Timestamp)
					assert.NotNil(t, event.GetChainCommit().ChainIDHash)
					assert.NotNil(t, event.GetChainCommit().Signature)
					assert.NotNil(t, event.GetChainCommit().EntryHash)
					assert.NotNil(t, event.GetChainCommit().EntryCreditPublicKey)
					assert.NotNil(t, event.GetChainCommit().Credits)
					assert.NotNil(t, event.GetChainCommit().Weld)
					assert.Equal(t, eventmessages.EntityState_REJECTED, event.GetChainCommit().EntityState)
				}
			},
		},
		"StateChangeEventFromCommitEntryResend": {
			Input:                    events.NewStateChangeEventFromMsg(eventmessages.EventSource_REPLAY_BOOT, eventmessages.EntityState_REJECTED, newTestCommitEntry()),
			BroadcastContent:         eventservices.BroadcastAlways,
			EventReplayDuringStartup: false,
			Assertion: func(t *testing.T, event *eventmessages.FactomEvent) {
				assert.Equal(t, eventmessages.EventSource_REPLAY_BOOT, event.EventSource)
				if assert.NotNil(t, event.GetEntryCommit()) {
					assert.NotNil(t, event.GetEntryCommit().Version)
					assert.NotNil(t, event.GetEntryCommit().Timestamp)
					assert.NotNil(t, event.GetEntryCommit().Signature)
					assert.NotNil(t, event.GetEntryCommit().EntryHash)
					assert.NotNil(t, event.GetEntryCommit().EntryCreditPublicKey)
					assert.NotNil(t, event.GetEntryCommit().Credits)
					assert.Equal(t, eventmessages.EntityState_REJECTED, event.GetEntryCommit().EntityState)
				}
			},
		},
		"StateChangeEventEntryRevealResend": {
			Input:                    events.NewStateChangeEventFromMsg(eventmessages.EventSource_REPLAY_BOOT, eventmessages.EntityState_REJECTED, newTestEntryReveal()),
			BroadcastContent:         eventservices.BroadcastAlways,
			EventReplayDuringStartup: false,
			Assertion: func(t *testing.T, event *eventmessages.FactomEvent) {
				assert.Equal(t, eventmessages.EventSource_REPLAY_BOOT, event.EventSource)
				if assert.NotNil(t, event.GetEntryReveal()) {
					assert.NotNil(t, event.GetEntryReveal().GetEntry())
					assert.NotNil(t, event.GetEntryReveal().GetTimestamp())
					assert.NotNil(t, eventmessages.EntityState_REJECTED, event.GetEntryReveal().GetEntityState())
				}
			},
		},
		"StateChangeEventDirectoryBlock": {
			Input:                    events.NewStateChangeEvent(eventmessages.EventSource_REPLAY_BOOT, eventmessages.EntityState_REJECTED, newTestDBState()),
			BroadcastContent:         eventservices.BroadcastAlways,
			EventReplayDuringStartup: true,
			Assertion: func(t *testing.T, event *eventmessages.FactomEvent) {
				assert.Equal(t, eventmessages.EventSource_REPLAY_BOOT, event.EventSource)
				if assert.NotNil(t, event.GetDirectoryBlockCommit()) {
					assert.NotNil(t, event.GetDirectoryBlockCommit().GetDirectoryBlock())
					assert.NotNil(t, event.GetDirectoryBlockCommit().GetEntryBlockEntries())
					assert.NotNil(t, event.GetDirectoryBlockCommit().GetEntryBlocks())
					assert.NotNil(t, event.GetDirectoryBlockCommit().GetFactoidBlock())
				}
			},
		},
		"StateChangeEventDirectoryBlockContent": {
			Input:                    events.NewStateChangeEvent(eventmessages.EventSource_REPLAY_BOOT, eventmessages.EntityState_REJECTED, newTestDBState()),
			BroadcastContent:         eventservices.BroadcastAlways,
			EventReplayDuringStartup: true,
			Assertion: func(t *testing.T, event *eventmessages.FactomEvent) {
				assert.Equal(t, eventmessages.EventSource_REPLAY_BOOT, event.EventSource)
				if assert.NotNil(t, event.GetDirectoryBlockCommit()) {
					assert.NotNil(t, event.GetDirectoryBlockCommit().GetDirectoryBlock())
					assert.NotNil(t, event.GetDirectoryBlockCommit().GetEntryBlockEntries())
					assert.NotNil(t, event.GetDirectoryBlockCommit().GetEntryBlocks())
					assert.NotNil(t, event.GetDirectoryBlockCommit().GetFactoidBlock())
				}
			},
		},
		"ProcessMessage": {
			Input:            events.ProcessInfoMessage(eventmessages.EventSource_REPLAY_BOOT, eventmessages.ProcessCode_NEW_MINUTE, "test message"),
			BroadcastContent: eventservices.BroadcastAlways,
			Assertion: func(t *testing.T, event *eventmessages.FactomEvent) {
				assert.Equal(t, eventmessages.EventSource_REPLAY_BOOT, event.EventSource)
				if assert.NotNil(t, event.GetProcessMessage()) {
					assert.NotNil(t, event.GetProcessMessage().Level)
					assert.Equal(t, eventmessages.ProcessCode_NEW_MINUTE, event.GetProcessMessage().ProcessCode)
					assert.Equal(t, "test message", event.GetProcessMessage().MessageText)
				}
			},
		},
		"NodeMessage": {
			Input:            events.NodeInfoMessage(eventmessages.NodeMessageCode_STARTED, "test message"),
			BroadcastContent: eventservices.BroadcastAlways,
			Assertion: func(t *testing.T, event *eventmessages.FactomEvent) {
				if assert.NotNil(t, event.GetNodeMessage()) {
					assert.NotNil(t, event.GetNodeMessage().Level)
					assert.NotNil(t, event.GetNodeMessage().MessageCode)
					assert.Equal(t, "test message", event.GetNodeMessage().MessageText)
				}
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			event, err := eventservices.MapToFactomEvent(testCase.Input, testCase.BroadcastContent, testCase.EventReplayDuringStartup)
			assert.Nil(t, err)

			if assert.NotNil(t, event) {
				assert.NotNil(t, event.FactomNodeName)
				assert.NotNil(t, event.EventSource)

				testCase.Assertion(t, event)
			}
		})
	}
}

func TestMapToFactomEventRevealNoContentRegistration(t *testing.T) {
	// same test as TestMapToFactomEvent, except for a registration event with an EntryReveal registration
	// then no value will be set
	input := events.NewRegistrationEvent(eventmessages.EventSource_REPLAY_BOOT, newTestEntryReveal())
	event, err := eventservices.MapToFactomEvent(input, eventservices.BroadcastNever, true)
	assert.Nil(t, err)
	assert.Nil(t, event)
}

func setServiceState(state eventservices.BroadcastContent) func(t *testing.T) {
	return func(t *testing.T) {
		params := &eventservices.EventServiceParams{
			BroadcastContent: state,
		}
		testState.EventsService, testState.EventsServiceControl = eventservices.NewEventServiceTo(testState, params)
	}
}

func newTestCommitChain() interfaces.IMsg {
	msg := new(messages.CommitChainMsg)
	msg.CommitChain = entryCreditBlock.NewCommitChain()
	msg.CommitChain.ChainIDHash.SetBytes([]byte(""))
	msg.CommitChain.ECPubKey = new(primitives.ByteSlice32)
	msg.CommitChain.Sig = new(primitives.ByteSlice64)
	msg.CommitChain.Weld.SetBytes([]byte("1"))
	return msg
}

func newTestCommitEntry() interfaces.IMsg {
	msg := messages.NewCommitEntryMsg()
	msg.CommitEntry = entryCreditBlock.NewCommitEntry()
	msg.CommitEntry.Init()
	msg.CommitEntry.EntryHash = msg.CommitEntry.Hash()
	return msg
}

func newTestEntryReveal() interfaces.IMsg {
	msg := messages.NewRevealEntryMsg()
	msg.Entry = engine.RandomEntry()
	msg.Timestamp = primitives.NewTimestampNow()
	return msg
}

func newTestDBState() interfaces.IDBState {
	set := testHelper.CreateTestBlockSet(nil)
	set = testHelper.CreateTestBlockSet(set)

	msg := new(state.DBState)
	msg.DirectoryBlock = set.DBlock
	msg.AdminBlock = set.ABlock
	msg.FactoidBlock = set.FBlock
	msg.EntryCreditBlock = set.ECBlock

	return msg
}
