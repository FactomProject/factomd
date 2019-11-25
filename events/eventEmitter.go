package events

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/events/eventinput"
	"github.com/FactomProject/factomd/events/eventmessages/generated/eventmessages"
)

type eventEmitter struct {
	parentState   StateEventServices
	eventsService EventService
}

func NewEventEmitter() *eventEmitter {
	return new(eventEmitter)
}

func (eventEmitter *eventEmitter) EmitRegistrationEvent(msg interfaces.IMsg) {
	if eventEmitter.eventsService != nil {
		switch msg.(type) { // Do not fill the channel with message we don't need (like EOM's)
		case *messages.CommitChainMsg, *messages.CommitEntryMsg, *messages.RevealEntryMsg:
			event := eventinput.NewRegistrationEvent(eventEmitter.GetStreamSource(), msg)
			eventEmitter.eventsService.Send(event)
		}
	}
}

func (eventEmitter *eventEmitter) EmitStateChangeEvent(msg interfaces.IMsg, entityState eventmessages.EntityState) {
	eventEmitter.emitStateChangeEvent(msg, entityState, eventEmitter.GetStreamSource())
}

func (eventEmitter *eventEmitter) EmitDBStateEvent(dbState interfaces.IDBState, entityState eventmessages.EntityState) {
	if eventEmitter.eventsService != nil {
		event := eventinput.NewStateChangeEvent(eventEmitter.GetStreamSource(), entityState, dbState)
		eventEmitter.eventsService.Send(event)
	}
}

func (eventEmitter *eventEmitter) EmitReplayStateChangeEvent(msg interfaces.IMsg, entityState eventmessages.EntityState) {
	eventEmitter.emitStateChangeEvent(msg, entityState, eventmessages.EventSource_REPLAY_BOOT)
}

func (eventEmitter *eventEmitter) emitStateChangeEvent(msg interfaces.IMsg, entityState eventmessages.EntityState, streamSource eventmessages.EventSource) {
	if eventEmitter.eventsService != nil {
		switch msg.(type) {
		case *messages.CommitChainMsg, *messages.CommitEntryMsg, *messages.RevealEntryMsg, *messages.DBStateMsg:
			event := eventinput.NewStateChangeEventFromMsg(streamSource, entityState, msg)
			eventEmitter.eventsService.Send(event)
		}
	}
}

func (eventEmitter *eventEmitter) EmitDBAnchorEvent(dirBlockInfo interfaces.IDirBlockInfo) {
	if eventEmitter.eventsService != nil {
		event := eventinput.NewAnchorEvent(eventEmitter.GetStreamSource(), dirBlockInfo)
		eventEmitter.eventsService.Send(event)
	}
}

func (eventEmitter *eventEmitter) EmitProcessListEventNewBlock(newBlockHeight uint32) {
	if eventEmitter.eventsService != nil {
		event := eventinput.ProcessListEventNewBlock(eventEmitter.GetStreamSource(), newBlockHeight)
		eventEmitter.eventsService.Send(event)
	}
}

func (eventEmitter *eventEmitter) EmitProcessListEventNewMinute(newMinute int, blockHeight uint32) {
	if eventEmitter.eventsService != nil {
		event := eventinput.ProcessListEventNewMinute(eventEmitter.GetStreamSource(), newMinute, blockHeight)
		eventEmitter.eventsService.Send(event)
	}
}

func (eventEmitter *eventEmitter) EmitNodeInfoMessage(messageCode eventmessages.NodeMessageCode, message string) {
	if eventEmitter.eventsService != nil {
		event := eventinput.NodeInfoMessageF(messageCode, message)
		eventEmitter.eventsService.Send(event)
	}
}

func (eventEmitter *eventEmitter) EmitNodeInfoMessageF(messageCode eventmessages.NodeMessageCode, format string, values ...interface{}) {
	if eventEmitter.eventsService != nil {
		event := eventinput.NodeInfoMessageF(messageCode, format, values...)
		eventEmitter.eventsService.Send(event)
	}
}

func (eventEmitter *eventEmitter) EmitNodeErrorMessage(messageCode eventmessages.NodeMessageCode, message string, values interface{}) {
	if eventEmitter.eventsService != nil {
		event := eventinput.NodeErrorMessage(messageCode, message, values)
		eventEmitter.eventsService.Send(event)
	}
}

func (eventEmitter *eventEmitter) GetStreamSource() eventmessages.EventSource {
	if eventEmitter.parentState.IsRunLeader() {
		return eventmessages.EventSource_LIVE
	} else {
		return eventmessages.EventSource_REPLAY_BOOT
	}
}

func AttachEventServiceToEventEmitter(parentState StateEventServices, eventsService EventService) {
	eventEmitter := parentState.GetEvents().(*eventEmitter)
	eventEmitter.parentState = parentState
	eventEmitter.eventsService = eventsService
}
