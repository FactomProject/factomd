// Copyright 2015 FactomProject Authors. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package state

import (
	"log"
	"fmt"
	"sync"
	"github.com/FactomProject/factomd/common/directoryBlock/dbInfo"
	"github.com/FactomProject/factomd/common/interfaces"
	"time"
)

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
	multex		*sync.Mutex
	base 		uint32
	complete    uint32
	DBStates 	[] *DBState
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

func (list *DBStateList) Put (dbstate *DBState) {
	list.multex.Lock()	
	defer list.multex.Unlock()
	
	dbheight := dbstate.DirectoryBlock.GetHeader().GetDBHeight()
		
	index := int(dbheight)-int(list.base) 
	for len(list.DBStates) <= index {
		list.DBStates = append(list.DBStates,nil)
	}
	if index >= 0 {
		list.DBStates[index]=dbstate
	}
}
							  
		
func (list *DBStateList) Get(height uint32) *DBState {
	list.multex.Lock()	
	defer list.multex.Unlock()
	
	i := int(height)-int(list.base)
	if i < 0 || i >= len(list.DBStates) {
		return nil
	}
	return list.DBStates[i]
}

func (list *DBStateList) Process(state interfaces.IState) {
	
	for int(list.complete+1) < len(list.DBStates) {
		d := list.DBStates[list.complete+1]
		
		s := state.(*State)
		s.AddAdminBlock(d.AdminBlock)
		
		if d.isNew {
			state.GetDB().SaveDirectoryBlockHead(d.DirectoryBlock)
			state.GetDB().ProcessDBlockBatch(d.DirectoryBlock)
			state.GetDB().ProcessABlockBatch(d.AdminBlock)
			state.GetDB().ProcessFBlockBatch(d.FactoidBlock)
			state.GetDB().ProcessECBlockBatch(d.EntryCreditBlock)
			state.GetAnchor().UpdateDirBlockInfoMap(dbInfo.NewDirBlockInfoFromDirBlock(d.DirectoryBlock))
		}else{
			fs := state.GetFactoidState()
			fs.AddTransactionBlock(d.FactoidBlock)
			fs.AddECBlock(d.EntryCreditBlock)
		}
		
		list.complete++
	}
}

/****************************************
 * Support
 ****************************************/

func NewDBState(isNew            bool,
				DirectoryBlock   interfaces.IDirectoryBlock,
				AdminBlock       interfaces.IAdminBlock,
				FactoidBlock     interfaces.IFBlock,
				EntryCreditBlock interfaces.IEntryCreditBlock) *DBState {
	dbstate := new(DBState)

	dbstate.isNew	 		 = isNew
	dbstate.DirectoryBlock	 = DirectoryBlock
	dbstate.AdminBlock       = AdminBlock
	dbstate.FactoidBlock     = FactoidBlock
	dbstate.EntryCreditBlock = EntryCreditBlock
	fmt.Println(DirectoryBlock)
	return dbstate
}
