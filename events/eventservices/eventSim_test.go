package eventservices_test

import (
	"bufio"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/PaulSnow/factom2d/common/constants/runstate"
	log "github.com/sirupsen/logrus"
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

var sim *EventServerSim
var cmd *exec.Cmd
var stdin io.Writer
var scanner *bufio.Scanner

func init() {
	log.Info("Event server simulator")
	sim = &EventServerSim{}
	flag.StringVar(&sim.Protocol, "protocol", "tcp", "Protocol")
	flag.StringVar(&sim.Address, "address", "", "Binding adress")
	flag.IntVar(&sim.ExpectedEvents, "expectedevents", 0, "Expected events")
	flag.Parse()
	scanner = bufio.NewScanner(os.Stdin)
}

func (sim *EventServerSim) StartExternal() error {
	sim.CorrectSendEvents = 0
	goPath := os.Getenv("GOROOT") + "/bin/go"
	cmd = exec.Command(goPath, "test", "-v", "./", "-run", "TestRunExternal", "-protocol="+sim.Protocol,
		"-address="+sim.Address, "-expectedevents="+strconv.Itoa(sim.ExpectedEvents))
	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr
	reader, _ := cmd.StdoutPipe()
	scanner = bufio.NewScanner(reader)
	stdin, _ = cmd.StdinPipe()
	if err := cmd.Start(); err != nil {
		return err
	}
	if _, err := waitForResponse("running"); err != nil {
		return err
	}
	return nil
}

func TestRunExternal(t *testing.T) {
	defer func() { fmt.Println("exit") }()

	sim.test = t
	if sim.ExpectedEvents == 0 || len(sim.Address) == 0 {
		fmt.Println("commandline parameters not set, ignoring test")
	} else {
		fmt.Println("Starting simulator on", sim.Address)
		sim.Start()
		fmt.Println("running")
		sim.waitForExpectedEvents()
		fmt.Println("CorrectSendEvents:", sim.CorrectSendEvents)
		sim.Stop()
	}
}

func waitForResponse(response string) (string, error) {
	for scanner.Scan() {
		text := scanner.Text()
		fmt.Println("####", text)

		if strings.HasPrefix(text, response) {
			return text, nil
		} else if text == "exit" {
			return "", errors.New("sim process exited prematurely")
		}
	}
	return "", errors.New("sim process disappeared")
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
	if cmd != nil {
		sim.stopExternalInstance()
	} else {
		sim.stopLocalInstance()
	}
}

func (sim *EventServerSim) stopLocalInstance() {
	if sim.runState < runstate.Stopping {
		sim.runState = runstate.Stopping
	}
	for sim.runState < runstate.Stopped {
		time.Sleep(10 * time.Millisecond)
	}
}

func (sim *EventServerSim) stopExternalInstance() {
	response, err := waitForResponse("CorrectSendEvents")
	if err == nil {
		s := response[19:]
		i, _ := strconv.Atoi(s)
		sim.CorrectSendEvents = int32(i)
		scanner = nil
		stdin = nil
		cmd = nil
	}
}

func (sim *EventServerSim) waitForExpectedEvents() {
	timeOut := 0
	for {
		if sim.CorrectSendEvents >= int32(sim.ExpectedEvents) || timeOut > sim.ExpectedEvents {
			break
		}
		time.Sleep(2 * time.Second)
		timeOut++
	}
}

func (sim *EventServerSim) waitForConnection() {
	var err error
	sim.runState = runstate.Booting
	sim.connection, err = sim.listener.Accept()
	if err != nil && sim.runState < runstate.Stopping {
		sim.test.Fatalf("failed to accept connection: %v", err)
	}
	log.Info("Accepted incoming connection")
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
		log.Infof("read event: %d/%d", i, sim.ExpectedEvents)
		protocolVersion, err := reader.ReadByte()
		if err != nil {
			log.Errorf("failed to read protocol version: %v", err)
			return
		}
		if protocolVersion != supportedProtocolVersion {
			log.Errorf("unsupported protocol version: %d", protocolVersion)
			return
		}

		var dataSize int32
		if err := binary.Read(reader, binary.LittleEndian, &dataSize); err != nil {
			log.Errorf("failed to read data size: %v", err)
		}

		if dataSize < 1 {
			log.Errorf("data size incorrect: %d", dataSize)
		}
		data := make([]byte, dataSize)
		bytesRead, err := io.ReadFull(reader, data)
		if err != nil {
			log.Errorf("failed to read data: %v", err)
		}
		sim.test.Logf("read %d bytes", bytesRead)
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
