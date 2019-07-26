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
		FactoidBlock:      mapFactoidBlock(dbStateMessage.FactoidBlock),
		EntryBlocks:       mapEntryBlocks(dbStateMessage.EBlocks),
		EntryBlockEntries: mapEntryBlockEntries(dbStateMessage.Entries),
	}}
	return event
}

func mapFactoidBlock(block interfaces.IFBlock) *eventmessages.FactoidBlock {
	result := &eventmessages.FactoidBlock{
		BodyMerkleRoot: &eventmessages.Hash{
			HashValue: block.GetBodyMR().Bytes(),
		},
		PreviousKeyMerkleRoot: &eventmessages.Hash{
			HashValue: block.GetPrevKeyMR().Bytes(),
		},
		PreviousLedgerKeyMerkleRoot: &eventmessages.Hash{
			HashValue: block.GetLedgerKeyMR().Bytes(),
		},
		ExchRate:     block.GetExchRate(),
		BlockHeight:  block.GetDBHeight(),
		Transactions: mapTransactions(block.GetTransactions()),
	}
	return result
}

func mapTransactions(transactions []interfaces.ITransaction) []*eventmessages.Transaction {
	result := make([]*eventmessages.Transaction, 0)
	for _, transaction := range transactions {
		err := transaction.ValidateSignatures()
		if err == nil {
			result = append(result, mapTransaction(transaction))
		}
	}
	return result
}

func mapTransaction(transaction interfaces.ITransaction) *eventmessages.Transaction {
	result := &eventmessages.Transaction{
		TransactionId: &eventmessages.Hash{
			HashValue: transaction.GetSigHash().Bytes(),
		},
		BlockHeight:        transaction.GetBlockHeight(),
		Timestamp:          convertTimeToTimestamp(transaction.GetTimestamp().GetTime()),
		Inputs:             mapTransactionAddresses(transaction.GetInputs()),
		Outputs:            mapTransactionAddresses(transaction.GetOutputs()),
		OutputEntryCredits: mapTransactionAddresses(transaction.GetECOutputs()),
		RCds:               mapRCDs(transaction.GetRCDs()),
		SignatureBlocks:    mapSignatureBlocks(transaction.GetSignatureBlocks()),
	}
	return result
}

func mapTransactionAddresses(inputs []interfaces.ITransAddress) []*eventmessages.TransactionAddress {
	result := make([]*eventmessages.TransactionAddress, len(inputs))
	for i, input := range inputs {
		result[i] = mapTransactionAddress(input)
	}
	return result
}

func mapTransactionAddress(address interfaces.ITransAddress) *eventmessages.TransactionAddress {
	result := &eventmessages.TransactionAddress{
		Amount: address.GetAmount(),
		Address: &eventmessages.Hash{
			HashValue: address.GetAddress().Bytes(),
		},
	}
	return result
}

func mapRCDs(rcds []interfaces.IRCD) []*eventmessages.RCD {
	result := make([]*eventmessages.RCD, len(rcds))
	for i, rcd := range rcds {
		result[i] = mapRCD(rcd)
	}
	return result

}

func mapRCD(rcd interfaces.IRCD) *eventmessages.RCD {
	result := &eventmessages.RCD{}
	/* TODO research this more, the rcd1/2 structs and interfaces make me dizzy
	https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#factoid-transaction

	switch rcd.(type) {
	case factoid.IRCD_1:
		rcd1 := rcd.(factoid.IRCD_1)
		result.Value = &eventmessages.RCD_Rcd1{
			Rcd1: &eventmessages.RCD1{
				PublicKey: rcd1.GetPublicKey(),
			},
		}

	case factoid.IRCD:
		rcd2 := rcd.(factoid.IRCD)
		evRcd2 := eventmessages.RCD2{
			M:          0,
			N:          0,
			NAddresses: rcd2.GetHash().Bytes(),
		}
		evRcd2Value := &eventmessages.RCD_Rcd2{Rcd2: evRcd2}
		result.Value = evRcd2Value
	}*/
	return result
}

func mapSignatureBlocks(blocks []interfaces.ISignatureBlock) []*eventmessages.FactoidSignatureBlock {
	result := make([]*eventmessages.FactoidSignatureBlock, len(blocks))
	for i, block := range blocks {
		result[i] = mapSignatureBlock(block)
	}
	return result
}

func mapSignatureBlock(block interfaces.ISignatureBlock) *eventmessages.FactoidSignatureBlock {
	result := &eventmessages.FactoidSignatureBlock{
		Signature: mapFactoidSignatureBlockSingatures(block.GetSignatures()),
	}
	return result
}

func mapFactoidSignatureBlockSingatures(signatures []interfaces.ISignature) []*eventmessages.FactoidSignature {
	result := make([]*eventmessages.FactoidSignature, len(signatures))
	for i, signature := range signatures {
		result[i] = &eventmessages.FactoidSignature{
			SignatureValue: signature.Bytes(),
		}
	}
	return result
}

func mapDirBlock(block interfaces.IDirectoryBlock) *eventmessages.DirectoryBlock {
	result := &eventmessages.DirectoryBlock{Header: mapDirHeader(block.GetHeader()),
		Entries: mapDirEntries(block.GetDBEntries())}
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
		Timestamp:   convertTimeToTimestamp(header.GetTimestamp().GetTime()),
		BlockHeight: header.GetDBHeight(),
		BlockCount:  header.GetBlockCount(),
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
