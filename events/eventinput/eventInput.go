package eventinput

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/events/eventmessages/generated/eventmessages"
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
