package eventservices

import (
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/events/eventmessages/generated/eventmessages"
)

func mapCommitEntryEvent(entityState eventmessages.EntityState, msg interfaces.IMsg) *eventmessages.FactomEvent_EntryCommit {
	commitEntry := msg.(*messages.CommitEntryMsg).CommitEntry
	ecPubKey := commitEntry.ECPubKey.Fixed()
	sig := commitEntry.Sig

	result := &eventmessages.FactomEvent_EntryCommit{
		EntryCommit: &eventmessages.EntryCommit{
			EntityState: entityState,
			EntryHash: &eventmessages.Hash{
				HashValue: commitEntry.EntryHash.Bytes(),
			},
			Timestamp:            convertByteSlice6ToTimestamp(commitEntry.MilliTime),
			Credits:              uint32(commitEntry.Credits),
			EntryCreditPublicKey: ecPubKey[:],
			Signature:            sig[:],
			Version:              uint32(commitEntry.Version),
		}}
	return result
}

func mapCommitEntryEventState(state eventmessages.EntityState, msg interfaces.IMsg) *eventmessages.FactomEvent_StateChange {
	commitEntry := msg.(*messages.CommitEntryMsg).CommitEntry
	result := &eventmessages.FactomEvent_StateChange{
		StateChange: &eventmessages.StateChange{
			EntityHash: &eventmessages.Hash{
				HashValue: commitEntry.EntryHash.Bytes()},
			EntityState: state,
		},
	}
	return result
}

func mapRevealEntryEvent(entityState eventmessages.EntityState, msg interfaces.IMsg) *eventmessages.FactomEvent_EntryReveal {
	revealEntry := msg.(*messages.RevealEntryMsg)
	return &eventmessages.FactomEvent_EntryReveal{
		EntryReveal: &eventmessages.EntryReveal{
			EntityState: entityState,
			Entry:       mapEntryBlockEntry(revealEntry.Entry, true),
			Timestamp:   convertTimeToTimestamp(revealEntry.Timestamp.GetTime()),
		},
	}
}

func mapRevealEntryEventState(state eventmessages.EntityState, msg interfaces.IMsg) *eventmessages.FactomEvent_StateChange {
	revealEntry := msg.(*messages.RevealEntryMsg)
	result := &eventmessages.FactomEvent_StateChange{
		StateChange: &eventmessages.StateChange{
			EntityHash: &eventmessages.Hash{
				HashValue: revealEntry.Entry.GetHash().Bytes()},
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

func mapEntryBlockEntry(iEntry interfaces.IEBEntry, shouldIncludeContent bool) *eventmessages.EntryBlockEntry {
	entry := iEntry.(*entryBlock.Entry)

	blockEntry := &eventmessages.EntryBlockEntry{
		Hash: &eventmessages.Hash{HashValue: entry.GetHash().Bytes()},
	}
	if shouldIncludeContent {
		blockEntry.ExternalIDs = mapExternalIds(entry.ExternalIDs())
		blockEntry.Content = &eventmessages.Content{BinaryValue: entry.GetContent()}
		blockEntry.Version = uint32(entry.Version)
	}
	return blockEntry
}

func mapExternalIds(externalIds [][]byte) []*eventmessages.ExternalId {
	result := make([]*eventmessages.ExternalId, len(externalIds))
	for i, extId := range externalIds {
		result[i] = &eventmessages.ExternalId{BinaryValue: extId}
	}
	return result
}
