package eventservices_test

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/FactomProject/factomd/common/constants/runstate"
	"github.com/FactomProject/factomd/events"
	"github.com/FactomProject/factomd/events/eventmessages"
	"github.com/FactomProject/factomd/events/eventservices"
	state2 "github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
	"net"
	"sync/atomic"
	"testing"
	"time"
)

var (
	entries  = 10000
	testHash = []byte("12345678901234567890123456789012")
)

func TestNoReceivingServer(t *testing.T) {
	protocol := "tcp"
	address := ":12410"

	state := &state2.State{}
	state.RunState = runstate.Running
	eventService, _ := eventservices.NewEventServiceTo(protocol, address, state)
	msgs := testHelper.CreateTestDBStateList()

	msg := msgs[0]
	event := events.EventFromMessage(eventmessages.EventSource_ADD_TO_PROCESSLIST, msg)
	eventService.Send(event)

	time.Sleep(2 * time.Second) // sleep less than the retry * redail sleep duration

	// listen for results
	var correctSendEvents int32 = 0
	listener, err := net.Listen(protocol, address)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	go listenForEvents(t, listener, &correctSendEvents, len(msgs))

	waitOnEvents(&correctSendEvents, 1, 10*time.Second)
	assert.EqualValues(t, 1, correctSendEvents, "failed to receive the correct number of events %d != %d", 1, correctSendEvents)
}

func TestEventProxy_Send(t *testing.T) {
	protocol := "tcp"
	address := ":12409"

	state := &state2.State{}
	state.RunState = runstate.Running
	eventService, _ := eventservices.NewEventServiceTo(protocol, address, state)
	msgs := testHelper.CreateTestDBStateList()

	// listen for results
	var correctSendEvents int32 = 0
	listener, err := net.Listen(protocol, address)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	go listenForEvents(t, listener, &correctSendEvents, len(msgs))

	// send messages
	for _, msg := range msgs {
		event := events.EventFromMessage(eventmessages.EventSource_ADD_TO_PROCESSLIST, msg)
		eventService.Send(event)
	}

	waitOnEvents(&correctSendEvents, len(msgs), 10*time.Second)

	assert.EqualValues(t, len(msgs), correctSendEvents, "failed to receive the correct number of events %d != %d", len(msgs), correctSendEvents)
}

func waitOnEvents(correctSendEvents *int32, n int, timeLimit time.Duration) {
	deadline := time.Now().Add(timeLimit)
	for int(atomic.LoadInt32(correctSendEvents)) != n && time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)
	}
}

func listenForEvents(t *testing.T, listener net.Listener, correctSendEvents *int32, n int) {
	conn, err := listener.Accept()
	if err != nil {
		fmt.Printf("failed to accept connection: %v\n", err)
		return
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for i := atomic.LoadInt32(correctSendEvents); i < int32(n); i++ {
		fmt.Printf("read event: %d/%d\n", i, n)
		var dataSize int32
		if err := binary.Read(reader, binary.LittleEndian, &dataSize); err != nil {
			fmt.Printf("failed to read data size: %v\n", err)
		}

		if dataSize < 1 {
			fmt.Printf("data size incorrect: %d\n", dataSize)
		}
		data := make([]byte, dataSize)
		bytesRead, err := reader.Read(data)
		if err != nil {
			fmt.Printf("failed to read data: %v\n", err)
		}

		t.Logf("%v", data[0:bytesRead])
		atomic.AddInt32(correctSendEvents, 1)
	}
}

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
	msg, _ := json.Marshal(event)
	fmt.Println("Message size", len(msg))

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(event)
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

func mockAnchorEvent() *eventmessages.AnchoredEvent {
	result := &eventmessages.AnchoredEvent{}
	result.DirectoryBlock = mockDirectoryBlock()
	return result
}

func mockDirectoryBlock() *eventmessages.DirectoryBlock {
	result := &eventmessages.DirectoryBlock{}
	result.Header = mockDirHeader()
	result.Entries = mockDirEntries()
	return result
}

func mockDirHeader() *eventmessages.DirectoryBlockHeader {
	t := time.Now()
	result := &eventmessages.DirectoryBlockHeader{
		BodyMerkleRoot: &eventmessages.Hash{
			HashValue: testHash,
		},
		PreviousKeyMerkleRoot: &eventmessages.Hash{
			HashValue: testHash,
		},
		PreviousFullHash: &eventmessages.Hash{
			HashValue: testHash,
		},
		Timestamp:   &types.Timestamp{Seconds: int64(t.Second()), Nanos: int32(t.Nanosecond())},
		BlockHeight: 123,
		BlockCount:  456,
	}
	return result
}

func mockDirEntries() []*eventmessages.Entry {
	result := make([]*eventmessages.Entry, entries)
	for i := 0; i < entries; i++ {
		result[i] = mockDirEntry()

	}
	return result
}

func mockDirEntry() *eventmessages.Entry {
	result := &eventmessages.Entry{
		ChainID: &eventmessages.Hash{
			HashValue: testHash,
		},
		KeyMerkleRoot: &eventmessages.Hash{
			HashValue: testHash,
		},
	}
	return result
}
