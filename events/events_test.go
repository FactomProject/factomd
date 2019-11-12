package events_test

import (
	"encoding/json"
	"fmt"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/events"
	"github.com/FactomProject/factomd/events/eventmessages/generated/eventmessages"
	"github.com/FactomProject/factomd/testHelper"
	"github.com/stretchr/testify/assert"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

func TestEmitRegistrationEvents(t *testing.T) {
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
			waitOnReceivedEvents(&mockService.EventsReceived, 1, 1*time.Minute)

			if assert.Equal(t, int32(1), mockService.EventsReceived) {
				fmt.Println(mockService.Events[0])

				event := mockService.Events[0]
				assert.Equal(t, eventmessages.EventSource_LIVE, event.GetStreamSource())

				if registrionEvent, ok := event.(*events.RegistrationEvent); assert.True(t, ok, "event received has wrong type: %s event: %+v", reflect.TypeOf(event), event) {
					assert.NotNil(t, registrionEvent)
					assert.EqualValues(t, testCase.Message, registrionEvent.GetPayload())
				}
			}
		})
	}
}

func waitOnReceivedEvents(eventsReceived *int32, maxEvents int, duration time.Duration) {
	deadline := time.Now().Add(duration)
	for int(atomic.LoadInt32(eventsReceived)) != maxEvents && time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)
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
