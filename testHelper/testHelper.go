package testHelper

//A package for functions used multiple times in tests that aren't useful in production code.

import (
	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/mapdb"
	//"github.com/FactomProject/factomd/engine"
	//"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/state"
	"time"
	//"fmt"
)

var BlockCount int = 10
var DefaultCoinbaseAmount uint64 = 100000000

func CreateEmptyTestState() *state.State {
	s := new(state.State)
	s.LoadConfig("", "")
	s.Init()
	state.LoadDatabase(s)
	return s
}

func CreateAndPopulateTestState() *state.State {
	s := new(state.State)
	s.DB = CreateAndPopulateTestDatabaseOverlay()
	s.LoadConfig("", "")
	s.Init()
	/*err := s.RecalculateBalances()
	if err != nil {
		panic(err)
	}*/
	s.SetFactoshisPerEC(1)
	state.LoadDatabase(s)
	s.UpdateState()
	go s.ValidatorLoop()
	time.Sleep(30 * time.Millisecond)

	return s
}

func CreateTestDBStateList() []interfaces.IMsg {
	answer := make([]interfaces.IMsg, BlockCount)
	var prev *BlockSet = nil

	for i := 0; i < BlockCount; i++ {
		prev = CreateTestBlockSet(prev)

		timestamp := primitives.NewTimestampNow()
		timestamp.SetTime(uint64(i * 1000 * 60 * 60 * 6)) //6 hours of difference between messages

		answer[i] = messages.NewDBStateMsg(timestamp, prev.DBlock, prev.ABlock, prev.FBlock, prev.ECBlock)
	}
	return answer
}

func CreateTestLogFileString() string {
	messages := CreateTestDBStateList()
	answer := ""
	st := CreateEmptyTestState()
	for _, v := range messages {
		s := st.MessageToLogString(v)
		answer = answer + s
	}
	return answer
}

func CreateAndPopulateTestDatabaseOverlay() *databaseOverlay.Overlay {
	dbo := CreateEmptyTestDatabaseOverlay()

	var prev *BlockSet = nil

	var err error

	for i := 0; i < BlockCount; i++ {
		dbo.StartMultiBatch()
		prev = CreateTestBlockSet(prev)

		err = dbo.ProcessABlockMultiBatch(prev.ABlock)
		if err != nil {
			panic(err)
		}

		err = dbo.ProcessEBlockMultiBatch(prev.EBlock, false)
		if err != nil {
			panic(err)
		}

		err = dbo.ProcessEBlockMultiBatch(prev.AnchorEBlock, false)
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
			err = dbo.InsertEntry(entry)
			if err != nil {
				panic(err)
			}
		}

		if err := dbo.ExecuteMultiBatch(); err != nil {
			panic(err)
		}
	}

	err = dbo.RebuildDirBlockInfo()
	if err != nil {
		panic(err)
	}

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
	answer.FBlock = CreateTestFactoidBlock(prev.FBlock)

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

	answer.DBlock = CreateTestDirectoryBlock(prev.DBlock)
	err = answer.DBlock.SetDBEntries(dbEntries)
	if err != nil {
		panic(err)
	}

	return answer
}

func CreateEmptyTestDatabaseOverlay() *databaseOverlay.Overlay {
	return databaseOverlay.NewOverlay(new(mapdb.MapDB))
}

func CreateTestAdminBlock(prev *adminBlock.AdminBlock) *adminBlock.AdminBlock {
	block := new(adminBlock.AdminBlock)
	block.SetHeader(CreateTestAdminHeader(prev))
	block.GetHeader().SetMessageCount(uint32(len(block.GetABEntries())))
	return block
}

func CreateTestAdminHeader(prev *adminBlock.AdminBlock) *adminBlock.ABlockHeader {
	header := new(adminBlock.ABlockHeader)

	if prev == nil {
		header.PrevFullHash = primitives.NewZeroHash()
		header.DBHeight = 0
	} else {
		keyMR, err := prev.GetKeyMR()
		if err != nil {
			panic(err)
		}
		header.PrevFullHash = keyMR
		header.DBHeight = prev.Header.GetDBHeight() + 1
	}

	header.HeaderExpansionSize = 5
	header.HeaderExpansionArea = []byte{0x00, 0x01, 0x02, 0x03, 0x04}
	header.MessageCount = 0
	header.BodySize = 0

	return header
}

func CreateTestDirectoryBlock(prevBlock *directoryBlock.DirectoryBlock) *directoryBlock.DirectoryBlock {
	dblock := new(directoryBlock.DirectoryBlock)

	dblock.SetHeader(CreateTestDirectoryBlockHeader(prevBlock))

	de := new(directoryBlock.DBEntry)
	de.ChainID = primitives.NewZeroHash()
	de.KeyMR = primitives.NewZeroHash()

	err := dblock.SetDBEntries(append(make([]interfaces.IDBEntry, 0, 5), de))
	if err != nil {
		panic(err)
	}
	//dblock.GetHeader().SetBlockCount(uint32(len(dblock.GetDBEntries())))

	return dblock
}

func CreateTestDirectoryBlockHeader(prevBlock *directoryBlock.DirectoryBlock) *directoryBlock.DBlockHeader {
	header := new(directoryBlock.DBlockHeader)

	header.SetBodyMR(primitives.Sha(primitives.NewZeroHash().Bytes()))
	header.SetBlockCount(0)
	header.SetNetworkID(0xffff)

	if prevBlock == nil {
		header.SetDBHeight(0)
		header.SetPrevFullHash(primitives.NewZeroHash())
		header.SetPrevKeyMR(primitives.NewZeroHash())
		header.SetTimestamp(primitives.NewTimestampFromMinutes(1234))
	} else {
		header.SetDBHeight(prevBlock.Header.GetDBHeight() + 1)
		header.SetPrevFullHash(prevBlock.GetHash())
		keyMR, err := prevBlock.BuildKeyMerkleRoot()
		if err != nil {
			panic(err)
		}
		header.SetPrevKeyMR(keyMR)
		header.SetTimestamp(primitives.NewTimestampFromMinutes(prevBlock.Header.GetTimestamp().GetTimeMinutesUInt32() + 1))
	}

	header.SetVersion(1)

	return header
}
