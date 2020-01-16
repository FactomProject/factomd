package simulation

//A package for functions used multiple times in tests that aren't useful in production code.

import (
	"bytes"
	"encoding/binary"
	"github.com/FactomProject/factomd/pubsub"
	"github.com/FactomProject/factomd/testHelper"

	"github.com/FactomProject/factomd/registry"
	"github.com/FactomProject/factomd/worker"

	"github.com/FactomProject/factom"
	"github.com/FactomProject/factomd/common/primitives"
	"time"

	"fmt"
	"os"

	"github.com/FactomProject/factomd/state"

	"github.com/FactomProject/factomd/common/messages/electionMsgs"
)

var BlockCount int = 10
var DefaultCoinbaseAmount uint64 = 100000000

func CreateAndPopulateStaleHolding() *state.State {
	s := CreateAndPopulateTestState()

	// TODO: refactor into test helpers
	a := AccountFromFctSecret("Fs2zQ3egq2j99j37aYzaCddPq9AF3mgh64uG9gRaDAnrkjRx3eHs")

	encode := func(s string) []byte {
		b := bytes.Buffer{}
		b.WriteString(s)
		return b.Bytes()
	}

	id := "92475004e70f41b94750f4a77bf7b430551113b25d3d57169eadca5692bb043d"
	extids := [][]byte{encode(fmt.Sprintf("makeStaleMessages"))}

	e := factom.Entry{
		ChainID: id,
		ExtIDs:  extids,
		Content: encode(fmt.Sprintf("this is a stale message")),
	}

	// create stale MilliTime
	mockTime := func() (r []byte) {
		buf := new(bytes.Buffer)
		t := time.Now().UnixNano()
		m := t/1e6 - state.FilterTimeLimit // make msg too old
		binary.Write(buf, binary.BigEndian, m)
		return buf.Bytes()[2:]
	}

	// adding a commit w/ no REVEAL
	m, _ := ComposeCommitEntryMsg(a.Priv, e)
	copy(m.CommitEntry.MilliTime[:], mockTime())

	// add commit to holding
	s.Hold.Add(m.GetMsgHash().Fixed(), m)

	return s
}

func CreateAndPopulateTestState() *state.State {
	pubsub.Reset() // clear existing pubsub paths between tests
	s := new(state.State)
	s.BindPublishers()
	s.TimestampAtBoot = new(primitives.Timestamp)
	s.TimestampAtBoot.SetTime(0)
	s.SetLeaderTimestamp(primitives.NewTimestampFromMilliseconds(0))
	s.DB = testHelper.CreateAndPopulateTestDatabaseOverlay()
	s.LoadConfig("", "")

	s.DirectoryBlockInSeconds = 20

	s.Network = "LOCAL"
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "enablenet", false))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "database", s.DBType))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "database for clones", s.CloneDBType))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%d\"\n", "port", s.PortNumber))
	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "block time", s.DirectoryBlockInSeconds))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "Network", s.Network))
	s.LogPath = "stdout"

	p := registry.New()
	p.Register(func(w *worker.Thread) { s.Initialize(w, new(electionMsgs.ElectionsFactory)) })
	go p.Run()
	p.WaitForRunning()

	s.Network = "LOCAL"
	/*err := s.RecalculateBalances()
	if err != nil {
		panic(err)
	}*/
	s.SetFactoshisPerEC(1)
	s.MMRDummy() // Need to start MMR to ensure queues don't fill up
	s.LoadDatabase()
	s.Process()
	s.UpdateState()

	return s
}
