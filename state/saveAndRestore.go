// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"bytes"
	"fmt"
	"sort"

	. "github.com/FactomProject/factomd/common/identity"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// Because we have to go back to a previous state should the network be partictoned and we are on a separate
// brach, we need to log our state periodically so we can reset to a state prior to the network partitioin.
// The need to go back to a SaveState should be rare.  And even more rare would be the need to go back two
// levels.   However, it is possible that a minority particion is able to see a level of consensus and save
// a state to disk that the majority of the nodes did not see.  However it is not possible for this to occur
// more than once.  This is because any consensus a node can see requires that all the nodes saw the previous
// consensus.
type SaveState struct {
	DBHeight uint32

	// Don't need fields from the DBState because once we make it to the DBState.ProcessBlock() call, the
	// DBState settings are fixed and cannot change going forward.  Any DBState objects higher than the
	// DBHeight here must be tossed and reconstructed.

	FedServers   []interfaces.IServer
	AuditServers []interfaces.IServer

	// The old balances must be restored
	FactoidBalancesP map[[32]byte]int64
	ECBalancesP      map[[32]byte]int64

	Identities           []*Identity  // Identities of all servers in management chain
	Authorities          []*Authority // Identities of all servers in management chain
	AuthorityServerCount int          // number of federated or audit servers allowed

	// Server State
	LLeaderHeight uint32
	Leader        bool
	LeaderVMIndex int
	LeaderPL      *ProcessList
	CurrentMinute int

	EOMsyncing bool

	EOM          bool // Set to true when the first EOM is encountered
	EOMLimit     int
	EOMProcessed int
	EOMDone      bool
	EOMMinute    int
	EOMSys       bool // At least one EOM has covered the System List

	DBSig          bool
	DBSigLimit     int
	DBSigProcessed int // Number of DBSignatures received and processed.
	DBSigDone      bool
	DBSigSys       bool // At least one DBSig has covered the System List

	Newblk  bool // True if we are starting a new block, and a dbsig is needed.
	Saving  bool // True if we are in the process of saving to the database
	Syncing bool // Looking for messages from leaders to sync

	Replay *Replay

	LeaderTimestamp interfaces.Timestamp

	Holding map[[32]byte]interfaces.IMsg // Hold Messages
	XReview []interfaces.IMsg            // After the EOM, we must review the messages in Holding
	Acks    map[[32]byte]interfaces.IMsg // Hold Acknowledgemets
	Commits *SafeMsgMap                  // map[[32]byte]interfaces.IMsg // Commit Messages

	InvalidMessages map[[32]byte]interfaces.IMsg

	// DBlock Height at which node has a complete set of eblocks+entries
	EntryBlockDBHeightComplete uint32
	// DBlock Height at which we have started asking for entry blocks
	EntryBlockDBHeightProcessing uint32
	// Entry Blocks we don't have that we are asking our neighbors for
	MissingEntryBlocks []MissingEntryBlock

	// DBlock Height at which node has a complete set of eblocks+entries
	EntryDBHeightComplete uint32
	// Height in the DBlock where we have all the entries
	EntryHeightComplete int
	// DBlock Height at which we have started asking for or have all entries
	EntryDBHeightProcessing uint32
	// Height in the Directory Block where we have
	// Entries we don't have that we are asking our neighbors for
	MissingEntries []MissingEntry

	// FER section
	FactoshisPerEC                 uint64
	FERChainId                     string
	ExchangeRateAuthorityPublicKey string

	FERChangeHeight      uint32
	FERChangePrice       uint64
	FERPriority          uint32
	FERPrioritySetHeight uint32
}

var _ interfaces.BinaryMarshallable = (*SaveState)(nil)
var _ interfaces.Printable = (*SaveState)(nil)

func (ss *SaveState) Init() {
	if ss.FactoidBalancesP == nil {
		ss.FactoidBalancesP = map[[32]byte]int64{}
	}
	if ss.ECBalancesP == nil {
		ss.ECBalancesP = map[[32]byte]int64{}
	}
	if ss.Holding == nil {
		ss.Holding = map[[32]byte]interfaces.IMsg{}
	}
	if ss.Acks == nil {
		ss.Acks = map[[32]byte]interfaces.IMsg{}
	}
	if ss.Commits == nil {
		ss.Commits = NewSafeMsgMap() // map[[32]byte]interfaces.IMsg{}
	}
	if ss.InvalidMessages == nil {
		ss.InvalidMessages = map[[32]byte]interfaces.IMsg{}
	}
}

func (a *SaveState) IsSameAs(b *SaveState) bool {
	if a == nil || b == nil {
		if a == nil && b == nil {
			return true
		}
		return false
	}

	if a.DBHeight != b.DBHeight {
		return false
	}

	//FedServers   []interfaces.IServer
	//AuditServers []interfaces.IServer

	if len(a.FactoidBalancesP) != len(b.FactoidBalancesP) {
		return false
	}
	for k := range a.FactoidBalancesP {
		if a.FactoidBalancesP[k] != b.FactoidBalancesP[k] {
			return false
		}
	}
	if len(a.ECBalancesP) != len(b.ECBalancesP) {
		return false
	}
	for k := range a.ECBalancesP {
		if a.ECBalancesP[k] != b.ECBalancesP[k] {
			return false
		}
	}

	if len(a.Identities) != len(b.Identities) {
		return false
	}
	for i := range a.Identities {
		if a.Identities[i].IsSameAs(b.Identities[i]) == false {
			fmt.Printf("%v: %v vs %v\n", i, a.Identities[i].String(), b.Identities[i].String())
			return false
		}
	}
	if len(a.Authorities) != len(b.Authorities) {
		return false
	}
	for i := range a.Authorities {
		if a.Authorities[i].IsSameAs(b.Authorities[i]) == false {
			return false
		}
	}
	if a.AuthorityServerCount != b.AuthorityServerCount {
		return false
	}

	if a.LLeaderHeight != b.LLeaderHeight {
		return false
	}
	if a.Leader != b.Leader {
		return false
	}
	if a.LeaderVMIndex != b.LeaderVMIndex {
		return false
	}
	//LeaderPL      *ProcessList
	if a.CurrentMinute != b.CurrentMinute {
		return false
	}

	if a.EOMsyncing != b.EOMsyncing {
		return false
	}

	if a.EOM != b.EOM {
		return false
	}
	if a.EOMLimit != b.EOMLimit {
		return false
	}
	if a.EOMProcessed != b.EOMProcessed {
		return false
	}
	if a.EOMDone != b.EOMDone {
		return false
	}
	if a.EOMMinute != b.EOMMinute {
		return false
	}
	if a.EOMSys != b.EOMSys {
		return false
	}

	if a.DBSig != b.DBSig {
		return false
	}
	if a.DBSigLimit != b.DBSigLimit {
		return false
	}
	if a.DBSigProcessed != b.DBSigProcessed {
		return false
	}
	if a.DBSigDone != b.DBSigDone {
		return false
	}
	if a.DBSigSys != b.DBSigSys {
		return false
	}

	if a.Newblk != b.Newblk {
		return false
	}
	if a.Saving != b.Saving {
		return false
	}
	if a.Syncing != b.Syncing {
		return false
	}

	//Replay *Replay

	if a.LeaderTimestamp.IsSameAs(b.LeaderTimestamp) == false {
		return false
	}

	//Holding map[[32]byte]interfaces.IMsg
	//XReview []interfaces.IMsg
	//Acks    map[[32]byte]interfaces.IMsg
	//Commits map[[32]byte][]interfaces.IMsg

	//InvalidMessages map[[32]byte]interfaces.IMsg

	if a.EntryBlockDBHeightComplete != b.EntryBlockDBHeightComplete {
		return false
	}
	if a.EntryBlockDBHeightProcessing != b.EntryBlockDBHeightProcessing {
		return false
	}
	//MissingEntryBlocks []MissingEntryBlock

	if a.EntryDBHeightComplete != b.EntryDBHeightComplete {
		return false
	}
	if a.EntryHeightComplete != b.EntryHeightComplete {
		return false
	}
	if a.EntryDBHeightProcessing != b.EntryDBHeightProcessing {
		return false
	}
	//MissingEntries []MissingEntry

	if a.FactoshisPerEC != b.FactoshisPerEC {
		return false
	}
	if a.FERChainId != b.FERChainId {
		return false
	}
	if a.ExchangeRateAuthorityPublicKey != b.ExchangeRateAuthorityPublicKey {
		return false
	}

	if a.FERChangeHeight != b.FERChangeHeight {
		return false
	}
	if a.FERChangePrice != b.FERChangePrice {
		return false
	}
	if a.FERPriority != b.FERPriority {
		return false
	}
	if a.FERPrioritySetHeight != b.FERPrioritySetHeight {
		return false
	}

	return true
}

func SaveFactomdState(state *State, d *DBState) (ss *SaveState) {
	ss = new(SaveState)
	ss.DBHeight = d.DirectoryBlock.GetHeader().GetDBHeight()
	pl := state.ProcessLists.Get(ss.DBHeight)

	if pl == nil {
		return nil
	}

	//Only check if we're not loading from the database
	if state.DBFinished == true {
		// If the timestamp is over a day old, then there is really no point in saving the state of
		// historical data.
		if int(state.GetHighestKnownBlock())-int(state.GetHighestSavedBlk()) > 144 {
			return nil
		}
	}

	// state.AddStatus(fmt.Sprintf("Save state at dbht: %d", ss.DBHeight))

	ss.Replay = state.Replay.Save()
	ss.LeaderTimestamp = d.DirectoryBlock.GetTimestamp()

	ss.FedServers = append(ss.FedServers, pl.FedServers...)
	ss.AuditServers = append(ss.AuditServers, pl.AuditServers...)

	state.FactoidBalancesPMutex.Lock()
	ss.FactoidBalancesP = make(map[[32]byte]int64)
	for k := range state.FactoidBalancesP {
		ss.FactoidBalancesP[k] = state.FactoidBalancesP[k]
	}
	state.FactoidBalancesPMutex.Unlock()

	state.ECBalancesPMutex.Lock()
	ss.ECBalancesP = make(map[[32]byte]int64)
	for k := range state.ECBalancesP {
		ss.ECBalancesP[k] = state.ECBalancesP[k]
	}
	state.ECBalancesPMutex.Unlock()

	ss.Identities = append(ss.Identities, state.Identities...)
	ss.Authorities = append(ss.Authorities, state.Authorities...)
	ss.AuthorityServerCount = state.AuthorityServerCount

	ss.LLeaderHeight = state.LLeaderHeight
	ss.Leader = state.Leader
	ss.LeaderVMIndex = state.LeaderVMIndex
	ss.LeaderPL = state.LeaderPL
	ss.CurrentMinute = state.CurrentMinute

	ss.EOMsyncing = state.EOMsyncing

	ss.EOM = state.EOM
	ss.EOMLimit = state.EOMLimit
	ss.EOMProcessed = state.EOMProcessed
	ss.EOMDone = state.EOMDone
	ss.EOMMinute = state.EOMMinute
	ss.EOMSys = state.EOMSys
	ss.DBSig = state.DBSig
	ss.DBSigLimit = state.DBSigLimit
	ss.DBSigProcessed = state.DBSigProcessed
	ss.DBSigDone = state.DBSigDone
	ss.DBSigSys = state.DBSigSys
	ss.Saving = state.Saving
	ss.Syncing = state.Syncing

	ss.Holding = make(map[[32]byte]interfaces.IMsg)
	//for k := range state.Holding {
	//ss.Holding[k] = state.Holding[k]
	//}

	ss.XReview = append(ss.XReview, state.XReview...)

	ss.Acks = make(map[[32]byte]interfaces.IMsg)
	//for k := range state.Acks {
	//	ss.Acks[k] = state.Acks[k]
	//}

	ss.Commits = state.Commits.Copy()
	// for k, c := range state.Commits {
	// 	ss.Commits[k] = c
	// }

	ss.InvalidMessages = make(map[[32]byte]interfaces.IMsg)
	for k := range state.InvalidMessages {
		ss.InvalidMessages[k] = state.InvalidMessages[k]
	}

	ss.FactoshisPerEC = state.FactoshisPerEC
	ss.FERChainId = state.FERChainId
	ss.ExchangeRateAuthorityPublicKey = state.ExchangeRateAuthorityPublicKey

	ss.FERChangeHeight = state.FERChangeHeight
	ss.FERChangePrice = state.FERChangePrice
	ss.FERPriority = state.FERPriority
	ss.FERPrioritySetHeight = state.FERPrioritySetHeight

	/*
		err := SaveTheState(ss)
		if err != nil {
			panic(err)
		}
	*/

	return
}

func (ss *SaveState) TrimBack(state *State, d *DBState) {
	pdbstate := d
	d = state.DBStates.Get(int(ss.DBHeight + 1))
	if pdbstate == nil {
		return
	}
	// Don't do anything until we are within the current day
	if state.GetHighestKnownBlock()-state.GetHighestSavedBlk() > 144 {
		return
	}

	pss := pdbstate.SaveStruct
	if pss == nil {
		return
	}
	ppl := state.ProcessLists.Get(ss.DBHeight)
	if ppl == nil {
		return
	}
	pl := state.ProcessLists.Get(ss.DBHeight + 1)
	if pl == nil {
		return
	}

	for _, vm := range pl.VMs {
		vm.LeaderMinute = 0
		if vm.Height > 0 {
			vm.Signed = true
			vm.Synced = true
			vm.Height = 0
			vm.List = vm.List[:0]
			vm.ListAck = vm.ListAck[:0]
		} else {
			vm.Signed = false
			vm.Synced = false
			vm.List = vm.List[:0]
			vm.ListAck = vm.ListAck[:0]
		}
	}

	ss.EOMsyncing = state.EOMsyncing

	state.EOM = pss.EOM
	state.EOMLimit = pss.EOMLimit
	state.EOMProcessed = pss.EOMProcessed
	state.EOMDone = pss.EOMDone
	state.EOMMinute = pss.EOMMinute
	state.EOMSys = pss.EOMSys
	state.DBSig = pss.DBSig
	state.DBSigLimit = pss.DBSigLimit
	state.DBSigProcessed = pss.DBSigProcessed
	state.DBSigDone = pss.DBSigDone
	state.DBSigSys = pss.DBSigSys
	state.Saving = pss.Saving
	state.Syncing = pss.Syncing

	state.Replay = pss.Replay.Save()

	return
	/*
		pl.FedServers = append(pl.FedServers[0:], ppl.FedServers...)
		pl.AuditServers = append(pl.AuditServers[0:], ppl.AuditServers...)

		//state.Identities = append(state.Identities[:0], pss.Identities...)
		//state.Authorities = append(state.Authorities[:0], pss.Authorities...)
		//state.AuthorityServerCount = pss.AuthorityServerCount

		state.Holding = make(map[[32]byte]interfaces.IMsg)
		for k := range ss.Holding {
			state.Holding[k] = pss.Holding[k]
		}
		state.XReview = append(state.XReview[:0], pss.XReview...)
	*/

	/**
	ss.EOMsyncing = state.EOMsyncing

	state.EOM = pss.EOM
	state.EOMLimit = pss.EOMLimit
	state.EOMProcessed = pss.EOMProcessed
	state.EOMDone = pss.EOMDone
	state.EOMMinute = pss.EOMMinute
	state.EOMSys = pss.EOMSys
	state.DBSig = pss.DBSig
	state.DBSigLimit = pss.DBSigLimit
	state.DBSigProcessed = pss.DBSigProcessed
	state.DBSigDone = pss.DBSigDone
	state.DBSigSys = pss.DBSigSys
	state.Newblk = pss.Newblk
	state.Saving = pss.Saving
	state.Syncing = pss.Syncing

	state.Holding = make(map[[32]byte]interfaces.IMsg)
	for k := range ss.Holding {
		state.Holding[k] = pss.Holding[k]
	}
	state.XReview = append(state.XReview[:0], pss.XReview...)

	state.Acks = make(map[[32]byte]interfaces.IMsg)
	for k := range pss.Acks {
		state.Acks[k] = pss.Acks[k]
	}

	state.Commits = make(map[[32]byte][]interfaces.IMsg)
	for k := range pss.Commits {
		var c []interfaces.IMsg
		state.Commits[k] = append(c, pss.Commits[k]...)
	}

	state.InvalidMessages = make(map[[32]byte]interfaces.IMsg)
	for k := range pss.InvalidMessages {
		state.InvalidMessages[k] = pss.InvalidMessages[k]
	}

	// DBlock Height at which node has a complete set of eblocks+entries
	state.EntryBlockDBHeightComplete = pss.EntryBlockDBHeightComplete
	state.EntryBlockDBHeightProcessing = pss.EntryBlockDBHeightProcessing
	state.MissingEntryBlocks = append(state.MissingEntryBlocks[:0], pss.MissingEntryBlocks...)

	state.EntryBlockDBHeightComplete = pss.EntryDBHeightComplete
	state.EntryDBHeightComplete = pss.EntryDBHeightComplete
	state.EntryHeightComplete = pss.EntryHeightComplete
	state.EntryDBHeightProcessing = pss.EntryBlockDBHeightProcessing
	state.MissingEntries = append(state.MissingEntries[:0], pss.MissingEntries...)

	state.FactoshisPerEC = pss.FactoshisPerEC
	state.FERChainId = pss.FERChainId
	state.ExchangeRateAuthorityAddress = pss.ExchangeRateAuthorityAddress

	state.FERChangeHeight = pss.FERChangeHeight
	state.FERChangePrice = pss.FERChangePrice
	state.FERPriority = pss.FERPriority
	state.FERPrioritySetHeight = pss.FERPrioritySetHeight

	**/
}

func (ss *SaveState) RestoreFactomdState(state *State) { //, d *DBState) {
	// We trim away the ProcessList under construction (and any others) so we can
	// rebuild afresh.
	index := int(state.ProcessLists.DBHeightBase) - int(ss.DBHeight)
	if index < 0 {
		index = 0
	} else {
		fmt.Println(state.ProcessLists.String())

		if len(state.ProcessLists.Lists) > index+1 {
			state.ProcessLists.Lists = state.ProcessLists.Lists[:index+2]
			pln := state.ProcessLists.Lists[index+1]
			for _, vm := range pln.VMs {
				vm.LeaderMinute = 0
				if vm.Height > 0 {
					vm.Signed = true
					vm.Synced = true
					vm.Height = 0
					vm.List = vm.List[:0]
					vm.ListAck = vm.ListAck[:0]
				} else {
					vm.Signed = false
					vm.Synced = false
					vm.List = vm.List[:0]
					vm.ListAck = vm.ListAck[:0]
				}
			}
		}
	}
	pl := state.ProcessLists.Get(ss.DBHeight)

	// state.AddStatus(fmt.Sprintln("Index: ", index, "dbht:", ss.DBHeight, "lleaderheight", state.LLeaderHeight))

	dindex := ss.DBHeight - state.DBStates.Base
	state.DBStates.DBStates = state.DBStates.DBStates[:dindex]
	//state.AddStatus(fmt.Sprintf("SAVESTATE Restoring the State to dbht: %d", ss.DBHeight))

	state.Replay = ss.Replay.Save()
	state.LeaderTimestamp = ss.LeaderTimestamp

	pl.FedServers = []interfaces.IServer{}
	pl.AuditServers = []interfaces.IServer{}
	pl.FedServers = append(pl.FedServers, ss.FedServers...)
	pl.AuditServers = append(pl.AuditServers, ss.AuditServers...)

	state.FactoidBalancesPMutex.Lock()
	state.FactoidBalancesP = make(map[[32]byte]int64, 0)
	for k := range ss.FactoidBalancesP {
		state.FactoidBalancesP[k] = ss.FactoidBalancesP[k]
	}
	state.FactoidBalancesPMutex.Unlock()

	state.ECBalancesPMutex.Lock()
	state.ECBalancesP = make(map[[32]byte]int64, 0)
	for k := range ss.ECBalancesP {
		state.ECBalancesP[k] = ss.ECBalancesP[k]
	}
	state.ECBalancesPMutex.Unlock()

	state.Identities = append(state.Identities[:0], ss.Identities...)
	state.Authorities = append(state.Authorities[:0], ss.Authorities...)
	state.AuthorityServerCount = ss.AuthorityServerCount

	state.LLeaderHeight = ss.LLeaderHeight
	state.Leader = ss.Leader
	state.LeaderVMIndex = ss.LeaderVMIndex
	state.LeaderPL = ss.LeaderPL
	state.CurrentMinute = ss.CurrentMinute

	ss.EOMsyncing = state.EOMsyncing

	state.EOM = false
	state.EOMLimit = ss.EOMLimit
	state.EOMProcessed = ss.EOMProcessed
	state.EOMDone = ss.EOMDone
	state.EOMMinute = ss.EOMMinute
	state.EOMSys = ss.EOMSys
	state.DBSig = false
	state.DBSigLimit = ss.DBSigLimit
	state.DBSigProcessed = ss.DBSigProcessed
	state.DBSigDone = ss.DBSigDone
	state.DBSigSys = ss.DBSigSys
	state.Saving = true
	state.Syncing = false
	state.HighestAck = ss.DBHeight + 1
	state.HighestKnown = ss.DBHeight + 2
	state.Holding = make(map[[32]byte]interfaces.IMsg)
	for k := range ss.Holding {
		state.Holding[k] = ss.Holding[k]
	}
	state.XReview = append(state.XReview[:0], ss.XReview...)

	state.Acks = make(map[[32]byte]interfaces.IMsg)
	for k := range ss.Acks {
		state.Acks[k] = ss.Acks[k]
	}

	state.Commits = ss.Commits.Copy() // make(map[[32]byte]interfaces.IMsg)
	// for k, c := range ss.Commits {
	// 	state.Commits[k] = c
	// }

	state.InvalidMessages = make(map[[32]byte]interfaces.IMsg)
	for k := range ss.InvalidMessages {
		state.InvalidMessages[k] = ss.InvalidMessages[k]
	}

	state.FactoshisPerEC = ss.FactoshisPerEC
	state.FERChainId = ss.FERChainId
	state.ExchangeRateAuthorityPublicKey = ss.ExchangeRateAuthorityPublicKey

	state.FERChangeHeight = ss.FERChangeHeight
	state.FERChangePrice = ss.FERChangePrice
	state.FERPriority = ss.FERPriority
	state.FERPrioritySetHeight = ss.FERPrioritySetHeight
}

func (ss *SaveState) MarshalBinary() ([]byte, error) {
	buf := primitives.NewBuffer(nil)

	err := buf.PushUInt32(ss.DBHeight)
	if err != nil {
		return nil, err
	}

	l := len(ss.FedServers)
	err = buf.PushVarInt(uint64(l))
	if err != nil {
		return nil, err
	}
	for _, v := range ss.FedServers {
		err = buf.PushBinaryMarshallable(v)
		if err != nil {
			return nil, err
		}
	}

	l = len(ss.AuditServers)
	err = buf.PushVarInt(uint64(l))
	if err != nil {
		return nil, err
	}
	for _, v := range ss.AuditServers {
		err = buf.PushBinaryMarshallable(v)
		if err != nil {
			return nil, err
		}
	}

	err = PushBalanceMap(buf, ss.FactoidBalancesP)
	if err != nil {
		return nil, err
	}

	err = PushBalanceMap(buf, ss.ECBalancesP)
	if err != nil {
		return nil, err
	}

	l = len(ss.Identities)
	err = buf.PushVarInt(uint64(l))
	if err != nil {
		return nil, err
	}
	for _, v := range ss.Identities {
		err = buf.PushBinaryMarshallable(v)
		if err != nil {
			return nil, err
		}
	}

	l = len(ss.Authorities)
	err = buf.PushVarInt(uint64(l))
	if err != nil {
		return nil, err
	}
	for _, v := range ss.Authorities {
		err = buf.PushBinaryMarshallable(v)
		if err != nil {
			return nil, err
		}
	}

	err = buf.PushVarInt(uint64(ss.AuthorityServerCount))
	if err != nil {
		return nil, err
	}

	err = buf.PushUInt32(ss.LLeaderHeight)
	if err != nil {
		return nil, err
	}
	err = buf.PushBool(ss.Leader)
	if err != nil {
		return nil, err
	}
	err = buf.PushVarInt(uint64(ss.LeaderVMIndex))
	if err != nil {
		return nil, err
	}
	//TODO: handle LeaderPL      *ProcessList
	err = buf.PushVarInt(uint64(ss.CurrentMinute))
	if err != nil {
		return nil, err
	}

	err = buf.PushBool(ss.EOMsyncing)
	if err != nil {
		return nil, err
	}

	err = buf.PushBool(ss.EOM)
	if err != nil {
		return nil, err
	}
	err = buf.PushVarInt(uint64(ss.EOMLimit))
	if err != nil {
		return nil, err
	}
	err = buf.PushVarInt(uint64(ss.EOMProcessed))
	if err != nil {
		return nil, err
	}
	err = buf.PushBool(ss.EOMDone)
	if err != nil {
		return nil, err
	}
	err = buf.PushVarInt(uint64(ss.EOMMinute))
	if err != nil {
		return nil, err
	}
	err = buf.PushBool(ss.EOMSys)
	if err != nil {
		return nil, err
	}

	err = buf.PushBool(ss.DBSig)
	if err != nil {
		return nil, err
	}
	err = buf.PushVarInt(uint64(ss.DBSigLimit))
	if err != nil {
		return nil, err
	}
	err = buf.PushVarInt(uint64(ss.DBSigProcessed))
	if err != nil {
		return nil, err
	}
	err = buf.PushBool(ss.DBSigDone)
	if err != nil {
		return nil, err
	}
	err = buf.PushBool(ss.DBSigSys)
	if err != nil {
		return nil, err
	}

	err = buf.PushBool(ss.Newblk)
	if err != nil {
		return nil, err
	}
	err = buf.PushBool(ss.Saving)
	if err != nil {
		return nil, err
	}
	err = buf.PushBool(ss.Syncing)
	if err != nil {
		return nil, err
	}

	err = buf.PushBinaryMarshallable(ss.Replay)
	if err != nil {
		return nil, err
	}

	err = buf.PushBinaryMarshallable(ss.LeaderTimestamp)
	if err != nil {
		return nil, err
	}
	/*
		Holding map[[32]byte]interfaces.IMsg   // Hold Messages
		XReview []interfaces.IMsg              // After the EOM, we must review the messages in Holding
		Acks    map[[32]byte]interfaces.IMsg   // Hold Acknowledgemets
		Commits map[[32]byte][]interfaces.IMsg // Commit Messages

		InvalidMessages map[[32]byte]interfaces.IMsg
	*/

	err = buf.PushUInt32(ss.EntryBlockDBHeightComplete)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(ss.EntryBlockDBHeightProcessing)
	if err != nil {
		return nil, err
	}
	l = len(ss.MissingEntryBlocks)
	err = buf.PushVarInt(uint64(l))
	if err != nil {
		return nil, err
	}
	for _, v := range ss.MissingEntryBlocks {
		err = buf.PushBinaryMarshallable(&v)
		if err != nil {
			return nil, err
		}
	}

	err = buf.PushUInt32(ss.EntryDBHeightComplete)
	if err != nil {
		return nil, err
	}
	err = buf.PushVarInt(uint64(ss.EntryHeightComplete))
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(ss.EntryDBHeightProcessing)
	if err != nil {
		return nil, err
	}
	l = len(ss.MissingEntries)
	err = buf.PushVarInt(uint64(l))
	if err != nil {
		return nil, err
	}
	for _, v := range ss.MissingEntries {
		err = buf.PushBinaryMarshallable(&v)
		if err != nil {
			return nil, err
		}
	}

	err = buf.PushVarInt(ss.FactoshisPerEC)
	if err != nil {
		return nil, err
	}
	err = buf.PushString(ss.FERChainId)
	if err != nil {
		return nil, err
	}
	err = buf.PushString(ss.ExchangeRateAuthorityPublicKey)
	if err != nil {
		return nil, err
	}

	err = buf.PushUInt32(ss.FERChangeHeight)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt64(ss.FERChangePrice)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(ss.FERPriority)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(ss.FERPrioritySetHeight)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (ss *SaveState) UnmarshalBinaryData(p []byte) (newData []byte, err error) {
	ss.FactoidBalancesP = map[[32]byte]int64{}
	ss.ECBalancesP = map[[32]byte]int64{}
	ss.Holding = map[[32]byte]interfaces.IMsg{}
	ss.Acks = map[[32]byte]interfaces.IMsg{}
	ss.Commits = NewSafeMsgMap()
	ss.InvalidMessages = map[[32]byte]interfaces.IMsg{}

	ss.FedServers = []interfaces.IServer{}
	ss.AuditServers = []interfaces.IServer{}

	ss.Identities = []*Identity{}
	ss.Authorities = []*Authority{}

	newData = p
	buf := primitives.NewBuffer(p)

	ss.DBHeight, err = buf.PopUInt32()
	if err != nil {
		return
	}

	l, err := buf.PopVarInt()
	if err != nil {
		return
	}
	for i := 0; i < int(l); i++ {
		s := new(Server)
		err = buf.PopBinaryMarshallable(s)
		if err != nil {
			return
		}
		ss.FedServers = append(ss.FedServers, s)
	}

	l, err = buf.PopVarInt()
	if err != nil {
		return
	}
	for i := 0; i < int(l); i++ {
		s := new(Server)
		err = buf.PopBinaryMarshallable(s)
		if err != nil {
			return
		}
		ss.AuditServers = append(ss.AuditServers, s)
	}

	ss.FactoidBalancesP, err = PopBalanceMap(buf)
	if err != nil {
		return
	}

	ss.ECBalancesP, err = PopBalanceMap(buf)
	if err != nil {
		return
	}

	l, err = buf.PopVarInt()
	if err != nil {
		return
	}
	for i := 0; i < int(l); i++ {
		s := new(Identity)
		err = buf.PopBinaryMarshallable(s)
		if err != nil {
			return
		}
		ss.Identities = append(ss.Identities, s)
	}

	l, err = buf.PopVarInt()
	if err != nil {
		return
	}
	for i := 0; i < int(l); i++ {
		s := new(Authority)
		err = buf.PopBinaryMarshallable(s)
		if err != nil {
			return
		}
		ss.Authorities = append(ss.Authorities, s)
	}
	l, err = buf.PopVarInt()
	if err != nil {
		return
	}
	ss.AuthorityServerCount = int(l)

	ss.LLeaderHeight, err = buf.PopUInt32()
	if err != nil {
		return
	}
	ss.Leader, err = buf.PopBool()
	if err != nil {
		return
	}
	l, err = buf.PopVarInt()
	if err != nil {
		return
	}
	ss.LeaderVMIndex = int(l)

	//TODO: handle LeaderPL      *ProcessList
	l, err = buf.PopVarInt()
	if err != nil {
		return
	}
	ss.CurrentMinute = int(l)

	ss.EOMsyncing, err = buf.PopBool()
	if err != nil {
		return
	}
	ss.EOM, err = buf.PopBool()
	if err != nil {
		return
	}

	l, err = buf.PopVarInt()
	if err != nil {
		return
	}
	ss.EOMLimit = int(l)

	l, err = buf.PopVarInt()
	if err != nil {
		return
	}
	ss.EOMProcessed = int(l)

	ss.EOMDone, err = buf.PopBool()
	if err != nil {
		return
	}

	l, err = buf.PopVarInt()
	if err != nil {
		return
	}
	ss.EOMMinute = int(l)

	ss.EOMSys, err = buf.PopBool()
	if err != nil {
		return
	}

	ss.DBSig, err = buf.PopBool()
	if err != nil {
		return
	}

	l, err = buf.PopVarInt()
	if err != nil {
		return
	}
	ss.DBSigLimit = int(l)

	l, err = buf.PopVarInt()
	if err != nil {
		return
	}
	ss.DBSigProcessed = int(l)

	ss.DBSigDone, err = buf.PopBool()
	if err != nil {
		return
	}
	ss.DBSigSys, err = buf.PopBool()
	if err != nil {
		return
	}

	ss.Newblk, err = buf.PopBool()
	if err != nil {
		return
	}
	ss.Saving, err = buf.PopBool()
	if err != nil {
		return
	}
	ss.Syncing, err = buf.PopBool()
	if err != nil {
		return
	}

	ss.Replay = new(Replay)
	err = buf.PopBinaryMarshallable(ss.Replay)
	if err != nil {
		return
	}

	ss.LeaderTimestamp = primitives.NewTimestampFromMilliseconds(0)
	err = buf.PopBinaryMarshallable(ss.LeaderTimestamp)
	if err != nil {
		return
	}

	/*
		Holding map[[32]byte]interfaces.IMsg   // Hold Messages
		XReview []interfaces.IMsg              // After the EOM, we must review the messages in Holding
		Acks    map[[32]byte]interfaces.IMsg   // Hold Acknowledgemets
		Commits map[[32]byte][]interfaces.IMsg // Commit Messages

		InvalidMessages map[[32]byte]interfaces.IMsg
	*/

	ss.EntryBlockDBHeightComplete, err = buf.PopUInt32()
	if err != nil {
		return
	}
	ss.EntryBlockDBHeightProcessing, err = buf.PopUInt32()
	if err != nil {
		return
	}

	l, err = buf.PopVarInt()
	if err != nil {
		return
	}
	for i := 0; i < int(l); i++ {
		s := new(MissingEntryBlock)
		err = buf.PopBinaryMarshallable(s)
		if err != nil {
			return
		}
		ss.MissingEntryBlocks = append(ss.MissingEntryBlocks, *s)
	}

	ss.EntryDBHeightComplete, err = buf.PopUInt32()
	if err != nil {
		return
	}

	l, err = buf.PopVarInt()
	if err != nil {
		return
	}
	ss.EntryHeightComplete = int(l)

	ss.EntryDBHeightProcessing, err = buf.PopUInt32()
	if err != nil {
		return
	}

	l, err = buf.PopVarInt()
	if err != nil {
		return
	}
	for i := 0; i < int(l); i++ {
		s := new(MissingEntry)
		err = buf.PopBinaryMarshallable(s)
		if err != nil {
			return
		}
		ss.MissingEntries = append(ss.MissingEntries, *s)
	}

	ss.FactoshisPerEC, err = buf.PopVarInt()
	if err != nil {
		return
	}
	ss.FERChainId, err = buf.PopString()
	if err != nil {
		return
	}
	ss.ExchangeRateAuthorityPublicKey, err = buf.PopString()
	if err != nil {
		return
	}

	ss.FERChangeHeight, err = buf.PopUInt32()
	if err != nil {
		return
	}
	ss.FERChangePrice, err = buf.PopUInt64()
	if err != nil {
		return
	}
	ss.FERPriority, err = buf.PopUInt32()
	if err != nil {
		return
	}
	ss.FERPrioritySetHeight, err = buf.PopUInt32()
	if err != nil {
		return
	}

	newData = buf.DeepCopyBytes()
	return
}

func (ss *SaveState) UnmarshalBinary(p []byte) error {
	_, err := ss.UnmarshalBinaryData(p)
	return err
}

func (e *SaveState) String() string {
	str, _ := e.JSONString()
	return str
}

func (e *SaveState) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *SaveState) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func PushBalanceMap(b *primitives.Buffer, m map[[32]byte]int64) error {
	l := len(m)
	err := b.PushVarInt(uint64(l))
	if err != nil {
		return err
	}

	keys := [][32]byte{}
	for k := range m {
		keys = append(keys, k)
	}

	sort.Sort(ByKey(keys))

	for _, k := range keys {
		err = b.Push(k[:])
		if err != nil {
			return err
		}
		err = b.PushInt64(m[k])
		if err != nil {
			return err
		}
	}
	return nil
}

type ByKey [][32]byte

func (f ByKey) Len() int {
	return len(f)
}
func (f ByKey) Less(i, j int) bool {
	return bytes.Compare(f[i][:], f[j][:]) < 0
}
func (f ByKey) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func PopBalanceMap(buf *primitives.Buffer) (map[[32]byte]int64, error) {
	m := map[[32]byte]int64{}
	k := make([]byte, 32)
	l, err := buf.PopVarInt()
	if err != nil {
		return nil, err
	}
	for i := 0; i < int(l); i++ {
		var b [32]byte
		err = buf.Pop(k)
		if err != nil {
			return nil, err
		}
		copy(b[:], k)
		v, err := buf.PopInt64()
		if err != nil {
			return nil, err
		}
		m[b] = v
	}
	return m, nil
}
