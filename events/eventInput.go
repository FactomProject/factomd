package events

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/events/eventmessages"
	log "github.com/sirupsen/logrus"
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

func EventFromMessage(eventSource eventmessages.EventSource, msg interfaces.IMsg) *ProcessEvent {
	return &ProcessEvent{
		eventSource: eventSource,
		payload:     msg}
}

func ProcessInfoEvent(message string) *NodeEvent {
	return &NodeEvent{
		eventSource: eventmessages.EventSource_PROCESS_INFO,
		payload:     message,
	}
}

func ProcessInfoEventF(format string, values ...interface{}) *NodeEvent {
	return &NodeEvent{
		eventSource: eventmessages.EventSource_PROCESS_INFO,
		payload:     fmt.Sprintf(format, values),
	}
}

func NodeInfoEvent(message string) *NodeEvent {
	return &NodeEvent{
		eventSource: eventmessages.EventSource_NODE_INFO,
		payload:     message,
	}
}

func NodeInfoEventF(format string, values ...interface{}) *NodeEvent {
	return &NodeEvent{
		eventSource: eventmessages.EventSource_NODE_INFO,
		payload:     fmt.Sprintf(format, values),
	}
}

func NodeErrorEvent(message string, error interface{}) *NodeEvent {
	errorMsg := fmt.Sprint(message, error)
	event := &NodeEvent{
		eventSource: eventmessages.EventSource_NODE_ERROR,
		payload:     errorMsg,
	}
	log.Errorln(errorMsg)
	return event
}
