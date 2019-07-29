package eventservices

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/events/eventmessages"
)

func mapCommitEntryEvent(msg interfaces.IMsg) *eventmessages.FactomEvent_CommitEntry {
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
		BlockHeight:           header.GetDBHeight(),
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
