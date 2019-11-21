package testHelper

//A package for functions used multiple times in tests that aren't useful in production code.

import (
	"bytes"
	"encoding/binary"
	"os/exec"
	"regexp"
	"runtime"

	"github.com/FactomProject/factom"
	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/mapdb"

	"time"

	"fmt"
	"os"

	"github.com/FactomProject/factomd/state"

	"github.com/FactomProject/factomd/common/messages/electionMsgs"
)

var BlockCount int = 10
var DefaultCoinbaseAmount uint64 = 100000000

func CreateEmptyTestState() *state.State {
	s := new(state.State)
	s.TimestampAtBoot = new(primitives.Timestamp)
	s.TimestampAtBoot.SetTime(0)
	s.EFactory = new(electionMsgs.ElectionsFactory)
	s.LoadConfig("", "")
	s.Network = "LOCAL"
	s.LogPath = "stdout"
	s.Init()
	s.Network = "LOCAL"
	s.CheckChainHeads.CheckChainHeads = false
	state.LoadDatabase(s)
	s.Process()
	s.DBFinished = true
	return s
}

func CreateAndPopulateTestStateAndStartValidator() *state.State {
	s := CreateAndPopulateTestState()
	go s.ValidatorLoop()
	time.Sleep(30 * time.Millisecond)

	return s
}

func CreatePopulateAndExecuteTestState() *state.State {
	s := CreateAndPopulateTestState()
	ExecuteAllBlocksFromDatabases(s)
	go s.ValidatorLoop()
	time.Sleep(30 * time.Millisecond)

	return s
}

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
	s := new(state.State)
	s.TimestampAtBoot = new(primitives.Timestamp)
	s.TimestampAtBoot.SetTime(0)
	s.EFactory = new(electionMsgs.ElectionsFactory)
	s.SetLeaderTimestamp(primitives.NewTimestampFromMilliseconds(0))
	s.DB = CreateAndPopulateTestDatabaseOverlay()
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

	s.Init()
	s.Network = "LOCAL"
	/*err := s.RecalculateBalances()
	if err != nil {
		panic(err)
	}*/
	s.SetFactoshisPerEC(1)
	s.MMRDummy() // Need to start MMR to ensure queues don't fill up
	state.LoadDatabase(s)
	s.Process()
	s.UpdateState()

	return s
}

func GetAllDBStateMsgsFromDatabase(s *state.State) []interfaces.IMsg {
	timestamp := primitives.NewTimestampNow()
	i := uint32(0)
	var msgs []interfaces.IMsg
	for {
		timestamp.SetTimeSeconds(timestamp.GetTimeSeconds() + 60)

		d, err := s.DB.FetchDBlockByHeight(i)
		if err != nil || d == nil {
			break
		}

		a, err := s.DB.FetchABlockByHeight(i)
		if err != nil || a == nil {
			break
		}
		f, err := s.DB.FetchFBlockByHeight(i)
		if err != nil || f == nil {
			break
		}
		ec, err := s.DB.FetchECBlockByHeight(i)
		if err != nil || ec == nil {
			break
		}

		var eblocks []interfaces.IEntryBlock
		var entries []interfaces.IEBEntry

		ebs := d.GetEBlockDBEntries()
		for _, eb := range ebs {
			eblock, _ := s.DB.FetchEBlock(eb.GetKeyMR())
			if eblock != nil {
				eblocks = append(eblocks, eblock)
				for _, e := range eblock.GetEntryHashes() {
					ent, _ := s.DB.FetchEntry(e)
					if ent != nil {
						entries = append(entries, ent)
					}
				}
			}
		}

		dbs := messages.NewDBStateMsg(timestamp, d, a, f, ec, eblocks, entries, nil)
		i++
		msgs = append(msgs, dbs)
	}
	return msgs
}

func ExecuteAllBlocksFromDatabases(s *state.State) {
	msgs := GetAllDBStateMsgsFromDatabase(s)
	for _, dbs := range msgs {
		dbs.(*messages.DBStateMsg).IgnoreSigs = true
		dbs.(*messages.DBStateMsg).IsInDB = true

		s.FollowerExecuteDBState(dbs)
	}
}

func CreateTestBlockCommitList() []interfaces.IMsg {
	answer := make([]interfaces.IMsg, BlockCount)
	var prev *BlockSet = nil

	for i := 0; i < BlockCount; i++ {
		prev = CreateTestBlockSet(prev)

		timestamp := primitives.NewTimestampNow()
		timestamp.SetTime(uint64(i * 1000 * 60 * 60 * 6)) //6 hours of difference between messages

		answer[i] = messages.NewDBStateMsg(timestamp, prev.DBlock, prev.ABlock, prev.FBlock, prev.ECBlock, nil, nil, nil)
	}
	return answer
}

func PopulateTestDatabaseOverlay(dbo *databaseOverlay.Overlay) {
	var prev *BlockSet = nil
	var err error

	for i := 0; i < BlockCount; i++ {
		dbo.StartMultiBatch()
		prev = CreateTestBlockSet(prev)

		err = dbo.ProcessABlockMultiBatch(prev.ABlock)
		if err != nil {
			panic(err)
		}

		err = dbo.ProcessEBlockMultiBatch(prev.EBlock, true)
		if err != nil {
			panic(err)
		}

		err = dbo.ProcessEBlockMultiBatch(prev.AnchorEBlock, true)
		if err != nil {
			panic(err)
		}

		err = dbo.ProcessECBlockMultiBatch(prev.ECBlock, false)
		if err != nil {
			panic(err)
		}

		err = dbo.ProcessFBlockMultiBatch(prev.FBlock)
		if err != nil {
			panic(err)
		}

		err = dbo.ProcessDBlockMultiBatch(prev.DBlock)
		if err != nil {
			panic(err)
		}

		for _, entry := range prev.Entries {
			err = dbo.InsertEntryMultiBatch(entry)
			if err != nil {
				panic(err)
			}
		}

		if err := dbo.ExecuteMultiBatch(); err != nil {
			panic(err)
		}
	}
	/*
		err = dbo.ReparseAnchorChains()
		if err != nil {
			panic(err)
		}
	*/
}

func CreateAndPopulateTestDatabaseOverlay() *databaseOverlay.Overlay {
	dbo := CreateEmptyTestDatabaseOverlay()
	PopulateTestDatabaseOverlay(dbo)
	return dbo
}

type BlockSet struct {
	DBlock       *directoryBlock.DirectoryBlock
	ABlock       *adminBlock.AdminBlock
	ECBlock      interfaces.IEntryCreditBlock
	FBlock       interfaces.IFBlock
	EBlock       *entryBlock.EBlock
	AnchorEBlock *entryBlock.EBlock
	Entries      []*entryBlock.Entry
	Height       int
}

func newBlockSet() *BlockSet {
	bs := new(BlockSet)
	bs.DBlock = nil
	bs.ABlock = nil
	bs.ECBlock = nil
	bs.FBlock = nil
	bs.EBlock = nil
	bs.AnchorEBlock = nil
	bs.Entries = nil
	return bs
}

func CreateFullTestBlockSet() []*BlockSet {
	answer := make([]*BlockSet, BlockCount)
	var prev *BlockSet = nil

	for i := 0; i < BlockCount; i++ {
		prev = CreateTestBlockSet(prev)
		answer[i] = prev
	}

	return answer
}

func CreateTestBlockSet(prev *BlockSet) *BlockSet {
	return CreateTestBlockSetWithNetworkID(prev, constants.LOCAL_NETWORK_ID, true)
}

// Transactions says whether or not to add a transaction
func CreateTestBlockSetWithNetworkID(prev *BlockSet, networkID uint32, transactions bool) *BlockSet {
	var err error
	height := 0
	if prev != nil {
		height = prev.Height + 1
	}

	if prev == nil {
		prev = newBlockSet()
	}
	answer := new(BlockSet)
	answer.Height = height

	dbEntries := []interfaces.IDBEntry{}
	//ABlock
	answer.ABlock = CreateTestAdminBlock(prev.ABlock)

	de := new(directoryBlock.DBEntry)
	de.ChainID, err = primitives.NewShaHash(answer.ABlock.GetChainID().Bytes())
	if err != nil {
		panic(err)
	}
	de.KeyMR = answer.ABlock.DatabasePrimaryIndex()
	dbEntries = append(dbEntries, de)

	//FBlock
	if transactions {
		answer.FBlock = CreateTestFactoidBlock(prev.FBlock)
	} else {
		answer.FBlock = CreateTestFactoidBlockWithCoinbase(prev.FBlock, NewFactoidAddress(0), DefaultCoinbaseAmount)
	}

	de = new(directoryBlock.DBEntry)
	de.ChainID, err = primitives.NewShaHash(answer.FBlock.GetChainID().Bytes())
	if err != nil {
		panic(err)
	}
	de.KeyMR = answer.FBlock.DatabasePrimaryIndex()
	dbEntries = append(dbEntries, de)

	//EBlock
	answer.EBlock, answer.Entries = CreateTestEntryBlock(prev.EBlock)

	de = new(directoryBlock.DBEntry)
	de.ChainID, err = primitives.NewShaHash(answer.EBlock.GetChainID().Bytes())
	if err != nil {
		panic(err)
	}
	de.KeyMR = answer.EBlock.DatabasePrimaryIndex()
	dbEntries = append(dbEntries, de)

	//Anchor EBlock
	anchor, entries := CreateTestAnchorEntryBlock(prev.AnchorEBlock, prev.DBlock)
	answer.AnchorEBlock = anchor
	answer.Entries = append(answer.Entries, entries...)

	de = new(directoryBlock.DBEntry)
	de.ChainID, err = primitives.NewShaHash(answer.AnchorEBlock.GetChainID().Bytes())
	if err != nil {
		panic(err)
	}
	de.KeyMR = answer.AnchorEBlock.DatabasePrimaryIndex()
	dbEntries = append(dbEntries, de)

	//ECBlock
	answer.ECBlock = CreateTestEntryCreditBlock(prev.ECBlock)
	ecEntries := createECEntriesfromBlocks(answer.FBlock, []*entryBlock.EBlock{answer.EBlock, answer.AnchorEBlock}, height)
	answer.ECBlock.GetBody().SetEntries(ecEntries)

	de = new(directoryBlock.DBEntry)
	de.ChainID, err = primitives.NewShaHash(answer.ECBlock.GetChainID().Bytes())
	if err != nil {
		panic(err)
	}
	de.KeyMR = answer.ECBlock.DatabasePrimaryIndex()
	dbEntries = append(dbEntries[:1], append([]interfaces.IDBEntry{de}, dbEntries[1:]...)...)

	answer.DBlock = CreateTestDirectoryBlockWithNetworkID(prev.DBlock, networkID)
	err = answer.DBlock.SetDBEntries(dbEntries)
	if err != nil {
		panic(err)
	}

	return answer
}

func CreateEmptyTestDatabaseOverlay() *databaseOverlay.Overlay {
	return databaseOverlay.NewOverlay(new(mapdb.MapDB), nil)
}

func PrintList(title string, list map[string]uint64) {
	for addr, amt := range list {
		fmt.Printf("%v - %v:%v\n", title, addr, amt)
	}
}

func SystemCall(cmd string) []byte {
	fmt.Println("SystemCall(\"", cmd, "\")")
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		foo := err.Error()
		fmt.Println(foo)
		os.Exit(1)
		panic(err)
	}
	fmt.Print(string(out))
	return out
}

var testNameRe = regexp.MustCompile(`\.Test\w+$`)

// find Test Function name in stack
func GetTestName() (name string) {
	targetFrameIndex := 4 // limit caller frame depth to check for a test name

	programCounters := make([]uintptr, targetFrameIndex+2)
	n := runtime.Callers(0, programCounters)

	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])
		var frameCandidate runtime.Frame
		for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
			frameCandidate, more = frames.Next()
			if testNameRe.MatchString(frameCandidate.Function) {
				return testNameRe.FindString(frameCandidate.Function)[1:]
			}
		}
	}

	return name
}
