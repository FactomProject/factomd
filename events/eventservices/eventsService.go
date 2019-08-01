package eventservices

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/FactomProject/factomd/common/constants/runstate"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/events"
	"github.com/FactomProject/factomd/events/eventmessages"
	"github.com/FactomProject/factomd/p2p"
	"github.com/gogo/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"net"
	"time"
)

var eventService events.EventService
var eventServiceControl events.EventServiceControl

const (
	defaultConnectionProtocol = "tcp"
	defaultConnectionHost     = "127.0.0.1"
	defaultConnectionPort     = "8040"
	sendRetries               = 3
	dialRetryPostponeDuration = 5 * time.Minute
	redialSleepDuration       = 10 * time.Second
)

type eventServiceInstance struct {
	eventsOutQueue       chan *eventmessages.FactomEvent
	postponeSendingUntil time.Time
	connection           net.Conn
	protocol             string
	address              string
	owningState          interfaces.IState
}

func NewEventService(state interfaces.IState) (events.EventService, events.EventServiceControl) {
	return NewEventServiceTo(defaultConnectionProtocol, fmt.Sprintf("%s:%s", defaultConnectionHost, defaultConnectionPort), state)
}

func NewEventServiceTo(protocol string, address string, state interfaces.IState) (events.EventService, events.EventServiceControl) {
	if eventService == nil {
		eventServiceInstance := &eventServiceInstance{
			eventsOutQueue: make(chan *eventmessages.FactomEvent, p2p.StandardChannelSize),
			protocol:       protocol,
			address:        address,
			owningState:    state,
		}
		eventService = eventServiceInstance
		eventServiceControl = eventServiceInstance
		go eventServiceInstance.processEventsChannel()
	}
	return eventService, eventServiceControl
}

func (ep *eventServiceInstance) Send(event events.EventInput) error {
	if ep.owningState.GetRunState() > runstate.Running { // Stop queuing messages to the events channel when shutting down
		return nil
	}

	factomEvent, err := MapToFactomEvent(event)
	if err != nil {
		return fmt.Errorf("failed to map to factom event: %v\n", err)
	}

	select {
	case ep.eventsOutQueue <- factomEvent:
	default:
	}

	return nil
}

func (ep *eventServiceInstance) processEventsChannel() {
	ep.connect()

	for event := range ep.eventsOutQueue {
		if ep.postponeSendingUntil.IsZero() || ep.postponeSendingUntil.Before(time.Now()) {
			ep.sendEvent(event)
		}
	}
}

func (ep *eventServiceInstance) sendEvent(event *eventmessages.FactomEvent) {
	data, err := ep.marshallEvent(event)
	if err != nil {
		log.Errorf("An error occurred while serializing factom event of type %s: %v", event.EventSource.String(), err)
		return
	}

	// retry sending event ... times
	sendSuccessful := false
	for retry := 0; retry < sendRetries && !sendSuccessful; retry++ {
		if err = ep.connect(); err != nil {
			log.Errorf("An error occurred while connecting to receiver %s: %v, retry %d", ep.address, err, retry)
			time.Sleep(redialSleepDuration)
			continue
		}

		// send the factom event to the live api
		if err = ep.writeEvent(data); err == nil {
			sendSuccessful = true
		} else {
			log.Errorf("An error occurred while sending a message to receiver %s: %v, retry %d", ep.address, err, retry)

			// reset connection and retry
			ep.disconnect()
			time.Sleep(redialSleepDuration)
			ep.connection = nil
		}
	}

	if !sendSuccessful {
		ep.postponeSendingUntil = time.Now().Add(dialRetryPostponeDuration)
	}
}

func (ep *eventServiceInstance) connect() error {
	defer catchConnectPanics()

	if ep.connection == nil {
		conn, err := net.Dial(ep.protocol, ep.address)
		if err != nil {
			return fmt.Errorf("failed to connect: %v", err)
		}
		ep.connection = conn
		ep.postponeSendingUntil = time.Time{}
	}
	return nil
}

func catchConnectPanics() error {
	if r := recover(); r != nil {
		return errors.New(fmt.Sprintf("failed to connect to receiver: %v", r))
	}
	return nil
}

func (ep *eventServiceInstance) disconnect() {
	log.Infoln("Closing connection to receiver", ep.address)
	err := ep.connection.Close()
	if err != nil {
		log.Warnln("An error occurred while closing connection to receiver", ep.address)
	}
}

func (ep *eventServiceInstance) marshallEvent(event *eventmessages.FactomEvent) (data []byte, err error) {
	data, err = proto.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshell event: %v", err)
	}
	return data, err
}

func (ep *eventServiceInstance) writeEvent(data []byte) (err error) {
	defer catchSendPanics()

	writer := bufio.NewWriter(ep.connection)

	dataSize := int32(len(data))
	err = binary.Write(writer, binary.LittleEndian, dataSize)
	if err != nil {
		return fmt.Errorf("failed to write data size header: %v", err)
	}

	_, err = writer.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data: %v", err)
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

func (ep *eventServiceInstance) Shutdown() {
	log.Infoln("Waiting until queued event messages have been dispatched.")
	for len(ep.eventsOutQueue) > 0 {
		time.Sleep(25 * time.Millisecond)
	}
	ep.disconnect()
}
