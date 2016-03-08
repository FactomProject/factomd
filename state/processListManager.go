package state

import (
	"fmt"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"log"
)

var _ = fmt.Print
var _ = log.Print

type ProcessLists struct {
	State        *State         // Pointer to the state object
	DBHeightBase uint32         // Height of the first Process List in this structure.
	Lists        []*ProcessList // Pointer to the ProcessList structure for each DBHeight under construction
}

// Returns the height of the Process List under construction.  There
// can be another list under construction (because of missing messages.), but
// this is the one just above the DBStates list.
//
// Note if we are waiting on DBStates, this routine returns zero
func (lists *ProcessLists) GetBuildingBlock() uint32 {
	
	highestRecordedBlock := lists.State.GetHighestRecordedBlock()
	if lists.DBHeightBase != highestRecordedBlock-1 { 
		return 0
	}
	
	last := lists.DBHeightBase
	for _,list := range lists.Lists {
		if list.Complete() {
			last++
		}
	}
	return last
}

// The highest block for which we have received a message.  Sometimes the same as
// BuildingBlock(), but can be different depending or the order messages are recieved.
func (lists *ProcessLists) GetHighestKnownBlock() 		uint32 {
	last := lists.DBHeightBase
	if last == 0 { 
		return 0
	}
	last -= 1
	for _,list := range lists.Lists {
		if list.HasMessage() {
			last++
		}
	}
	return last
}

func (lists *ProcessLists) UpdateState() {

	buildingBlock := lists.GetBuildingBlock()
	
	if buildingBlock == 0 {
		return
	}
	
	pl := lists.Get(buildingBlock)

	diff := buildingBlock - lists.DBHeightBase
	if diff > 0 {
		lists.DBHeightBase += diff
		lists.Lists = lists.Lists[diff:]
	}

	//*******************************************************************//
	// Do initialization of blocks for the next Process List level here
	//*******************************************************************
	if pl.DirectoryBlock == nil {
		dbstate := lists.State.DBStates.Last()
		pl.DirectoryBlock = directoryBlock.NewDirectoryBlock(buildingBlock, nil)
		pl.FactoidBlock = lists.State.GetFactoidState().GetCurrentBlock()
		pl.AdminBlock = lists.State.NewAdminBlock(buildingBlock)
		var err error
		pl.EntryCreditBlock, err = entryCreditBlock.NextECBlock(dbstate.EntryCreditBlock)
		if err != nil {
			panic(err.Error())
		}

	}
	// Create DState blocks for all completed Process Lists
	pl.Process(lists.State)

	// Only when we are sig complete that we can move on.
	if pl.Complete() {
		lists.State.DBStates.NewDBState(true, pl.DirectoryBlock, pl.AdminBlock, pl.FactoidBlock, pl.EntryCreditBlock)
		pln := lists.Get(buildingBlock + 1)
		for _, srv := range pl.FedServers { // Bring forward the current Federated Servers
			pln.AddFedServer(srv.(*interfaces.Server))
		}
	}
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

	return pls
}
