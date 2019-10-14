package eventservices_test

import (
	"encoding/json"
	"fmt"
	"github.com/FactomProject/factomd/common/constants/runstate"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/events"
	"github.com/FactomProject/factomd/events/eventmessages/generated/eventmessages"
	"github.com/FactomProject/factomd/events/eventoutputformat"
	"github.com/FactomProject/factomd/events/eventservices"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

var (
	entries  = 10000
	testHash = []byte("12345678901234567890123456789012")
	tcpPort  = 12408
)

func testEventServiceProtobuf(t *testing.T) {
	outputFormat := eventoutputformat.Protobuf
	t.Run("Event service sim-tests protobuf", func(t *testing.T) {
		blockCommitList := testHelper.CreateTestBlockCommitList()

		testSend(t, blockCommitList, outputFormat)
		testLateReceivingServer(t, blockCommitList, outputFormat)
		testReceivingServerRestart(t, blockCommitList, outputFormat)
	})
}

func testEventServiceJson(t *testing.T) {
	outputFormat := eventoutputformat.Json
	t.Run("Event service sim-tests json", func(t *testing.T) {
		blockCommitList := testHelper.CreateTestBlockCommitList()

		testSend(t, blockCommitList, outputFormat)
		testLateReceivingServer(t, blockCommitList, outputFormat)
		testReceivingServerRestart(t, blockCommitList, outputFormat)
	})
}

func testSend(t *testing.T, msgs []interfaces.IMsg, outputFormat eventoutputformat.Format) {
	t.Run("Test receiving running normally", func(t *testing.T) {
		state := &state.State{
			IdentityChainID: primitives.NewZeroHash(),
			RunState:        runstate.Running,
		}

		tcpPort = tcpPort + 1
		sim := &EventServerSim{
			Protocol:       "tcp",
			Address:        ":" + strconv.Itoa(tcpPort),
			ExpectedEvents: len(msgs),
			test:           t,
		}
		sim.Start()
		eventService, eventServiceControl := eventservices.NewEventServiceTo(state, buildParams(sim, outputFormat))
		defer eventServiceControl.Shutdown()

		// send messages
		for _, msg := range msgs {
			event := events.NewStateChangeEventFromMsg(eventmessages.EventSource_LIVE, eventmessages.EntityState_COMMITTED_TO_DIRECTORY_BLOCK, msg)
			eventService.Send(event)
		}

		waitOnEvents(&sim.CorrectSendEvents, len(msgs), 10*time.Second)
		assert.EqualValues(t, len(msgs), sim.CorrectSendEvents,
			"failed to receive the correct number of events %d != %d", len(msgs), sim.CorrectSendEvents)
	})
}

func testLateReceivingServer(t *testing.T, msgs []interfaces.IMsg, outputFormat eventoutputformat.Format) {
	t.Run("Test receiving late start", func(t *testing.T) {
		state := &state.State{
			IdentityChainID: primitives.NewZeroHash(),
			RunState:        runstate.Running,
		}
		msgs := testHelper.CreateTestBlockCommitList()

		tcpPort = tcpPort + 1
		sim := &EventServerSim{
			Protocol:       "tcp",
			Address:        ":" + strconv.Itoa(tcpPort),
			ExpectedEvents: len(msgs),
			test:           t,
		}
		eventService, eventServiceControl := eventservices.NewEventServiceTo(state, buildParams(sim, outputFormat))
		defer eventServiceControl.Shutdown()

		msg := msgs[0]
		event := events.NewStateChangeEventFromMsg(eventmessages.EventSource_LIVE, eventmessages.EntityState_ACCEPTED, msg)
		eventService.Send(event)

		time.Sleep(2 * time.Second) // sleep less than the retry * redial sleep duration
		sim.Start()
		waitOnEvents(&sim.CorrectSendEvents, 1, 25*time.Second)
		assert.EqualValues(t, 1, sim.CorrectSendEvents,
			"failed to receive the correct number of events %d != %d", 1, sim.CorrectSendEvents)
	})
}

func testReceivingServerRestart(t *testing.T, msgs []interfaces.IMsg, outputFormat eventoutputformat.Format) {
	t.Run("Test receiving server restart", func(t *testing.T) {

		state := &state.State{
			IdentityChainID: primitives.NewZeroHash(),
			RunState:        runstate.Running,
		}
		msgs := testHelper.CreateTestBlockCommitList()

		tcpPort = tcpPort + 1
		sim := &EventServerSim{
			Protocol:       "tcp",
			Address:        ":" + strconv.Itoa(tcpPort),
			ExpectedEvents: 1,
			test:           t,
		}
		if err := sim.StartExternal(); err != nil {
			t.Fatal("Could not launch external eventservice sim", err)
		}
		eventService, eventServiceControl := eventservices.NewEventServiceTo(state, buildParams(sim, outputFormat))
		defer eventServiceControl.Shutdown()

		msg := msgs[0]
		event := events.NewStateChangeEventFromMsg(eventmessages.EventSource_LIVE, eventmessages.EntityState_ACCEPTED, msg)
		eventService.Send(event)

		// Restart the simulator
		sim.Stop()
		time.Sleep(5 * time.Second)

		correctEventsFromFirstSession := sim.CorrectSendEvents
		sim.CorrectSendEvents = 0
		sim.ExpectedEvents = 2
		sim.Start()
		defer sim.Stop()

		msg = msgs[1]
		event = events.NewStateChangeEventFromMsg(eventmessages.EventSource_LIVE, eventmessages.EntityState_ACCEPTED, msg)
		eventService.Send(event)
		msg = msgs[2]
		event = events.NewStateChangeEventFromMsg(eventmessages.EventSource_LIVE, eventmessages.EntityState_ACCEPTED, msg)
		eventService.Send(event)
		waitOnEvents(&sim.CorrectSendEvents, 2, 25*time.Second)
		assert.EqualValues(t, 3, sim.CorrectSendEvents+correctEventsFromFirstSession,
			"failed to receive the correct number of events %d != %d", 1, sim.CorrectSendEvents)

	})
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
	event := mockBlockCommitEvent()
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
	event := mockBlockCommitEvent()
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
		mockBlockCommitEvent()
	}
}

func mockBlockCommitEvent() *eventmessages.DirectoryBlockCommit {
	result := &eventmessages.DirectoryBlockCommit{}
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

func mockDirEntries() []*eventmessages.DirectoryBlockEntry {
	result := make([]*eventmessages.DirectoryBlockEntry, entries)
	for i := 0; i < entries; i++ {
		result[i] = mockDirEntry()

	}
	return result
}

func mockDirEntry() *eventmessages.DirectoryBlockEntry {
	result := &eventmessages.DirectoryBlockEntry{
		ChainID: &eventmessages.Hash{
			HashValue: testHash,
		},
		KeyMerkleRoot: &eventmessages.Hash{
			HashValue: testHash,
		},
	}
	return result
}

func buildParams(sim *EventServerSim, format eventoutputformat.Format) *eventservices.EventServiceParams {
	params := &eventservices.EventServiceParams{
		EnableLiveFeedAPI:            true,
		Protocol:                     sim.Protocol,
		Address:                      sim.Address,
		OutputFormat:                 format,
		MuteEventReplayDuringStartup: false,
	}
	return params
}
