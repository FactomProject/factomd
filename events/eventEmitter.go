package events

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants/runstate"
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/events/eventinput"
	"github.com/FactomProject/factomd/events/eventmessages/generated/eventmessages"
	"github.com/FactomProject/factomd/events/eventservices"
	"github.com/FactomProject/factomd/util"
)

type EventService interface {
	ConfigService(state StateEventServices, config *util.FactomdConfig, factomParams *globals.FactomParams)
	ConfigSender(state StateEventServices, sender eventservices.EventSender)
	EmitRegistrationEvent(msg interfaces.IMsg)
	EmitStateChangeEvent(msg interfaces.IMsg, entityState eventmessages.EntityState)
	EmitDirectoryBlockCommitEvent(dbState interfaces.IDBState)
	EmitDirectoryBlockAnchorEvent(dirBlockInfo interfaces.IDirBlockInfo)
	EmitReplayDirectoryBlockCommit(msg interfaces.IMsg)
	EmitProcessListEventNewBlock(newBlockHeight uint32)
	EmitProcessListEventNewMinute(newMinute int, blockHeight uint32)
	EmitNodeInfoMessage(messageCode eventmessages.NodeMessageCode, message string)
	EmitNodeInfoMessageF(messageCode eventmessages.NodeMessageCode, format string, values ...interface{})
	EmitNodeErrorMessage(messageCode eventmessages.NodeMessageCode, message string, values interface{})
}

type eventEmitter struct {
	parentState StateEventServices
	eventSender eventservices.EventSender
}

func NewEventService() EventService {
	return new(eventEmitter)
}

func (eventEmitter *eventEmitter) ConfigService(state StateEventServices, config *util.FactomdConfig, factomParams *globals.FactomParams) {
	eventEmitter.parentState = state
	eventEmitter.eventSender = eventservices.NewEventSender(config, factomParams)
}

func (eventEmitter *eventEmitter) ConfigSender(state StateEventServices, eventSender eventservices.EventSender) {
	eventEmitter.parentState = state
	eventEmitter.eventSender = eventSender
}

func (eventEmitter *eventEmitter) Send(event eventinput.EventInput) error {
	if eventEmitter.parentState.GetRunState() > runstate.Running { // Stop queuing messages to the events channel when shutting down
		return nil
	}

	// Only send info messages when EventReplayDuringStartup is disabled
	if !eventEmitter.eventSender.ReplayDuringStartup() && !eventEmitter.parentState.IsRunLeader() {
		switch event.(type) {
		case *eventinput.ProcessListEvent:
		case *eventinput.NodeMessageEvent:
		default:
			return nil
		}
	}

	broadcastContent := eventEmitter.eventSender.GetBroadcastContent()
	sendStateChangeEvents := eventEmitter.eventSender.IsSendStateChangeEvents()
	factomEvent, err := eventservices.MapToFactomEvent(event, broadcastContent, sendStateChangeEvents)
	if err != nil {
		return fmt.Errorf("failed to map to factom event: %v\n", err)
	}
	if factomEvent == nil {
		return nil
	}

	factomEvent.IdentityChainID = eventEmitter.parentState.GetIdentityChainID().Bytes()
	select {
	case eventEmitter.eventSender.GetEventQueue() <- factomEvent:
	default:
		eventEmitter.eventSender.IncreaseDroppedFromQueueCounter()
	}
	return nil
}

func (eventEmitter *eventEmitter) EmitRegistrationEvent(msg interfaces.IMsg) {
	if eventEmitter.eventSender != nil {
		switch msg.(type) { // Do not fill the channel with message we don't need (like EOM's)
		case *messages.CommitChainMsg, *messages.CommitEntryMsg, *messages.RevealEntryMsg:
			event := eventinput.NewRegistrationEvent(eventEmitter.GetStreamSource(), msg)
			eventEmitter.Send(event)
		}
	}
}

func (eventEmitter *eventEmitter) EmitStateChangeEvent(msg interfaces.IMsg, entityState eventmessages.EntityState) {
	if eventEmitter.eventSender != nil {
		switch msg.(type) {
		case *messages.CommitChainMsg, *messages.CommitEntryMsg, *messages.RevealEntryMsg, *messages.DBStateMsg:
			event := eventinput.NewStateChangeEvent(eventEmitter.GetStreamSource(), entityState, msg)
			eventEmitter.Send(event)
		}
	}
}

func (eventEmitter *eventEmitter) EmitDirectoryBlockCommitEvent(dbState interfaces.IDBState) {
	if eventEmitter.eventSender != nil {
		event := eventinput.NewDirectoryBlockEvent(eventEmitter.GetStreamSource(), dbState)
		eventEmitter.Send(event)
	}
}

func (eventEmitter *eventEmitter) EmitDirectoryBlockAnchorEvent(dirBlockInfo interfaces.IDirBlockInfo) {
	if eventEmitter.eventSender != nil {
		event := eventinput.NewAnchorEvent(eventEmitter.GetStreamSource(), dirBlockInfo)
		eventEmitter.Send(event)
	}
}

func (eventEmitter *eventEmitter) EmitReplayDirectoryBlockCommit(msg interfaces.IMsg) {
	if eventEmitter.eventSender != nil {
		event := eventinput.NewReplayDirectoryBlockEvent(eventmessages.EventSource_REPLAY_BOOT, msg)
		eventEmitter.Send(event)
	}
}

func (eventEmitter *eventEmitter) EmitProcessListEventNewBlock(newBlockHeight uint32) {
	if eventEmitter.eventSender != nil {
		event := eventinput.ProcessListEventNewBlock(eventEmitter.GetStreamSource(), newBlockHeight)
		eventEmitter.Send(event)
	}
}

func (eventEmitter *eventEmitter) EmitProcessListEventNewMinute(newMinute int, blockHeight uint32) {
	if eventEmitter.eventSender != nil {
		event := eventinput.ProcessListEventNewMinute(eventEmitter.GetStreamSource(), newMinute, blockHeight)
		eventEmitter.Send(event)
	}
}

func (eventEmitter *eventEmitter) EmitNodeInfoMessage(messageCode eventmessages.NodeMessageCode, message string) {
	if eventEmitter.eventSender != nil {
		event := eventinput.NodeInfoMessageF(messageCode, message)
		eventEmitter.Send(event)
	}
}

func (eventEmitter *eventEmitter) EmitNodeInfoMessageF(messageCode eventmessages.NodeMessageCode, format string, values ...interface{}) {
	if eventEmitter.eventSender != nil {
		event := eventinput.NodeInfoMessageF(messageCode, format, values...)
		eventEmitter.Send(event)
	}
}

func (eventEmitter *eventEmitter) EmitNodeErrorMessage(messageCode eventmessages.NodeMessageCode, message string, values interface{}) {
	if eventEmitter.eventSender != nil {
		event := eventinput.NodeErrorMessage(messageCode, message, values)
		eventEmitter.Send(event)
	}
}

func (eventEmitter *eventEmitter) GetStreamSource() eventmessages.EventSource {
	if eventEmitter.parentState == nil {
		return -1
	}

	if eventEmitter.parentState.IsRunLeader() {
		return eventmessages.EventSource_LIVE
	} else {
		return eventmessages.EventSource_REPLAY_BOOT
	}
}
