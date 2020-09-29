package eventservices

import (
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/modules/events/eventmessages/generated/eventmessages"
)

func mapCommitEntryEvent(entityState eventmessages.EntityState, commitEntryMsg *messages.CommitEntryMsg) *eventmessages.FactomEvent_EntryCommit {
	commitEntry := commitEntryMsg.CommitEntry
	ecPubKey := commitEntry.ECPubKey.Fixed()
	sig := commitEntry.Sig

	result := &eventmessages.FactomEvent_EntryCommit{
		EntryCommit: &eventmessages.EntryCommit{
			EntityState:          entityState,
			EntryHash:            commitEntry.EntryHash.Bytes(),
			Timestamp:            convertByteSlice6ToTimestamp(commitEntry.MilliTime),
			Credits:              uint32(commitEntry.Credits),
			EntryCreditPublicKey: ecPubKey[:],
			Signature:            sig[:],
			Version:              uint32(commitEntry.Version),
		}}
	return result
}

func mapCommitEntryEventState(state eventmessages.EntityState, commitEntryMsg *messages.CommitEntryMsg) *eventmessages.FactomEvent_StateChange {
	commitEntry := commitEntryMsg.CommitEntry
	result := &eventmessages.FactomEvent_StateChange{
		StateChange: &eventmessages.StateChange{
			EntityHash:  commitEntry.EntryHash.Bytes(),
			EntityState: state,
		},
	}
	return result
}

func mapRevealEntryEvent(entityState eventmessages.EntityState, revealEntry *messages.RevealEntryMsg) *eventmessages.FactomEvent_EntryReveal {
	return &eventmessages.FactomEvent_EntryReveal{
		EntryReveal: &eventmessages.EntryReveal{
			EntityState: entityState,
			Entry:       mapEntryBlockEntry(revealEntry.Entry, true),
			Timestamp:   ConvertTimeToTimestamp(revealEntry.Timestamp.GetTime()),
		},
	}
}

func mapRevealEntryEventState(state eventmessages.EntityState, revealEntry *messages.RevealEntryMsg) *eventmessages.FactomEvent_StateChange {
	result := &eventmessages.FactomEvent_StateChange{
		StateChange: &eventmessages.StateChange{
			EntityHash:  revealEntry.Entry.GetHash().Bytes(),
			EntityState: state,
		},
	}
	return result
}

func mapEntryBlocks(blocks []interfaces.IEntryBlock) []*eventmessages.EntryBlock {
	result := make([]*eventmessages.EntryBlock, len(blocks))
	for i, block := range blocks {
		result[i] = &eventmessages.EntryBlock{
			Header:      mapEntryBlockHeader(block.GetHeader()),
			EntryHashes: mapEntryBlockHashes(block.GetBody().GetEBEntries()),
		}
	}
	return result
}

func mapEntryBlockHashes(entries []interfaces.IHash) [][]byte {
	result := make([][]byte, len(entries))
	for i, entry := range entries {
		result[i] = entry.Bytes()
	}
	return result
}

func mapEntryBlockHeader(header interfaces.IEntryBlockHeader) *eventmessages.EntryBlockHeader {
	return &eventmessages.EntryBlockHeader{
		BodyMerkleRoot:        header.GetBodyMR().Bytes(),
		ChainID:               header.GetChainID().Bytes(),
		PreviousKeyMerkleRoot: header.GetPrevKeyMR().Bytes(),
		PreviousFullHash:      header.GetPrevFullHash().Bytes(),
		BlockHeight:           header.GetDBHeight(),
		BlockSequence:         header.GetEBSequence(),
		EntryCount:            header.GetEntryCount(),
	}
}

func mapEntryBlockEntries(entries []interfaces.IEBEntry, shouldIncludeContent bool) []*eventmessages.EntryBlockEntry {
	result := make([]*eventmessages.EntryBlockEntry, len(entries))
	for i, entry := range entries {
		result[i] = mapEntryBlockEntry(entry, shouldIncludeContent)
	}
	return result
}

func mapEntryBlockEntry(entry interfaces.IEBEntry, shouldIncludeContent bool) *eventmessages.EntryBlockEntry {
	blockEntry := &eventmessages.EntryBlockEntry{
		Hash:    entry.GetHash().Bytes(),
		ChainID: entry.GetChainIDHash().Bytes(),
	}
	if shouldIncludeContent {
		blockEntry.ExternalIDs = entry.ExternalIDs()
		blockEntry.Content = entry.GetContent()

		if e, ok := entry.(*entryBlock.Entry); ok {
			blockEntry.Version = uint32(e.Version)
		}
	}
	return blockEntry
}
