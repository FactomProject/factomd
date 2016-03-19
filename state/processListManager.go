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


// UpdateState is executed from a Follower's perspective.  So the block we are building
// is always the block above the HighestRecordedBlock.  
func (lists *ProcessLists) UpdateState() {

	buildingBlock := lists.State.GetHighestRecordedBlock()+1
	pl := lists.Get(buildingBlock)
	
	// Look and see if we need to toss some previous blocks under construction.
	diff := buildingBlock - lists.DBHeightBase
	if diff >= 1 && len(lists.Lists)>=2{
		lists.DBHeightBase += (diff)
		lists.Lists = lists.Lists[(diff):]
	}
	lists.State.Println("=================================Process State DBHeight: ",buildingBlock, " base: ", lists.DBHeightBase)
	// Create DState blocks for all completed Process Lists
	pl.Process(lists.State)
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

	if pl == nil {
		pl = NewProcessList(lists.State, 1, dbheight)
		lists.Lists[i] = pl
	}
	return pl
}

func (lists *ProcessLists) String() string {
	str := "Process Lists"
	str = fmt.Sprintf("%s  DBBase: %d\n", str, lists.DBHeightBase)
	for i, pl := range lists.Lists {
		str = fmt.Sprintf("%s ht: %d pl: %s\n", str, uint32(i)+lists.DBHeightBase, pl.String())
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
