package eventinput

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/events/eventmessages/generated/eventmessages"
	"github.com/prometheus/common/log"
)

type EventInput interface {
	GetStreamSource() eventmessages.EventSource
}

type RegistrationEvent struct {
	EventSource eventmessages.EventSource
	Payload     interfaces.IMsg
}

type StateChangeMsgEvent struct {
	EventSource eventmessages.EventSource
	EntityState eventmessages.EntityState
	Payload     interfaces.IMsg
}

type StateChangeEvent struct {
	EventSource eventmessages.EventSource
	EntityState eventmessages.EntityState
	Payload     interfaces.IDBState
}

type AnchorEvent struct {
	EventSource eventmessages.EventSource
	Payload     interfaces.IDirBlockInfo
}

type ProcessListEvent struct {
	EventSource              eventmessages.EventSource
	ProcessListEventInstance *eventmessages.ProcessListEvent
}

type NodeMessageEvent struct {
	EventSource eventmessages.EventSource
	NodeMessage *eventmessages.NodeMessage
}

func (event RegistrationEvent) GetStreamSource() eventmessages.EventSource {
	return event.EventSource
}

func (event RegistrationEvent) GetPayload() interfaces.IMsg {
	return event.Payload
}

func (event StateChangeMsgEvent) GetStreamSource() eventmessages.EventSource {
	return event.EventSource
}

func (event StateChangeMsgEvent) GetEntityState() eventmessages.EntityState {
	return event.EntityState
}

func (event StateChangeMsgEvent) GetPayload() interfaces.IMsg {
	return event.Payload
}

func (event StateChangeEvent) GetStreamSource() eventmessages.EventSource {
	return event.EventSource
}

func (event StateChangeEvent) GetEntityState() eventmessages.EntityState {
	return event.EntityState
}

func (event StateChangeEvent) GetPayload() interfaces.IDBState {
	return event.Payload
}

func (event AnchorEvent) GetStreamSource() eventmessages.EventSource {
	return event.EventSource
}

func (event AnchorEvent) GetPayload() interfaces.IDirBlockInfo {
	return event.Payload
}

func (event ProcessListEvent) GetStreamSource() eventmessages.EventSource {
	return event.EventSource
}

func (event ProcessListEvent) GetProcessListEvent() *eventmessages.ProcessListEvent {
	return event.ProcessListEventInstance
}

func (event NodeMessageEvent) GetStreamSource() eventmessages.EventSource {
	return event.EventSource
}

func (event NodeMessageEvent) GetNodeMessage() *eventmessages.NodeMessage {
	return event.NodeMessage
}

func NewRegistrationEvent(streamSource eventmessages.EventSource, msg interfaces.IMsg) *RegistrationEvent {
	return &RegistrationEvent{
		EventSource: streamSource,
		Payload:     msg}
}

func NewStateChangeEventFromMsg(streamSource eventmessages.EventSource, entityState eventmessages.EntityState, msg interfaces.IMsg) *StateChangeMsgEvent {
	return &StateChangeMsgEvent{
		EventSource: streamSource,
		EntityState: entityState,
		Payload:     msg}
}

func NewStateChangeEvent(streamSource eventmessages.EventSource, entityState eventmessages.EntityState, dbState interfaces.IDBState) *StateChangeEvent {
	return &StateChangeEvent{
		EventSource: streamSource,
		EntityState: entityState,
		Payload:     dbState}
}

func NewAnchorEvent(streamSource eventmessages.EventSource, dbDirBlockInfo interfaces.IDirBlockInfo) *AnchorEvent {
	return &AnchorEvent{
		EventSource: streamSource,
		Payload:     dbDirBlockInfo,
	}
}

func ProcessListEventNewBlock(streamSource eventmessages.EventSource, newBlockHeight uint32) *ProcessListEvent {
	return &ProcessListEvent{
		EventSource: streamSource,
		ProcessListEventInstance: &eventmessages.ProcessListEvent{
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
		EventSource: streamSource,
		ProcessListEventInstance: &eventmessages.ProcessListEvent{
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
		EventSource: eventmessages.EventSource_LIVE,
		NodeMessage: &eventmessages.NodeMessage{
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
		NodeMessage: &eventmessages.NodeMessage{
			MessageCode: messageCode,
			Level:       eventmessages.Level_ERROR,
			MessageText: errorMsg,
		},
	}
	log.Errorln(errorMsg)
	return event
}
