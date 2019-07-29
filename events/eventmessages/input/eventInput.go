package eventsinput

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/events/eventmessages"
)

type EventInput struct {
	eventSource    eventmessages.EventSource
	processMessage interfaces.IMsg
	nodeMessage    string
}

func (srcEvent *EventInput) GetEventSource() eventmessages.EventSource {
	return srcEvent.eventSource
}

func (srcEvent *EventInput) GetProcessMessage() interfaces.IMsg {
	return srcEvent.processMessage
}

func (srcEvent *EventInput) GetNodeMessage() string {
	return srcEvent.nodeMessage
}

func EventFromMessage(eventSource eventmessages.EventSource, msg interfaces.IMsg) *EventInput {
	return &EventInput{
		eventSource:    eventSource,
		processMessage: msg}
}

func NewInfoEvent(message string) *EventInput {
	return &EventInput{
		eventSource: eventmessages.EventSource_NODE_INFO,
		nodeMessage: message,
	}
}

func NewInfoEventF(format string, values ...interface{}) *EventInput {
	return &EventInput{
		eventSource: eventmessages.EventSource_NODE_INFO,
		nodeMessage: fmt.Sprintf(format, values),
	}
}

func NewErrorEvent(message string, error interface{}) *EventInput {
	return &EventInput{
		eventSource: eventmessages.EventSource_NODE_ERROR,
		nodeMessage: fmt.Sprint(message, error),
	}
}
