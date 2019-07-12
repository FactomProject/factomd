package eventMessages

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/gogo/protobuf/types"
)

func FromDBState(dbStateMessage *messages.DBStateMsg) *AnchoredEvent {
	event := &AnchoredEvent{}
	event.DirectoryBlock = mapDirBlock(dbStateMessage.DirectoryBlock)
	return event
}

func mapDirBlock(block interfaces.IDirectoryBlock) *DirectoryBlock {
	result := &DirectoryBlock{}
	result.Header = mapDirHeader(block.GetHeader())
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
