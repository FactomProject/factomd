package eventservices

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/modules/events/eventconfig"
	"github.com/FactomProject/factomd/modules/events/eventinput"
	"github.com/FactomProject/factomd/modules/events/eventmessages/generated/eventmessages"
	"github.com/gogo/protobuf/types"
)

func MapToFactomEvent(eventInput eventinput.EventInput, broadcastContent eventconfig.BroadcastContent, sendStateChangeEvents bool) (*eventmessages.FactomEvent, error) {
	switch eventInput.(type) {
	case *eventinput.RegistrationEvent:
		registrationEvent := eventInput.(*eventinput.RegistrationEvent)
		return mapRegistrationEvent(registrationEvent, broadcastContent)
	case *eventinput.StateChangeEvent:
		stateChangeEvent := eventInput.(*eventinput.StateChangeEvent)
		return mapStateChangeEvent(stateChangeEvent, broadcastContent, sendStateChangeEvents)
	case *eventinput.DirectoryBlockEvent:
		directoryBlockEvent := eventInput.(*eventinput.DirectoryBlockEvent)
		return mapDirectoryBlockEvent(directoryBlockEvent, broadcastContent)
	case *eventinput.ReplayDirectoryBlockEvent:
		directoryBlockEvent := eventInput.(*eventinput.ReplayDirectoryBlockEvent)
		return mapReplayDirectoryBlockEvent(directoryBlockEvent, broadcastContent)
	case *eventinput.AnchorEvent:
		anchorEvent := eventInput.(*eventinput.AnchorEvent)
		return mapAnchorEvent(anchorEvent)
	case *eventinput.ProcessListEvent:
		processMessageEvent := eventInput.(*eventinput.ProcessListEvent)
		return mapProcessMessageEvent(processMessageEvent)
	case *eventinput.NodeMessageEvent:
		nodeMessageEvent := eventInput.(*eventinput.NodeMessageEvent)
		return mapNodeMessageEvent(nodeMessageEvent)
	default:
		return nil, errors.New("no payload found in source event")
	}
}

func mapRegistrationEvent(registrationEvent *eventinput.RegistrationEvent, broadcastContent eventconfig.BroadcastContent) (*eventmessages.FactomEvent, error) {
	event := &eventmessages.FactomEvent{}
	event.EventSource = registrationEvent.GetStreamSource()
	msg := registrationEvent.GetPayload()
	if msg != nil {
		shouldIncludeContent := broadcastContent > eventconfig.BroadcastNever

		switch msg.(type) {
		case *messages.CommitChainMsg:
			commitChainMsg := msg.(*messages.CommitChainMsg)
			event.Event = mapCommitChain(eventmessages.EntityState_REQUESTED, commitChainMsg)
		case *messages.CommitEntryMsg:
			commitEntryMsg := msg.(*messages.CommitEntryMsg)
			event.Event = mapCommitEntryEvent(eventmessages.EntityState_REQUESTED, commitEntryMsg)
		case *messages.RevealEntryMsg:
			revealEntryMsg := msg.(*messages.RevealEntryMsg)
			if shouldIncludeContent {
				event.Event = mapRevealEntryEvent(eventmessages.EntityState_REQUESTED, revealEntryMsg)
			} else {
				return nil, nil
			}
		default:
			return nil, errors.New("unknown message type")
		}
	}
	return event, nil
}

func mapStateChangeEvent(stateChangeEvent *eventinput.StateChangeEvent, broadcastContent eventconfig.BroadcastContent, sendStateChangeEvents bool) (*eventmessages.FactomEvent, error) {
	event := &eventmessages.FactomEvent{}
	event.EventSource = stateChangeEvent.GetStreamSource()
	msg := stateChangeEvent.GetPayload()
	if msg != nil {
		shouldIncludeContent := broadcastContent > eventconfig.BroadcastOnce

		switch msg.(type) {
		case *messages.CommitChainMsg:
			commitChainMsg := msg.(*messages.CommitChainMsg)
			if sendStateChangeEvents {
				event.Event = mapCommitChainState(stateChangeEvent.GetEntityState(), commitChainMsg)
			} else {
				event.Event = mapCommitChain(stateChangeEvent.GetEntityState(), commitChainMsg)
			}
		case *messages.CommitEntryMsg:
			commitEntryMsg := msg.(*messages.CommitEntryMsg)
			if sendStateChangeEvents {
				event.Event = mapCommitEntryEventState(stateChangeEvent.GetEntityState(), commitEntryMsg)
			} else {
				event.Event = mapCommitEntryEvent(stateChangeEvent.GetEntityState(), commitEntryMsg)
			}
		case *messages.RevealEntryMsg:
			revealEntryMsg := msg.(*messages.RevealEntryMsg)
			if sendStateChangeEvents {
				event.Event = mapRevealEntryEventState(stateChangeEvent.GetEntityState(), revealEntryMsg)
			} else if shouldIncludeContent {
				event.Event = mapRevealEntryEvent(stateChangeEvent.GetEntityState(), revealEntryMsg)
			}
		default:
			return nil, errors.New("unknown message type")
		}
	}
	return event, nil
}

func mapReplayDirectoryBlockEvent(directoryBlockEvent *eventinput.ReplayDirectoryBlockEvent, broadcastContent eventconfig.BroadcastContent) (*eventmessages.FactomEvent, error) {
	msg := directoryBlockEvent.GetPayload()
	dbStateMessage, ok := msg.(*messages.DBStateMsg)
	if !ok {
		return nil, fmt.Errorf("unknown message type of replay directory block event: %v", directoryBlockEvent)
	}

	shouldIncludeContent := broadcastContent > eventconfig.BroadcastOnce
	event := &eventmessages.FactomEvent{}
	event.EventSource = directoryBlockEvent.GetStreamSource()
	event.Event = mapDBStateFromMsg(dbStateMessage, shouldIncludeContent)
	return event, nil
}

func mapDirectoryBlockEvent(directoryBlockEvent *eventinput.DirectoryBlockEvent, broadcastContent eventconfig.BroadcastContent) (*eventmessages.FactomEvent, error) {
	event := &eventmessages.FactomEvent{}
	event.EventSource = directoryBlockEvent.GetStreamSource()
	directoryBlock := directoryBlockEvent.GetPayload()

	if directoryBlock != nil {
		shouldIncludeContent := broadcastContent > eventconfig.BroadcastOnce
		event.Event = mapDirectoryBlockState(directoryBlock, shouldIncludeContent)
	}
	return event, nil
}

func mapAnchorEvent(anchorEvent *eventinput.AnchorEvent) (*eventmessages.FactomEvent, error) {
	event := &eventmessages.FactomEvent{}
	event.EventSource = anchorEvent.GetStreamSource()
	dirBlockInfo := anchorEvent.GetPayload()
	if dirBlockInfo != nil {
		event.Event = mapDirectoryBlockInfo(dirBlockInfo)
	}
	return event, nil
}

func mapProcessMessageEvent(processMessageEvent *eventinput.ProcessListEvent) (*eventmessages.FactomEvent, error) {
	event := &eventmessages.FactomEvent{
		EventSource: processMessageEvent.GetStreamSource(),
		Event: &eventmessages.FactomEvent_ProcessListEvent{
			ProcessListEvent: processMessageEvent.GetProcessListEvent(),
		},
	}
	return event, nil
}

func mapNodeMessageEvent(nodeMessageEvent *eventinput.NodeMessageEvent) (*eventmessages.FactomEvent, error) {
	event := &eventmessages.FactomEvent{
		EventSource: nodeMessageEvent.GetStreamSource(),
		Event: &eventmessages.FactomEvent_NodeMessage{
			NodeMessage: nodeMessageEvent.GetNodeMessage(),
		},
	}
	return event, nil
}

func mapDBStateFromMsg(dbStateMessage *messages.DBStateMsg, shouldIncludeContent bool) *eventmessages.FactomEvent_DirectoryBlockCommit {
	event := &eventmessages.FactomEvent_DirectoryBlockCommit{DirectoryBlockCommit: &eventmessages.DirectoryBlockCommit{
		DirectoryBlock:    mapDirectoryBlock(dbStateMessage.DirectoryBlock),
		AdminBlock:        MapAdminBlock(dbStateMessage.AdminBlock),
		FactoidBlock:      mapFactoidBlock(dbStateMessage.FactoidBlock),
		EntryCreditBlock:  mapEntryCreditBlock(dbStateMessage.EntryCreditBlock),
		EntryBlocks:       mapEntryBlocks(dbStateMessage.EBlocks),
		EntryBlockEntries: mapEntryBlockEntries(dbStateMessage.Entries, shouldIncludeContent),
	}}
	return event
}

func mapDirectoryBlockState(dbState interfaces.IDBState, shouldIncludeContent bool) *eventmessages.FactomEvent_DirectoryBlockCommit {
	event := &eventmessages.FactomEvent_DirectoryBlockCommit{DirectoryBlockCommit: &eventmessages.DirectoryBlockCommit{
		DirectoryBlock:    mapDirectoryBlock(dbState.GetDirectoryBlock()),
		AdminBlock:        MapAdminBlock(dbState.GetAdminBlock()),
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
	return ConvertTimeToTimestamp(t)
}

func ConvertTimeToTimestamp(t time.Time) *types.Timestamp {
	return &types.Timestamp{Seconds: t.Unix(), Nanos: int32(t.Nanosecond())}
}
