package blockchainState

import (
	"fmt"
)

var Expired int = 0
var LatestReveal int = 0
var TotalEntries int = 0

var MES *MissingEntries

func init() {
	MES = new(MissingEntries)
	MES.missing = map[string]*Entry{}
	MES.found = []*Entry{}
}

type MissingEntries struct {
	missing map[string]*Entry
	found   []*Entry
}

func (me *MissingEntries) IsEntryMissing(h string) bool {
	if me.missing[h] != nil {
		return true
	}
	return false
}

func (me *MissingEntries) NewMissing(entryID, dBlock string, height uint32) {
	fmt.Printf("New missing - %v\n", entryID)
	if me.missing[entryID] != nil {
		panic("Duplicate missing entry " + entryID)
	}
	e := new(Entry)
	e.EntryID = entryID
	e.EntryDBlock = dBlock
	e.EntryDBlockHeight = height

	me.missing[entryID] = e
}

func (me *MissingEntries) FoundMissing(entryID, commitID, dBlock string, height uint32) {
	fmt.Printf("Found missing - %v\n", entryID)
	e := me.missing[entryID]
	if e == nil {
		panic("Found non-missing entry! " + entryID)
	}
	e.CommitID = commitID
	e.CommitDBlock = dBlock
	e.CommitHeight = height

	me.found = append(me.found, e)
	//delete(me.missing, entryID)
}

func (me *MissingEntries) Print() {
	fmt.Printf("Missing:\n")
	for _, v := range me.missing {
		fmt.Printf("%v\t%v\t%v\n", v.EntryID, v.EntryDBlock, v.EntryDBlockHeight)
	}
	fmt.Printf("Found:\n")
	for _, v := range me.found {
		fmt.Printf("%v\t%v\t%v\t%v\t%v\t%v\n", v.EntryID, v.EntryDBlock, v.EntryDBlockHeight, v.CommitID, v.CommitDBlock, v.CommitHeight)
	}
}

type Entry struct {
	EntryID           string
	EntryDBlock       string
	EntryDBlockHeight uint32

	CommitID     string
	CommitDBlock string
	CommitHeight uint32
}
