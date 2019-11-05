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

type StateChangeMsgEvent struct {
	eventSource eventmessages.EventSource
	entityState eventmessages.EntityState
	payload     interfaces.IMsg
}

type StateChangeEvent struct {
	eventSource eventmessages.EventSource
	entityState eventmessages.EntityState
	payload     interfaces.IDBState
}

type ProcessListEvent struct {
	eventSource              eventmessages.EventSource
	processListEventInstance *eventmessages.ProcessListEvent
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

func (event StateChangeMsgEvent) GetStreamSource() eventmessages.EventSource {
	return event.eventSource
}

func (event StateChangeMsgEvent) GetEntityState() eventmessages.EntityState {
	return event.entityState
}

func (event StateChangeMsgEvent) GetPayload() interfaces.IMsg {
	return event.payload
}

func (event StateChangeEvent) GetStreamSource() eventmessages.EventSource {
	return event.eventSource
}

func (event StateChangeEvent) GetEntityState() eventmessages.EntityState {
	return event.entityState
}

func (event StateChangeEvent) GetPayload() interfaces.IDBState {
	return event.payload
}

func (event ProcessListEvent) GetStreamSource() eventmessages.EventSource {
	return event.eventSource
}

func (event ProcessListEvent) GetProcessListEvent() *eventmessages.ProcessListEvent {
	return event.processListEventInstance
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

func NewStateChangeEventFromMsg(streamSource eventmessages.EventSource, entityState eventmessages.EntityState, msg interfaces.IMsg) *StateChangeMsgEvent {
	return &StateChangeMsgEvent{
		eventSource: streamSource,
		entityState: entityState,
		payload:     msg}
}

func NewStateChangeEvent(streamSource eventmessages.EventSource, entityState eventmessages.EntityState, dbState interfaces.IDBState) *StateChangeEvent {
	return &StateChangeEvent{
		eventSource: streamSource,
		entityState: entityState,
		payload:     dbState}
}

func ProcessListEventNewBlock(streamSource eventmessages.EventSource, newBlockHeight uint32) *ProcessListEvent {
	return &ProcessListEvent{
		eventSource: streamSource,
		processListEventInstance: &eventmessages.ProcessListEvent{
			ProcessListEvent: &eventmessages.ProcessListEvent_NewBlockEvent{
				NewBlockEvent: &eventmessages.NewBlockEvent{
					NewBlockHeight: newBlockHeight,
				},
			},
		},
	}
}

func ProcessListEventNewMinute(streamSource eventmessages.EventSource, newMinute int, blockHeight uint32) *ProcessListEvent {
	return &ProcessListEvent{
		eventSource: streamSource,
		processListEventInstance: &eventmessages.ProcessListEvent{
			ProcessListEvent: &eventmessages.ProcessListEvent_NewMinuteEvent{
				NewMinuteEvent: &eventmessages.NewMinuteEvent{
					NewMinute:   uint32(newMinute),
					BlockHeight: blockHeight,
				},
			},
		},
	}
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
	errorMsg := fmt.Sprintf(message, values)
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
