package events

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/events/eventinput"
	"github.com/FactomProject/factomd/events/eventmessages/generated/eventmessages"
	"github.com/prometheus/common/log"
)

func EmitRegistrationEvent(msg interfaces.IMsg, state IStateEventServices) {
	if state.GetEventsService() != nil {
		switch msg.(type) { // Do not fill the channel with message we don't want anyway (like EOM's)
		case *messages.CommitChainMsg, *messages.CommitEntryMsg, *messages.RevealEntryMsg:
			event := NewRegistrationEvent(GetStreamSource(state), msg)
			state.GetEventsService().Send(event)
		}
	}
}

func EmitStateChangeEvent(msg interfaces.IMsg, entityState eventmessages.EntityState, state IStateEventServices) {
	if state.GetEventsService() != nil {
		switch msg.(type) {
		case *messages.CommitChainMsg, *messages.CommitEntryMsg, *messages.RevealEntryMsg, *messages.DBStateMsg:
			event := NewStateChangeEventFromMsg(GetStreamSource(state), entityState, msg)
			state.GetEventsService().Send(event)
		}
	}
}

func EmitDBStateEvent(dbState interfaces.IDBState, entityState eventmessages.EntityState, state IStateEventServices) {
	if state.GetEventsService() != nil {
		event := NewStateChangeEvent(GetStreamSource(state), entityState, dbState)
		state.GetEventsService().Send(event)
	}
}

func EmitDBAnchorEvent(dirBlockInfo interfaces.IDirBlockInfo, state IStateEventServices) {
	if state.GetEventsService() != nil {
		event := NewAnchorEvent(GetStreamSource(state), dirBlockInfo)
		state.GetEventsService().Send(event)
	}
}

func GetStreamSource(state IStateEventServices) eventmessages.EventSource {
	if state.IsRunLeader() {
		return eventmessages.EventSource_LIVE
	} else {
		return eventmessages.EventSource_REPLAY_BOOT
	}
}

func NewRegistrationEvent(streamSource eventmessages.EventSource, msg interfaces.IMsg) *eventinput.RegistrationEvent {
	return &eventinput.RegistrationEvent{
		EventSource: streamSource,
		Payload:     msg}
}

func NewStateChangeEventFromMsg(streamSource eventmessages.EventSource, entityState eventmessages.EntityState, msg interfaces.IMsg) *eventinput.StateChangeMsgEvent {
	return &eventinput.StateChangeMsgEvent{
		EventSource: streamSource,
		EntityState: entityState,
		Payload:     msg}
}

func NewStateChangeEvent(streamSource eventmessages.EventSource, entityState eventmessages.EntityState, dbState interfaces.IDBState) *eventinput.StateChangeEvent {
	return &eventinput.StateChangeEvent{
		EventSource: streamSource,
		EntityState: entityState,
		Payload:     dbState}
}

func NewAnchorEvent(streamSource eventmessages.EventSource, dbDirBlockInfo interfaces.IDirBlockInfo) *eventinput.AnchorEvent {
	return &eventinput.AnchorEvent{
		EventSource: streamSource,
		Payload:     dbDirBlockInfo,
	}
}

func ProcessListEventNewBlock(streamSource eventmessages.EventSource, newBlockHeight uint32) *eventinput.ProcessListEvent {
	return &eventinput.ProcessListEvent{
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

func ProcessListEventNewMinute(streamSource eventmessages.EventSource, newMinute int, blockHeight uint32) *eventinput.ProcessListEvent {
	return &eventinput.ProcessListEvent{
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

func NodeInfoMessage(messageCode eventmessages.NodeMessageCode, message string) *eventinput.NodeMessageEvent {
	return &eventinput.NodeMessageEvent{
		EventSource: eventmessages.EventSource_LIVE,
		NodeMessage: &eventmessages.NodeMessage{
			MessageCode: messageCode,
			Level:       eventmessages.Level_INFO,
			MessageText: message,
		},
	}
}

func NodeInfoMessageF(messageCode eventmessages.NodeMessageCode, format string, values ...interface{}) *eventinput.NodeMessageEvent {
	return NodeInfoMessage(messageCode, fmt.Sprintf(format, values...))
}

func NodeErrorMessage(messageCode eventmessages.NodeMessageCode, message string, values interface{}) *eventinput.NodeMessageEvent {
	errorMsg := fmt.Sprintf(message, values)
	event := &eventinput.NodeMessageEvent{
		NodeMessage: &eventmessages.NodeMessage{
			MessageCode: messageCode,
			Level:       eventmessages.Level_ERROR,
			MessageText: errorMsg,
		},
	}
	log.Errorln(errorMsg)
	return event
}
