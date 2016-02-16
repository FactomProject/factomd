// Copyright 2015 FactomProject Authors. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package state

import (
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/factomd/common/directoryBlock/dbInfo"
	"github.com/FactomProject/factomd/common/interfaces"
	"log"
	"sync"
	"time"
)

var _ = hex.EncodeToString
var _ = fmt.Print
var _ = time.Now()
var _ = log.Print

type DBState struct {
	isNew            bool
	DirectoryBlock   interfaces.IDirectoryBlock
	AdminBlock       interfaces.IAdminBlock
	FactoidBlock     interfaces.IFBlock
	EntryCreditBlock interfaces.IEntryCreditBlock
}

type DBStateList struct {
	multex   *sync.Mutex
	state    *State
	base     uint32
	complete uint32
	DBStates []*DBState
}

func (list *DBStateList) GetDBHeight() uint32 {
	db := list.Last()
	if db == nil {
		return 0
	}
	return db.DirectoryBlock.GetHeader().GetDBHeight()
}

func (list *DBStateList) Length() int {
	return len(list.DBStates)
}

func (list *DBStateList) Last() *DBState {
	last := len(list.DBStates)
	if last == 0 {
		return nil
	}
	return list.DBStates[last-1]
}

func (list *DBStateList) Put(dbstate *DBState) {
	list.multex.Lock()
	defer list.multex.Unlock()

	dblk := dbstate.DirectoryBlock
	dbheight := dblk.GetHeader().GetDBHeight()

	index := int(dbheight) - int(list.base)
	fmt.Println("iiiiiiiiiiiiiiii index", index, int(dbheight), int(list.base), len(list.DBStates))
	for len(list.DBStates) <= index {
		list.DBStates = append(list.DBStates, nil)
	}
	if index >= 0 {
		list.DBStates[index] = dbstate
	}

	hash, _ := dbstate.AdminBlock.GetKeyMR()
	dbstate.DirectoryBlock.GetDBEntries()[0].SetKeyMR(hash)
	hash, _ = dbstate.EntryCreditBlock.Hash()
	dbstate.DirectoryBlock.GetDBEntries()[1].SetKeyMR(hash)
	hash = dbstate.FactoidBlock.GetHash()
	dbstate.DirectoryBlock.GetDBEntries()[2].SetKeyMR(hash)

	if dbheight > 0 {
		prev := list.Getul(dbheight - 1)
		dbstate.DirectoryBlock.GetHeader().SetPrevKeyMR(prev.DirectoryBlock.GetKeyMR())
		dbstate.DirectoryBlock.GetHeader().SetPrevFullHash(prev.DirectoryBlock.GetHash())
	}
	
	
	
	fmt.Println("iiiiiiiiiiiiiiii index", index, int(dbheight), int(list.base), len(list.DBStates))
}

func (list *DBStateList) Get(height uint32) *DBState {
	list.multex.Lock()
	defer list.multex.Unlock()

	return list.Getul(height)
}

func (list *DBStateList) Getul(height uint32) *DBState {
	i := int(height) - int(list.base)
	if i < 0 || i >= len(list.DBStates) {
		return nil
	}
	return list.DBStates[i]
}

func (list *DBStateList) Process() {

	for int(list.complete) < len(list.DBStates) {
		d := list.DBStates[list.complete]

		// Make sure the directory block is properly synced up with the prior block, if there
		// is one.
		if list.complete > 0 {
			p := list.DBStates[list.complete-1]
			d.DirectoryBlock.GetHeader().SetPrevFullHash(p.DirectoryBlock.GetHeader().GetFullHash())
			d.DirectoryBlock.GetHeader().SetPrevKeyMR(p.DirectoryBlock.GetKeyMR())
		}

		if d.isNew {
			fmt.Println("Save new blocks")
			err := list.state.GetDB().ProcessDBlockBatch(d.DirectoryBlock)
			if err != nil {
				panic(err.Error())
			}

			list.state.GetDB().ProcessABlockBatch(d.AdminBlock)

			list.state.GetDB().ProcessFBlockBatch(d.FactoidBlock)

			list.state.GetDB().ProcessECBlockBatch(d.EntryCreditBlock)

			list.state.GetDB().SaveDirectoryBlockHead(d.DirectoryBlock)

			list.state.GetAnchor().UpdateDirBlockInfoMap(dbInfo.NewDirBlockInfoFromDirBlock(d.DirectoryBlock))

		} else {
			fmt.Println("loading into Factom, no save")
			fs := list.state.GetFactoidState()
			fs.AddTransactionBlock(d.FactoidBlock)
			fs.AddECBlock(d.EntryCreditBlock)
		}
		list.complete++
		list.state.DBHeight = list.complete
	}
}

func (list *DBStateList) NewDBState(isNew bool,
	DirectoryBlock interfaces.IDirectoryBlock,
	AdminBlock interfaces.IAdminBlock,
	FactoidBlock interfaces.IFBlock,
	EntryCreditBlock interfaces.IEntryCreditBlock) *DBState {

	fmt.Println("Added new state at height", DirectoryBlock.GetHeader().GetDBHeight())

	dbstate := new(DBState)

	dbstate.isNew = isNew
	dbstate.DirectoryBlock = DirectoryBlock
	dbstate.AdminBlock = AdminBlock
	dbstate.FactoidBlock = FactoidBlock
	dbstate.EntryCreditBlock = EntryCreditBlock

	list.Put(dbstate)
	return dbstate
}
