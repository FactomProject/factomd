package events

import (
	"testing"

	"github.com/FactomProject/factomd/common/constants/runstate"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/events/eventconfig"
	"github.com/FactomProject/factomd/events/eventinput"
	"github.com/FactomProject/factomd/events/eventmessages/generated/eventmessages"
	"github.com/FactomProject/factomd/p2p"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

func TestEventEmitter_Send(t *testing.T) {
	testCases := map[string]struct {
		Emitter   *eventEmitter
		Event     eventinput.EventInput
		Assertion func(*testing.T, *mockEventSender, error)
	}{
		"queue-filled": {
			Emitter: &eventEmitter{
				parentState: StateMock{
					IdentityChainID: primitives.NewZeroHash(),
				},
				eventSender: &mockEventSender{
					eventsOutQueue:          make(chan *eventmessages.FactomEvent, 0),
					droppedFromQueueCounter: prometheus.NewCounter(prometheus.CounterOpts{}),
				},
			},
			Event: eventinput.NodeInfoMessageF(eventmessages.NodeMessageCode_GENERAL, "test message of node: %s", "node name"),
			Assertion: func(t *testing.T, eventService *mockEventSender, err error) {
				assert.NoError(t, err)
				assert.Equal(t, 0, len(eventService.eventsOutQueue))
				assert.Equal(t, float64(1), getCounterValue(t, eventService.droppedFromQueueCounter))
			},
		},
		"not-running": {
			Emitter: &eventEmitter{
				eventSender: &mockEventSender{
					eventsOutQueue: make(chan *eventmessages.FactomEvent, p2p.StandardChannelSize),
				},
				parentState: StateMock{
					RunState: runstate.Stopping,
				},
			},
			Event: nil,
			Assertion: func(t *testing.T, eventService *mockEventSender, err error) {
				assert.NoError(t, err)
				assert.Equal(t, 0, len(eventService.eventsOutQueue))
			},
		},
		"nil-event": {
			Emitter: &eventEmitter{
				eventSender: &mockEventSender{
					eventsOutQueue:      make(chan *eventmessages.FactomEvent, p2p.StandardChannelSize),
					replayDuringStartup: true,
				},
				parentState: StateMock{},
			},
			Event: nil,
			Assertion: func(t *testing.T, eventService *mockEventSender, err error) {
				assert.Error(t, err)
				assert.Equal(t, 0, len(eventService.eventsOutQueue))
			},
		},
		"mute-replay-starting": {
			Emitter: &eventEmitter{
				eventSender: &mockEventSender{
					eventsOutQueue:      make(chan *eventmessages.FactomEvent, p2p.StandardChannelSize),
					replayDuringStartup: false,
				},
				parentState: StateMock{
					RunLeader: false,
				},
			},
			Event: eventinput.NewStateChangeEvent(eventmessages.EventSource_REPLAY_BOOT, eventmessages.EntityState_REJECTED, nil),
			Assertion: func(t *testing.T, eventService *mockEventSender, err error) {
				assert.NoError(t, err)
				assert.Equal(t, 0, len(eventService.eventsOutQueue))
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			err := testCase.Emitter.Send(testCase.Event)
			testCase.Assertion(t, testCase.Emitter.eventSender.(*mockEventSender), err)
		})
	}
}

func TestEventsService_SendFillupQueue(t *testing.T) {
	n := 3

	eventSender := &mockEventSender{
		eventsOutQueue:          make(chan *eventmessages.FactomEvent, n),
		droppedFromQueueCounter: prometheus.NewCounter(prometheus.CounterOpts{}),
	}
	eventEmitter := &eventEmitter{
		parentState: StateMock{
			IdentityChainID: primitives.NewZeroHash(),
		},
		eventSender: eventSender,
	}

	event := eventinput.NodeInfoMessageF(eventmessages.NodeMessageCode_GENERAL, "test message of node: %s", "node name")
	for i := 0; i < n+1; i++ {
		err := eventEmitter.Send(event)
		assert.Nil(t, err)
	}

	assert.Equal(t, float64(1), getCounterValue(t, eventSender.droppedFromQueueCounter))
}

func getCounterValue(t *testing.T, counter prometheus.Counter) float64 {
	metric := &dto.Metric{}
	err := counter.Write(metric)
	if err != nil {
		assert.Fail(t, "fail to retrieve prometheus counter: %v", err)
	}
	return *metric.Counter.Value
}

type mockEventSender struct {
	eventsOutQueue          chan *eventmessages.FactomEvent
	droppedFromQueueCounter prometheus.Counter
	notSentCounter          prometheus.Counter
	replayDuringStartup     bool
}

func (m *mockEventSender) GetBroadcastContent() eventconfig.BroadcastContent {
	return eventconfig.BroadcastAlways
}
func (m *mockEventSender) IsSendStateChangeEvents() bool {
	return true
}
func (m *mockEventSender) ReplayDuringStartup() bool {
	return m.replayDuringStartup
}
func (m *mockEventSender) IncreaseDroppedFromQueueCounter() {
	m.droppedFromQueueCounter.Inc()
}
func (m *mockEventSender) GetEventQueue() chan *eventmessages.FactomEvent {
	return m.eventsOutQueue
}

func (m *mockEventSender) Shutdown() {}

type StateMock struct {
	IdentityChainID interfaces.IHash
	RunState        runstate.RunState
	RunLeader       bool
	Service         EventService
}

func (s StateMock) GetRunState() runstate.RunState {
	return s.RunState
}

func (s StateMock) GetIdentityChainID() interfaces.IHash {
	return s.IdentityChainID
}

func (s StateMock) IsRunLeader() bool {
	return s.RunLeader
}

func (s StateMock) GetEventService() EventService {
	return s.Service
}
