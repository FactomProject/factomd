package state

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"log"
)

var _ = fmt.Print
var _ = log.Print

type ProcessLists struct {
	State        *State         // Pointer to the state object
	DBHeightBase uint32         // Height of the first Process List in this structure.
	Lists        []*ProcessList // Pointer to the ProcessList structure for each DBHeight under construction

	Acks *map[[32]byte]interfaces.IMsg // acknowlegments by hash
	Msgs *map[[32]byte]interfaces.IMsg // messages by hash

}

func (lists *ProcessLists) LastList() *ProcessList {
	return lists.Lists[len(lists.Lists)-1]
}

// UpdateState is executed from a Follower's perspective.  So the block we are building
// is always the block above the HighestRecordedBlock, but we only care about messages that
// are at the highest known block, as long as that is above the highest recorded block.
func (lists *ProcessLists) UpdateState(dbheight uint32) (progress bool) {

	pl := lists.Get(dbheight)

	// Look and see if we need to toss some previous blocks under construction.
	diff := dbheight - lists.DBHeightBase
	if diff > 1 && len(lists.Lists) > 1 {
		progress = true
		lists.DBHeightBase += (diff - 1)
		var newlist []*ProcessList
		newlist = append(newlist, lists.Lists[(diff-1):]...)
		lists.Lists = newlist
	}
	// Create DState blocks for all completed Process Lists
	p2 := pl.Process(lists.State)
	progress = p2 || progress
	return
}

func (lists *ProcessLists) Get(dbheight uint32) *ProcessList {

	i := int(dbheight) - int(lists.DBHeightBase)

	if i < 0 {
		return nil
	}
	for len(lists.Lists) <= i {
		lists.Lists = append(lists.Lists, nil)
	}
	pl := lists.Lists[i]

	var prev *ProcessList

	if dbheight > 0 {
		prev = lists.Get(dbheight - 1)
	}
	if pl == nil {
		pl = NewProcessList(lists.State, prev, dbheight)
		lists.Lists[i] = pl
	}
	return pl
}

func (lists *ProcessLists) String() string {
	str := "Process Lists"
	for _, pl := range lists.Lists {
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
		panic("Failed to initalize Process Lists because the wrong state object was used")
	}
	pls.State = s
	pls.DBHeightBase = 0
	pls.Lists = make([]*ProcessList, 0)

	pls.Acks = new(map[[32]byte]interfaces.IMsg)
	pls.Msgs = new(map[[32]byte]interfaces.IMsg)

	return pls
}
