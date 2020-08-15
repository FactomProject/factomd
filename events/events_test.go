package events_test

import (
	"bytes"
	"encoding/binary"
	"reflect"

	"github.com/PaulSnow/factom2d/common/adminBlock"
	"github.com/PaulSnow/factom2d/common/directoryBlock"
	"github.com/PaulSnow/factom2d/common/directoryBlock/dbInfo"
	"github.com/PaulSnow/factom2d/common/entryBlock"
	"github.com/PaulSnow/factom2d/common/entryCreditBlock"
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/messages"
	"github.com/PaulSnow/factom2d/common/primitives"
	"github.com/PaulSnow/factom2d/database/databaseOverlay"
	"github.com/PaulSnow/factom2d/events/eventconfig"
	"github.com/PaulSnow/factom2d/events/eventmessages/generated/eventmessages"
	"github.com/PaulSnow/factom2d/p2p"
	"github.com/PaulSnow/factom2d/testHelper"
	"github.com/stretchr/testify/assert"

	"testing"
	"time"
)

func TestUpdateState(t *testing.T) {
	eventQueue := make(chan *eventmessages.FactomEvent, p2p.StandardChannelSize)
	mockSender := &mockEventSender{
		eventsOutQueue:      eventQueue,
		replayDuringStartup: true,
		sendStateChange:     true,
	}

	s := testHelper.CreateAndPopulateTestState()
	s.EventService.ConfigSender(s, mockSender)

	eBlocks := []interfaces.IEntryBlock{}
	entries := []interfaces.IEBEntry{}
	dBlock := directoryBlock.NewDirectoryBlock(s.LeaderPL.DirectoryBlock)
	aBlock := adminBlock.NewAdminBlock(s.LeaderPL.AdminBlock)
	ecBlock := entryCreditBlock.NewECBlock()

	dbState := s.AddDBState(true, dBlock, aBlock, s.GetFactoidState().GetCurrentBlock(), ecBlock, eBlocks, entries)
	dbState.ReadyToSave = true
	dbState.Signed = true
	dbState.Repeat = true

	s.DBStates.SavedHeight = dbState.DirectoryBlock.GetHeader().GetDBHeight()

	progress := s.DBStates.UpdateState()

	assert.True(t, progress)
	if assert.Equal(t, 1, len(eventQueue)) {
		event := <-eventQueue
		assert.Equal(t, eventmessages.EventSource_REPLAY_BOOT, event.GetEventSource())
		assert.Equal(t, s.IdentityChainID.Bytes(), event.IdentityChainID)
		assert.NotNil(t, event.Event)

		stateChangeEvent := event.GetDirectoryBlockCommit()
		if assert.NotNil(t, stateChangeEvent, "event received has wrong type: %s event: %+v", reflect.TypeOf(event.GetEvent()), event) {
			assert.NotNil(t, stateChangeEvent.AdminBlock)
			assert.NotNil(t, stateChangeEvent.DirectoryBlock)
			assert.NotNil(t, stateChangeEvent.FactoidBlock)
			assert.NotNil(t, stateChangeEvent.EntryCreditBlock)
		}
	}
}

func TestAddToAndDeleteFromHolding(t *testing.T) {
	eventQueue := make(chan *eventmessages.FactomEvent, p2p.StandardChannelSize)
	mockSender := &mockEventSender{
		eventsOutQueue:      eventQueue,
		replayDuringStartup: true,
		sendStateChange:     true,
	}

	s := testHelper.CreateAndPopulateTestState()
	s.EventService.ConfigSender(s, mockSender)

	msg := &messages.CommitChainMsg{CommitChain: entryCreditBlock.NewCommitChain()}
	msg.CommitChain.MilliTime = createByteSlice6Timestamp(-2 * 1e3)

	s.AddToHolding(msg.GetMsgHash().Fixed(), msg)
	s.DeleteFromHolding(msg.GetMsgHash().Fixed(), msg, "to unit test")

	// assertions
	if assert.Equal(t, 2, len(eventQueue)) {
		addToHoldingEvent := <-eventQueue
		assert.Equal(t, eventmessages.EventSource_REPLAY_BOOT, addToHoldingEvent.GetEventSource())
		assert.Equal(t, s.IdentityChainID.Bytes(), addToHoldingEvent.IdentityChainID)
		assert.NotNil(t, addToHoldingEvent.Event)

		registrationEvent := addToHoldingEvent.GetChainCommit()
		if assert.NotNil(t, registrationEvent, "event received has wrong type: %s event: %+v", reflect.TypeOf(addToHoldingEvent.GetEvent()), addToHoldingEvent) {
			assert.Equal(t, msg.CommitChain.ChainIDHash.Bytes(), registrationEvent.GetChainIDHash())
		}

		deleteFromHoldingEvent := <-eventQueue
		assert.Equal(t, eventmessages.EventSource_REPLAY_BOOT, deleteFromHoldingEvent.GetEventSource())

		stateChangeEvent := deleteFromHoldingEvent.GetStateChange()
		if assert.NotNil(t, stateChangeEvent, "event received has wrong type: %s event: %+v", reflect.TypeOf(deleteFromHoldingEvent.GetEvent()), deleteFromHoldingEvent) {
			assert.Equal(t, eventmessages.EntityState_REJECTED, stateChangeEvent.GetEntityState())
			assert.NotNil(t, stateChangeEvent.GetEntityHash())
			assert.NotNil(t, stateChangeEvent.GetEntityState())
			assert.NotNil(t, stateChangeEvent.GetBlockHeight())
		}
	}
}

func TestAddToProcessList(t *testing.T) {
	eventQueue := make(chan *eventmessages.FactomEvent, p2p.StandardChannelSize)
	mockSender := &mockEventSender{
		eventsOutQueue:      eventQueue,
		replayDuringStartup: true,
		sendStateChange:     true,
	}

	s := testHelper.CreateAndPopulateTestState()
	s.EventService.ConfigSender(s, mockSender)
	s.SetLeaderTimestamp(primitives.NewTimestampNow())

	msg := &messages.CommitChainMsg{CommitChain: entryCreditBlock.NewCommitChain()}
	msg.CommitChain.MilliTime = createByteSlice6Timestamp(-2 * 1e3)

	ack := new(messages.Ack)
	ack.Timestamp = msg.GetTimestamp()
	ack.LeaderChainID = msg.GetLeaderChainID()
	ack.MessageHash = msg.GetMsgHash()
	ack.SerialHash = primitives.RandomHash() //primitives.NewHash([]byte("serial"))
	ack.Timestamp.SetTimeNow()

	for _, processList := range s.ProcessLists.Lists {
		processList.AddToProcessList(s, ack, msg)
	}

	// assertions
	if assert.Equal(t, 1, len(eventQueue)) {
		event := <-eventQueue
		assert.Equal(t, eventmessages.EventSource_REPLAY_BOOT, event.GetEventSource())
		assert.Equal(t, s.IdentityChainID.Bytes(), event.IdentityChainID)

		stateChangeEvent := event.GetStateChange()
		if assert.NotNil(t, stateChangeEvent, "event received has wrong type: %s event: %+v", reflect.TypeOf(event.GetEvent()), event) {
			assert.Equal(t, eventmessages.EntityState_ACCEPTED, stateChangeEvent.GetEntityState())
			assert.NotNil(t, stateChangeEvent.GetEntityHash())
			assert.NotNil(t, stateChangeEvent.GetEntityState())
			assert.NotNil(t, stateChangeEvent.GetBlockHeight())
		}
	}
}

func TestEmitDirectoryBlockEventsFromHeightRange(t *testing.T) {
	eventQueue := make(chan *eventmessages.FactomEvent, p2p.StandardChannelSize)
	mockSender := &mockEventSender{
		eventsOutQueue:      eventQueue,
		replayDuringStartup: true,
		sendStateChange:     true,
	}

	s := testHelper.CreateAndPopulateTestState()
	s.EventService.ConfigSender(s, mockSender)
	s.RunLeader = true

	s.EmitDirectoryBlockEventsFromHeight(0, 7)

	if assert.Equal(t, 8, len(eventQueue)) {
		for i := 0; i < 8; i++ {
			event := <-eventQueue
			assert.Equal(t, eventmessages.EventSource_REPLAY_BOOT, event.GetEventSource())
			assert.Equal(t, s.IdentityChainID.Bytes(), event.IdentityChainID)

			directoryBlockCommit := event.GetDirectoryBlockCommit()
			if assert.NotNil(t, directoryBlockCommit, "event received has wrong type: %s event: %+v", reflect.TypeOf(event.GetEvent()), event) {
				assert.NotNil(t, directoryBlockCommit.GetAdminBlock())
				assert.NotNil(t, directoryBlockCommit.GetDirectoryBlock())
				assert.NotNil(t, directoryBlockCommit.GetEntryCreditBlock())
				assert.NotNil(t, directoryBlockCommit.GetFactoidBlock())
				assert.NotNil(t, directoryBlockCommit.GetEntryBlocks())
				assert.NotNil(t, directoryBlockCommit.GetEntryBlockEntries())
			}
		}
	}
}

func TestEmitDirectoryBlockStateEventsFromHeight(t *testing.T) {
	eventQueue := make(chan *eventmessages.FactomEvent, p2p.StandardChannelSize)
	mockSender := &mockEventSender{
		eventsOutQueue:      eventQueue,
		replayDuringStartup: true,
		sendStateChange:     true,
	}

	s := testHelper.CreateAndPopulateTestState()
	s.EventService.ConfigSender(s, mockSender)
	s.RunLeader = true

	s.EmitDirectoryBlockEventsFromHeight(3, 100)

	if assert.Equal(t, 7, len(eventQueue)) {
		for i := 0; i < 7; i++ {
			event := <-eventQueue
			assert.Equal(t, eventmessages.EventSource_REPLAY_BOOT, event.GetEventSource())
			assert.Equal(t, s.IdentityChainID.Bytes(), event.IdentityChainID)

			directoryBlockCommit := event.GetDirectoryBlockCommit()
			if assert.NotNil(t, directoryBlockCommit, "event received has wrong type: %s event: %+v", reflect.TypeOf(event.GetEvent()), event) {
				assert.NotNil(t, directoryBlockCommit.GetAdminBlock())
				assert.NotNil(t, directoryBlockCommit.GetDirectoryBlock())
				assert.NotNil(t, directoryBlockCommit.GetEntryCreditBlock())
				assert.NotNil(t, directoryBlockCommit.GetFactoidBlock())
				assert.NotNil(t, directoryBlockCommit.GetEntryBlocks())
				assert.NotNil(t, directoryBlockCommit.GetEntryBlockEntries())
			}
		}
	}
}

func TestEmitDirectoryBlockAnchorEvent(t *testing.T) {
	eventQueue := make(chan *eventmessages.FactomEvent, p2p.StandardChannelSize)
	mockSender := &mockEventSender{
		eventsOutQueue:      eventQueue,
		replayDuringStartup: true,
		sendStateChange:     true,
	}

	s := testHelper.CreateAndPopulateTestState()
	s.EventService.ConfigSender(s, mockSender)

	directoryBlockInfo := dbInfo.NewDirBlockInfo()

	db := databaseOverlay.NewOverlayWithState(s.DB.(*databaseOverlay.Overlay), s)
	db.ProcessDirBlockInfoMultiBatch(directoryBlockInfo)
	db.ProcessDirBlockInfoBatch(directoryBlockInfo)

	if assert.Equal(t, 2, len(eventQueue)) {
		for i := 0; i < 2; i++ {
			event := <-eventQueue
			assert.Equal(t, eventmessages.EventSource_REPLAY_BOOT, event.GetEventSource())
			assert.Equal(t, s.IdentityChainID.Bytes(), event.IdentityChainID)

			directoryBlockAnchorEvent := event.GetDirectoryBlockAnchor()
			if assert.NotNil(t, directoryBlockAnchorEvent, "event received has wrong type: %s event: %+v", reflect.TypeOf(event.GetEvent()), event) {
				assert.NotNil(t, directoryBlockAnchorEvent.GetDirectoryBlockHash())
			}
		}
	}
}

func TestExecuteMessage(t *testing.T) {
	testCases := map[string]struct {
		Message   interfaces.IMsg
		Assertion func(t *testing.T, event *eventmessages.FactomEvent)
	}{
		"chain-commit": {
			Message: &messages.CommitChainMsg{CommitChain: entryCreditBlock.NewCommitChain()},
			Assertion: func(t *testing.T, event *eventmessages.FactomEvent) {
				assert.NotNil(t, event.GetChainCommit(), "event received has wrong type: %s event: %+v", reflect.TypeOf(event.GetEvent()), event)
			},
		},
		"entry-commit": {
			Message: &messages.CommitEntryMsg{CommitEntry: entryCreditBlock.NewCommitEntry()},
			Assertion: func(t *testing.T, event *eventmessages.FactomEvent) {
				assert.NotNil(t, event.GetEntryCommit(), "event received has wrong type: %s event: %+v", reflect.TypeOf(event.GetEvent()), event)
			},
		},
		"entry-reveal": {
			Message: &messages.RevealEntryMsg{Entry: entryBlock.NewEntry(), CommitChain: &messages.CommitChainMsg{CommitChain: entryCreditBlock.NewCommitChain()}, Timestamp: primitives.NewTimestampNow()},
			Assertion: func(t *testing.T, event *eventmessages.FactomEvent) {
				assert.NotNil(t, event.GetEntryReveal(), "event received has wrong type: %s event: %+v", reflect.TypeOf(event.GetEvent()), event)
			},
		},
	}

	s := testHelper.CreateAndPopulateTestState()
	s.RunLeader = true // state is on the latest block making the event source LIVE instead of a REPLAY

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			// set mock service to receive events
			eventQueue := make(chan *eventmessages.FactomEvent, p2p.StandardChannelSize)
			mockSender := &mockEventSender{
				eventsOutQueue:      eventQueue,
				replayDuringStartup: true,
				sendStateChange:     true,
				broadcastContent:    eventconfig.BroadcastAlways,
			}
			s.EventService.ConfigSender(s, mockSender)

			// test
			testCase.Message.FollowerExecute(s)

			// assertions
			if assert.Equal(t, 1, len(eventQueue)) {
				event := <-eventQueue
				assert.Equal(t, eventmessages.EventSource_LIVE, event.GetEventSource())
				assert.Equal(t, s.IdentityChainID.Bytes(), event.IdentityChainID)

				testCase.Assertion(t, event)
			}
		})
	}
}

func TestEmitProcessListEvent(t *testing.T) {
	eventQueue := make(chan *eventmessages.FactomEvent, p2p.StandardChannelSize)
	mockSender := &mockEventSender{
		eventsOutQueue:      eventQueue,
		replayDuringStartup: true,
		sendStateChange:     true,
	}

	s := testHelper.CreateAndPopulateTestState()
	s.EventService.ConfigSender(s, mockSender)

	blockHeight := s.LLeaderHeight

	s.MoveStateToHeight(blockHeight, 1)

	currentMinute := s.CurrentMinute + 1
	s.MoveStateToHeight(s.LLeaderHeight, currentMinute)

	if assert.Equal(t, 2, len(eventQueue)) {
		event := <-eventQueue
		assert.Equal(t, eventmessages.EventSource_REPLAY_BOOT, event.GetEventSource())
		assert.Equal(t, s.IdentityChainID.Bytes(), event.IdentityChainID)

		processNewBlockEvent := event.GetProcessListEvent()
		if assert.NotNil(t, processNewBlockEvent, "event received has wrong type: %s event: %+v", reflect.TypeOf(event.GetEvent()), event) {
			if assert.NotNil(t, processNewBlockEvent.GetNewBlockEvent(), "%v", processNewBlockEvent) {
				assert.Equal(t, blockHeight, processNewBlockEvent.GetNewBlockEvent().GetNewBlockHeight())
			}
		}

		event = <-eventQueue
		assert.Equal(t, eventmessages.EventSource_REPLAY_BOOT, event.GetEventSource())
		assert.Equal(t, s.IdentityChainID.Bytes(), event.IdentityChainID)

		processNewMinuteEvent := event.GetProcessListEvent()
		if assert.NotNil(t, processNewMinuteEvent, "event received has wrong type: %s event: %+v", reflect.TypeOf(event.GetEvent()), event) {
			if assert.NotNil(t, processNewMinuteEvent.GetNewMinuteEvent(), "%v", processNewMinuteEvent) {
				assert.EqualValues(t, currentMinute, processNewMinuteEvent.GetNewMinuteEvent().GetNewMinute())
				assert.Equal(t, s.LLeaderHeight, processNewMinuteEvent.GetNewMinuteEvent().GetBlockHeight())
			}
		}
	}
}

type mockEventSender struct {
	eventsOutQueue      chan *eventmessages.FactomEvent
	broadcastContent    eventconfig.BroadcastContent
	sendStateChange     bool
	replayDuringStartup bool
}

func (m *mockEventSender) GetBroadcastContent() eventconfig.BroadcastContent {
	return m.broadcastContent
}
func (m *mockEventSender) IsSendStateChangeEvents() bool {
	return m.sendStateChange
}
func (m *mockEventSender) ReplayDuringStartup() bool {
	return m.replayDuringStartup
}
func (m *mockEventSender) GetEventQueue() chan *eventmessages.FactomEvent {
	return m.eventsOutQueue
}
func (m *mockEventSender) Shutdown() {}

func (m *mockEventSender) IncreaseDroppedFromQueueCounter() {}

func createByteSlice6Timestamp(offset int64) *primitives.ByteSlice6 {
	buf := new(bytes.Buffer)
	t := time.Now().UnixNano()
	m := t/1e6 + offset
	binary.Write(buf, binary.BigEndian, m)

	var b6 primitives.ByteSlice6
	copy(b6[:], buf.Bytes()[2:])

	return &b6
}
