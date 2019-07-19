package events

import (
	"bufio"
	"encoding/binary"
	"fmt"
	eventMessages "github.com/FactomProject/factomd/common/messages/eventmessages"
	"github.com/FactomProject/factomd/p2p"
	"github.com/gogo/protobuf/proto"
	"github.com/prometheus/common/log"
	"io"
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

type EventProxy struct {
	eventsOutQueue     chan eventMessages.Event
	postponeRetryUntil time.Time
	connection         net.Conn
}

func (ep *EventProxy) Init() *EventProxy {
	ep.eventsOutQueue = make(chan eventMessages.Event, p2p.StandardChannelSize)
	return ep
}

func (ep *EventProxy) StartProxy() *EventProxy {
	go ep.processEventsChannel()
	return ep
}

func (ep *EventProxy) Send(event eventMessages.Event) {
	select {
	case ep.eventsOutQueue <- event:
	default:
	}
}

func (ep *EventProxy) processEventsChannel() {
	ep.dialServer()

	for event := range ep.eventsOutQueue {

		if ep.connection == nil {
			if ep.postponeRetryUntil.IsZero() || time.Now().After(ep.postponeRetryUntil) {
				ep.dialServer()
				if ep.connection == nil {
					ep.postponeRetryUntil = time.Now().Add(dialRetryPostponeDuration)
					continue
				}
			}
			continue
		}

		factomEvent := eventMessages.WrapInFactomEvent(event)
		ep.sendEvent(factomEvent)
		ep.postponeRetryUntil = time.Now().Add(dialRetryPostponeDuration)
	}
}

func (ep *EventProxy) sendEvent(event eventMessages.Event) {
	writer := bufio.NewWriter(ep.connection)
	retry := 0
	for {
		messageBuffer, err := proto.Marshal(event)
		if err != nil {
			log.Error("An error occurred when marshalling an event to a protocol buffer", err)
			break
		}

		i := int32(len(messageBuffer))
		err = binary.Write(writer, binary.LittleEndian, i)
		if err != nil {
			log.Error(err)
			break
		}
		_, err = writer.Write(messageBuffer)
		writer.Flush()
		if err == io.EOF {
			ep.redial()
			if ep.connection == nil {
				time.Sleep(redialSleepDuration)
			}
			retry++
			if retry < sendRetries {
				continue
			}
		} else if err != nil {
			fmt.Println("Event network error", err)
		}
		break
	}
}

func (ep *EventProxy) dialServer() {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Unable to dial event server", r)
		}
	}()

	var err error
	ep.connection, err = net.Dial(defaultConnectionProtocol, fmt.Sprintf("%s:%s", defaultConnectionHost, defaultConnectionPort))
	if err != nil {
		ep.connection = nil
		panic(err)
	}
}

func (ep *EventProxy) redial() {
	ep.connection.Close()
	ep.dialServer()
}
