package eventMessages

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/gogo/protobuf/types"
	"github.com/golang/protobuf/proto"
)

func AnchoredEventFromDBState(dbStateMessage *messages.DBStateMsg) *AnchoredEvent {
	event := &AnchoredEvent{}
	event.DirectoryBlock = mapDirBlock(dbStateMessage.DirectoryBlock)
	return event
}

func mapDirBlock(block interfaces.IDirectoryBlock) *DirectoryBlock {
	result := &DirectoryBlock{}
	result.Header = mapDirHeader(block.GetHeader())
	result.Entries = mapDirEntries(block.GetDBEntries())
	return result
}

func mapDirHeader(header interfaces.IDirectoryBlockHeader) *DirectoryBlockHeader {

	time := header.GetTimestamp().GetTime()
	result := &DirectoryBlockHeader{
		BodyMerkleRoot: &Hash{
			HashValue: header.GetBodyMR().Bytes(),
		},
		PreviousKeyMerkleRoot: &Hash{
			HashValue: header.GetPrevKeyMR().Bytes(),
		},
		PreviousFullHash: &Hash{
			HashValue: header.GetPrevFullHash().Bytes(),
		},
		Timestamp:  &types.Timestamp{Seconds: int64(time.Second()), Nanos: int32(time.Nanosecond())},
		DBHeight:   header.GetDBHeight(),
		BlockCount: header.GetBlockCount(),
	}

	return result
}

func mapDirEntries(entries []interfaces.IDBEntry) []*Entry {
	result := make([]*Entry, len(entries))
	for i, entry := range entries {
		result[i] = mapDirEntry(entry)
	}
	return result
}

func mapDirEntry(entry interfaces.IDBEntry) *Entry {
	result := &Entry{
		ChainID: &Hash{
			HashValue: entry.GetChainID().Bytes(),
		},
		KeyMerkleRoot: &Hash{
			HashValue: entry.GetKeyMR().Bytes(),
		},
	}
	return result
}

type Event interface {
	Reset()
	String() string
	ProtoMessage()
	XXX_Unmarshal(b []byte) error
	XXX_Marshal(b []byte, deterministic bool) ([]byte, error)
	XXX_Merge(src proto.Message)
	XXX_Size() int
	XXX_DiscardUnknown()
}
