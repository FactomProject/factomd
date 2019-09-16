package eventservices

import (
	"encoding/binary"
	"errors"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/events"
	"github.com/FactomProject/factomd/events/contentfiltermode"
	"github.com/FactomProject/factomd/events/eventmessages/generated/eventmessages"
	graphqlproto_types "github.com/bi-foundation/protobuf-graphql-extension/graphqlproto/types"
	"time"
)

type EventMapper interface {
	MapToFactomEvent(eventInput events.EventInput) (*eventmessages.FactomEvent, error)
}

func MapToFactomEvent(eventInput events.EventInput) (*eventmessages.FactomEvent, error) {
	switch eventInput.(type) {
	case *events.RegistrationEvent:
		registrationEvent := eventInput.(*events.RegistrationEvent)
		return mapRegistrationEvent(registrationEvent)
	case *events.StateChangeEvent:
		stateChangeEvent := eventInput.(*events.StateChangeEvent)
		return mapStateChangeEvent(stateChangeEvent)
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

func mapRegistrationEvent(registrationEvent *events.RegistrationEvent) (*eventmessages.FactomEvent, error) {
	event := &eventmessages.FactomEvent{}
	event.StreamSource = registrationEvent.GetStreamSource()
	msg := registrationEvent.GetPayload()
	if msg != nil {
		shouldIncludeContent := eventServiceControl.GetContentFilterMode() > contentfiltermode.SendNever

		switch msg.(type) {
		case *messages.CommitChainMsg:
			event.Value = mapCommitChain(eventmessages.EntityState_REQUESTED, msg)
		case *messages.CommitEntryMsg:
			event.Value = mapCommitEntryEvent(eventmessages.EntityState_REQUESTED, msg)
		case *messages.RevealEntryMsg:
			if shouldIncludeContent {
				event.Value = mapRevealEntryEvent(eventmessages.EntityState_REQUESTED, msg)
			}
		default:
			return nil, errors.New("unknown message type")
		}
	}
	return event, nil
}

func mapStateChangeEvent(stateChangeEvent *events.StateChangeEvent) (*eventmessages.FactomEvent, error) {
	event := &eventmessages.FactomEvent{}
	event.StreamSource = stateChangeEvent.GetStreamSource()
	msg := stateChangeEvent.GetPayload()
	if msg != nil {
		shouldIncludeContent := eventServiceControl.GetContentFilterMode() > contentfiltermode.SendOnRegistration
		resendRegistrations := eventServiceControl.IsResendRegistrationsOnStateChange()
		switch msg.(type) {
		case *messages.CommitChainMsg:
			if resendRegistrations {
				event.Value = mapCommitChain(stateChangeEvent.GetEntityState(), msg)
			} else {
				event.Value = mapCommitChainState(stateChangeEvent.GetEntityState(), msg)
			}
		case *messages.CommitEntryMsg:
			if resendRegistrations {
				event.Value = mapCommitEntryEvent(stateChangeEvent.GetEntityState(), msg)
			} else {
				event.Value = mapCommitEntryEventState(stateChangeEvent.GetEntityState(), msg)
			}
		case *messages.RevealEntryMsg:
			if resendRegistrations {
				event.Value = mapRevealEntryEvent(stateChangeEvent.GetEntityState(), msg)
			} else {
				event.Value = mapRevealEntryEventState(stateChangeEvent.GetEntityState(), msg)
			}
		case *messages.DBStateMsg:
			event.Value = mapDBState(msg, shouldIncludeContent)
		default:
			return nil, errors.New("unknown message type")
		}
	}
	return event, nil
}

func mapProcessMessageEvent(processMessageEvent *events.ProcessMessageEvent) (*eventmessages.FactomEvent, error) {
	event := &eventmessages.FactomEvent{
		StreamSource: processMessageEvent.GetStreamSource(),
		Value: &eventmessages.FactomEvent_ProcessMessage{
			ProcessMessage: processMessageEvent.GetProcessMessage(),
		},
	}
	return event, nil
}

func mapNodeMessageEvent(nodeMessageEvent *events.NodeMessageEvent) (*eventmessages.FactomEvent, error) {
	event := &eventmessages.FactomEvent{
		StreamSource: nodeMessageEvent.GetStreamSource(),
		Value: &eventmessages.FactomEvent_NodeMessage{
			NodeMessage: nodeMessageEvent.GetNodeMessage(),
		},
	}
	return event, nil
}

func mapDBState(msg interfaces.IMsg, shouldIncludeContent bool) *eventmessages.FactomEvent_DirectoryBlockCommit {
	dbStateMessage := msg.(*messages.DBStateMsg)
	event := &eventmessages.FactomEvent_DirectoryBlockCommit{DirectoryBlockCommit: &eventmessages.DirectoryBlockCommit{
		DirectoryBlock:    mapDirBlock(dbStateMessage.DirectoryBlock),
		FactoidBlock:      mapFactoidBlock(dbStateMessage.FactoidBlock),
		EntryBlocks:       mapEntryBlocks(dbStateMessage.EBlocks),
		EntryBlockEntries: mapEntryBlockEntries(dbStateMessage.Entries, shouldIncludeContent),
	}}
	return event
}

func convertByteSlice6ToTimestamp(milliTime *primitives.ByteSlice6) *graphqlproto_types.Timestamp {
	// TODO Is there an easier way to do this?
	slice8 := make([]byte, 8)
	copy(slice8[2:], milliTime[:])
	millis := int64(binary.BigEndian.Uint64(slice8))
	t := time.Unix(0, millis*1000000)
	return convertTimeToTimestamp(t)
}

func convertTimeToTimestamp(time time.Time) *graphqlproto_types.Timestamp {
	return &graphqlproto_types.Timestamp{Seconds: int64(time.Second()), Nanos: int32(time.Nanosecond())}
}
