package eventservices

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/events/eventmessages"
)

func mapCommitEntryEvent(entityState eventmessages.EntityState, msg interfaces.IMsg) *eventmessages.FactomEvent_EntryRegistration {
	commitEntry := msg.(*messages.CommitEntryMsg).CommitEntry
	ecPubKey := commitEntry.ECPubKey.Fixed()
	sig := commitEntry.Sig

	result := &eventmessages.FactomEvent_EntryRegistration{
		EntryRegistration: &eventmessages.EntryRegistration{
			EntityState: entityState,
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

func mapRevealEntryEvent(entityState eventmessages.EntityState, msg interfaces.IMsg, shouldIncludeContent bool) *eventmessages.FactomEvent_EntryContentRegistration {
	revealEntry := msg.(*messages.RevealEntryMsg)
	return &eventmessages.FactomEvent_EntryContentRegistration{
		EntryContentRegistration: &eventmessages.EntryContentRegistration{
			EntityState: entityState,
			Entry:       mapEntryBlockEntry(revealEntry.Entry, shouldIncludeContent),
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
		Hash: &eventmessages.Hash{HashValue: entry.GetHash().Bytes()},
	}
	if shouldIncludeContent {
		blockEntry.ExternalIDs = mapExternalIds(entry.ExternalIDs())
		blockEntry.Content = &eventmessages.Content{BinaryValue: entry.GetContent()}
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
