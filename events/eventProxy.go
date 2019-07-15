package events

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/FactomProject/factomd/common/messages/eventMsgs"
	"github.com/FactomProject/factomd/p2p"
	"github.com/gogo/protobuf/proto"
	"io"
	"net"
)

type EventProxy struct {
	eventsOutQueue chan eventMessages.Event
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
	conn := dialServer()

	for event := range ep.eventsOutQueue {

		if conn == nil {
			conn = dialServer()
		}
		writer := bufio.NewWriter(conn)

		// TODO implement give-up / retry policy
		retry := 3
		for {
			messageBuffer, err := proto.Marshal(event)
			if err == nil {
				i := int32(len(messageBuffer))
				err := binary.Write(writer, binary.LittleEndian, i)
				if err != nil {
					fmt.Println(err)
				}
				_, err = writer.Write(messageBuffer)
				writer.Flush()
				if err == io.EOF {
					conn = redial(conn)
					retry--
					if retry > 0 {
						continue
					}
				} else if err != nil {
					fmt.Println("Event network error", err)
				}
				break
			}
		}
	}
}

func dialServer() net.Conn {
	conn, err := net.Dial("tcp", "127.0.0.1:8040")
	if err != nil {
		fmt.Println("Unable to dial event server")
		return nil
	}

	return conn
}

func redial(conn net.Conn) net.Conn {
	conn.Close()
	return dialServer()
}
