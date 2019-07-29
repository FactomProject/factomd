package eventsinput

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/events/eventmessages"
)

type EventInput struct {
	eventSource    eventmessages.EventSource
	messagePayload interfaces.IMsg
	message        string
}

func (srcEvent *EventInput) GetEventSource() eventmessages.EventSource {
	return srcEvent.eventSource
}

func (srcEvent *EventInput) GetMessagePayload() interfaces.IMsg {
	return srcEvent.messagePayload
}

func (srcEvent *EventInput) GetMessage() string {
	return srcEvent.message
}

func EventFromMessage(eventSource eventmessages.EventSource, msg interfaces.IMsg) *EventInput {
	return &EventInput{
		eventSource:    eventSource,
		messagePayload: msg}
}

func NewInfoEvent(message string) *EventInput {
	return &EventInput{
		eventSource: eventmessages.EventSource_NODE_INFO,
		message:     message,
	}
}

func NewInfoEventF(format string, values ...interface{}) *EventInput {
	return &EventInput{
		eventSource: eventmessages.EventSource_NODE_INFO,
		message:     fmt.Sprintf(format, values),
	}
}

func NewErrorEvent(message string, error interface{}) *EventInput {
	return &EventInput{
		eventSource: eventmessages.EventSource_NODE_ERROR,
		message:     fmt.Sprint(message, error),
	}
}
