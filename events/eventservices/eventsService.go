package eventservices

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/FactomProject/factomd/common/constants/runstate"
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/events"
	"github.com/FactomProject/factomd/events/eventmessages/generated/eventmessages"
	"github.com/FactomProject/factomd/events/eventoutputformat"
	"github.com/FactomProject/factomd/p2p"
	"github.com/FactomProject/factomd/util"
	"github.com/gogo/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"net"
	"reflect"
	"time"
)

var eventService EventService
var eventServiceControl EventServiceControl

const (
	defaultProtocol       = "tcp"
	defaultConnectionHost = "127.0.0.1"
	defaultConnectionPort = 8040
	defaultOutputFormat   = eventoutputformat.Protobuf
	protocolVersion       = byte(1)
)

var (
	dialRetryPostponeDuration = 5 * time.Minute
	redialSleepDuration       = 10 * time.Second
	sendRetries               = 3
)

type eventServiceInstance struct {
	params                  *EventServiceParams
	eventsOutQueue          chan *eventmessages.FactomEvent
	postponeSendingUntil    time.Time
	connection              net.Conn
	owningState             interfaces.IState
	droppedFromQueueCounter prometheus.Counter
	notSentCounter          prometheus.Counter
}

func NewEventService(state interfaces.IState, config *util.FactomdConfig, factomParams *globals.FactomParams) (EventService, EventServiceControl) {
	return NewEventServiceTo(state, selectParameters(factomParams, config))
}

func NewEventServiceTo(state interfaces.IState, params *EventServiceParams) (EventService, EventServiceControl) {
	if eventService == nil {
		eventServiceInstance := &eventServiceInstance{
			eventsOutQueue: make(chan *eventmessages.FactomEvent, p2p.StandardChannelSize),
			params:         params,
			owningState:    state,
		}
		eventService = eventServiceInstance
		eventServiceControl = eventServiceInstance

		eventServiceInstance.droppedFromQueueCounter = prometheus.NewCounter(prometheus.CounterOpts{
			Name: "factomd_livefeed_dropped_from_queue_counter",
			Help: "Number of times we dropped events due of a full the event queue",
		})
		eventServiceInstance.notSentCounter = prometheus.NewCounter(prometheus.CounterOpts{
			Name: "factomd_livefeed_not_send_counter",
			Help: "Number of times we couldn't send out an event",
		})

		go eventServiceInstance.processEventsChannel()
	}
	return eventService, eventServiceControl
}

func (esi *eventServiceInstance) Send(event events.EventInput) error {
	if esi.owningState.GetRunState() > runstate.Running { // Stop queuing messages to the events channel when shutting down
		return nil
	}

	// Only send info messages when EventReplayDuringStartup is disabled
	if !esi.params.ReplayDuringStartup && !esi.owningState.IsRunLeader() {
		switch event.(type) {
		case *events.ProcessMessageEvent:
		case *events.NodeMessageEvent:
		default:
			return nil
		}
	}

	broadcastContent := eventServiceControl.GetBroadcastContent()
	sendStateChangeEvents := eventServiceControl.IsSendStateChangeEvents()
	factomEvent, err := MapToFactomEvent(event, broadcastContent, sendStateChangeEvents)
	if err != nil {
		return fmt.Errorf("failed to map to factom event: %v\n", err)
	}
	if factomEvent == nil {
		return nil
	}

	factomEvent.IdentityChainID = &eventmessages.Hash{
		HashValue: esi.owningState.GetIdentityChainID().Bytes(),
	}
	select {
	case esi.eventsOutQueue <- factomEvent:
	default:
		esi.droppedFromQueueCounter.Inc()
	}
	return nil
}

// TODO describe choice of dropping events.
func (esi *eventServiceInstance) processEventsChannel() {
	esi.connect()

	for event := range esi.eventsOutQueue {
		if esi.postponeSendingUntil.IsZero() || esi.postponeSendingUntil.Before(time.Now()) {
			esi.sendEvent(event)
		} else {
			esi.notSentCounter.Inc()
		}
	}
}

func (esi *eventServiceInstance) sendEvent(event *eventmessages.FactomEvent) {
	data, err := esi.marshallMessage(event)
	if err != nil {
		log.Errorf("An error occurred while serializing factom event of type %s: %v", reflect.TypeOf(event), err)
		esi.notSentCounter.Inc()
		return
	}

	// retry sending event ... times
	sendSuccessful := false
	for retry := 0; retry < sendRetries && !sendSuccessful; retry++ {
		if err = esi.connect(); err != nil {
			log.Errorf("An error occurred while connecting to receiver %s: %v, retry %d", esi.params.Address, err, retry)
			time.Sleep(redialSleepDuration)
			continue
		}

		// send the factom event to the live api
		if err = esi.writeEvent(data); err == nil {
			sendSuccessful = true
		} else {
			log.Errorf("An error occurred while sending a message to receiver %s: %v, retry %d", esi.params.Address, err, retry)

			// reset connection and retry
			esi.disconnect()
			esi.connection = nil
			time.Sleep(redialSleepDuration)
		}
	}

	if !sendSuccessful {
		esi.notSentCounter.Inc()
		esi.postponeSendingUntil = time.Now().Add(dialRetryPostponeDuration)
	}
}

func (esi *eventServiceInstance) marshallMessage(event *eventmessages.FactomEvent) ([]byte, error) {
	var data []byte
	var err error
	switch esi.params.OutputFormat {
	case eventoutputformat.Protobuf:
		data, err = esi.marshallEvent(event)
	case eventoutputformat.Json:
		data, err = json.Marshal(event)
	default:
		return nil, errors.New("unsupported event format: " + esi.params.OutputFormat.String())
	}
	return data, err
}

func (esi *eventServiceInstance) connect() error {
	defer catchConnectPanics()

	if esi.connection == nil {
		log.Infoln("Connecting to ", esi.params.Address)
		conn, err := net.Dial(esi.params.Protocol, esi.params.Address)
		if err != nil {
			return fmt.Errorf("failed to connect: %v", err)
		}
		esi.connection = conn
		esi.postponeSendingUntil = time.Time{}
	}
	return nil
}

func catchConnectPanics() error {
	if r := recover(); r != nil {
		return errors.New(fmt.Sprintf("failed to connect to receiver: %v", r))
	}
	return nil
}

func (esi *eventServiceInstance) disconnect() {
	if esi.connection != nil {
		log.Infoln("Closing connection to receiver", esi.params.Address)
		err := esi.connection.Close()
		if err != nil {
			log.Warnln("An error occurred while closing connection to receiver", esi.params.Address)
		}
	}
}

func (esi *eventServiceInstance) marshallEvent(event *eventmessages.FactomEvent) (data []byte, err error) {
	data, err = proto.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshell event: %v", err)
	}
	return data, err
}

func (esi *eventServiceInstance) writeEvent(data []byte) (err error) {
	defer catchSendPanics()

	writer := bufio.NewWriter(esi.connection)
	writer.WriteByte(protocolVersion)
	writer.Flush() // Flush this already to expedite a possible broken pipe which will only be detected in the second flush (unless there hasn't been any traffic for a few minutes)

	dataSize := int32(len(data))
	err = binary.Write(writer, binary.LittleEndian, dataSize)
	if err != nil {
		return fmt.Errorf("failed to write data size header: %v", err)
	}

	bytesWritten, err := writer.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data: %v. Bytes written: %d", err, bytesWritten)
	}
	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("failed to write data: %v", err)
	}
	return nil
}

func catchSendPanics() error {
	if r := recover(); r != nil {
		return errors.New(fmt.Sprintf("failed to write data: %v", r))
	}
	return nil
}

func (esi *eventServiceInstance) GetBroadcastContent() BroadcastContent {
	return esi.params.BroadcastContent
}

func (esi *eventServiceInstance) IsSendStateChangeEvents() bool {
	return esi.params.SendStateChangeEvents
}

func (esi *eventServiceInstance) Shutdown() {
	log.Infoln("Waiting until queued event messages have been dispatched.")
	for len(esi.eventsOutQueue) > 0 {
		time.Sleep(25 * time.Millisecond)
	}
	close(esi.eventsOutQueue)
	esi.disconnect()
	eventService = nil
	eventServiceControl = nil
}
