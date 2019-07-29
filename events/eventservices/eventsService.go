package eventservices

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/FactomProject/factomd/common/constants/runstate"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/events"
	"github.com/FactomProject/factomd/events/eventmessages"
	"github.com/FactomProject/factomd/p2p"
	"github.com/gogo/protobuf/proto"
	"net"
	"time"
)

const (
	defaultConnectionProtocol = "tcp"
	defaultConnectionHost     = "127.0.0.1"
	defaultConnectionPort     = "8040"
	sendRetries               = 3
	dialRetryPostponeDuration = time.Minute
	redialSleepDuration       = 5 * time.Second
)

type eventServiceInstance struct {
	eventsOutQueue     chan *eventmessages.FactomEvent
	postponeRetryUntil time.Time
	connection         net.Conn
	protocol           string
	address            string
	owningState        interfaces.IState
}

func NewEventService(state interfaces.IState) events.EventService {
	return NewEventServiceTo(defaultConnectionProtocol, fmt.Sprintf("%s:%s", defaultConnectionHost, defaultConnectionPort), state)
}

func NewEventServiceTo(protocol string, address string, state interfaces.IState) events.EventService {
	eventServiceInstance := &eventServiceInstance{
		eventsOutQueue: make(chan *eventmessages.FactomEvent, p2p.StandardChannelSize),
		protocol:       protocol,
		address:        address,
		owningState:    state,
	}
	go eventServiceInstance.processEventsChannel()
	return eventServiceInstance
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
	for event := range ep.eventsOutQueue {
		ep.sendEvent(event)
	}
}

func (ep *eventServiceInstance) sendEvent(event *eventmessages.FactomEvent) {
	data, err := ep.marshallEvent(event)
	if err != nil {
		fmt.Printf("TODO error logging: %v", err)
		return
	}

	// retry sending event ... times
	sendSuccessful := false
	for retry := 0; retry < sendRetries && !sendSuccessful; retry++ {
		if err = ep.connect(); err != nil {
			// TODO handle error
			fmt.Printf("TODO error logging: %v", err)
			time.Sleep(redialSleepDuration)
			continue
		}

		// send the factom event to the live api
		if err = ep.writeEvent(data); err == nil {
			sendSuccessful = true
		} else {
			// TODO handle / log error
			fmt.Printf("TODO error logging: %v\n", err)

			// reset connection and retry
			time.Sleep(redialSleepDuration)
			ep.connection = nil
		}
	}
}

func (ep *eventServiceInstance) connect() error {
	if ep.connection == nil {
		conn, err := net.Dial(ep.protocol, ep.address)
		if err != nil {
			return fmt.Errorf("failed to connect: %v", err)
		}
		ep.connection = conn
		ep.postponeRetryUntil = time.Unix(0, 0)
	}
	return nil
}

func (ep *eventServiceInstance) marshallEvent(event *eventmessages.FactomEvent) (data []byte, err error) {
	data, err = proto.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshell event: %v", err)
	}
	return data, err
}

func (ep *eventServiceInstance) writeEvent(data []byte) (err error) {
	writer := bufio.NewWriter(ep.connection)

	dataSize := int32(len(data))
	err = binary.Write(writer, binary.LittleEndian, dataSize)
	if err != nil {
		return fmt.Errorf("failed to write data size: %v", err)
	}

	_, err = writer.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data: %v", err)
	}
	err = writer.Flush()
	return nil
}

func (ep *eventServiceInstance) HasQueuedMessages() bool {
	return len(ep.eventsOutQueue) > 0
}

func (ep *eventServiceInstance) WaitForQueuedMessages() {
	for ep.HasQueuedMessages() {
		time.Sleep(25 * time.Millisecond)
	}
}
