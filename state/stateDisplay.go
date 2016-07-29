package state

// The Control Panel needs access to the State, so a deep copy of the elements needed
// will be constructed and sent over a channel. Guards are in place to prevent a full
// channel from hanging. This fixes any concurrency issue on the control panel side.

import (
	"fmt"

	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

var ControlPanelAllowedSize int = 10

// This struct will contain all information wanted by the control panel from the state.
type DisplayState struct {
	NodeName string

	ControlPanelPort    int
	ControlPanelPath    string
	ControlPanelSetting int

	// DB Info
	CurrentNodeHeight   uint32
	CurrentLeaderHeight uint32
	CurrentEBDBHeight   uint32
	LastDirectoryBlock  interfaces.IDirectoryBlock

	// Identity Info
	IdentityChainID interfaces.IHash
	Identities      []Identity  // Identities of all servers in management chain
	Authorities     []Authority // Identities of all servers in management chain
	PublicKey       *primitives.PublicKey
}

func NewDisplayState() *DisplayState {
	d := new(DisplayState)
	d.Identities = make([]Identity, 0)
	d.Authorities = make([]Authority, 0)
	d.PublicKey = new(primitives.PublicKey)
	d.LastDirectoryBlock = nil

	return d
}

func (s *State) CopyStateToControlPanel() error {
	if s.ControlPanelSetting == 2 {
		return nil
	}
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
	ds.ControlPanelPath = s.ControlPanelPath
	ds.ControlPanelSetting = s.ControlPanelSetting

	// DB Info
	ds.CurrentNodeHeight = s.GetHighestRecordedBlock()
	ds.CurrentLeaderHeight = s.GetLeaderHeight()
	ds.CurrentEBDBHeight = s.GetEBDBHeightComplete()
	if dir := s.GetDirectoryBlock(); dir != nil {
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

	return ds, nil
}

func (d *DisplayState) Clone() *DisplayState {
	ds := NewDisplayState()

	ds.NodeName = d.NodeName
	ds.ControlPanelPort = d.ControlPanelPort
	ds.ControlPanelPath = d.ControlPanelPath
	ds.ControlPanelSetting = d.ControlPanelSetting

	// DB Info
	ds.CurrentNodeHeight = d.CurrentNodeHeight
	ds.CurrentLeaderHeight = d.CurrentLeaderHeight
	ds.CurrentEBDBHeight = d.CurrentEBDBHeight

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

	return ds
}
