package eventservices_test

import (
	"fmt"
	"github.com/FactomProject/factomd/common/directoryBlock/dbInfo"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/engine"
	"github.com/FactomProject/factomd/events/eventconfig"
	"github.com/FactomProject/factomd/events/eventinput"
	"github.com/FactomProject/factomd/events/eventmessages/generated/eventmessages"
	"github.com/FactomProject/factomd/events/eventservices"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDirectoryBlockMapping(t *testing.T) {
	msg := newDBStateMsg()
	data, _ := msg.MarshalBinary()
	assert.Len(t, data, 2409, msgChangedMessage("DBStateMsg"))

	inputEvent := eventinput.NewReplayDirectoryBlockEvent(eventmessages.EventSource_LIVE, msg)
	event, err := eventservices.MapToFactomEvent(inputEvent, eventconfig.BroadcastAlways, true)
	if err != nil {
		t.Error(err)
	}

	assert.IsType(t, &eventmessages.FactomEvent_DirectoryBlockCommit{}, event.Event)
	anchorEvent := event.Event.(*eventmessages.FactomEvent_DirectoryBlockCommit).DirectoryBlockCommit
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

func TestCommitChainMapping(t *testing.T) {
	msg := newCommitChainMsg()
	data, _ := msg.MarshalBinary()
	assert.Len(t, data, 201, msgChangedMessage("CommitChainMsg"))

	inputEvent := eventinput.NewRegistrationEvent(eventmessages.EventSource_LIVE, msg)
	event, err := eventservices.MapToFactomEvent(inputEvent, eventconfig.BroadcastAlways, true)
	if err != nil {
		t.Error(err)
	}

	assert.IsType(t, &eventmessages.FactomEvent_ChainCommit{}, event.Event)
	commitChainEvent := event.Event.(*eventmessages.FactomEvent_ChainCommit).ChainCommit
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

func TestCommitEntryMapping(t *testing.T) {
	msg := newCommitEntryMsg()
	data, _ := msg.MarshalBinary()
	assert.Len(t, data, 137, msgChangedMessage("CommitEntryMsg"))

	inputEvent := eventinput.NewRegistrationEvent(eventmessages.EventSource_LIVE, msg)
	event, err := eventservices.MapToFactomEvent(inputEvent, eventconfig.BroadcastAlways, true)
	if err != nil {
		t.Error(err)
	}

	assert.IsType(t, &eventmessages.FactomEvent_EntryCommit{}, event.Event)
	commitEntryEvent := event.Event.(*eventmessages.FactomEvent_EntryCommit).EntryCommit
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

func TestEntryRevealMapping(t *testing.T) {
	msg := newEntryRevealMsg()
	data, _ := msg.MarshalBinary()
	assert.Len(t, data, 60, msgChangedMessage("RevealEntryMsg"))

	inputEvent := eventinput.NewRegistrationEvent(eventmessages.EventSource_LIVE, msg)
	event, err := eventservices.MapToFactomEvent(inputEvent, eventconfig.BroadcastAlways, true)
	if err != nil {
		t.Error(err)
	}

	assert.IsType(t, &eventmessages.FactomEvent_EntryReveal{}, event.Event)
	entryCommit := event.Event.(*eventmessages.FactomEvent_EntryReveal).EntryReveal
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

func TestStateChangeMapping(t *testing.T) {
	msg := newCommitEntryMsg()
	data, _ := msg.MarshalBinary()
	assert.Len(t, data, 137, msgChangedMessage("CommitEntryMsg"))

	inputEvent := eventinput.NewStateChangeEvent(eventmessages.EventSource_LIVE, eventmessages.EntityState_ACCEPTED, msg)
	event, err := eventservices.MapToFactomEvent(inputEvent, eventconfig.BroadcastAlways, true)
	if err != nil {
		t.Error(err)
	}

	assert.IsType(t, &eventmessages.FactomEvent_StateChange{}, event.Event)
	stateChangedEvent := event.Event.(*eventmessages.FactomEvent_StateChange).StateChange
	assert.NotNil(t, stateChangedEvent)
	assert.EqualValues(t, eventmessages.EntityState_ACCEPTED, stateChangedEvent.EntityState)
	entityHash := msg.CommitEntry.EntryHash.Bytes()
	assert.EqualValues(t, entityHash, stateChangedEvent.EntityHash)
}

func msgChangedMessage(msgName string) string {
	return fmt.Sprintf("%s changed, please reevalate properties used by this event and adjust the expected message length.", msgName)
}

func TestAnchorEventMapping(t *testing.T) {
	dirBlockInfo := testHelper.CreateTestDirBlockInfo(&dbInfo.DirBlockInfo{DBHeight: 100})
	inputEvent := eventinput.NewAnchorEvent(eventmessages.EventSource_LIVE, dirBlockInfo)
	event, err := eventservices.MapToFactomEvent(inputEvent, eventconfig.BroadcastAlways, true)
	if err != nil {
		t.Error(err)
	}
	assert.IsType(t, &eventmessages.FactomEvent_DirectoryBlockAnchor{}, event.Event)
	dirBlockAnchorEvent := event.Event.(*eventmessages.FactomEvent_DirectoryBlockAnchor).DirectoryBlockAnchor
	assert.NotNil(t, dirBlockAnchorEvent)
	assert.EqualValues(t, int32(101000), dirBlockAnchorEvent.Timestamp.Nanos)
	assert.NotNil(t, dirBlockAnchorEvent.BtcTxHash)
	assert.NotNil(t, dirBlockAnchorEvent.BtcBlockHash)
	assert.EqualValues(t, 101, dirBlockAnchorEvent.BtcBlockHeight)
	assert.EqualValues(t, 101, dirBlockAnchorEvent.BtcTxOffset)
	assert.True(t, dirBlockAnchorEvent.BtcConfirmed)
	assert.NotNil(t, 101, dirBlockAnchorEvent.EthereumAnchorRecordEntryHash)
	assert.True(t, dirBlockAnchorEvent.EthereumConfirmed)
}

func TestMapToFactomEvent(t *testing.T) {
	testCases := map[string]struct {
		Input                    eventinput.EventInput
		BroadcastContent         eventconfig.BroadcastContent
		EventReplayDuringStartup bool
		Assertion                func(t *testing.T, event *eventmessages.FactomEvent)
	}{
		"RegistrationEvent": {
			Input:            eventinput.NewRegistrationEvent(eventmessages.EventSource_REPLAY_BOOT, newTestCommitChain()),
			BroadcastContent: eventconfig.BroadcastAlways,
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
			Input:            eventinput.NewRegistrationEvent(eventmessages.EventSource_REPLAY_BOOT, newTestEntryReveal()),
			BroadcastContent: eventconfig.BroadcastAlways,
			Assertion: func(t *testing.T, event *eventmessages.FactomEvent) {
				assert.Equal(t, eventmessages.EventSource_REPLAY_BOOT, event.EventSource)
				if assert.NotNil(t, event.GetEntryReveal()) {
					assert.NotNil(t, event.GetEntryReveal().Timestamp)
					assert.NotNil(t, event.GetEntryReveal().Entry)
					assert.Equal(t, eventmessages.EntityState_REQUESTED, event.GetEntryReveal().EntityState)
				}
			},
		},
		"StateChangeEventFromCommitChain": {
			Input:                    eventinput.NewStateChangeEvent(eventmessages.EventSource_REPLAY_BOOT, eventmessages.EntityState_REJECTED, newTestCommitChain()),
			BroadcastContent:         eventconfig.BroadcastOnce,
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
			Input:                    eventinput.NewStateChangeEvent(eventmessages.EventSource_REPLAY_BOOT, eventmessages.EntityState_REJECTED, newTestCommitEntry()),
			BroadcastContent:         eventconfig.BroadcastOnce,
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
			Input:                    eventinput.NewStateChangeEvent(eventmessages.EventSource_REPLAY_BOOT, eventmessages.EntityState_REJECTED, newTestEntryReveal()),
			BroadcastContent:         eventconfig.BroadcastOnce,
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
			BroadcastContent:         eventconfig.BroadcastAlways,
			EventReplayDuringStartup: false,
			Input:                    eventinput.NewStateChangeEvent(eventmessages.EventSource_REPLAY_BOOT, eventmessages.EntityState_REJECTED, newTestCommitChain()),
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
			Input:                    eventinput.NewStateChangeEvent(eventmessages.EventSource_REPLAY_BOOT, eventmessages.EntityState_REJECTED, newTestCommitEntry()),
			BroadcastContent:         eventconfig.BroadcastAlways,
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
			Input:                    eventinput.NewStateChangeEvent(eventmessages.EventSource_REPLAY_BOOT, eventmessages.EntityState_REJECTED, newTestEntryReveal()),
			BroadcastContent:         eventconfig.BroadcastAlways,
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
		"DirectoryBlockEvent": {
			Input:                    eventinput.NewDirectoryBlockEvent(eventmessages.EventSource_REPLAY_BOOT, newTestDirectoryBlockState()),
			BroadcastContent:         eventconfig.BroadcastNever,
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
		"DirectoryBlockContent": {
			Input:                    eventinput.NewDirectoryBlockEvent(eventmessages.EventSource_REPLAY_BOOT, newTestDirectoryBlockState()),
			BroadcastContent:         eventconfig.BroadcastAlways,
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
		"DirectoryBlockAnchor": {
			Input:                    eventinput.NewAnchorEvent(eventmessages.EventSource_REPLAY_BOOT, newTestDirectoryBlockInfo()),
			BroadcastContent:         eventconfig.BroadcastAlways,
			EventReplayDuringStartup: true,
			Assertion: func(t *testing.T, event *eventmessages.FactomEvent) {
				assert.Equal(t, eventmessages.EventSource_REPLAY_BOOT, event.EventSource)
				if assert.NotNil(t, event.GetDirectoryBlockAnchor()) {
					assert.NotNil(t, event.GetDirectoryBlockAnchor().GetDirectoryBlockHash())
					assert.NotNil(t, event.GetDirectoryBlockAnchor().GetDirectoryBlockMerkleRoot())
					assert.NotNil(t, event.GetDirectoryBlockAnchor().GetBlockHeight())
					assert.NotNil(t, event.GetDirectoryBlockAnchor().GetBtcBlockHash())
					assert.NotNil(t, event.GetDirectoryBlockAnchor().GetBtcBlockHeight())
					assert.NotNil(t, event.GetDirectoryBlockAnchor().GetBtcBlockHash())
					assert.NotNil(t, event.GetDirectoryBlockAnchor().GetBtcConfirmed())
					assert.NotNil(t, event.GetDirectoryBlockAnchor().GetBtcTxHash())
					assert.NotNil(t, event.GetDirectoryBlockAnchor().GetBtcTxOffset())
					assert.NotNil(t, event.GetDirectoryBlockAnchor().GetEthereumAnchorRecordEntryHash())
					assert.NotNil(t, event.GetDirectoryBlockAnchor().GetEthereumConfirmed())
					assert.NotNil(t, event.GetDirectoryBlockAnchor().GetTimestamp())
				}
			},
		},
		"ReplayDirectoryBlock": {
			Input:                    eventinput.NewReplayDirectoryBlockEvent(eventmessages.EventSource_REPLAY_BOOT, newDBStateMsg()),
			BroadcastContent:         eventconfig.BroadcastOnce,
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
		"ProcessListEvent": {
			Input:            eventinput.ProcessListEventNewMinute(eventmessages.EventSource_REPLAY_BOOT, 2, 123),
			BroadcastContent: eventconfig.BroadcastAlways,
			Assertion: func(t *testing.T, event *eventmessages.FactomEvent) {
				assert.Equal(t, eventmessages.EventSource_REPLAY_BOOT, event.EventSource)
				if assert.NotNil(t, event.GetProcessListEvent()) && assert.NotNil(t, event.GetProcessListEvent().GetNewMinuteEvent()) {
					assert.Equal(t, uint32(2), event.GetProcessListEvent().GetNewMinuteEvent().GetNewMinute())
					assert.Equal(t, uint32(123), event.GetProcessListEvent().GetNewMinuteEvent().GetBlockHeight())
				}
			},
		},
		"NodeMessage": {
			Input:            eventinput.NodeInfoMessage(eventmessages.NodeMessageCode_STARTED, "test message"),
			BroadcastContent: eventconfig.BroadcastAlways,
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
	input := eventinput.NewRegistrationEvent(eventmessages.EventSource_REPLAY_BOOT, newTestEntryReveal())
	event, err := eventservices.MapToFactomEvent(input, eventconfig.BroadcastNever, true)
	assert.Nil(t, err)
	assert.Nil(t, event)
}

func TestConvertTimeToTimestamp(t *testing.T) {
	// 2019-10-24 11:56:18.338002 = 1571910978 and 338001966 nanos
	loc := time.FixedZone("UTC-8", -8*60*60)
	now := time.Date(2019, 10, 24, 11, 56, 18, 338001966, loc)
	timestamp := eventservices.ConvertTimeToTimestamp(now)

	assert.NotNil(t, timestamp)
	assert.Equal(t, int64(1571946978), timestamp.Seconds)
	assert.Equal(t, int32(338001966), timestamp.Nanos)
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

func newTestDirectoryBlockState() interfaces.IDBState {
	set := testHelper.CreateTestBlockSet(nil)
	set = testHelper.CreateTestBlockSet(set)

	msg := new(state.DBState)
	msg.DirectoryBlock = set.DBlock
	msg.AdminBlock = set.ABlock
	msg.FactoidBlock = set.FBlock
	msg.EntryCreditBlock = set.ECBlock

	return msg
}

func newTestDirectoryBlockInfo() interfaces.IDirBlockInfo {
	return testHelper.CreateTestDirBlockInfo(&dbInfo.DirBlockInfo{DBHeight: 910})
}
