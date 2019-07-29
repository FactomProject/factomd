package events

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/events/eventmessages"
)

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
