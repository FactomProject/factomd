package eventservices

import (
	"encoding/binary"
	"errors"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/events"
	"github.com/FactomProject/factomd/events/eventmessages"
	"github.com/gogo/protobuf/types"
	"time"
)

type EventMapper interface {
	MapToFactomEvent(eventInput events.EventInput) (*eventmessages.FactomEvent, error)
}

func MapToFactomEvent(eventInput events.EventInput) (*eventmessages.FactomEvent, error) {
	switch eventInput.(type) {
	case *events.ProcessEvent:
		processEvent := eventInput.(*events.ProcessEvent)
		return mapProcessEvent(processEvent)
	case *events.NodeEvent:
		nodeEvent := eventInput.(*events.NodeEvent)
		return mapNodeEvent(nodeEvent)
	default:
		return nil, errors.New("no payload found in source event")
	}
}

func mapProcessEvent(processEvent *events.ProcessEvent) (*eventmessages.FactomEvent, error) {
	event := &eventmessages.FactomEvent{}
	event.EventSource = processEvent.GetEventSource()
	msg := processEvent.GetPayload()
	switch msg.(type) {
	case *messages.DBStateMsg:
		event.Value = mapDBState(msg.(*messages.DBStateMsg))
	case *messages.CommitChainMsg:
		event.Value = mapCommitChain(msg)
	case *messages.CommitEntryMsg:
		event.Value = mapCommitEntryEvent(msg)
	case *messages.RevealEntryMsg:
		event.Value = mapRevealEntryEvent(msg)
	default:
		return nil, errors.New("unknown message type")
	}
	return event, nil
}

func mapNodeEvent(nodeEvent *events.NodeEvent) (*eventmessages.FactomEvent, error) {
	event := &eventmessages.FactomEvent{
		EventSource: nodeEvent.GetEventSource(),
		Value:       &eventmessages.FactomEvent_Message{Message: nodeEvent.GetPayload()},
	}
	return event, nil
}

func mapDBState(dbStateMessage *messages.DBStateMsg) *eventmessages.FactomEvent_AnchorEvent {
	event := &eventmessages.FactomEvent_AnchorEvent{AnchorEvent: &eventmessages.AnchoredEvent{
		DirectoryBlock:    mapDirBlock(dbStateMessage.DirectoryBlock),
		FactoidBlock:      mapFactoidBlock(dbStateMessage.FactoidBlock),
		EntryBlocks:       mapEntryBlocks(dbStateMessage.EBlocks),
		EntryBlockEntries: mapEntryBlockEntries(dbStateMessage.Entries),
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
