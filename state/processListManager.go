// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
)

type ProcessLists struct {
	State        *State         // Pointer to the state object
	DBHeightBase uint32         // Height of the first Process List in this structure.
	Lists        []*ProcessList // Pointer to the ProcessList structure for each DBHeight under construction
	SetString    bool
	Str          string
}

func (lists *ProcessLists) LastList() *ProcessList {
	return lists.Lists[len(lists.Lists)-1]
}

// UpdateState is executed from a Follower's perspective.  So the block we are building
// is always the block above the HighestRecordedBlock, but we only care about messages that
// are at the highest known block, as long as that is above the highest recorded block.
func (lists *ProcessLists) UpdateState(dbheight uint32) (progress bool) {
	// Look and see if we need to toss some previous blocks under construction.
	diff := int(dbheight) - int(lists.DBHeightBase)
	//TODO: Maybe the test about len(lists.list) is pointless
	if diff > 1 && len(lists.Lists) > 1 {
		diff = diff - 1
		progress = true
		lists.DBHeightBase += uint32(diff)

		newlist := append([]*ProcessList{}, lists.Lists[diff:]...)
		lists.Lists = newlist
	}
	dbstate := lists.State.DBStates.Get(int(dbheight))
	pl := lists.Get(lists.State.LLeaderHeight)
	for pl.Complete() || (dbstate != nil && dbstate.Locked && dbstate.Signed) {
		dbheight++
		pl = lists.Get(dbheight)
		lists.State.MoveStateToHeight(dbheight, 0)
		dbstate = lists.State.DBStates.Get(int(dbheight))
	}
	if pl == nil {
		return false
	}
	if dbheight > lists.State.LLeaderHeight {
		s := lists.State
		//fmt.Println(fmt.Sprintf("EOM PROCESS: %10s ProcessListManager: !s.EOM(%v)", s.FactomNodeName, s.EOM))

		s.MoveStateToHeight(dbheight, 0)
	}
	lists.State.AddStatus(fmt.Sprintf("UpdateState: ProcessList Height %d", pl.DBHeight))
	return pl.Process(lists.State)
}

// Only gets an existing process list
func (lists *ProcessLists) GetSafe(dbheight uint32) (pl *ProcessList) {
	var i int

	getindex := func() bool {
		i = int(dbheight) - int(lists.DBHeightBase)

		if i < 0 {
			return false
		}
		if len(lists.Lists) <= i {
			return false
		}
		return true
	}
	if getindex() {
		return lists.Lists[i]
	}
	return nil
}

func (lists *ProcessLists) Get(dbheight uint32) *ProcessList {
	if dbheight < lists.DBHeightBase {
		return nil
	}

	i := int(dbheight) - int(lists.DBHeightBase)

	s := lists.State
	_ = s

	// Only allocate a pl I have a hope of using. If too high, ignore.
	highestCompletedBlk := lists.State.GetHighestCompletedBlk()
	if dbheight >= highestCompletedBlk+200 {
		return nil
	}
	//TODO: Actually allocate the PL here !!!
	for len(lists.Lists) <= i {
		lists.Lists = append(lists.Lists, nil)
	}

	pl := lists.Lists[i]

	var prev *ProcessList

	if pl == nil {
		if dbheight == 0 {
			prev = nil
		} else {
			prev = lists.Get(dbheight - 1)
		}
		pl = NewProcessList(lists.State, prev, dbheight)
		lists.Lists[i] = pl
	}

	return pl
}

func (lists *ProcessLists) String() string {
	str := "Process Lists"
	for i, pl := range lists.Lists {
		if pl == nil {
			continue
		}
		if len(lists.Lists)-i > 3 {
			continue
		}
		str = fmt.Sprintf("%s  DBBase: %d\n", str, lists.DBHeightBase)
		str = fmt.Sprintf("%s ht: %d pl: %s\n", str, pl.DBHeight, pl.String())
	}
	return str
}

/************************************************
 * Support
 ************************************************/

func NewProcessLists(state interfaces.IState) *ProcessLists {
	pls := new(ProcessLists)

	s, ok := state.(*State)
	if !ok {
		panic("Failed to initialize Process Lists because the wrong state object was used")
	}
	pls.State = s
	pls.DBHeightBase = 0
	pls.Lists = make([]*ProcessList, 0)

	return pls
}
