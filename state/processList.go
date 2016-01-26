package state

import (
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

var _ = fmt.Print

type ProcessList struct {
	dBHeight 	int								// The directory block height for these lists
	lists       [][]interfaces.IMsg				// Lists of acknowledged messages
	heights     []int							// Height of messages that have been processed
	EomComplete []bool							// Lists that are end of minute complete
	SigComplete []bool							// Lists that are signature complete
	acks        *map[[32]byte] interfaces.IMsg	// acknowlegments by hash
	msgs        *map[[32]byte] interfaces.IMsg	// messages by hash
}

func (p *ProcessList) GetLen(list int) int {
	if list >= len(p.lists) {
		return -1
	}
	l := len(p.lists[list])
	return l
}

// Test if the process list is complete.  Return true if all messages
// have been recieved, and we have all the signaures for the directory blocks.
func (p *ProcessList) Complete() bool {
	if p == nil {
		return true
	}
	for _,c := range p.SigComplete {
		if !c {
			return false
		}
	}
	return true
}

func (p *ProcessList) Process(state interfaces.IState) {
	for i := 0; i < len(p.lists); i++ {
		plist := p.lists[i]
		for j := p.heights[i]; j < len(plist); j++ {
			if plist[j] == nil {
				break
			}
			p.heights[i] = j + 1  // Don't process it again.
			plist[j].Process(state) // Process this entry
			
			eom, ok := plist[j].(*messages.EOM)
			if ok && eom.Minute == 9 {
				p.EomComplete[i]=true
			}
			
			_,ok = plist[j].(*messages.DirectoryBlockSignature)
			if ok {
				p.SigComplete[i]=true
			}
			
		}
	}
}

func (p *ProcessList) AddToProcessList( ack *messages.Ack, m interfaces.IMsg) {
	processlist := p.lists[ack.ServerIndex]
	for len(processlist) <= ack.Height {
		processlist = append(processlist, nil)
	}
	processlist[ack.Height] = m
	p.lists[ack.ServerIndex] = processlist
}


/************************************************
 * Support
 ************************************************/

func NewProcessList(state interfaces.IState) *ProcessList {
	numberServers := state.GetTotalServers()
	
	pl := new(ProcessList)
	pl.lists   = make([][]interfaces.IMsg, numberServers)
	pl.heights = make([]int, numberServers)
	pl.EomComplete = make([]bool,numberServers)
	pl.SigComplete = make([]bool,numberServers)
	pl.dBHeight = int(state.GetDBHeight())
	pl.acks = new(map[[32]byte] interfaces.IMsg)
	pl.msgs = new(map[[32]byte] interfaces.IMsg)
	
	return pl
}

	