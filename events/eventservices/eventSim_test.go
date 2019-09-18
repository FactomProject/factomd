package eventservices_test

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/FactomProject/factomd/common/constants/runstate"
	"io"
	"net"
	"sync/atomic"
	"testing"
	"time"
)

const supportedProtocolVersion = 1

type EventServerSim struct {
	Protocol          string
	Address           string
	CorrectSendEvents int32
	ExpectedEvents    int
	test              *testing.T
	listener          net.Listener
	connection        net.Conn
	runState          runstate.RunState
}

func (sim *EventServerSim) Start() {
	var err error
	sim.runState = runstate.New
	sim.CorrectSendEvents = 0
	sim.listener, err = net.Listen(sim.Protocol, sim.Address)
	if err != nil {
		sim.test.Fatal(err)
	}
	go sim.waitForConnection()
}

func (sim *EventServerSim) Stop() {
	sim.runState = runstate.Stopping
	for sim.runState < runstate.Stopped {
		time.Sleep(1 * time.Millisecond)
	}
}

func (sim *EventServerSim) waitForConnection() {
	var err error
	sim.runState = runstate.Booting
	sim.connection, err = sim.listener.Accept()
	if err != nil {
		sim.test.Fatalf("failed to accept connection: %v\n", err)
	}
	sim.listenForEvents()
}

func (sim *EventServerSim) disconnect() {
	if sim.connection != nil {
		sim.connection.Close()
		sim.connection = nil
	}
	if sim.listener != nil {
		sim.listener.Close()
		sim.listener = nil
	}
}

func (sim *EventServerSim) listenForEvents() {
	defer sim.finalize()
	sim.runState = runstate.Running
	reader := bufio.NewReader(sim.connection)

	for i := atomic.LoadInt32(&sim.CorrectSendEvents); i < int32(sim.ExpectedEvents) && sim.runState < runstate.Stopping; i++ {
		fmt.Printf("read event: %d/%d\n", i, sim.ExpectedEvents)
		protocolVersion, err := reader.ReadByte()
		if err != nil {
			fmt.Printf("failed to read protocol version: %v\n", err)
			return
		}
		if protocolVersion != supportedProtocolVersion {
			fmt.Printf("unsupported protocol version: %d\n", protocolVersion)
			return
		}

		var dataSize int32
		if err := binary.Read(reader, binary.LittleEndian, &dataSize); err != nil {
			fmt.Printf("failed to read data size: %v\n", err)
		}

		if dataSize < 1 {
			fmt.Printf("data size incorrect: %d\n", dataSize)
		}
		data := make([]byte, dataSize)
		bytesRead, err := io.ReadFull(reader, data)
		if err != nil {
			fmt.Printf("failed to read data: %v\n", err)
		}

		sim.test.Logf("%v", data[0:bytesRead])
		atomic.AddInt32(&sim.CorrectSendEvents, 1)
	}
	return
}

func (sim *EventServerSim) finalize() {
	if r := recover(); r != nil {
		sim.test.Fatalf("Event simulator failed: %v", r)
	}
	sim.disconnect()
	sim.runState = runstate.Stopped
}
