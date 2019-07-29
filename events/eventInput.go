package events

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/events/eventmessages"
)

type EventInput interface {
	GetEventSource() eventmessages.EventSource
}

type ProcessEvent struct {
	eventSource eventmessages.EventSource
	payload     interfaces.IMsg
}

type NodeEvent struct {
	eventSource eventmessages.EventSource
	payload     string
}

func (processEvent ProcessEvent) GetEventSource() eventmessages.EventSource {
	return processEvent.eventSource
}

func (processEvent ProcessEvent) GetPayload() interfaces.IMsg {
	return processEvent.payload
}

func (nodeEvent NodeEvent) GetEventSource() eventmessages.EventSource {
	return nodeEvent.eventSource
}
func (nodeEvent NodeEvent) GetPayload() string {
	return nodeEvent.payload
}

func (nodeEvent NodeEvent) Println() {
	fmt.Println(nodeEvent.GetPayload())
}

func EventFromMessage(eventSource eventmessages.EventSource, msg interfaces.IMsg) *ProcessEvent {
	return &ProcessEvent{
		eventSource: eventSource,
		payload:     msg}
}

func NewInfoEvent(message string) *NodeEvent {
	return &NodeEvent{
		eventSource: eventmessages.EventSource_NODE_INFO,
		payload:     message,
	}
}

func NewInfoEventF(format string, values ...interface{}) *NodeEvent {
	return &NodeEvent{
		eventSource: eventmessages.EventSource_NODE_INFO,
		payload:     fmt.Sprintf(format, values),
	}
}

func NewErrorEvent(message string, error interface{}) *NodeEvent {
	return &NodeEvent{
		eventSource: eventmessages.EventSource_NODE_ERROR,
		payload:     fmt.Sprint(message, error),
	}
}
