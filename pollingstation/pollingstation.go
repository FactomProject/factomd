package pollingstation

import (
	//"github.com/FactomProject/electiontesting/election"
	//"github.com/FactomProject/electiontesting/imessage"
	"github.com/FactomProject/electiontesting/messages"
	"github.com/FactomProject/electiontesting/primitives"
	//"github.com/FactomProject/electiontesting/util"
)

const (
	_ int = iota
	PollingStationState_Active
)

// PollingStation controls minutes 0 to 10 (where 0 & 10 is dbsigs). It needs all dbsigs from the previous
// PollingStation to begin
type PollingStation struct {
	Height     int
	PrevDBSigs map[primitives.Hash]map[primitives.Identity]messages.DbsigMessage

	Self  primitives.Identity
	State int
}

// NewPollingStation takes the initial authset that starts the block
func NewPollingStation(identity primitives.Identity, authset primitives.AuthSet, height int,
	PrevDBSigs map[primitives.Hash]map[primitives.Identity]messages.DbsigMessage) *PollingStation {
	p := new(PollingStation)

	return p
}
