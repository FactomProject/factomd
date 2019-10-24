package eventservices

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/events/eventmessages/generated/eventmessages"
)

func mapDirectoryBlock(block interfaces.IDirectoryBlock) *eventmessages.DirectoryBlock {
	result := &eventmessages.DirectoryBlock{Header: mapDirectoryBlockHeader(block.GetHeader()),
		Entries: mapDirectoryBlockEntries(block.GetDBEntries())}
	return result
}

func mapDirectoryBlockHeader(header interfaces.IDirectoryBlockHeader) *eventmessages.DirectoryBlockHeader {
	result := &eventmessages.DirectoryBlockHeader{
		BodyMerkleRoot:        header.GetBodyMR().Bytes(),
		PreviousKeyMerkleRoot: header.GetPrevKeyMR().Bytes(),
		PreviousFullHash:      header.GetPrevFullHash().Bytes(),
		Timestamp:             ConvertTimeToTimestamp(header.GetTimestamp().GetTime()),
		BlockHeight:           header.GetDBHeight(),
		BlockCount:            header.GetBlockCount(),
		Version:               uint32(header.GetVersion()),
		NetworkID:             header.GetNetworkID(),
	}
	return result
}

func mapDirectoryBlockEntries(entries []interfaces.IDBEntry) []*eventmessages.DirectoryBlockEntry {
	result := make([]*eventmessages.DirectoryBlockEntry, len(entries))
	for i, entry := range entries {
		result[i] = mapDirectoryBlockEntry(entry)
	}
	return result
}

func mapDirectoryBlockEntry(entry interfaces.IDBEntry) *eventmessages.DirectoryBlockEntry {
	result := &eventmessages.DirectoryBlockEntry{
		ChainID:       entry.GetChainID().Bytes(),
		KeyMerkleRoot: entry.GetKeyMR().Bytes(),
	}
	return result
}
