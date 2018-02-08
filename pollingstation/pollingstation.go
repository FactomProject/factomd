package pollingstation

import (
	//"github.com/FactomProject/electiontesting/election"
	"github.com/FactomProject/electiontesting/imessage"
	"github.com/FactomProject/electiontesting/messages"
	"github.com/FactomProject/electiontesting/primitives"
	"github.com/FactomProject/electiontesting/util"
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

	NumberOfFeds int
	// Each minute has an original authset, and it's new one after elections
	OriginalAuthSet []*primitives.AuthSet
	NewAuthSet      []*primitives.AuthSet

	// A collection of DBSigs for each prevHash
	DBSigs map[primitives.Hash]map[primitives.Identity]messages.DbsigMessage

	// If we a set of trump eom's for a vm/min we always send that back
	TrumpEOMs [][]*messages.EomMessage
	// This EOM's are subject to change
	CurrentEOMs [][]*messages.EomMessage
	// What EOM we are currently working on
	WorkingEOM int

	Self  primitives.Identity
	State int
}

// NewPollingStation takes the initial authset that starts the block
func NewPollingStation(identity primitives.Identity, authset primitives.AuthSet, height int,
	PrevDBSigs map[primitives.Hash]map[primitives.Identity]messages.DbsigMessage) *PollingStation {
	p := new(PollingStation)
	p.OriginalAuthSet = make([]*primitives.AuthSet, primitives.NumberOfMinutes+1)
	p.NewAuthSet = make([]*primitives.AuthSet, primitives.NumberOfMinutes+1)
	// 0th index is the dbsig level
	p.OriginalAuthSet[0] = &authset
	p.OriginalAuthSet[1] = &authset
	p.NewAuthSet[0] = &authset

	p.DBSigs = make(map[primitives.Hash]map[primitives.Identity]messages.DbsigMessage)
	p.Self = identity
	p.WorkingEOM = 1

	p.NumberOfFeds = len(authset.GetFeds())
	p.TrumpEOMs = make([][]*messages.EomMessage, primitives.NumberOfMinutes+1, p.NumberOfFeds)
	p.CurrentEOMs = make([][]*messages.EomMessage, primitives.NumberOfMinutes+1, p.NumberOfFeds)
	return p
}

func (p *PollingStation) Execute(msg imessage.IMessage) []imessage.IMessage {
	switch msg.(type) {
	case messages.EomMessage:
		// An EOM from outside means it is the leader in charge. We need to put it in our CurrentEOM collection
		// if it matches our authset position
		// TODO: If the EOM is 2 behind us and does not match, send out our EOM and discard this one

		eom := msg.(messages.EomMessage)
		if eom.Minute > p.WorkingEOM {
			// TODO: What do we do for EOM's from the future?
			return []imessage.IMessage{}
		}

		vm := util.GetVMForMsg(msg, *p.OriginalAuthSet[eom.Minute], primitives.MinuteLocation{eom.Minute, p.Height})
		if vm != eom.Vm {
			// TODO: What do we do for EOM's with the wrong VM number? Probably a different consensus
			return []imessage.IMessage{}
		}

		// TODO: Add EOM hash for auth set

		// If here we assume we agree, and it's a round 0 trump
		p.addEOMFromInternal(eom)
	case messages.DbsigMessage:

		dbs := msg.(messages.DbsigMessage)
		// A DBSig Majority will kill this Polling Station
		p.DBSigs[dbs.Prev][dbs.Signer] = dbs
		if len(p.DBSigs[dbs.Prev]) > p.NewAuthSet[9].Majority() {
			// Found a majority
			// TODO: Never go back
			myDBSig := messages.NewDBSigMessage(p.Self, dbs.Eom, dbs.Prev)
			return imessage.MakeMessageArray(myDBSig, dbs)
		}

	// Election Messages
	default:
	}

	return nil
}

// addEOMFromInternal will add to our current EOM collection. If the collection is complete
// we will move the round to trumping.
// 		Returns true if the minute is complete, false if the minute is not
func (p *PollingStation) addEOMFromInternal(eom messages.EomMessage) bool {
	if p.CurrentEOMs[eom.Minute][eom.Vm] != nil && *p.CurrentEOMs[eom.Minute][eom.Vm] != eom {
		// TODO: This should never happen
	}

	p.CurrentEOMs[eom.Minute][eom.Vm] = &eom
	// Check if the minute is complete
	for _, e := range p.CurrentEOMs[eom.Minute] {
		if e == nil {
			return false
		}
	}

	// If we got here the minute is complete
	if eom.Minute >= 2 {
		// TODO: If less than 2 it affects the previous Polling station
	}

	for i, e := range p.CurrentEOMs[eom.Minute-2] {
		p.TrumpEOMs[eom.Minute-2][i] = e
	}

	return true
}
