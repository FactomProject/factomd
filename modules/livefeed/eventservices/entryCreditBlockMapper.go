package eventservices

import (
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/modules/livefeed/eventmessages/generated/eventmessages"
)

func mapEntryCreditBlock(block interfaces.IEntryCreditBlock) *eventmessages.EntryCreditBlock {
	return &eventmessages.EntryCreditBlock{
		Header:  mapEntryCreditBlockHeader(block.GetHeader()),
		Entries: mapEntryCreditBlockEntries(block.GetEntries()),
	}
}

func mapEntryCreditBlockHeader(header interfaces.IECBlockHeader) *eventmessages.EntryCreditBlockHeader {
	return &eventmessages.EntryCreditBlockHeader{
		BodyHash:           header.GetBodyHash().Bytes(),
		PreviousHeaderHash: header.GetPrevHeaderHash().Bytes(),
		PreviousFullHash:   header.GetPrevHeaderHash().Bytes(),
		BlockHeight:        header.GetDBHeight(),
		ObjectCount:        header.GetObjectCount(),
	}
}

func mapEntryCreditBlockEntries(entries []interfaces.IECBlockEntry) []*eventmessages.EntryCreditBlockEntry {
	result := make([]*eventmessages.EntryCreditBlockEntry, len(entries))
	for i, entry := range entries {
		result[i] = mapEntryCreditBlockEntry(entry)
	}
	return result
}

func mapEntryCreditBlockEntry(entry interfaces.IECBlockEntry) *eventmessages.EntryCreditBlockEntry {
	result := &eventmessages.EntryCreditBlockEntry{}
	switch entry.(type) {
	case *entryCreditBlock.CommitChain:
		result.EntryCreditBlockEntry = mapEntryCreditCommitChain(entry)
	case *entryCreditBlock.CommitEntry:
		result.EntryCreditBlockEntry = mapEntryCreditCommitEntry(entry)
	case *entryCreditBlock.IncreaseBalance:
		result.EntryCreditBlockEntry = mapEntryCreditIncreaseBalance(entry)
	case *entryCreditBlock.MinuteNumber:
		result.EntryCreditBlockEntry = mapEntryCreditMinuteNumber(entry)
	case *entryCreditBlock.ServerIndexNumber:
		result.EntryCreditBlockEntry = mapEntryCreditServerIndexNumber(entry)
	}
	return result
}

func mapEntryCreditCommitChain(entry interfaces.IECBlockEntry) *eventmessages.EntryCreditBlockEntry_ChainCommit {
	commitChain, _ := entry.(*entryCreditBlock.CommitChain)
	return &eventmessages.EntryCreditBlockEntry_ChainCommit{
		ChainCommit: &eventmessages.ChainCommit{
			EntityState:          eventmessages.EntityState_COMMITTED_TO_DIRECTORY_BLOCK,
			ChainIDHash:          commitChain.ChainIDHash.Bytes(),
			EntryHash:            commitChain.EntryHash.Bytes(),
			Weld:                 commitChain.Weld.Bytes(),
			Timestamp:            convertByteSlice6ToTimestamp(commitChain.MilliTime),
			Credits:              uint32(commitChain.Credits),
			EntryCreditPublicKey: commitChain.ECPubKey[:],
			Signature:            commitChain.Sig[:],
			Version:              uint32(commitChain.Version),
		},
	}
}

func mapEntryCreditCommitEntry(entry interfaces.IECBlockEntry) *eventmessages.EntryCreditBlockEntry_EntryCommit {
	commitEntry, _ := entry.(*entryCreditBlock.CommitEntry)
	return &eventmessages.EntryCreditBlockEntry_EntryCommit{
		EntryCommit: &eventmessages.EntryCommit{
			EntityState:          eventmessages.EntityState_COMMITTED_TO_DIRECTORY_BLOCK,
			EntryHash:            commitEntry.EntryHash.Bytes(),
			Timestamp:            convertByteSlice6ToTimestamp(commitEntry.MilliTime),
			Credits:              uint32(commitEntry.Credits),
			EntryCreditPublicKey: commitEntry.ECPubKey[:],
			Signature:            commitEntry.Sig[:],
			Version:              uint32(commitEntry.Version),
		},
	}
}
func mapEntryCreditIncreaseBalance(entry interfaces.IECBlockEntry) *eventmessages.EntryCreditBlockEntry_IncreaseBalance {
	increaseBalance, _ := entry.(*entryCreditBlock.IncreaseBalance)
	return &eventmessages.EntryCreditBlockEntry_IncreaseBalance{
		IncreaseBalance: &eventmessages.IncreaseBalance{
			EntryCreditPublicKey: increaseBalance.ECPubKey[:],
			TransactionID:        increaseBalance.TXID.Bytes(),
			Index:                increaseBalance.Index,
			Amount:               increaseBalance.NumEC,
		},
	}
}

func mapEntryCreditMinuteNumber(entry interfaces.IECBlockEntry) *eventmessages.EntryCreditBlockEntry_MinuteNumber {
	minuteNumber, _ := entry.(*entryCreditBlock.MinuteNumber)
	return &eventmessages.EntryCreditBlockEntry_MinuteNumber{
		MinuteNumber: &eventmessages.MinuteNumber{
			MinuteNumber: uint32(minuteNumber.Number),
		},
	}
}

func mapEntryCreditServerIndexNumber(entry interfaces.IECBlockEntry) *eventmessages.EntryCreditBlockEntry_ServerIndexNumber {
	serverIndexNumber, _ := entry.(*entryCreditBlock.ServerIndexNumber)
	return &eventmessages.EntryCreditBlockEntry_ServerIndexNumber{
		ServerIndexNumber: &eventmessages.ServerIndexNumber{
			ServerIndexNumber: uint32(serverIndexNumber.ServerIndexNumber),
		},
	}
}
