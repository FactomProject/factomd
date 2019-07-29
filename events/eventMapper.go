package events

import (
	"encoding/binary"
	"errors"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/events/eventmessages"
	eventsinput "github.com/FactomProject/factomd/events/eventmessages/input"
	"github.com/gogo/protobuf/types"
	"time"
)

type EventMapper interface {
	MapToFactomEvent(eventInput *eventsinput.EventInput) (*eventmessages.FactomEvent, error)
}

func MapToFactomEvent(eventInput *eventsinput.EventInput) (*eventmessages.FactomEvent, error) {
	if eventInput.GetMessagePayload() != nil {
		return msgToFactomEvent(eventInput.GetEventSource(), eventInput.GetMessagePayload())
	} else if len(eventInput.GetMessage()) > 0 {
		return stringToFactomEvent(eventInput.GetEventSource(), eventInput.GetMessage())
	} else {
		return nil, errors.New("no payload found in source event")
	}

}

func msgToFactomEvent(eventSource eventmessages.EventSource, msg interfaces.IMsg) (*eventmessages.FactomEvent, error) {
	event := &eventmessages.FactomEvent{}
	event.EventSource = eventSource
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

func stringToFactomEvent(eventSource eventmessages.EventSource, message string) (*eventmessages.FactomEvent, error) {
	event := &eventmessages.FactomEvent{
		EventSource: eventSource,
		Value:       &eventmessages.FactomEvent_NodeMessage{NodeMessage: message},
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
