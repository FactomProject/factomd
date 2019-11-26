package eventservices

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/FactomProject/factomd/common/constants/runstate"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/events"
	"github.com/FactomProject/factomd/events/eventconfig"
	"github.com/FactomProject/factomd/events/eventinput"
	"github.com/FactomProject/factomd/events/eventmessages/generated/eventmessages"
	"github.com/FactomProject/factomd/p2p"
	"github.com/FactomProject/factomd/util/atomic"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

func init() {
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)
}

func TestEventsService_Send(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	var finished atomic.AtomicBool
	finished.Store(false)

	expectedMessage := `{"identityChainID":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","Event":{"nodeMessage":{"messageText":"test message of node: node name"}}}`

	// mock server by reading everything until stop byte is found
	// use the stop byte to stop as soon as possible, note: don't use stop byte in test message before the end
	receivingMessage := bytes.NewBufferString("")
	go func() {
		reader := bufio.NewReader(server)
		for {
			b, err := reader.ReadByte()
			if err == io.ErrClosedPipe || err == io.EOF {
				// test finished, test stopped and probably closed the connection
				break
			}
			if err != nil {
				fmt.Println(err)
			}
			receivingMessage.WriteByte(b)

			// stop if message received
			if strings.HasSuffix(receivingMessage.String(), expectedMessage) {
				break
			}
		}
		finished.Store(true)
	}()

	state := &StateMock{
		IdentityChainID: primitives.NewZeroHash(),
		RunState:        runstate.Running,
		Events:          events.NewEventEmitter(),
	}

	params := &EventServiceParams{
		OutputFormat: eventconfig.Json,
	}
	NewEventServiceTo(state, params)

	// set connection

	serviceInstance.connection = client

	// create test factom event
	event := eventinput.NodeInfoMessageF(eventmessages.NodeMessageCode_GENERAL, "test message of node: %s", "node name")

	// test send
	err := serviceInstance.Send(event)

	assert.NoError(t, err)

	// wait max 1 second until the server has read the bytes
	for i := 0; !finished.Load() && i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
	}

	assert.True(t, finished.Load())

	discardReceivedMetadata(receivingMessage)
	assert.JSONEq(t, expectedMessage, receivingMessage.String(), "%s != %s", expectedMessage, receivingMessage.String())
	assert.Equal(t, float64(0), getCounterValue(t, serviceInstance.notSentCounter))
}

func TestEventsService_SendFillupQueue(t *testing.T) {
	n := 3
	eventService := &eventServiceInstance{
		eventsOutQueue: make(chan *eventmessages.FactomEvent, n),
		parentState: StateMock{
			IdentityChainID: primitives.NewZeroHash(),
		},
		params:                  &EventServiceParams{},
		droppedFromQueueCounter: prometheus.NewCounter(prometheus.CounterOpts{}),
	}

	event := eventinput.NodeInfoMessageF(eventmessages.NodeMessageCode_GENERAL, "test message of node: %s", "node name")
	for i := 0; i < n+1; i++ {
		eventService.Send(event)
	}

	assert.Equal(t, float64(1), getCounterValue(t, eventService.droppedFromQueueCounter))
}

func TestEventsService_SendNoStartupMessages(t *testing.T) {
	testCases := map[string]struct {
		Service   *eventServiceInstance
		Event     eventinput.EventInput
		Assertion func(*testing.T, *eventServiceInstance, error)
	}{
		"queue-filled": {
			Service: &eventServiceInstance{
				eventsOutQueue: make(chan *eventmessages.FactomEvent, 0),
				parentState: StateMock{
					IdentityChainID: primitives.NewZeroHash(),
				},
				params:                  &EventServiceParams{},
				droppedFromQueueCounter: prometheus.NewCounter(prometheus.CounterOpts{}),
			},
			Event: eventinput.NodeInfoMessageF(eventmessages.NodeMessageCode_GENERAL, "test message of node: %s", "node name"),
			Assertion: func(t *testing.T, eventService *eventServiceInstance, err error) {
				assert.NoError(t, err)
				assert.Equal(t, 0, len(eventService.eventsOutQueue))
				assert.Equal(t, float64(1), getCounterValue(t, eventService.droppedFromQueueCounter))
			},
		},
		"not-running": {
			Service: &eventServiceInstance{
				eventsOutQueue: make(chan *eventmessages.FactomEvent, p2p.StandardChannelSize),
				parentState: StateMock{
					RunState: runstate.Stopping,
				},
			},
			Event: nil,
			Assertion: func(t *testing.T, eventService *eventServiceInstance, err error) {
				assert.NoError(t, err)
				assert.Equal(t, 0, len(eventService.eventsOutQueue))
			},
		},
		"nil-event": {
			Service: &eventServiceInstance{
				eventsOutQueue: make(chan *eventmessages.FactomEvent, p2p.StandardChannelSize),
				parentState:    StateMock{},
				params: &EventServiceParams{
					ReplayDuringStartup: true,
				},
			},
			Event: nil,
			Assertion: func(t *testing.T, eventService *eventServiceInstance, err error) {
				assert.Error(t, err)
				assert.Equal(t, 0, len(eventService.eventsOutQueue))
			},
		},
		"mute-replay-starting": {
			Service: &eventServiceInstance{
				eventsOutQueue: make(chan *eventmessages.FactomEvent, p2p.StandardChannelSize),
				parentState: StateMock{
					RunLeader: false,
				},
				params: &EventServiceParams{
					ReplayDuringStartup: false,
				},
			},
			Event: eventinput.NewStateChangeEvent(eventmessages.EventSource_REPLAY_BOOT, eventmessages.EntityState_REJECTED, nil),
			Assertion: func(t *testing.T, eventService *eventServiceInstance, err error) {
				assert.NoError(t, err)
				assert.Equal(t, 0, len(eventService.eventsOutQueue))
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			err := testCase.Service.Send(testCase.Event)
			testCase.Assertion(t, testCase.Service, err)
		})
	}
}

func TestEventService_ProcessEventsChannelNoSent(t *testing.T) {
	redialSleepDuration = 1 * time.Millisecond
	sendRetries = 1

	eventQueue := make(chan *eventmessages.FactomEvent, p2p.StandardChannelSize)
	eventService := &eventServiceInstance{
		eventsOutQueue: eventQueue,
		params: &EventServiceParams{
			OutputFormat: eventconfig.Json,
		},
		notSentCounter: prometheus.NewCounter(prometheus.CounterOpts{}),
	}

	factomEvent := &eventmessages.FactomEvent{
		EventSource:    eventmessages.EventSource_REPLAY_BOOT,
		FactomNodeName: "test",
	}

	n := 3
	for i := 0; i < n; i++ {
		eventQueue <- factomEvent
	}

	go eventService.processEventsChannel()

	// wait reasonable time until the prccess queue is empty
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 0, len(eventQueue))
	assert.Equal(t, float64(n), getCounterValue(t, eventService.notSentCounter))
}

func TestEventsService_SendEvent(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	eventService := &eventServiceInstance{
		params: &EventServiceParams{
			OutputFormat: eventconfig.Json,
		},
		connection:     client,
		notSentCounter: prometheus.NewCounter(prometheus.CounterOpts{}),
	}

	expectedMessage := `{"eventSource":1,"factomNodeName":"test","Event":null}`

	// mock server by reading everything until stop byte is found
	// use the stop byte to stop as soon as possible, note: don't use stop byte in test message before the end
	receivingMessage := bytes.NewBufferString("")
	var finished atomic.AtomicBool
	finished.Store(false)
	go func() {
		reader := bufio.NewReader(server)
		for {
			b, err := reader.ReadByte()
			if err == io.ErrClosedPipe || err == io.EOF {
				// test finished, test stopped and probably closed the connection
				break
			}
			if err != nil {
				fmt.Println(err)
			}
			receivingMessage.WriteByte(b)

			// stop if message received
			if strings.HasSuffix(receivingMessage.String(), expectedMessage) {
				break
			}
		}
		finished.Store(true)
	}()

	factomEvent := &eventmessages.FactomEvent{
		EventSource:    eventmessages.EventSource_REPLAY_BOOT,
		FactomNodeName: "test",
	}

	// test send event
	eventService.sendEvent(factomEvent)

	// wait max 1 second until the server has read the bytes
	for i := 0; !finished.Load() && i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
	}

	discardReceivedMetadata(receivingMessage)

	assert.True(t, finished.Load())
	assert.JSONEq(t, expectedMessage, receivingMessage.String(), "%s != %s", expectedMessage, receivingMessage.String())
	assert.Equal(t, float64(0), getCounterValue(t, eventService.notSentCounter))
}

func TestEventsService_SendEventBreakdown(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	redialSleepDuration = 10 * time.Millisecond
	sendRetries = 3

	eventService := &eventServiceInstance{
		params: &EventServiceParams{
			OutputFormat: eventconfig.Json,
		},
		connection:     client,
		notSentCounter: prometheus.NewCounter(prometheus.CounterOpts{}),
	}

	expectedMessage := `{"eventSource":1,"factomNodeName":"test","Event":null}`

	// mock server by reading everything until stop byte is found
	// use the stop byte to stop as soon as possible, note: don't use stop byte in test message before the end
	receivingMessage := bytes.NewBufferString("")
	var finished atomic.AtomicBool
	finished.Store(false)
	go func() {
		reader := bufio.NewReader(server)
		for i := 0; i < 1; i++ {
			b, err := reader.ReadByte()
			if err == io.ErrClosedPipe || err == io.EOF {
				// test finished, test stopped and probably closed the connection
				break
			}
			if err != nil {
				fmt.Println(err)
				continue
			}
			receivingMessage.WriteByte(b)

			// stop if message received
			if strings.HasSuffix(receivingMessage.String(), expectedMessage) {
				break
			}
		}

		// this breaks the connection
		_ = server.Close()
		receivingMessage = bytes.NewBufferString("")

		// wait until the other thread handles the broken connection
		time.Sleep(1 * time.Millisecond)

		// re-establish a connection, by setting the new connection
		server, client = net.Pipe()
		eventService.connection = client

		// continue reading the messages
		reader = bufio.NewReader(server)
		for {
			b, err := reader.ReadByte()
			if err == io.ErrClosedPipe { //|| err == io.EOF {
				fmt.Println(err)
				continue
			}
			if err != nil {
				fmt.Println(err)
				time.Sleep(10 * time.Millisecond)
				continue
			}
			receivingMessage.WriteByte(b)

			// stop if message received
			if strings.HasSuffix(receivingMessage.String(), expectedMessage) {
				break
			}
		}
		finished.Store(true)
	}()

	factomEvent := &eventmessages.FactomEvent{
		EventSource:    eventmessages.EventSource_REPLAY_BOOT,
		FactomNodeName: "test",
	}

	// test send event
	eventService.sendEvent(factomEvent)

	// wait max 5 second until the server has read the bytes
	for i := 0; !finished.Load() && i < 50; i++ {
		time.Sleep(100 * time.Millisecond)
	}

	discardReceivedMetadata(receivingMessage)

	assert.True(t, finished.Load())
	assert.JSONEq(t, expectedMessage, receivingMessage.String(), "%s != %s", expectedMessage, receivingMessage.String())
	assert.Equal(t, float64(0), getCounterValue(t, eventService.notSentCounter))
}

func TestEventsService_SendEventMarshallingError(t *testing.T) {
	eventService := &eventServiceInstance{
		params:         &EventServiceParams{},
		notSentCounter: prometheus.NewCounter(prometheus.CounterOpts{}),
	}
	factomEvent := &eventmessages.FactomEvent{}
	eventService.sendEvent(factomEvent)

	assert.Equal(t, float64(1), getCounterValue(t, eventService.notSentCounter))
}

func TestEventsService_SendEventNoConnection(t *testing.T) {
	redialSleepDuration = 1 * time.Millisecond
	eventService := &eventServiceInstance{
		params: &EventServiceParams{
			OutputFormat: eventconfig.Json,
		},
		notSentCounter: prometheus.NewCounter(prometheus.CounterOpts{}),
	}
	factomEvent := &eventmessages.FactomEvent{}
	eventService.sendEvent(factomEvent)

	assert.Equal(t, float64(1), getCounterValue(t, eventService.notSentCounter))
}

func TestEventService_ConnectAndShutdown(t *testing.T) {
	address := ":1444"
	listener, err := net.Listen("tcp", address)

	if err != nil {
		t.Fatalf("setup test failed: %v", err)
	}
	var finished atomic.AtomicBool
	finished.Store(false)

	go func() {
		conn, _ := listener.Accept()
		reader := bufio.NewReader(conn)
		for {
			b, err := reader.ReadByte()
			// wait until there is an EOF, which means the connection is closed remotely.
			if err == io.EOF {
				finished.Store(true)
				break
			}
			fmt.Print(b)
		}
	}()

	eventService := &eventServiceInstance{
		eventsOutQueue: make(chan *eventmessages.FactomEvent, p2p.StandardChannelSize),
		params: &EventServiceParams{
			Protocol: "tcp",
			Address:  address,
		},
	}

	err = eventService.connect()

	assert.NoError(t, err)

	eventService.Shutdown()

	// wait max 1 second until disconnect
	for i := 0; !finished.Load() && i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
	}
	assert.True(t, finished.Load())
}

func TestEventsService_ConnectNoConnection(t *testing.T) {
	address := ":1445"
	eventService := &eventServiceInstance{
		params: &EventServiceParams{
			Protocol: "tcp",
			Address:  address,
		},
	}
	err := eventService.connect()

	assert.EqualError(t, err, fmt.Sprintf("failed to connect: dial tcp %s: connect: connection refused", address))
}

func TestEventsService_MarshallMessage(t *testing.T) {
	testCases := map[string]struct {
		Event        *eventmessages.FactomEvent
		OutputFormat eventconfig.EventFormat
		Assertion    func(*testing.T, []byte, error)
	}{
		"protobuf": {
			Event: &eventmessages.FactomEvent{
				EventSource:    eventmessages.EventSource_REPLAY_BOOT,
				FactomNodeName: "test",
			},
			OutputFormat: eventconfig.Protobuf,
			Assertion: func(t *testing.T, data []byte, err error) {
				assert.NoError(t, err)
				// the first byte is the indication of the eventSource following the eventSource
				assert.Equal(t, byte(eventmessages.EventSource_REPLAY_BOOT), data[1])
				// the third byte is the indication of the factomNodeName following the name
				assert.Equal(t, []byte("test"), data[4:])
			},
		},
		"json": {
			Event: &eventmessages.FactomEvent{
				EventSource:    eventmessages.EventSource_REPLAY_BOOT,
				FactomNodeName: "test",
			},
			OutputFormat: eventconfig.Json,
			Assertion: func(t *testing.T, data []byte, err error) {
				assert.NoError(t, err)
				assert.JSONEq(t, `{"eventSource":1,"factomNodeName":"test","Event":null}`, string(data))
			},
		},
		"empty-protobuf": {
			Event:        &eventmessages.FactomEvent{},
			OutputFormat: eventconfig.Protobuf,
			Assertion: func(t *testing.T, data []byte, err error) {
				assert.NoError(t, err)
				assert.Equal(t, []byte{}, data)
			},
		},
		"empty-json": {
			Event:        &eventmessages.FactomEvent{},
			OutputFormat: eventconfig.Json,
			Assertion: func(t *testing.T, data []byte, err error) {
				assert.NoError(t, err)
				assert.JSONEq(t, `{"Event": null}`, string(data))
			},
		},
		"EventFormat-issue": {
			Event:        &eventmessages.FactomEvent{},
			OutputFormat: 3,
			Assertion: func(t *testing.T, data []byte, err error) {
				assert.EqualError(t, err, "unsupported event format: unknown format 3")
				assert.Nil(t, data)
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			eventService := &eventServiceInstance{
				params: &EventServiceParams{
					OutputFormat: testCase.OutputFormat,
				},
			}
			data, err := eventService.marshallMessage(testCase.Event)

			testCase.Assertion(t, data, err)
		})
	}

}

func TestEventsService_MarshallEvent(t *testing.T) {
	factomEvent := &eventmessages.FactomEvent{
		EventSource:    eventmessages.EventSource_REPLAY_BOOT,
		FactomNodeName: "test",
	}

	eventService := &eventServiceInstance{}
	data, err := eventService.marshallEvent(factomEvent)

	assert.NoError(t, err)
	// the first byte is the indication of the eventSource following the eventSource
	assert.Equal(t, byte(eventmessages.EventSource_REPLAY_BOOT), data[1])
	// the third byte is the indication of the factomNodeName following the name
	assert.Equal(t, []byte("test"), data[4:])
}

func TestEventsService_WriteEvent(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	eventService := &eventServiceInstance{
		connection: client,
	}

	stopByte := byte(0xFF)
	testMessage := []byte("tests")
	testMessage = append(testMessage, stopByte)

	// mock server by reading everything until stop byte is found
	// use the stop byte to stop as soon as possible, note: don't use stop byte in test message before the end
	receivingMessage := bytes.NewBufferString("")
	var finished atomic.AtomicBool
	finished.Store(false)
	go func() {
		reader := bufio.NewReader(server)
		for {
			b, err := reader.ReadByte()
			if err == io.ErrClosedPipe {
				// test finished
				break
			}
			if err != nil {
				fmt.Println(err)
				continue
			}
			receivingMessage.WriteByte(b)
			if b == stopByte {
				break
			}
		}
		finished.Store(true)
		_ = server.Close()
	}()
	// test write event
	err := eventService.writeEvent(testMessage)

	// wait max 1 second until the server has read the bytes
	for i := 0; !finished.Load() && i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
	}

	discardReceivedMetadata(receivingMessage)
	assert.NoError(t, err)
	assert.Equalf(t, testMessage, receivingMessage.Bytes(), "expected: %s\n actual: %s", testMessage, receivingMessage.String())
}

func TestEventService_GetBroadcastContent(t *testing.T) {
	eventService := &eventServiceInstance{
		params: &EventServiceParams{
			BroadcastContent: eventconfig.BroadcastNever,
		},
	}
	broadcastContent := eventService.GetBroadcastContent()
	assert.Equal(t, eventconfig.BroadcastNever, broadcastContent)
}

func TestEventService_IsSendStateChangeEvents(t *testing.T) {
	eventService := &eventServiceInstance{
		params: &EventServiceParams{
			SendStateChangeEvents: true,
		},
	}
	sendStateChangeEvents := eventService.IsSendStateChangeEvents()
	assert.True(t, sendStateChangeEvents)
}

func BenchmarkEventService_Send(b *testing.B) {
	listener, _ := net.Listen("tcp", ":2135")
	client, _ := net.Dial("tcp", "127.0.0.1:2135")

	event := eventinput.NodeInfoMessageF(eventmessages.NodeMessageCode_GENERAL, "benchmark message of node: %s", "node name")

	state := &StateMock{
		IdentityChainID: primitives.NewZeroHash(),
		RunState:        runstate.Running,
	}

	params := &EventServiceParams{
		OutputFormat:          eventconfig.Protobuf,
		SendStateChangeEvents: false,
		BroadcastContent:      eventconfig.BroadcastAlways,
	}

	eventService := &eventServiceInstance{
		eventsOutQueue:          make(chan *eventmessages.FactomEvent, p2p.StandardChannelSize),
		connection:              client,
		params:                  params,
		parentState:             state,
		droppedFromQueueCounter: prometheus.NewCounter(prometheus.CounterOpts{}),
		notSentCounter:          prometheus.NewCounter(prometheus.CounterOpts{}),
	}

	i := atomic.AtomicInt(0)
	go func() {
		server, _ := listener.Accept()
		reader := bufio.NewReader(server)
		for {
			// consume every byte that is send
			_, err := reader.ReadByte()
			if err != nil {
				fmt.Println(err)
				continue
			}
			i++
		}
	}()
	time.Sleep(10 * time.Millisecond)
	go eventService.processEventsChannel()

	for n := 0; n < b.N; n++ {
		eventService.Send(event)
	}

	b.Logf("bytes received: %d", i.Load())
}

func discardReceivedMetadata(receivingMessage *bytes.Buffer) {
	// read 1 byte of protocol version
	_, _ = receivingMessage.ReadByte()
	// read 4 bytes of data length: int32
	_, _ = receivingMessage.ReadByte()
	_, _ = receivingMessage.ReadByte()
	_, _ = receivingMessage.ReadByte()
	_, _ = receivingMessage.ReadByte()
}

func getCounterValue(t *testing.T, counter prometheus.Counter) float64 {
	metric := &dto.Metric{}
	err := counter.Write(metric)
	if err != nil {
		assert.Fail(t, "fail to retrieve prometheus counter: %v", err)
	}
	return *metric.Counter.Value
}

type StateMock struct {
	IdentityChainID interfaces.IHash
	RunState        runstate.RunState
	RunLeader       bool
	Events          events.Events
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

func (s StateMock) GetEvents() events.Events {
	return s.Events
}
