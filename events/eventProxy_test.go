package events

import (
	"fmt"
	. "github.com/FactomProject/factomd/common/messages/eventMsgs"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"testing"
	time2 "time"
)

var (
	entries  = 10000
	testHash = []byte("12345678901234567890123456789012")
)

func BenchmarkMarshalAnchorEventToBinary(b *testing.B) {
	b.StopTimer()
	fmt.Println(fmt.Sprintf("Benchmarking AnchorEvent binary marshalling %d cycles with %d entries", b.N, entries))
	event := mockAnchorEvent()
	bytes, _ := proto.Marshal(event)
	fmt.Println("Message size", len(bytes))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		proto.Marshal(event)
	}
}

func BenchmarkMarshalAnchorEventToJSON(b *testing.B) {
	b.StopTimer()
	fmt.Println(fmt.Sprintf("Benchmarking AnchorEvent json marshalling %d cycles with %d entries", b.N, entries))
	event := mockAnchorEvent()
	marshaller := &jsonpb.Marshaler{}
	msg, _ := marshaller.MarshalToString(event)
	fmt.Println("Message size", len(msg))

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, err := marshaller.MarshalToString(event)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkMockAnchorEvents(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mockAnchorEvent()
	}
}

func mockAnchorEvent() *AnchoredEvent {
	result := &AnchoredEvent{}
	result.DirectoryBlock = mockDirectoryBlock()
	return result
}

func mockDirectoryBlock() *DirectoryBlock {
	result := &DirectoryBlock{}
	result.Header = mockDirHeader()
	result.Entries = mockDirEntries()
	return result
}

func mockDirHeader() *DirectoryBlockHeader {
	time := time2.Now()
	result := &DirectoryBlockHeader{
		BodyMerkleRoot: &Hash{
			HashValue: testHash,
		},
		PreviousKeyMerkleRoot: &Hash{
			HashValue: testHash,
		},
		PreviousFullHash: &Hash{
			HashValue: testHash,
		},
		Timestamp:  &types.Timestamp{Seconds: int64(time.Second()), Nanos: int32(time.Nanosecond())},
		DBHeight:   123,
		BlockCount: 456,
	}
	return result
}

func mockDirEntries() []*Entry {
	result := make([]*Entry, entries)
	for i := 0; i < entries; i++ {
		result[i] = mockDirEntry()

	}
	return result
}

func mockDirEntry() *Entry {
	result := &Entry{
		ChainID: &Hash{
			HashValue: testHash,
		},
		KeyMerkleRoot: &Hash{
			HashValue: testHash,
		},
	}
	return result
}
