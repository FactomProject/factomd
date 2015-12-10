package testHelper

//A package for functions used multiple times in tests that aren't useful in production code.

import (
	"fmt"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

func CreateTestEntryBlock(prev *entryBlock.EBlock) (*entryBlock.EBlock, []*entryBlock.Entry) {
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
		e.Header.SetPrevLedgerKeyMR(hash)
		e.Header.SetDBHeight(prev.Header.GetDBHeight() + 1)

		e.Header.SetChainID(prev.Header.GetChainID())
		entry := CreateTestEnry(e.Header.GetDBHeight())
		e.AddEBEntry(entry)
		entries = append(entries, entry)
	} else {
		e.Header.SetPrevKeyMR(primitives.NewZeroHash())
		e.Header.SetDBHeight(0)
		e.Header.SetChainID(getChainID())

		entry := CreateFirstTestEntry()
		e.AddEBEntry(entry)
		entries = append(entries, entry)
	}

	return e, entries
}

func getChainID() interfaces.IHash {
	return CreateFirstTestEntry().GetChainIDHash()
}

func CreateFirstTestEntry() *entryBlock.Entry {
	answer := new(entryBlock.Entry)

	answer.Version = 1
	answer.ExtIDs = [][]byte{[]byte("Test1"), []byte("Test2")}
	answer.Content = []byte("Test content, please ignore")
	answer.ChainID = entryBlock.NewChainID(answer)

	return answer
}

func CreateTestEnry(n uint32) *entryBlock.Entry {
	answer := entryBlock.NewEntry()

	answer.ChainID = getChainID()
	answer.Version = 1
	answer.ExtIDs = [][]byte{[]byte(fmt.Sprintf("ExtID %v", n))}
	answer.Content = []byte(fmt.Sprintf("Content %v", n))

	return answer
}
