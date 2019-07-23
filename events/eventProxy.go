package events

import (
	"bufio"
	"encoding/binary"
	"fmt"
	eventMessages "github.com/FactomProject/factomd/common/messages/eventmessages"
	eventsInput "github.com/FactomProject/factomd/common/messages/eventmessages/input"
	"github.com/FactomProject/factomd/p2p"
	"github.com/gogo/protobuf/proto"
	"io"
	"net"
	"strings"
	"sync/atomic"
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

type IEventProxy interface {
	Send(event *eventsInput.EventInput)
}

type EventProxy struct {
	eventsOutQueue     chan eventsInput.EventInput
	postponeRetryUntil time.Time
	connection         net.Conn
}

func (ep *EventProxy) Init() *EventProxy {
	ep.eventsOutQueue = make(chan eventsInput.EventInput, p2p.StandardChannelSize)
	return ep
}

func (ep *EventProxy) StartProxy() *EventProxy {
	go ep.processEventsChannel()
	return ep
}

func (ep *EventProxy) Send(event *eventsInput.EventInput) {
	select {
	case ep.eventsOutQueue <- *event:
	default:
	}
}

func (ep *EventProxy) processEventsChannel() {
	ep.dialServer()
	for event := range ep.eventsOutQueue {
		ep.processEvent(event)
	}
}

func (ep *EventProxy) processEvent(event eventsInput.EventInput) {
	defer handleEventError()
	if ep.connectionOk() {
		factomEvent := MapToFactomEvent(event)
		if factomEvent != nil {
			ep.sendEvent(factomEvent)
		}
	}
}

func (ep *EventProxy) connectionOk() bool {
	if ep.connection == nil {
		// We'll try to reconnect only once, if the receiver is still not there we try again later but we won't let the queue fill up.
		// These events will be skipped.
		if ep.postponeRetryUntil.IsZero() || time.Now().After(ep.postponeRetryUntil) {
			ep.dialServer()
			if ep.connection == nil {
				ep.postponeRetryUntil = time.Now().Add(dialRetryPostponeDuration)
				return false
			}
		}
		return false
	}
	ep.postponeRetryUntil = time.Unix(0, 0)
	return true
}

func (ep *EventProxy) sendEvent(event *eventMessages.FactomEvent) {
	writer := bufio.NewWriter(ep.connection)
	retry := uint32(0)
	sentOk := false
	for !sentOk {
		defer ep.handleSendError(&retry)

		messageBuffer := ep.marshallEvent(event)
		writeMessageLength(writer, len(messageBuffer))
		sentOk = ep.writeMessage(writer, messageBuffer)
	}
}

func (ep *EventProxy) marshallEvent(event *eventMessages.FactomEvent) []byte {
	messageBuffer, err := proto.Marshal(event)
	if err != nil {
		panic(fmt.Sprint("An error occurred when marshalling an event to a protocol buffer", err))
	}
	return messageBuffer
}

func writeMessageLength(writer *bufio.Writer, length int) {
	err := binary.Write(writer, binary.LittleEndian, int32(length))
	if err != nil {
		panic(err)
	}
}

func (ep *EventProxy) writeMessage(writer *bufio.Writer, messageBuffer []byte) bool {
	_, err := writer.Write(messageBuffer)
	ep.assertSendError(err)
	err = writer.Flush()
	ep.assertSendError(err)
	return true
}

func (ep *EventProxy) assertSendError(err error) {
	if err != nil {
		if err == io.EOF || strings.Contains(err.Error(), "broken pipe") {
			ep.redial()
			if ep.connection == nil {
				ep.postponeRetryUntil = time.Now().Add(dialRetryPostponeDuration)
			}
		}
		panic(fmt.Sprint("Event network error ", err))
	}
}

func (ep *EventProxy) handleSendError(retry *uint32) {
	atomic.AddUint32(retry, 1)
	if *retry < sendRetries && (ep.postponeRetryUntil.IsZero() || time.Now().After(ep.postponeRetryUntil)) {
		if r := recover(); r != nil {
			fmt.Println("Unable to send event", r)
		}
	}
}

func (ep *EventProxy) dialServer() {

	defer handleConnectError()

	var err error
	ep.connection, err = net.Dial(defaultConnectionProtocol, fmt.Sprintf("%s:%s", defaultConnectionHost, defaultConnectionPort))
	if err != nil {
		ep.connection = nil
		panic(err)
	}
}

func handleEventError() {
	if r := recover(); r != nil {
		fmt.Println("Unable to process event, proceeding with the next.", r)
	}
}

func handleConnectError() {
	if r := recover(); r != nil {
		fmt.Println("Unable to dial event server", r)
	}
}

func (ep *EventProxy) redial() {
	err := ep.connection.Close()
	if err != nil {
		fmt.Println("Close connection returned an error", err)
	}
	ep.connection = nil
	ep.dialServer()
}
