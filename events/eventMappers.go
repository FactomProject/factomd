package events

import (
	"encoding/binary"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/messages/eventmessages"
	eventinput "github.com/FactomProject/factomd/common/messages/eventmessages/input"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/gogo/protobuf/types"
	"time"
)

type EventMapper interface {
	MapToFactomEvent(eventInput eventinput.EventInput) *eventmessages.FactomEvent
}

func MapToFactomEvent(eventInput eventinput.EventInput) *eventmessages.FactomEvent {
	if eventInput.GetMessagePayload() != nil {
		return msgToFactomEvent(eventInput.GetEventSource(), eventInput.GetMessagePayload())
	}
	panic("No payload found in source event.")
}

func msgToFactomEvent(eventSource eventmessages.EventSource, msg interfaces.IMsg) *eventmessages.FactomEvent {
	event := &eventmessages.FactomEvent{}
	event.EventSource = eventSource
	switch msg.(type) {
	case *messages.DBStateMsg:
		event.Value = mapDBState(msg.(*messages.DBStateMsg))
	case *messages.CommitChainMsg:
		event.Value = mapCommitChain(msg)
	case *messages.CommitEntryMsg:
		event.Value = mapCommitEvent(msg)
	case *messages.RevealEntryMsg:
		event.Value = mapRevealEntryEvent(msg)
	default:
		return nil
	}
	return event
}

func mapDBState(dbStateMessage *messages.DBStateMsg) *eventmessages.FactomEvent_AnchorEvent {
	event := &eventmessages.FactomEvent_AnchorEvent{AnchorEvent: &eventmessages.AnchoredEvent{
		DirectoryBlock:    mapDirBlock(dbStateMessage.DirectoryBlock),
		EntryBlocks:       mapEntryBlocks(dbStateMessage.EBlocks),
		EntryBlockEntries: mapEntryBlockEntries(dbStateMessage.Entries),
	}}
	return event
}

func mapDirBlock(block interfaces.IDirectoryBlock) *eventmessages.DirectoryBlock {
	result := &eventmessages.DirectoryBlock{}
	result.Header = mapDirHeader(block.GetHeader())
	result.Entries = mapDirEntries(block.GetDBEntries())
	return result
}

func mapDirHeader(header interfaces.IDirectoryBlockHeader) *eventmessages.DirectoryBlockHeader {
	result := &eventmessages.DirectoryBlockHeader{
		BodyMerkleRoot: &eventmessages.Hash{
			HashValue: header.GetBodyMR().Bytes(),
		},
		PreviousKeyMerkleRoot: &eventmessages.Hash{
			HashValue: header.GetPrevKeyMR().Bytes(),
		},
		PreviousFullHash: &eventmessages.Hash{
			HashValue: header.GetPrevFullHash().Bytes(),
		},
		Timestamp:  convertTimeToTimestamp(header.GetTimestamp().GetTime()),
		DbHeight:   header.GetDBHeight(),
		BlockCount: header.GetBlockCount(),
	}
	return result
}

func mapDirEntries(entries []interfaces.IDBEntry) []*eventmessages.Entry {
	result := make([]*eventmessages.Entry, len(entries))
	for i, entry := range entries {
		result[i] = mapDirEntry(entry)
	}
	return result
}

func mapDirEntry(entry interfaces.IDBEntry) *eventmessages.Entry {
	result := &eventmessages.Entry{
		ChainID: &eventmessages.Hash{
			HashValue: entry.GetChainID().Bytes(),
		},
		KeyMerkleRoot: &eventmessages.Hash{
			HashValue: entry.GetKeyMR().Bytes(),
		},
	}
	return result
}

func mapCommitChain(msg interfaces.IMsg) *eventmessages.FactomEvent_CommitChain {
	commitChain := msg.(*messages.CommitChainMsg).CommitChain
	ecPubKey := commitChain.ECPubKey.Fixed()
	sig := commitChain.Sig

	result := &eventmessages.FactomEvent_CommitChain{
		CommitChain: &eventmessages.CommitChain{
			ChainIDHash: &eventmessages.Hash{
				HashValue: commitChain.ChainIDHash.Bytes()},
			EntryHash: &eventmessages.Hash{
				HashValue: commitChain.EntryHash.Bytes(),
			},
			Timestamp: convertByteSlice6ToTimestamp(commitChain.MilliTime),
			Credits:   uint32(commitChain.Credits),
			EcPubKey:  ecPubKey[:],
			Sig:       sig[:],
		}}
	return result
}

func mapCommitEvent(msg interfaces.IMsg) *eventmessages.FactomEvent_CommitEntry {
	commitEntry := msg.(*messages.CommitEntryMsg).CommitEntry
	ecPubKey := commitEntry.ECPubKey.Fixed()
	sig := commitEntry.Sig

	result := &eventmessages.FactomEvent_CommitEntry{
		CommitEntry: &eventmessages.CommitEntry{
			EntryHash: &eventmessages.Hash{
				HashValue: commitEntry.EntryHash.Bytes(),
			},
			Timestamp: convertByteSlice6ToTimestamp(commitEntry.MilliTime),
			Credits:   uint32(commitEntry.Credits),
			EcPubKey:  ecPubKey[:],
			Sig:       sig[:],
		}}
	return result
}

func mapRevealEntryEvent(msg interfaces.IMsg) *eventmessages.FactomEvent_RevealEntry {
	revealEntry := msg.(*messages.RevealEntryMsg)
	return &eventmessages.FactomEvent_RevealEntry{
		RevealEntry: &eventmessages.RevealEntry{
			Entry:     mapEntryBlockEntry(revealEntry.Entry),
			Timestamp: convertTimeToTimestamp(revealEntry.Timestamp.GetTime()),
		},
	}
}

func mapEntryBlocks(blocks []interfaces.IEntryBlock) []*eventmessages.EntryBlock {
	result := make([]*eventmessages.EntryBlock, len(blocks))
	for i, block := range blocks {
		result[i] = &eventmessages.EntryBlock{
			EntryBlockHeader: mapEntryBlockHeader(block.GetHeader()),
			EntryHashes:      mapEntryBlockHashes(block.GetBody().GetEBEntries()),
		}
	}
	return result
}

func mapEntryBlockHashes(entries []interfaces.IHash) []*eventmessages.Hash {
	result := make([]*eventmessages.Hash, len(entries))
	for i, entry := range entries {
		result[i] = &eventmessages.Hash{
			HashValue: entry.Bytes(),
		}
	}
	return result
}

func mapEntryBlockHeader(header interfaces.IEntryBlockHeader) *eventmessages.EntryBlockHeader {
	return &eventmessages.EntryBlockHeader{
		BodyMerkleRoot:        &eventmessages.Hash{HashValue: header.GetBodyMR().Bytes()},
		ChainID:               &eventmessages.Hash{HashValue: header.GetChainID().Bytes()},
		PreviousFullHash:      &eventmessages.Hash{HashValue: header.GetPrevFullHash().Bytes()},
		PreviousKeyMerkleRoot: &eventmessages.Hash{HashValue: header.GetPrevKeyMR().Bytes()},
		DbHeight:              header.GetDBHeight(),
		BlockSequence:         header.GetEBSequence(),
		EntryCount:            header.GetEntryCount(),
	}
}

func mapEntryBlockEntries(entries []interfaces.IEBEntry) []*eventmessages.EntryBlockEntry {
	result := make([]*eventmessages.EntryBlockEntry, len(entries))
	for i, entry := range entries {
		result[i] = mapEntryBlockEntry(entry)
	}
	return result
}

func mapEntryBlockEntry(entry interfaces.IEBEntry) *eventmessages.EntryBlockEntry {
	return &eventmessages.EntryBlockEntry{
		Hash:        &eventmessages.Hash{HashValue: entry.GetHash().Bytes()},
		ExternalIDs: mapExternalIds(entry.ExternalIDs()),
		Content:     &eventmessages.Content{BinaryValue: entry.GetContent()},
	}
}

func mapExternalIds(externalIds [][]byte) []*eventmessages.ExternalId {
	result := make([]*eventmessages.ExternalId, len(externalIds))
	for i, extId := range externalIds {
		result[i] = &eventmessages.ExternalId{BinaryValue: extId}
	}
	return result
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
