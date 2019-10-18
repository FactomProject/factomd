package eventservices

import (
	"encoding/binary"
	"errors"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/events"
	"github.com/FactomProject/factomd/events/eventmessages/generated/eventmessages"
	"github.com/gogo/protobuf/types"
	"time"
)

func MapToFactomEvent(eventInput events.EventInput, broadcastContent BroadcastContent, sendStateChangeEvents bool) (*eventmessages.FactomEvent, error) {
	switch eventInput.(type) {
	case *events.RegistrationEvent:
		registrationEvent := eventInput.(*events.RegistrationEvent)
		return mapRegistrationEvent(registrationEvent, broadcastContent)
	case *events.StateChangeMsgEvent:
		stateChangeEvent := eventInput.(*events.StateChangeMsgEvent)
		return mapStateChangeEvent(stateChangeEvent, broadcastContent, sendStateChangeEvents)
	case *events.StateChangeEvent:
		stateChangeEvent := eventInput.(*events.StateChangeEvent)
		return mapDBStateEvent(stateChangeEvent, broadcastContent)
	case *events.ProcessMessageEvent:
		processMessageEvent := eventInput.(*events.ProcessMessageEvent)
		return mapProcessMessageEvent(processMessageEvent)
	case *events.NodeMessageEvent:
		nodeMessageEvent := eventInput.(*events.NodeMessageEvent)
		return mapNodeMessageEvent(nodeMessageEvent)
	default:
		return nil, errors.New("no payload found in source event")
	}
}

func mapRegistrationEvent(registrationEvent *events.RegistrationEvent, broadcastContent BroadcastContent) (*eventmessages.FactomEvent, error) {
	event := &eventmessages.FactomEvent{}
	event.EventSource = registrationEvent.GetStreamSource()
	msg := registrationEvent.GetPayload()
	if msg != nil {
		shouldIncludeContent := broadcastContent > BroadcastNever

		switch msg.(type) {
		case *messages.CommitChainMsg:
			event.Value = mapCommitChain(eventmessages.EntityState_REQUESTED, msg)
		case *messages.CommitEntryMsg:
			event.Value = mapCommitEntryEvent(eventmessages.EntityState_REQUESTED, msg)
		case *messages.RevealEntryMsg:
			if shouldIncludeContent {
				event.Value = mapRevealEntryEvent(eventmessages.EntityState_REQUESTED, msg)
			} else {
				return nil, nil
			}
		default:
			return nil, errors.New("unknown message type")
		}
	}
	return event, nil
}

func mapStateChangeEvent(stateChangeEvent *events.StateChangeMsgEvent, broadcastContent BroadcastContent, sendStateChangeEvents bool) (*eventmessages.FactomEvent, error) {
	event := &eventmessages.FactomEvent{}
	event.EventSource = stateChangeEvent.GetStreamSource()
	msg := stateChangeEvent.GetPayload()
	if msg != nil {
		shouldIncludeContent := broadcastContent > BroadcastOnce

		switch msg.(type) {
		case *messages.CommitChainMsg:
			if sendStateChangeEvents {
				event.Value = mapCommitChainState(stateChangeEvent.GetEntityState(), msg)
			} else {
				event.Value = mapCommitChain(stateChangeEvent.GetEntityState(), msg)
			}
		case *messages.CommitEntryMsg:
			if sendStateChangeEvents {
				event.Value = mapCommitEntryEventState(stateChangeEvent.GetEntityState(), msg)
			} else {
				event.Value = mapCommitEntryEvent(stateChangeEvent.GetEntityState(), msg)
			}
		case *messages.RevealEntryMsg:
			if sendStateChangeEvents {
				event.Value = mapRevealEntryEventState(stateChangeEvent.GetEntityState(), msg)
			} else if shouldIncludeContent {
				event.Value = mapRevealEntryEvent(stateChangeEvent.GetEntityState(), msg)
			}
		case *messages.DBStateMsg:
			event.Value = mapDBStateFromMsg(msg, shouldIncludeContent)
		default:
			return nil, errors.New("unknown message type")
		}
	}
	return event, nil
}

func mapDBStateEvent(stateChangeEvent *events.StateChangeEvent, broadcastContent BroadcastContent) (*eventmessages.FactomEvent, error) {
	event := &eventmessages.FactomEvent{}
	event.EventSource = stateChangeEvent.GetStreamSource()
	state := stateChangeEvent.GetPayload()
	stateChangeEvent.GetEntityState()
	if state != nil {
		shouldIncludeContent := broadcastContent > BroadcastOnce
		event.Value = mapDBState(state, shouldIncludeContent)
	}
	return event, nil
}

func mapProcessMessageEvent(processMessageEvent *events.ProcessMessageEvent) (*eventmessages.FactomEvent, error) {
	event := &eventmessages.FactomEvent{
		EventSource: processMessageEvent.GetStreamSource(),
		Value: &eventmessages.FactomEvent_ProcessMessage{
			ProcessMessage: processMessageEvent.GetProcessMessage(),
		},
	}
	return event, nil
}

func mapNodeMessageEvent(nodeMessageEvent *events.NodeMessageEvent) (*eventmessages.FactomEvent, error) {
	event := &eventmessages.FactomEvent{
		EventSource: nodeMessageEvent.GetStreamSource(),
		Value: &eventmessages.FactomEvent_NodeMessage{
			NodeMessage: nodeMessageEvent.GetNodeMessage(),
		},
	}
	return event, nil
}

func mapDBStateFromMsg(msg interfaces.IMsg, shouldIncludeContent bool) *eventmessages.FactomEvent_DirectoryBlockCommit {
	dbStateMessage := msg.(*messages.DBStateMsg)
	event := &eventmessages.FactomEvent_DirectoryBlockCommit{DirectoryBlockCommit: &eventmessages.DirectoryBlockCommit{
		DirectoryBlock:    mapDirectoryBlock(dbStateMessage.DirectoryBlock),
		AdminBlock:        mapAdminBlock(dbStateMessage.AdminBlock),
		FactoidBlock:      mapFactoidBlock(dbStateMessage.FactoidBlock),
		EntryCreditBlock:  mapEntryCreditBlock(dbStateMessage.EntryCreditBlock),
		EntryBlocks:       mapEntryBlocks(dbStateMessage.EBlocks),
		EntryBlockEntries: mapEntryBlockEntries(dbStateMessage.Entries, shouldIncludeContent),
	}}
	return event
}

func mapDBState(dbState interfaces.IDBState, shouldIncludeContent bool) *eventmessages.FactomEvent_DirectoryBlockCommit {
	event := &eventmessages.FactomEvent_DirectoryBlockCommit{DirectoryBlockCommit: &eventmessages.DirectoryBlockCommit{
		DirectoryBlock:    mapDirectoryBlock(dbState.GetDirectoryBlock()),
		AdminBlock:        mapAdminBlock(dbState.GetAdminBlock()),
		FactoidBlock:      mapFactoidBlock(dbState.GetFactoidBlock()),
		EntryCreditBlock:  mapEntryCreditBlock(dbState.GetEntryCreditBlock()),
		EntryBlocks:       mapEntryBlocks(dbState.GetEntryBlocks()),
		EntryBlockEntries: mapEntryBlockEntries(dbState.GetEntries(), shouldIncludeContent),
	}}
	return event
}

func convertByteSlice6ToTimestamp(milliTime *primitives.ByteSlice6) *types.Timestamp {
	// TODO Is there an easier way to do this?
	slice8 := make([]byte, 8)
	copy(slice8[2:], milliTime[:])
	millis := int64(binary.BigEndian.Uint64(slice8))
	t := time.Unix(0, millis*1000000)
	return convertTimeToTimestamp(t)
}

func convertTimeToTimestamp(time time.Time) *types.Timestamp {
	return &types.Timestamp{Seconds: int64(time.Second()), Nanos: int32(time.Nanosecond())}
}
