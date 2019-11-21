package events_test

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/events"
	"github.com/FactomProject/factomd/events/eventmessages/generated/eventmessages"
	"github.com/FactomProject/factomd/testHelper"
	"github.com/stretchr/testify/assert"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

func TestUpdateState(t *testing.T) {
	mockService := &mockEventService{}

	s := testHelper.CreateAndPopulateTestState()
	s.EventsService = mockService

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
	if assert.Equal(t, int32(1), mockService.EventsReceived) {
		event := mockService.Events[0]
		assert.Equal(t, eventmessages.EventSource_REPLAY_BOOT, event.GetStreamSource())

		if stateChangeEvent, ok := event.(*events.StateChangeEvent); assert.True(t, ok, "event received has wrong type: %s event: %+v", reflect.TypeOf(event), event) {
			assert.NotNil(t, stateChangeEvent)
			assert.NotNil(t, stateChangeEvent.GetPayload())
			assert.Equal(t, eventmessages.EntityState_COMMITTED_TO_DIRECTORY_BLOCK, stateChangeEvent.GetEntityState())
		}
	}
}

func TestAddToAndDeleteFromHolding(t *testing.T) {
	mockService := &mockEventService{}

	s := testHelper.CreateAndPopulateTestState()
	s.EventsService = mockService

	msg := &messages.CommitChainMsg{CommitChain: entryCreditBlock.NewCommitChain()}
	msg.CommitChain.MilliTime = createByteSlice6Timestamp(-2 * 1e3)

	s.AddToHolding(msg.GetMsgHash().Fixed(), msg)
	s.DeleteFromHolding(msg.GetMsgHash().Fixed(), msg, "to unit test")

	// assertions
	if assert.Equal(t, int32(2), mockService.EventsReceived) {
		addToHoldingEvent := mockService.Events[0]
		assert.Equal(t, eventmessages.EventSource_REPLAY_BOOT, addToHoldingEvent.GetStreamSource())

		if registrationEvent, ok := addToHoldingEvent.(*events.RegistrationEvent); assert.True(t, ok, "event received has wrong type: %s event: %+v", reflect.TypeOf(addToHoldingEvent), addToHoldingEvent) {
			assert.NotNil(t, registrationEvent)
			assert.EqualValues(t, msg, registrationEvent.GetPayload())
		}

		deleteFromHoldingEvent := mockService.Events[1]
		assert.Equal(t, eventmessages.EventSource_REPLAY_BOOT, deleteFromHoldingEvent.GetStreamSource())

		if stateChangeEvent, ok := deleteFromHoldingEvent.(*events.StateChangeMsgEvent); assert.True(t, ok, "event received has wrong type: %s event: %+v", reflect.TypeOf(deleteFromHoldingEvent), deleteFromHoldingEvent) {
			assert.NotNil(t, stateChangeEvent)
			assert.Equal(t, eventmessages.EntityState_REJECTED, stateChangeEvent.GetEntityState())
			assert.EqualValues(t, msg, stateChangeEvent.GetPayload())
		}
	}
}

func TestAddToProcessList(t *testing.T) {
	mockService := &mockEventService{}

	s := testHelper.CreateAndPopulateTestState()
	s.SetLeaderTimestamp(primitives.NewTimestampNow())
	s.EventsService = mockService

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
	if assert.Equal(t, int32(1), mockService.EventsReceived) {
		event := mockService.Events[0]
		assert.Equal(t, eventmessages.EventSource_REPLAY_BOOT, event.GetStreamSource())

		if stateChangeEvent, ok := event.(*events.StateChangeMsgEvent); assert.True(t, ok, "event received has wrong type: %s event: %+v", reflect.TypeOf(event), event) {
			assert.NotNil(t, stateChangeEvent)
			assert.Equal(t, eventmessages.EntityState_ACCEPTED, stateChangeEvent.GetEntityState())
			assert.EqualValues(t, msg, stateChangeEvent.GetPayload())
		}
	}
}

func TestEmitDBStateEventsFromHeightRange(t *testing.T) {
	s := testHelper.CreateAndPopulateTestState()

	mockService := &mockEventService{}
	s.EventsService = mockService

	s.EmitDBStateEventsFromHeight(0, 7)

	if assert.Equal(t, int32(8), mockService.EventsReceived) {
		for _, event := range mockService.Events {
			assert.Equal(t, eventmessages.EventSource_REPLAY_BOOT, event.GetStreamSource())

			if stateChangeEvent, ok := event.(*events.StateChangeMsgEvent); assert.True(t, ok, "event received has wrong type: %s event: %+v", reflect.TypeOf(event), event) {
				assert.NotNil(t, stateChangeEvent)
				assert.Equal(t, eventmessages.EntityState_COMMITTED_TO_DIRECTORY_BLOCK, stateChangeEvent.GetEntityState())
				assert.NotNil(t, stateChangeEvent.GetPayload())
			}
		}
	}
}

func TestEmitDBStateEventsFromHeight(t *testing.T) {
	s := testHelper.CreateAndPopulateTestState()

	mockService := &mockEventService{}
	s.EventsService = mockService

	s.EmitDBStateEventsFromHeight(3, 100)

	if assert.Equal(t, int32(7), mockService.EventsReceived) {
		for _, event := range mockService.Events {
			//assert.Equal(t, eventmessages.EventSource_REPLAY_BOOT, event.GetStreamSource())

			if stateChangeEvent, ok := event.(*events.StateChangeMsgEvent); assert.True(t, ok, "event received has wrong type: %s event: %+v", reflect.TypeOf(event), event) {
				assert.NotNil(t, stateChangeEvent)
				assert.Equal(t, eventmessages.EntityState_COMMITTED_TO_DIRECTORY_BLOCK, stateChangeEvent.GetEntityState())
				assert.NotNil(t, stateChangeEvent.GetPayload())
			}
		}
	}
}

func TestExecuteMessage(t *testing.T) {
	testCases := map[string]struct {
		Message interfaces.IMsg
	}{
		"chain-commit": {&messages.CommitChainMsg{CommitChain: entryCreditBlock.NewCommitChain()}},
		"entry-commit": {&messages.CommitEntryMsg{CommitEntry: entryCreditBlock.NewCommitEntry()}},
		"entry-reveal": {&messages.RevealEntryMsg{Entry: entryBlock.NewEntry(), CommitChain: &messages.CommitChainMsg{CommitChain: entryCreditBlock.NewCommitChain()}}},
	}

	s := testHelper.CreateEmptyTestState()
	s.RunLeader = true // state is on the latest block making the event source LIVE instead of a REPLAY

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			// set mock service to receive events
			mockService := &mockEventService{}
			s.EventsService = mockService

			// test
			testCase.Message.FollowerExecute(s)

			// assertions
			if assert.Equal(t, int32(1), mockService.EventsReceived) {
				event := mockService.Events[0]
				assert.Equal(t, eventmessages.EventSource_LIVE, event.GetStreamSource())

				if registrationEvent, ok := event.(*events.RegistrationEvent); assert.True(t, ok, "event received has wrong type: %s event: %+v", reflect.TypeOf(event), event) {
					assert.NotNil(t, registrationEvent)
					assert.EqualValues(t, testCase.Message, registrationEvent.GetPayload())
				}
			}
		})
	}
}

type mockEventService struct {
	t              *testing.T
	Events         []events.EventInput
	EventsReceived int32
}

func (m *mockEventService) Send(event events.EventInput) error {
	m.Events = append(m.Events, event)
	atomic.AddInt32(&m.EventsReceived, 1)

	var err error
	if m.t != nil {
		var data []byte
		switch event.(type) {
		case *events.RegistrationEvent:
			data, err = json.Marshal(event.(*events.RegistrationEvent).GetPayload())
		case *events.StateChangeEvent:
			data, err = json.Marshal(event.(*events.StateChangeEvent).GetPayload())
		case *events.StateChangeMsgEvent:
			data, err = json.Marshal(event.(*events.StateChangeMsgEvent).GetPayload())
		case *events.ProcessListEvent:
			data, err = json.Marshal(event.(*events.ProcessListEvent).GetProcessListEvent())
		case *events.NodeMessageEvent:
			data, err = json.Marshal(event.(*events.NodeMessageEvent).GetNodeMessage())
		}

		m.t.Logf("incomming event: " + string(data))
	}
	return err
}

func createByteSlice6Timestamp(offset int64) *primitives.ByteSlice6 {
	buf := new(bytes.Buffer)
	t := time.Now().UnixNano()
	m := t/1e6 + offset
	binary.Write(buf, binary.BigEndian, m)

	var b6 primitives.ByteSlice6
	copy(b6[:], buf.Bytes()[2:])

	return &b6
}
