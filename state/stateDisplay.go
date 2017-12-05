// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

// The Control Panel needs access to the State, so a deep copy of the elements needed
// will be constructed and sent over a channel. Guards are in place to prevent a full
// channel from hanging. This fixes any concurrency issue on the control panel side.

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/directoryBlock"
	. "github.com/FactomProject/factomd/common/identity"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
)

var ControlPanelAllowedSize int = 2

// This struct will contain all information wanted by the control panel from the state.
type DisplayState struct {
	NodeName string

	ControlPanelPort    int
	ControlPanelSetting int

	// DB Info
	CurrentNodeHeight   uint32
	CurrentLeaderHeight uint32
	CurrentEBDBHeight   uint32
	LeaderHeight        uint32
	LastDirectoryBlock  interfaces.IDirectoryBlock

	// Identity Info
	IdentityChainID interfaces.IHash
	Identities      []*Identity
	Authorities     []*Authority
	PublicKey       *primitives.PublicKey

	// Process List
	PLFactoid []FactoidTransaction
	PLEntry   []EntryTransaction

	// DataDump
	RawSummary   string
	PrintMap     string
	ProcessList  string
	ProcessList2 string
}

type FactoidTransaction struct {
	TxID         string
	Hash         string
	TotalInput   string
	Status       string
	TotalInputs  int
	TotalOutputs int
}

type EntryTransaction struct {
	ChainID   string
	EntryHash string
}

func NewDisplayState() *DisplayState {
	d := new(DisplayState)
	d.Identities = make([]*Identity, 0)
	d.Authorities = make([]*Authority, 0)
	d.PublicKey = new(primitives.PublicKey)
	d.LastDirectoryBlock = nil
	d.PLEntry = make([]EntryTransaction, 0)
	d.PLFactoid = make([]FactoidTransaction, 0)

	return d
}

// Sends the copy of State over channel to control panel
func (s *State) CopyStateToControlPanel() error {
	if !s.ControlPanelDataRequest {
		return nil
	}
	s.ControlPanelDataRequest = false
	if len(s.ControlPanelChannel) < ControlPanelAllowedSize {
		ds, err := DeepStateDisplayCopy(s)
		if err != nil {
			return err
		}
		s.ControlPanelChannel <- *ds
		return nil
	} else {
		return fmt.Errorf("DisplayState Error: Control Panel channel has been filled to maximum allowed size.")
	}
	return fmt.Errorf("DisplayState Error: Reached unreachable code. Impressive")
}

func DeepStateDisplayCopy(s *State) (*DisplayState, error) {
	ds := NewDisplayState()

	ds.NodeName = s.GetFactomNodeName()
	ds.ControlPanelPort = s.ControlPanelPort
	ds.ControlPanelSetting = s.ControlPanelSetting

	// DB Info
	ds.CurrentNodeHeight = s.GetHighestSavedBlk()
	ds.CurrentLeaderHeight = s.GetLeaderHeight()
	ds.CurrentEBDBHeight = s.EntryDBHeightComplete
	ds.LeaderHeight = s.GetTrueLeaderHeight()
	dir := s.GetDirectoryBlockByHeight(s.GetLeaderHeight())
	if dir == nil {
		dir = s.GetDirectoryBlockByHeight(s.GetLeaderHeight() - 1)
	}
	if dir != nil {
		data, err := dir.MarshalBinary()
		if err != nil || dir == nil {
		} else {
			newDBlock, err := directoryBlock.UnmarshalDBlock(data)
			if err != nil {
				ds.LastDirectoryBlock = nil
			} else {
				ds.LastDirectoryBlock = newDBlock
			}
		}
	}

	// Identities
	ds.IdentityChainID = s.GetIdentityChainID().Copy()
	for _, id := range s.Identities {
		ds.Identities = append(ds.Identities, id)
	}
	for _, auth := range s.Authorities {
		ds.Authorities = append(ds.Authorities, auth)
	}
	if pubkey, err := s.GetServerPublicKey().Copy(); err != nil {
	} else {
		ds.PublicKey = pubkey
	}

	vms := s.LeaderPL.VMs
	for _, v := range vms {
		list := v.List
		for _, msg := range list {
			if msg == nil {
				continue
			}
			switch msg.Type() {
			case constants.REVEAL_ENTRY_MSG:
				data, err := msg.MarshalBinary()
				if err != nil {
					continue
				}
				rev := new(messages.RevealEntryMsg)
				err = rev.UnmarshalBinary(data)
				if rev.Entry == nil || err != nil {
					continue
				}

				var entry EntryTransaction
				entry.ChainID = "Processing..."
				entry.EntryHash = rev.Entry.GetHash().String()

				ds.PLEntry = append(ds.PLEntry, entry)
			case constants.FACTOID_TRANSACTION_MSG:
				data, err := msg.MarshalBinary()
				if err != nil {
					continue
				}
				transMsg := new(messages.FactoidTransaction)
				err = transMsg.UnmarshalBinary(data)
				if transMsg.Transaction == nil || err != nil {
					continue
				}
				trans := transMsg.Transaction
				input, err := trans.TotalInputs()
				if err != nil {
					continue
				}
				totalInputs := len(trans.GetInputs())
				totalOutputs := len(trans.GetECOutputs())
				totalOutputs = totalOutputs + len(trans.GetOutputs())
				inputStr := fmt.Sprintf("%f", float64(input)/1e8)

				ds.PLFactoid = append(ds.PLFactoid, struct {
					TxID         string
					Hash         string
					TotalInput   string
					Status       string
					TotalInputs  int
					TotalOutputs int
				}{trans.GetSigHash().String(), trans.GetHash().String(), inputStr, "Process List", totalInputs, totalOutputs})
			}
		}
	}

	prt := "===SummaryStart===\n"
	s.Status = 1
	prt = prt + fmt.Sprintf("%s \n", s.ShortString())
	fnodes := make([]*State, 0)
	fnodes = append(fnodes, s)
	prt = prt + messageLists(fnodes)
	prt = prt + "===SummaryEnd===\n"

	ds.RawSummary = prt

	b := s.GetHighestCompletedBlk() + 1
	pl := s.ProcessLists.Get(b)

	if pl != nil && pl.FedServers != nil {
		ds.PrintMap = pl.PrintMap()
		ds.ProcessList = pl.String()
	} else {
		ds.PrintMap = ""
		ds.ProcessList = ""
	}

	return ds, nil
}

// Used for display dump. Allows a clone of the display state to be made
func (d *DisplayState) Clone() *DisplayState {
	ds := NewDisplayState()

	ds.NodeName = d.NodeName
	ds.ControlPanelPort = d.ControlPanelPort
	ds.ControlPanelSetting = d.ControlPanelSetting

	// DB Info
	ds.CurrentNodeHeight = d.CurrentNodeHeight
	ds.CurrentLeaderHeight = d.CurrentLeaderHeight
	ds.CurrentEBDBHeight = d.CurrentEBDBHeight
	ds.LeaderHeight = d.LeaderHeight

	// Identities
	ds.IdentityChainID = d.IdentityChainID.Copy()
	for _, id := range d.Identities {
		ds.Identities = append(ds.Identities, id)
	}
	for _, auth := range d.Authorities {
		ds.Authorities = append(ds.Authorities, auth)
	}
	if pubkey, err := d.PublicKey.Copy(); err != nil {
	} else {
		ds.PublicKey = pubkey
	}

	ds.RawSummary = d.RawSummary
	ds.PrintMap = d.PrintMap
	ds.ProcessList = d.ProcessList

	return ds
}

// Data Dump String Creation
func messageLists(fnodes []*State) string {
	prt := ""
	list := ""
	fmtstr := "%22s%s\n"
	for i, _ := range fnodes {
		list = list + fmt.Sprintf(" %3d", i)
	}
	prt = prt + fmt.Sprintf(fmtstr, "", list)

	list = ""
	for _, f := range fnodes {
		list = list + fmt.Sprintf(" %3d", len(f.XReview))
	}
	prt = prt + fmt.Sprintf(fmtstr, "Review", list)

	list = ""
	for _, f := range fnodes {
		list = list + fmt.Sprintf(" %3d", len(f.Holding))
	}
	prt = prt + fmt.Sprintf(fmtstr, "Holding", list)

	list = ""
	for _, f := range fnodes {
		list = list + fmt.Sprintf(" %3d", len(f.Acks))
	}
	prt = prt + fmt.Sprintf(fmtstr, "Acks", list)

	prt = prt + "\n"

	list = ""
	for _, f := range fnodes {
		list = list + fmt.Sprintf(" %3d", len(f.MsgQueue()))
	}
	prt = prt + fmt.Sprintf(fmtstr, "MsgQueue", list)

	list = ""
	for _, f := range fnodes {
		list = list + fmt.Sprintf(" %3d", f.InMsgQueue().Length())
	}
	prt = prt + fmt.Sprintf(fmtstr, "InMsgQueue", list)

	list = ""
	for _, f := range fnodes {
		list = list + fmt.Sprintf(" %3d", f.APIQueue().Length())
	}
	prt = prt + fmt.Sprintf(fmtstr, "APIQueue", list)

	list = ""
	for _, f := range fnodes {
		list = list + fmt.Sprintf(" %3d", len(f.AckQueue()))
	}
	prt = prt + fmt.Sprintf(fmtstr, "AckQueue", list)

	prt = prt + "\n"

	list = ""
	for _, f := range fnodes {
		list = list + fmt.Sprintf(" %3d", len(f.TimerMsgQueue()))
	}
	prt = prt + fmt.Sprintf(fmtstr, "TimerMsgQueue", list)

	list = ""
	for _, f := range fnodes {
		list = list + fmt.Sprintf(" %3d", f.NetworkOutMsgQueue().Length())
	}
	prt = prt + fmt.Sprintf(fmtstr, "NetworkOutMsgQueue", list)

	list = ""
	for _, f := range fnodes {
		list = list + fmt.Sprintf(" %3d", len(f.NetworkInvalidMsgQueue()))
	}
	prt = prt + fmt.Sprintf(fmtstr, "NetworkInvalidMsgQueue", list)

	return prt
}
