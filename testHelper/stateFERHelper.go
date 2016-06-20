package testHelper

// A package for functions used multiple times in tests that aren't useful in production code.

import (
	//"github.com/FactomProject/factomd/common/adminBlock"
	//"github.com/FactomProject/factomd/common/directoryBlock"
	//"github.com/FactomProject/factomd/common/entryBlock"
	//"github.com/FactomProject/factomd/common/messages"
	//"github.com/FactomProject/factomd/common/primitives"
	//"github.com/FactomProject/factomd/database/databaseOverlay"
	//"github.com/FactomProject/factomd/database/mapdb"
	//"github.com/FactomProject/factomd/engine"
	//"github.com/FactomProject/factomd/log"
	// "github.com/FactomProject/goleveldb/leveldb/errors"
	//"fmt"

	ed "github.com/FactomProject/ed25519"
	"github.com/FactomProject/factom"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/state"

	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"time"
	//"github.com/FactomProject/FactomCode/common"
)

var _ = fmt.Print

type FEREntryWithHeight struct {
	AnFEREntry interfaces.IEBEntry
	Height     uint32
}

func MakeFEREntryWithHeightFromContent(passedResidentHeight uint32, passedTargetActivationHeight uint32,
	passedTargetPrice uint64, passedExpirationHeight uint32, passedPriority uint32) *FEREntryWithHeight {

	// Create and format the signing private key
	var signingPrivateKey [64]byte
	SigningPrivateKey := "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	signingBytes, err := hex.DecodeString(SigningPrivateKey)
	if err != nil {
		fmt.Println("Signing private key isn't parsable")
		return nil
	}
	copy(signingPrivateKey[:], signingBytes[:])
	_ = ed.GetPublicKey(&signingPrivateKey) // Needed to format the public half of the key set

	anFEREntry := new(state.FEREntry)

	anFEREntry.SetExpirationHeight(passedExpirationHeight)
	anFEREntry.SetTargetActivationHeight(passedTargetActivationHeight)
	anFEREntry.SetPriority(passedPriority)
	anFEREntry.SetTargetPrice(passedTargetPrice)

	entryJson, err := json.Marshal(anFEREntry)
	if err != nil {
		fmt.Println("Bad marshal of anFEREntry")
		return nil
	}

	// Create the factom entry with the signing private key
	signingSignature := ed.Sign(&signingPrivateKey, entryJson)

	// Make a new factom entry and populate it
	anEntry := new(factom.Entry)
	anEntry.ChainID = "eac57815972c504ec5ae3f9e5c1fe12321a3c8c78def62528fb74cf7af5e7389"
	anEntry.ExtIDs = append(anEntry.ExtIDs, signingSignature[:])
	anEntry.Content = entryJson

	// ce := common.NewEntry()
	emb, _ := anEntry.MarshalBinary()
	// ce.UnmarshalBinary(emb)

	EBEntry := entryBlock.NewEntry()
	_, err = EBEntry.UnmarshalBinaryData(emb)
	if err != nil {
		fmt.Println("Error 3:  couldn't unmarshal binary")
		return nil
	}

	ewh := new(FEREntryWithHeight)
	// Don't set the resident height in the actual FEREntry yet because the state validate loop will handle all that
	ewh.Height = passedResidentHeight
	ewh.AnFEREntry = EBEntry

	return ewh
}

func CreateAndPopulateTestStateForFER(testEntries []FEREntryWithHeight, desiredHeight int) *state.State {

	s := new(state.State)
	s.DB = CreateAndPopulateTestDatabaseOverlayForFER(testEntries, desiredHeight)
	s.LoadConfig("", "")
	s.Init()
	/*err := s.RecalculateBalances()
	if err != nil {
		panic(err)
	}*/
	s.SetFactoshisPerEC(1)
	state.LoadDatabase(s)
	s.FERChainId = "eac57815972c504ec5ae3f9e5c1fe12321a3c8c78def62528fb74cf7af5e7389"
	s.UpdateState()
	go s.ValidatorLoop()
	time.Sleep(20 * time.Millisecond)

	return s
}

func CreateAndPopulateTestDatabaseOverlayForFER(testEntries []FEREntryWithHeight, desiredHeight int) *databaseOverlay.Overlay {
	dbo := CreateEmptyTestDatabaseOverlay()

	var prev *BlockSet = nil
	var currentBlockSet *BlockSet = nil

	var err error

	if desiredHeight <= 0 {
		desiredHeight = 1
	}

	for i := 0; i < desiredHeight; i++ {

		fmt.Println("Making block number ", i)
		dbo.StartMultiBatch()
		currentBlockSet = CreateTestBlockSetForFER(prev, dbo, testEntries)

		err = dbo.ProcessABlockMultiBatch(currentBlockSet.ABlock)
		if err != nil {
			panic(err)
		}

		err = dbo.ProcessEBlockMultiBatch(currentBlockSet.EBlock, false)
		if err != nil {
			panic(err)
		}

		err = dbo.ProcessEBlockMultiBatch(currentBlockSet.AnchorEBlock, false)
		if err != nil {
			panic(err)
		}

		err = dbo.ProcessECBlockMultiBatch(currentBlockSet.ECBlock, false)
		if err != nil {
			panic(err)
		}

		err = dbo.ProcessFBlockMultiBatch(currentBlockSet.FBlock)
		if err != nil {
			panic(err)
		}

		err = dbo.ProcessDBlockMultiBatch(currentBlockSet.DBlock)
		if err != nil {
			panic(err)
		}

		for _, entry := range currentBlockSet.Entries {
			err = dbo.InsertEntry(entry)
			if err != nil {
				panic(err)
			}
		}
		prev = currentBlockSet

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

func CreateTestBlockSetForFER(prev *BlockSet, db *databaseOverlay.Overlay, testEntries []FEREntryWithHeight) *BlockSet {
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
	de.KeyMR, err = answer.ABlock.GetKeyMR()
	if err != nil {
		panic(err)
	}
	dbEntries = append(dbEntries, de)

	//FBlock
	answer.FBlock = CreateTestFactoidBlock(prev.FBlock)

	de = new(directoryBlock.DBEntry)
	de.ChainID, err = primitives.NewShaHash(answer.FBlock.GetChainID().Bytes())
	if err != nil {
		panic(err)
	}
	de.KeyMR = answer.FBlock.GetKeyMR()
	dbEntries = append(dbEntries, de)

	//EBlock
	answer.EBlock, answer.Entries = CreateTestEntryBlockForFER(prev.EBlock, uint32(height))

	//  Loop through the passed FEREntries and see which ones need to go into this EBlock
	for _, testEntry := range testEntries {
		if testEntry.Height == uint32(height) {
			answer.EBlock.AddEBEntry(testEntry.AnFEREntry)
			// db.InsertEntry(testEntry.AnFEREntry)    // I don't think I need this
		}
	}

	de = new(directoryBlock.DBEntry)
	de.ChainID, err = primitives.NewShaHash(answer.EBlock.GetChainID().Bytes())
	if err != nil {
		panic(err)
	}
	de.KeyMR, err = answer.EBlock.KeyMR()
	if err != nil {
		panic(err)
	}

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
	de.KeyMR, err = answer.AnchorEBlock.KeyMR()
	if err != nil {
		panic(err)
	}
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
	de.KeyMR, err = answer.ECBlock.GetFullHash()
	if err != nil {
		panic(err)
	}
	dbEntries = append(dbEntries[:1], append([]interfaces.IDBEntry{de}, dbEntries[1:]...)...)

	answer.DBlock = CreateTestDirectoryBlock(prev.DBlock)
	err = answer.DBlock.SetDBEntries(dbEntries)
	if err != nil {
		panic(err)
	}

	return answer
}

func CreateTestEntryBlockForFER(p interfaces.IEntryBlock, height uint32) (*entryBlock.EBlock, []*entryBlock.Entry) {
	prev, ok := p.(*entryBlock.EBlock)
	if ok == false {
		prev = nil
	}

	e := entryBlock.NewEBlock()
	entries := []*entryBlock.Entry{}

	if prev != nil {
		keyMR, err := prev.KeyMR()
		if err != nil {
			panic(err)
		}

		e.Header.SetPrevKeyMR(keyMR)
		hash, err := prev.Hash()
		if err != nil {
			panic(err)
		}
		e.Header.SetPrevFullHash(hash)
		e.Header.SetDBHeight(prev.GetHeader().GetDBHeight() + 1)

		e.Header.SetChainID(prev.GetHeader().GetChainID())
		entry := CreateTestEnry(e.Header.GetDBHeight())
		e.AddEBEntry(entry)
		entries = append(entries, entry)
	} else {
		e.Header.SetPrevKeyMR(primitives.NewZeroHash())
		e.Header.SetDBHeight(0)
		chainId := "eac57815972c504ec5ae3f9e5c1fe12321a3c8c78def62528fb74cf7af5e7389"
		hexBytes, _ := hex.DecodeString(chainId)
		chainIdHash := primitives.NewHash(hexBytes)
		e.Header.SetChainID(chainIdHash)

		entry := CreateFirstTestEntry()
		e.AddEBEntry(entry)
		entries = append(entries, entry)
	}

	return e, entries
}
