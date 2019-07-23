package eventsinput

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages/eventmessages"
)

type EventInput struct {
	eventSource    eventmessages.EventSource
	messagePayload interfaces.IMsg
}

func (srcEvent *EventInput) GetEventSource() eventmessages.EventSource {
	return srcEvent.eventSource
}

func (srcEvent *EventInput) GetMessagePayload() interfaces.IMsg {
	return srcEvent.messagePayload
}

func SourceEventFromMessage(eventSource eventmessages.EventSource, msg interfaces.IMsg) *EventInput {
	return &EventInput{
		eventSource:    eventSource,
		messagePayload: msg}
}
