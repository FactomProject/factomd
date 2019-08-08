/**
		===== IMPORTANT only run these tests one by one, not the entire package =====
**/

package eventservices_test

import (
	"encoding/json"
	"fmt"
	"github.com/FactomProject/factomd/common/constants/runstate"
	"github.com/FactomProject/factomd/events"
	"github.com/FactomProject/factomd/events/eventmessages"
	"github.com/FactomProject/factomd/events/eventoutputformat"
	"github.com/FactomProject/factomd/events/eventservices"
	state2 "github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
	"time"
)

var (
	entries  = 10000
	testHash = []byte("12345678901234567890123456789012")
)

func TestEventProxy_Send(t *testing.T) {
	state := &state2.State{}
	state.RunState = runstate.Running
	msgs := testHelper.CreateTestDBStateList()

	sim := &EventServerSim{
		Protocol:       "tcp",
		Address:        ":12409",
		ExpectedEvents: len(msgs),
		test:           t,
	}
	sim.Start()
	eventService, _ := eventservices.NewEventServiceTo(state, buildParams(sim))

	// send messages
	for _, msg := range msgs {
		event := events.EventFromMessage(eventmessages.EventSource_ADD_TO_PROCESSLIST, msg)
		eventService.Send(event)
	}

	waitOnEvents(&sim.CorrectSendEvents, len(msgs), 10*time.Second)

	assert.EqualValues(t, len(msgs), sim.CorrectSendEvents,
		"failed to receive the correct number of events %d != %d", len(msgs), sim.CorrectSendEvents)
}

func TestNoReceivingServer(t *testing.T) {
	state := &state2.State{}
	state.RunState = runstate.Running
	msgs := testHelper.CreateTestDBStateList()

	sim := &EventServerSim{
		Protocol:       "tcp",
		Address:        ":12410",
		ExpectedEvents: len(msgs),
		test:           t,
	}
	eventService, _ := eventservices.NewEventServiceTo(state, buildParams(sim))

	msg := msgs[0]
	event := events.EventFromMessage(eventmessages.EventSource_ADD_TO_PROCESSLIST, msg)
	eventService.Send(event)

	time.Sleep(2 * time.Second) // sleep less than the retry * redail sleep duration
	sim.Start()
	waitOnEvents(&sim.CorrectSendEvents, 1, 25*time.Second)
	assert.EqualValues(t, 1, sim.CorrectSendEvents,
		"failed to receive the correct number of events %d != %d", 1, sim.CorrectSendEvents)
}

func TestReceivingServerRestarted(t *testing.T) {
	state := &state2.State{}
	state.RunState = runstate.Running
	msgs := testHelper.CreateTestDBStateList()

	sim := &EventServerSim{
		Protocol:       "tcp",
		Address:        ":12411",
		ExpectedEvents: len(msgs),
		test:           t,
	}
	sim.Start()
	eventService, _ := eventservices.NewEventServiceTo(state, buildParams(sim))

	msg := msgs[0]
	event := events.EventFromMessage(eventmessages.EventSource_ADD_TO_PROCESSLIST, msg)
	eventService.Send(event)

	// Restart the simulator
	sim.Stop()

	// We have to wait quite some time for the listener to really die,
	// if we open a new listener too early the client's messages will go into the endless void
	// In real life when the process dies we won't see this issue
	time.Sleep(122 * time.Second)
	sim.Start()
	eventService.Send(event)
	waitOnEvents(&sim.CorrectSendEvents, 1, 25*time.Second)
	assert.EqualValues(t, 1, sim.CorrectSendEvents,
		"failed to receive the correct number of events %d != %d", 1, sim.CorrectSendEvents)
}

func waitOnEvents(correctSendEvents *int32, n int, timeLimit time.Duration) {
	deadline := time.Now().Add(timeLimit)
	for int(atomic.LoadInt32(correctSendEvents)) != n && time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)
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

func buildParams(sim *EventServerSim) *eventservices.EventServiceParams {
	params := &eventservices.EventServiceParams{
		EnableLiveFeedAPI:       true,
		Protocol:                sim.Protocol,
		Address:                 sim.Address,
		OutputFormat:            eventoutputformat.Protobuf,
		MuteEventsDuringStartup: false,
	}
	return params
}
