// Copyright 2015 FactomProject Authors. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package state

import (
	"fmt"
	"time"
	"github.com/FactomProject/factomd/common/interfaces"
)

var _ = time.Now()
var _ = fmt.Print

type DBState struct {
	isNew				bool
	DirectoryBlock		interfaces.IDirectoryBlock
	AdminBlock			interfaces.IAdminBlock
	FactoidBlock		interfaces.IFBlock
	EntryCreditBlock	interfaces.IEntryCreditBlock
}

func(d *DBState) Process(state interfaces.IState) {
	fs := state.GetFactoidState(d.DirectoryBlock.GetHeader().GetDBHeight())
	fs.AddTransactionBlock(d.FactoidBlock)
	fs.AddECBlock(d.EntryCreditBlock)

	state.AddAdminBlock(d.AdminBlock)
	
	if d.isNew {
		state.GetDB().ProcessDBlockBatch(d.DirectoryBlock)
		state.GetDB().ProcessABlockBatch(d.AdminBlock)
		state.GetDB().ProcessFBlockBatch(d.FactoidBlock)
		state.GetDB().ProcessECBlockBatch(d.EntryCreditBlock)
	}
}
