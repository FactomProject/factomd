package events

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/events/eventmessages/generated/eventmessages"
	log "github.com/sirupsen/logrus"
)

type EventInput interface {
	GetStreamSource() eventmessages.EventSource
}

type RegistrationEvent struct {
	eventSource eventmessages.EventSource
	payload     interfaces.IMsg
}

type StateChangeEvent struct {
	eventSource eventmessages.EventSource
	entityState eventmessages.EntityState
	payload     interfaces.IMsg
}

type ProcessMessageEvent struct {
	eventSource    eventmessages.EventSource
	processMessage *eventmessages.ProcessMessage
}

type NodeMessageEvent struct {
	eventSource eventmessages.EventSource
	nodeMessage *eventmessages.NodeMessage
}

func (event RegistrationEvent) GetStreamSource() eventmessages.EventSource {
	return event.eventSource
}

func (event RegistrationEvent) GetPayload() interfaces.IMsg {
	return event.payload
}

func (event StateChangeEvent) GetStreamSource() eventmessages.EventSource {
	return event.eventSource
}

func (event StateChangeEvent) GetEntityState() eventmessages.EntityState {
	return event.entityState
}

func (event StateChangeEvent) GetPayload() interfaces.IMsg {
	return event.payload
}

func (event ProcessMessageEvent) GetStreamSource() eventmessages.EventSource {
	return event.eventSource
}

func (event ProcessMessageEvent) GetProcessMessage() *eventmessages.ProcessMessage {
	return event.processMessage
}

func (event NodeMessageEvent) GetStreamSource() eventmessages.EventSource {
	return event.eventSource
}

func (event NodeMessageEvent) GetNodeMessage() *eventmessages.NodeMessage {
	return event.nodeMessage
}

func NewRegistrationEvent(streamSource eventmessages.EventSource, msg interfaces.IMsg) *RegistrationEvent {
	return &RegistrationEvent{
		eventSource: streamSource,
		payload:     msg}
}

func NewStateChangeEvent(streamSource eventmessages.EventSource, entityState eventmessages.EntityState, msg interfaces.IMsg) *StateChangeEvent {
	return &StateChangeEvent{
		eventSource: streamSource,
		entityState: entityState,
		payload:     msg}
}

func ProcessInfoMessage(streamSource eventmessages.EventSource, messageCode eventmessages.ProcessMessageCode, message string) *ProcessMessageEvent {
	return &ProcessMessageEvent{
		eventSource: streamSource,
		processMessage: &eventmessages.ProcessMessage{
			MessageCode: messageCode,
			Level:       eventmessages.Level_INFO,
			MessageText: message,
		},
	}
}

func ProcessInfoEventF(streamSource eventmessages.EventSource, messageCode eventmessages.ProcessMessageCode, format string, values ...interface{}) *ProcessMessageEvent {
	return ProcessInfoMessage(streamSource, messageCode, fmt.Sprintf(format, values...))
}

func NodeInfoMessage(messageCode eventmessages.NodeMessageCode, message string) *NodeMessageEvent {
	return &NodeMessageEvent{
		eventSource: eventmessages.EventSource_LIVE,
		nodeMessage: &eventmessages.NodeMessage{
			MessageCode: messageCode,
			Level:       eventmessages.Level_INFO,
			MessageText: message,
		},
	}
}

func NodeInfoMessageF(messageCode eventmessages.NodeMessageCode, format string, values ...interface{}) *NodeMessageEvent {
	return NodeInfoMessage(messageCode, fmt.Sprintf(format, values...))
}

func NodeErrorMessage(messageCode eventmessages.NodeMessageCode, message string, values interface{}) *NodeMessageEvent {
	errorMsg := fmt.Sprint(message, values)
	event := &NodeMessageEvent{
		nodeMessage: &eventmessages.NodeMessage{
			MessageCode: messageCode,
			Level:       eventmessages.Level_ERROR,
			MessageText: errorMsg,
		},
	}
	log.Errorln(errorMsg)
	return event
}
