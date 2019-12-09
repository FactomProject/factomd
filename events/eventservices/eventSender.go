package eventservices

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/events/eventconfig"
	"github.com/FactomProject/factomd/events/eventmessages/generated/eventmessages"
	"github.com/FactomProject/factomd/p2p"
	"github.com/FactomProject/factomd/util"
	"github.com/gogo/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"net"
	"reflect"
	"time"
)

var eventSenderInstance *eventSender

const (
	defaultProtocol       = "tcp"
	defaultConnectionHost = "127.0.0.1"
	defaultConnectionPort = 8040
	defaultOutputFormat   = eventconfig.Protobuf
	protocolVersion       = byte(1)
)

var (
	dialRetryPostponeDuration = 5 * time.Minute
	redialSleepDuration       = 10 * time.Second
	sendRetries               = 3
)

type EventSender interface {
	// Send(event eventinput.EventInput) error
	GetBroadcastContent() eventconfig.BroadcastContent
	Shutdown()
	IsSendStateChangeEvents() bool
	ReplayDuringStartup() bool
	GetEventQueue() chan *eventmessages.FactomEvent
	IncreaseDroppedFromQueueCounter()
}

type eventSender struct {
	params                  *EventServiceParams
	eventsOutQueue          chan *eventmessages.FactomEvent
	postponeSendingUntil    time.Time
	connection              net.Conn
	droppedFromQueueCounter prometheus.Counter
	notSentCounter          prometheus.Counter
}

func NewEventSender(config *util.FactomdConfig, factomParams *globals.FactomParams) EventSender {
	return NewEventSenderTo(selectParameters(factomParams, config))
}

func NewEventSenderTo(params *EventServiceParams) EventSender {
	if eventSenderInstance == nil {
		eventSenderInstance = &eventSender{
			eventsOutQueue: make(chan *eventmessages.FactomEvent, p2p.StandardChannelSize),
			params:         params,
		}

		eventSenderInstance.droppedFromQueueCounter = prometheus.NewCounter(prometheus.CounterOpts{
			Name: "factomd_livefeed_dropped_from_queue_counter",
			Help: "Number of times we dropped events due of a full the event queue",
		})
		eventSenderInstance.notSentCounter = prometheus.NewCounter(prometheus.CounterOpts{
			Name: "factomd_livefeed_not_send_counter",
			Help: "Number of times we couldn't send out an event",
		})

		go eventSenderInstance.processEventsChannel()
	}
	return eventSenderInstance
}

// TODO describe choice of dropping events.
func (eventSender *eventSender) processEventsChannel() {
	eventSender.connect()

	for event := range eventSender.eventsOutQueue {
		if eventSender.postponeSendingUntil.IsZero() || eventSender.postponeSendingUntil.Before(time.Now()) {
			eventSender.sendEvent(event)
		} else {
			eventSender.notSentCounter.Inc()
		}
	}
}

func (eventSender *eventSender) sendEvent(event *eventmessages.FactomEvent) {
	data, err := eventSender.marshallMessage(event)
	if err != nil {
		log.Errorf("An error occurred while serializing factom event of type %s: %v", reflect.TypeOf(event), err)
		eventSender.notSentCounter.Inc()
		return
	}

	// retry sending event ... times
	sendSuccessful := false
	for retry := 0; retry < sendRetries && !sendSuccessful; retry++ {
		if err = eventSender.connect(); err != nil {
			log.Errorf("An error occurred while connecting to receiver %s: %v, retry %d", eventSender.params.Address, err, retry)
			time.Sleep(redialSleepDuration)
			continue
		}

		// send the factom event to the live api
		if err = eventSender.writeEvent(data); err == nil {
			sendSuccessful = true
		} else {
			log.Errorf("An error occurred while sending a message to receiver %s: %v, retry %d", eventSender.params.Address, err, retry)

			// reset connection and retry
			eventSender.disconnect()
			eventSender.connection = nil
			time.Sleep(redialSleepDuration)
		}
	}

	if !sendSuccessful {
		eventSender.notSentCounter.Inc()
		eventSender.postponeSendingUntil = time.Now().Add(dialRetryPostponeDuration)
	}
}

func (eventSender *eventSender) marshallMessage(event *eventmessages.FactomEvent) ([]byte, error) {
	var data []byte
	var err error
	switch eventSender.params.OutputFormat {
	case eventconfig.Protobuf:
		data, err = eventSender.marshallEvent(event)
	case eventconfig.Json:
		data, err = json.Marshal(event)
	default:
		return nil, errors.New("unsupported event format: " + eventSender.params.OutputFormat.String())
	}
	return data, err
}

func (eventSender *eventSender) connect() error {
	defer catchConnectPanics()

	if eventSender.connection == nil {
		log.Infoln("Connecting to ", eventSender.params.Address)
		conn, err := net.Dial(eventSender.params.Protocol, eventSender.params.Address)
		if err != nil {
			return fmt.Errorf("failed to connect: %v", err)
		}
		eventSender.connection = conn
		eventSender.postponeSendingUntil = time.Time{}
	}
	return nil
}

func catchConnectPanics() error {
	if r := recover(); r != nil {
		return errors.New(fmt.Sprintf("failed to connect to receiver: %v", r))
	}
	return nil
}

func (eventSender *eventSender) disconnect() {
	if eventSender.connection != nil {
		log.Infoln("Closing connection to receiver", eventSender.params.Address)
		err := eventSender.connection.Close()
		if err != nil {
			log.Warnln("An error occurred while closing connection to receiver", eventSender.params.Address)
		}
	}
}

func (eventSender *eventSender) marshallEvent(event *eventmessages.FactomEvent) (data []byte, err error) {
	data, err = proto.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshell event: %v", err)
	}
	return data, err
}

func (eventSender *eventSender) writeEvent(data []byte) (err error) {
	defer catchSendPanics()

	writer := bufio.NewWriter(eventSender.connection)
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

func (eventSender *eventSender) GetEventQueue() chan *eventmessages.FactomEvent {
	return eventSender.eventsOutQueue
}

func (eventSender *eventSender) GetBroadcastContent() eventconfig.BroadcastContent {
	return eventSender.params.BroadcastContent
}

func (eventSender *eventSender) IsSendStateChangeEvents() bool {
	return eventSender.params.SendStateChangeEvents
}

func (eventSender *eventSender) ReplayDuringStartup() bool {
	return eventSender.params.ReplayDuringStartup
}

func (eventSender *eventSender) IncreaseDroppedFromQueueCounter() {
	eventSender.droppedFromQueueCounter.Inc()
}

func (eventSender *eventSender) Shutdown() {
	log.Infoln("Waiting until queued event messages have been dispatched.")
	for len(eventSender.eventsOutQueue) > 0 {
		time.Sleep(25 * time.Millisecond)
	}
	close(eventSender.eventsOutQueue)
	eventSender.disconnect()
	eventSenderInstance = nil
}
