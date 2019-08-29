package events

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/events/eventmessages"
	log "github.com/sirupsen/logrus"
)

type EventInput interface {
	GetStreamSource() eventmessages.StreamSource
}

type RegistrationEvent struct {
	streamSource eventmessages.StreamSource
	payload      interfaces.IMsg
}

type StateChangeEvent struct {
	streamSource eventmessages.StreamSource
	entityState  eventmessages.EntityState
	payload      interfaces.IMsg
}

type ProcessMessageEvent struct {
	streamSource   eventmessages.StreamSource
	processMessage *eventmessages.ProcessMessage
}

type NodeMessageEvent struct {
	streamSource eventmessages.StreamSource
	nodeMessage  *eventmessages.NodeMessage
}

func (event RegistrationEvent) GetStreamSource() eventmessages.StreamSource {
	return event.streamSource
}

func (event RegistrationEvent) GetPayload() interfaces.IMsg {
	return event.payload
}

func (event StateChangeEvent) GetStreamSource() eventmessages.StreamSource {
	return event.streamSource
}

func (event StateChangeEvent) GetEntityState() eventmessages.EntityState {
	return event.entityState
}

func (event StateChangeEvent) GetPayload() interfaces.IMsg {
	return event.payload
}

func (event ProcessMessageEvent) GetStreamSource() eventmessages.StreamSource {
	return event.streamSource
}

func (event ProcessMessageEvent) GetProcessMessage() *eventmessages.ProcessMessage {
	return event.processMessage
}

func (event NodeMessageEvent) GetStreamSource() eventmessages.StreamSource {
	return event.streamSource
}

func (event NodeMessageEvent) GetNodeMessage() *eventmessages.NodeMessage {
	return event.nodeMessage
}

func NewRegistrationEvent(streamSource eventmessages.StreamSource, msg interfaces.IMsg) *RegistrationEvent {
	return &RegistrationEvent{
		streamSource: streamSource,
		payload:      msg}
}

func NewStateChangeEvent(streamSource eventmessages.StreamSource, entityState eventmessages.EntityState, msg interfaces.IMsg) *StateChangeEvent {
	return &StateChangeEvent{
		streamSource: streamSource,
		payload:      msg}
}

func ProcessInfoMessage(streamSource eventmessages.StreamSource, messageCode eventmessages.ProcessMessageCode, message string) *ProcessMessageEvent {
	return &ProcessMessageEvent{
		streamSource: streamSource,
		processMessage: &eventmessages.ProcessMessage{
			MessageCode: messageCode,
			Level:       eventmessages.Level_INFO,
			MessageText: message,
		},
	}
}

func ProcessInfoEventF(streamSource eventmessages.StreamSource, messageCode eventmessages.ProcessMessageCode, format string, values ...interface{}) *ProcessMessageEvent {
	return ProcessInfoMessage(streamSource, messageCode, fmt.Sprintf(format, values))
}

func NodeInfoMessage(messageCode eventmessages.NodeMessageCode, message string) *NodeMessageEvent {
	return &NodeMessageEvent{
		streamSource: eventmessages.StreamSource_LIVE,
		nodeMessage: &eventmessages.NodeMessage{
			MessageCode: messageCode,
			Level:       eventmessages.Level_INFO,
			MessageText: message,
		},
	}
}

func NodeInfoMessageF(messageCode eventmessages.NodeMessageCode, format string, values ...interface{}) *NodeMessageEvent {
	return NodeInfoMessage(messageCode, fmt.Sprintf(format, values))
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
