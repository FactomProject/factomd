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
	eventSource    eventmessages.EventSource
	processMessage *eventmessages.ProcessMessage
	payload        interfaces.IMsg
}

type NodeEvent struct {
	eventSource eventmessages.EventSource
	nodeMessage *eventmessages.NodeMessage
	payload     interfaces.IMsg
}

func (processEvent ProcessEvent) GetEventSource() eventmessages.EventSource {
	return processEvent.eventSource
}

func (nodeEvent NodeEvent) GetEventSource() eventmessages.EventSource {
	return nodeEvent.eventSource
}

func (processEvent ProcessEvent) GetProcessMessage() *eventmessages.ProcessMessage {
	return processEvent.processMessage
}

func (processEvent ProcessEvent) GetPayload() interfaces.IMsg {
	return processEvent.payload
}

func (nodeEvent NodeEvent) GetNodeEvent() *eventmessages.NodeMessage {
	return nodeEvent.nodeMessage
}

func EventFromNetworkMessage(eventSource eventmessages.EventSource, msg interfaces.IMsg) *ProcessEvent {
	return &ProcessEvent{
		eventSource: eventSource,
		payload:     msg}
}

func ProcessInfoMessage(messageCode eventmessages.ProcessMessageCode, message string) *ProcessEvent {
	return &ProcessEvent{
		eventSource: eventmessages.EventSource_PROCESS_MESSAGE,
		processMessage: &eventmessages.ProcessMessage{
			MessageCode: messageCode,
			Level:       eventmessages.Level_INFO,
			MessageText: message,
		},
	}
}

func ProcessInfoEventF(messageCode eventmessages.ProcessMessageCode, format string, values ...interface{}) *ProcessEvent {
	return ProcessInfoMessage(messageCode, fmt.Sprintf(format, values))
}

func NodeInfoMessage(messageCode eventmessages.NodeMessageCode, message string) *NodeEvent {
	return &NodeEvent{
		eventSource: eventmessages.EventSource_NODE_MESSAGE,
		nodeMessage: &eventmessages.NodeMessage{
			MessageCode: messageCode,
			Level:       eventmessages.Level_INFO,
			MessageText: message,
		},
	}
}

func NodeInfoMessageF(messageCode eventmessages.NodeMessageCode, format string, values ...interface{}) *NodeEvent {
	return NodeInfoMessage(messageCode, fmt.Sprintf(format, values))
}

func NodeErrorMessage(messageCode eventmessages.NodeMessageCode, message string, values interface{}) *NodeEvent {
	errorMsg := fmt.Sprint(message, values)
	event := &NodeEvent{
		eventSource: eventmessages.EventSource_NODE_MESSAGE,
		nodeMessage: &eventmessages.NodeMessage{
			MessageCode: messageCode,
			Level:       eventmessages.Level_ERROR,
			MessageText: errorMsg,
		},
	}
	log.Errorln(errorMsg)
	return event
}
