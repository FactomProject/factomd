package testHelper

//A package for functions used multiple times in tests that aren't useful in production code.

import (
	"fmt"
	"github.com/FactomProject/factomd/anchor"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/log"
)

func CreateTestEntryBlock(prev interfaces.IEntryBlock) (*entryBlock.EBlock, []*entryBlock.Entry) {
	e := entryBlock.NewEBlock()
	entries := []*entryBlock.Entry{}

	log.Debug("prev - %v\n", prev)

	//if prev.(*entryBlock.EBlock) != nil {
	if prev != nil {
		log.Debug("Prev is not nil - %v\n", prev)
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
		e.Header.SetChainID(GetChainID())

		entry := CreateFirstTestEntry()
		e.AddEBEntry(entry)
		entries = append(entries, entry)
	}

	return e, entries
}

func CreateTestAnchorEntryBlock(prev interfaces.IEntryBlock, prevDBlock *directoryBlock.DirectoryBlock) (*entryBlock.EBlock, []*entryBlock.Entry) {
	if prevDBlock == nil && prev != nil {
		return nil, nil
	}
	e := entryBlock.NewEBlock()
	entries := []*entryBlock.Entry{}

	log.Debug("prev - %v\n", prev)
	if prev != nil {
		log.Debug("Prev is not nil - %v\n", prev)
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
		entry := CreateTestAnchorEnry(prevDBlock)
		e.AddEBEntry(entry)
		entries = append(entries, entry)
	} else {
		e.Header.SetPrevKeyMR(primitives.NewZeroHash())
		e.Header.SetDBHeight(0)
		e.Header.SetChainID(GetAnchorChainID())

		entry := CreateFirstAnchorEntry()
		e.AddEBEntry(entry)
		entries = append(entries, entry)
	}

	return e, entries
}

func GetChainID() interfaces.IHash {
	return CreateFirstTestEntry().GetChainIDHash()
}

func GetAnchorChainID() interfaces.IHash {
	return CreateFirstAnchorEntry().GetChainIDHash()
}

func CreateFirstTestEntry() *entryBlock.Entry {
	answer := new(entryBlock.Entry)

	answer.Version = 1
	answer.ExtIDs = [][]byte{[]byte("Test1"), []byte("Test2")}
	answer.Content = []byte("Test content, please ignore")
	answer.ChainID = entryBlock.NewChainID(answer)

	return answer
}

func CreateFirstAnchorEntry() *entryBlock.Entry {
	answer := new(entryBlock.Entry)

	answer.Version = 0
	answer.ExtIDs = [][]byte{[]byte("FactomAnchorChain")}
	answer.Content = []byte("This is the Factom anchor chain, which records the anchors Factom puts on Bitcoin and other networks.\n")
	answer.ChainID = entryBlock.NewChainID(answer)

	return answer
}

func CreateTestEnry(n uint32) *entryBlock.Entry {
	answer := entryBlock.NewEntry()

	answer.ChainID = GetChainID()
	answer.Version = 1
	answer.ExtIDs = [][]byte{[]byte(fmt.Sprintf("ExtID %v", n))}
	answer.Content = []byte(fmt.Sprintf("Content %v", n))

	return answer
}

func CreateTestAnchorEnry(dBlock *directoryBlock.DirectoryBlock) *entryBlock.Entry {
	answer := entryBlock.NewEntry()

	answer.ChainID = GetAnchorChainID()
	answer.Version = 0
	answer.ExtIDs = nil

	height := dBlock.GetHeader().GetDBHeight()

	ar := anchor.CreateAnchorRecordFromDBlock(dBlock)
	ar.Bitcoin.Address = "1HLoD9E4SDFFPDiYfNYnkBLQ85Y51J3Zb1"
	ar.Bitcoin.TXID = fmt.Sprintf("%x", IntToByteSlice(int(height)))
	ar.Bitcoin.BlockHeight = int32(height)
	ar.Bitcoin.BlockHash = fmt.Sprintf("%x", IntToByteSlice(255-int(height)))
	ar.Bitcoin.Offset = int32(height % 10)

	hex, err := ar.MarshalAndSign(NewPrimitivesPrivateKey(0))
	if err != nil {
		panic(err)
	}

	answer.Content = hex

	return answer
}
