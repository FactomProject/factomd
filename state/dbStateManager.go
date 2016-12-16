// Copyright 2015 FactomProject Authors. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package state

import (
	"encoding/hex"
	"fmt"
	"time"

	"bytes"
	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/log"
)

var _ = hex.EncodeToString
var _ = fmt.Print
var _ = time.Now()
var _ = log.Print

type DBState struct {
	isNew bool

	SaveStruct *SaveState

	DBHash interfaces.IHash
	ABHash interfaces.IHash
	FBHash interfaces.IHash
	ECHash interfaces.IHash

	DirectoryBlock   interfaces.IDirectoryBlock
	AdminBlock       interfaces.IAdminBlock
	FactoidBlock     interfaces.IFBlock
	EntryCreditBlock interfaces.IEntryCreditBlock

	EntryBlocks []interfaces.IEntryBlock
	Entries     []interfaces.IEBEntry

	Locked      bool
	ReadyToSave bool
	Saved       bool

	FinalExchangeRate uint64
	NextTimestamp     interfaces.Timestamp
}

type DBStateList struct {
	SrcNetwork          bool // True if I got this block from the network.
	LastTime            interfaces.Timestamp
	SecondsBetweenTests int
	Lastreq             int
	State               *State
	Base                uint32
	Complete            uint32
	DBStates            []*DBState
}

// Validate this directory block given the next Directory Block.  Need to check the
// signatures as being from the authority set, and valid. Also check that this DBState holds
// a previous KeyMR that matches the previous DBState KeyMR.
//
// Return a -1 on failure.
//
func (d *DBState) ValidNext(state *State, next *messages.DBStateMsg) int {
	dirblk := next.DirectoryBlock
	dbheight := dirblk.GetHeader().GetDBHeight()
	if dbheight == 0 {
		state.AddStatus(fmt.Sprintf("DBState.ValidNext: rtn 1 genesis block is valid dbht: %d", dbheight))
		// The genesis block is valid by definition.
		return 1
	}
	if d == nil || !d.Saved {
		state.AddStatus(fmt.Sprintf("DBState.ValidNext: rtn 0 dbstate is nil or not saved dbht: %d", dbheight))
		// Must be out of order.  Can't make the call if valid or not yet.
		return 0
	}

	if int(state.EntryBlockDBHeightComplete) < int(dbheight-1) {
		state.AddStatus(fmt.Sprintf("DBState.ValidNext: rtn 0s Don't have all the Entries we want dbht: %d", dbheight))
		return 0
	}

	// Get the keymr of the Previous DBState
	pkeymr := d.DirectoryBlock.GetKeyMR()
	// Get the Previous KeyMR pointer in the possible new Directory Block
	prevkeymr := dirblk.GetHeader().GetPrevKeyMR()
	if !pkeymr.IsSameAs(prevkeymr) {
		state.AddStatus(fmt.Sprintf("DBState.ValidNext: rtn -1 hashes don't match. dbht: %d dbstate had prev %x but we expected %x ",
			dbheight, prevkeymr.Bytes()[:3], pkeymr.Bytes()[:3]))
		// If not the same, this is a bad new Directory Block
		return -1
	}

	return 1

	admin := next.AdminBlock
	for _, entry := range admin.GetABEntries() {
		if addfed, ok := entry.(*adminBlock.AddFederatedServer); ok {
			if state.isIdentityChain(addfed.IdentityChainID) < 0 {
				state.AddStatus(fmt.Sprintf("DBState.ValidNext: rtn 0 Adding a Fed server in admin block without identity dbht: %d fed server: %x ",
					dbheight, addfed.IdentityChainID.Bytes()[:3]))
				return 0
			}
		}
		if addaudit, ok := entry.(*adminBlock.AddAuditServer); ok {
			if state.isIdentityChain(addaudit.IdentityChainID) < 0 {
				state.AddStatus(fmt.Sprintf("DBState.ValidNext: rtn 0 Adding a Audit server in admin block without identity dbht: %d fed server: %x ",
					dbheight, addaudit.IdentityChainID.Bytes()[:3]))
				return 0
			}
		}
	}

	return 1
}

func (list *DBStateList) String() string {
	str := "\n========DBStates Start=======\nddddd DBStates\n"
	str = fmt.Sprintf("dddd %s  Base      = %d\n", str, list.Base)
	ts := "-nil-"
	if list.LastTime != nil {
		ts = list.LastTime.String()
	}
	str = fmt.Sprintf("dddd %s  timestamp = %s\n", str, ts)
	str = fmt.Sprintf("dddd %s  Complete  = %d\n", str, list.Complete)
	rec := "M"
	last := ""
	for i, ds := range list.DBStates {
		rec = "?"
		if ds != nil {
			rec = "nil"
			if ds.DirectoryBlock != nil {
				rec = "x"

				dblk, _ := list.State.GetAndLockDB().FetchDBlock(ds.DirectoryBlock.GetKeyMR())
				defer list.State.UnlockDB()
				if dblk != nil {
					rec = "s"
				}

				if ds.Locked {
					rec = rec + "L"
				}

				if ds.ReadyToSave {
					rec = rec + "R"
				}

				if ds.Saved {
					rec = rec + "S"
				}
			}
		}
		if last != "" {
			str = last
		}
		str = fmt.Sprintf("dddd %s  %1s-DState ?-(DState nil) x-(Not in DB) s-(In DB) L-(Locked) R-(Ready to Save) S-(Saved)\n   DState Height: %d\n%s", str, rec, list.Base+uint32(i), ds.String())
		if rec == "?" && last == "" {
			last = str
		}
	}
	str = str + "dddd\n============DBStates End==========\n"
	return str
}

func (ds *DBState) String() string {
	str := ""
	if ds == nil {
		str = "  DBState = <nil>\n"
	} else if ds.DirectoryBlock == nil {
		str = "  Directory Block = <nil>\n"
	} else {

		str = fmt.Sprintf("%s      DBlk Height   = %v \n", str, ds.DirectoryBlock.GetHeader().GetDBHeight())
		str = fmt.Sprintf("%s      DBlock        = %x \n", str, ds.DirectoryBlock.GetHash().Bytes()[:5])
		str = fmt.Sprintf("%s      ABlock        = %x \n", str, ds.AdminBlock.GetHash().Bytes()[:5])
		str = fmt.Sprintf("%s      FBlock        = %x \n", str, ds.FactoidBlock.GetHash().Bytes()[:5])
		str = fmt.Sprintf("%s      ECBlock       = %x \n", str, ds.EntryCreditBlock.GetHash().Bytes()[:5])
	}
	return str
}

func (list *DBStateList) GetHighestSavedBlock() uint32 {
	ht := list.Base
	for i, dbstate := range list.DBStates {
		if dbstate != nil && dbstate.Locked {
			ht = list.Base + uint32(i)
		} else {
			if dbstate == nil {
				return ht
			}
		}
	}
	return ht
}

func (list *DBStateList) GetHighestCompletedBlock() uint32 {
	ht := list.Base
	for i, dbstate := range list.DBStates {
		if dbstate != nil && dbstate.Saved {
			ht = list.Base + uint32(i)
		} else {
			if dbstate == nil {
				return ht
			}
		}
	}
	return ht
}

// Once a second at most, we check to see if we need to pull down some blocks to catch up.
func (list *DBStateList) Catchup() {

	now := list.State.GetTimestamp()

	dbsHeight := list.GetHighestCompletedBlock()

	// We only check if we need updates once every so often.

	begin := -1
	end := -1

	// Find the first range of blocks that we don't have.
	for i, v := range list.DBStates {
		if (v == nil || v.DirectoryBlock == nil) && begin < 0 {
			begin = i
		}
		if v == nil {
			end = i
		}
	}

	if begin > 0 {
		begin += int(list.Base)
		end += int(list.Base)
	} else {
		plHeight := list.State.GetHighestKnownBlock()
		// Don't worry about the block initialization case.
		if plHeight < 1 {
			list.LastTime = nil
			return
		}

		if plHeight >= dbsHeight && plHeight-dbsHeight > 2 {
			begin = int(dbsHeight + 1)
			end = int(plHeight - 1)
		} else {
			list.LastTime = nil
			return
		}

		for list.State.ProcessLists.Get(uint32(begin)) != nil && list.State.ProcessLists.Get(uint32(begin)).Complete() {
			begin++
			if uint32(begin) >= plHeight || begin > end {
				list.LastTime = nil
				return
			}
		}
	}

	if begin > 0 {
		begin--
	}
	end++ // ask for one more, just in case.

	list.Lastreq = begin

	end2 := begin + 400
	if end < end2 {
		end2 = end
	}

	if list.LastTime == nil {
		list.LastTime = now
		return
	}

	if now.GetTimeMilli()-list.LastTime.GetTimeMilli() < 1500 {
		return
	}

	list.State.RunLeader = false

	list.LastTime = now

	msg := messages.NewDBStateMissing(list.State, uint32(begin), uint32(end2))

	if msg != nil {
		//		list.State.RunLeader = false
		//		list.State.StartDelay = list.State.GetTimestamp().GetTimeMilli()
		msg.SendOut(list.State, msg)
		msg.SendOut(list.State, msg)
		msg.SendOut(list.State, msg)
		list.LastTime = now
		list.State.DBStateAskCnt++
	}

}

// a contains b, returns true
func containsServer(haystack []interfaces.IFctServer, needle interfaces.IFctServer) bool {
	for _, hay := range haystack {
		if needle.GetChainID().IsSameAs(hay.GetChainID()) {
			return true
		}
	}
	return false
}

// p is previous, d is current
func (list *DBStateList) FixupLinks(p *DBState, d *DBState) (progress bool) {
	// If this block is new, then make sure all hashes are fully computed.
	if !d.isNew || p == nil {
		return
	}

	currentDBHeight := d.DirectoryBlock.GetHeader().GetDBHeight()
	previousDBHeight := p.DirectoryBlock.GetHeader().GetDBHeight()

	d.DirectoryBlock.MarshalBinary()

	hash, err := p.EntryCreditBlock.HeaderHash()
	if err != nil {
		panic(err.Error())
	}
	d.EntryCreditBlock.GetHeader().SetPrevHeaderHash(hash)

	hash, err = p.EntryCreditBlock.GetFullHash()
	if err != nil {
		panic(err.Error())
	}
	d.EntryCreditBlock.GetHeader().SetPrevFullHash(hash)
	d.EntryCreditBlock.GetHeader().SetDBHeight(currentDBHeight)

	// Admin Block Fixup
	previousPL := list.State.ProcessLists.Get(previousDBHeight)
	currentPL := list.State.ProcessLists.Get(currentDBHeight)

	// Servers
	previousFeds := previousPL.FedServers
	currentFeds := currentPL.FedServers
	currentAuds := currentPL.AuditServers

	// DB Sigs
	majority := (len(currentFeds) / 2) + 1
	lenDBSigs := len(list.State.ProcessLists.Get(currentDBHeight).DBSignatures)
	if lenDBSigs < majority {
		list.State.AddStatus(fmt.Sprintf("FIXUPLINKS: return without processing: lenDBSigs)(%v) < majority(%d)",
			lenDBSigs,
			majority))

		return false
	}
	list.State.AddStatus(fmt.Sprintf("FIXUPLINKS: Adding the first %d dbsigs",
		majority))

	for i, sig := range list.State.ProcessLists.Get(currentDBHeight).DBSignatures {
		if i < majority {
			d.AdminBlock.AddDBSig(sig.ChainID, sig.Signature)
		} else {
			break
		}
	}

	list.State.AddStatus("FIXUPLINKS: Adding the deltas to the Admin Block, if necessary")

	// Correcting Server Lists (Caused by Server Faults)
	for _, cf := range currentFeds {
		if !containsServer(previousFeds, cf) {
			// Promote to federated
			//index := list.State.isIdentityChain(cf.GetChainID())
			/*if index == -1 || !(list.State.Identities[index].Status == constants.IDENTITY_PENDING_FEDERATED_SERVER ||
			list.State.Identities[index].Status == constants.IDENTITY_FEDERATED_SERVER) ||
			list.State.Identities[index].Status == constants.IDENTITY_AUDIT_SERVER) {*/
			addEntry := adminBlock.NewAddFederatedServer(cf.GetChainID(), currentDBHeight+1)
			list.State.AddStatus(fmt.Sprintf("FIXUPLINKS: Adding delta to the Admin Block: %s", addEntry.String()))
			d.AdminBlock.AddFirstABEntry(addEntry)
			/*} else {
				list.State.AddStatus(fmt.Sprintf("FIXUPLINKS: Not Adding delta to the Admin Block: Idx: %d Status: %d", index, list.State.Identities[index].Status))
			}*/
		}
	}

	for _, pf := range previousFeds {
		if !containsServer(currentFeds, pf) {
			// Option 1: Remove as a server
			/*if list.State.isAuthorityChain(pf.GetChainID()) != -1 {
				removeEntry := adminBlock.NewRemoveFederatedServer(pf.GetChainID(), currentDBHeight+1)
				d.AdminBlock.AddFirstABEntry(removeEntry)
			}*/
			// Option 2: Demote to Audit if it is there
			if containsServer(currentAuds, pf) {
				demoteEntry := adminBlock.NewAddAuditServer(pf.GetChainID(), currentDBHeight+1)
				d.AdminBlock.AddFirstABEntry(demoteEntry)
			}
			_ = currentAuds
		}
	}

	hash, err = p.AdminBlock.BackReferenceHash()
	if err != nil {
		panic(err.Error())
	}
	d.AdminBlock.GetHeader().SetPrevBackRefHash(hash)

	p.FactoidBlock.SetDBHeight(previousDBHeight)
	d.FactoidBlock.SetDBHeight(currentDBHeight)
	d.FactoidBlock.SetPrevKeyMR(p.FactoidBlock.GetKeyMR())
	d.FactoidBlock.SetPrevLedgerKeyMR(p.FactoidBlock.GetLedgerKeyMR())

	fblock := d.FactoidBlock.(*factoid.FBlock)

	if len(fblock.Transactions) > 0 {
		coinbaseTx := fblock.Transactions[0]
		coinbaseTx.SetTimestamp(list.State.GetLeaderTimestamp())
		fblock.Transactions[0] = coinbaseTx
	}

	d.FactoidBlock = fblock

	d.DirectoryBlock.GetHeader().SetPrevFullHash(p.DirectoryBlock.GetFullHash())
	d.DirectoryBlock.GetHeader().SetPrevKeyMR(p.DirectoryBlock.GetKeyMR())
	d.DirectoryBlock.GetHeader().SetTimestamp(list.State.GetLeaderTimestamp())
	d.DirectoryBlock.GetHeader().SetNetworkID(list.State.GetNetworkID())

	d.DirectoryBlock.SetABlockHash(d.AdminBlock)
	d.DirectoryBlock.SetECBlockHash(d.EntryCreditBlock)
	d.DirectoryBlock.SetFBlockHash(d.FactoidBlock)

	pl := list.State.ProcessLists.Get(currentDBHeight)

	//for _, eb := range pl.NewEBlocks {
	//	eb.BuildHeader()
	//	eb.BodyKeyMR()
	//	eb.KeyMR()
	//}

	for _, eb := range pl.NewEBlocks {
		key, err := eb.KeyMR()
		if err != nil {
			panic(err.Error())
		}
		d.DirectoryBlock.AddEntry(eb.GetChainID(), key)
	}

	d.DirectoryBlock.BuildBodyMR()
	d.DirectoryBlock.MarshalBinary()

	progress = true
	d.isNew = false
	return
}

func (list *DBStateList) ProcessBlocks(d *DBState) (progress bool) {
	dbht := d.DirectoryBlock.GetHeader().GetDBHeight()

	if d.Locked || d.isNew {
		return
	}

	if dbht > 1 {
		pd := list.State.DBStates.Get(int(dbht - 1))
		if pd != nil && !pd.Saved {
			list.State.AddStatus(fmt.Sprintf("PROCESSBLOCKS:  Previous dbstate (%d) not saved", dbht-1))
			return
		}
	}

	list.LastTime = list.State.GetTimestamp() // If I saved or processed stuff, I'm good for a while

	// Bring the current federated servers and audit servers forward to the
	// next block.

	if list.State.DebugConsensus {
		PrintState(list.State)
	}

	// Saving our state so we can reset it if we need to.
	d.SaveStruct = SaveFactomdState(list.State, d)

	ht := d.DirectoryBlock.GetHeader().GetDBHeight()
	pl := list.State.ProcessLists.Get(ht)
	pln := list.State.ProcessLists.Get(ht + 1)

	var out bytes.Buffer
	out.WriteString("=== AdminBlock.UpdateState() Start ===\n")
	prt := func(lable string, pl *ProcessList) {
		out.WriteString(fmt.Sprintf("%19s %20s (%4d)", list.State.FactomNodeName, lable, pl.DBHeight))
		out.WriteString("Fed: ")
		for _, f := range pl.FedServers {
			out.WriteString(fmt.Sprintf("%x ", f.GetChainID().Bytes()[3:5]))
		}
		out.WriteString("---Audit: ")
		for _, f := range pl.AuditServers {
			out.WriteString(fmt.Sprintf("%x ", f.GetChainID().Bytes()[3:5]))
		}
		out.WriteString("\n")
	}

	prt("pl 1st", pl)
	prt("pln 1st", pln)

	//
	// ***** Apply the AdminBlock chainges to the next DBState
	//
	list.State.AddStatus(fmt.Sprintf("PROCESSBLOCKS:  Processing Admin Block at dbht: %d", d.AdminBlock.GetDBHeight()))
	d.AdminBlock.UpdateState(list.State)
	d.EntryCreditBlock.UpdateState(list.State)

	prt("pl 2st", pl)
	prt("pln 2st", pln)

	pln2 := list.State.ProcessLists.Get(ht + 2)
	pln2.FedServers = append(pln2.FedServers[:0], pln.FedServers...)
	pln2.AuditServers = append(pln2.AuditServers[:0], pln.AuditServers...)

	prt("pln2 3st", pln2)

	pln2.SortAuditServers()
	pln2.SortFedServers()

	pl.SortAuditServers()
	pl.SortFedServers()
	pln.SortAuditServers()
	pln.SortFedServers()

	prt("pl 4th", pl)
	prt("pln 4th", pln)
	prt("pln2 4th", pln2)

	out.WriteString("=== AdminBlock.UpdateState() End ===")
	fmt.Println(out.String())

	// Process the Factoid End of Block
	fs := list.State.GetFactoidState()
	fs.AddTransactionBlock(d.FactoidBlock)
	fs.AddECBlock(d.EntryCreditBlock)

	// Make the current exchange rate whatever we had in the previous block.
	// UNLESS there was a FER entry processed during this block  changeheight will be left at 1 on a change block
	if list.State.FERChangeHeight == 1 {
		list.State.FERChangeHeight = 0
	} else {
		if list.State.FactoshisPerEC != d.FactoidBlock.GetExchRate() {
			list.State.AddStatus(fmt.Sprint("PROCESSBLOCKS:  setting rate", list.State.FactoshisPerEC,
				" to ", d.FactoidBlock.GetExchRate(),
				" - Height ", d.DirectoryBlock.GetHeader().GetDBHeight()))
		}
		list.State.FactoshisPerEC = d.FactoidBlock.GetExchRate()
	}

	fs.ProcessEndOfBlock(list.State)

	// Promote the currently scheduled next FER

	list.State.ProcessRecentFERChainEntries()
	// Step my counter of Complete blocks
	i := d.DirectoryBlock.GetHeader().GetDBHeight() - list.Base
	if uint32(i) > list.Complete {
		list.Complete = uint32(i)
	}
	progress = true
	d.Locked = true // Only after all is done will I admit this state has been saved.

	pln.SortFedServers()
	pln.SortAuditServers()

	s := list.State
	// Time out commits every now and again.
	now := s.GetTimestamp()
	for k := range s.Commits {
		var keep []interfaces.IMsg
		commits := s.Commits[k]

		// Check to see if an entry Reveal has negated any pending commits.  All commits to the same EntryReveal
		// are discarded after we have recorded said Entry Reveal
		if len(commits) == 0 {
			delete(s.Commits, k)
		} else {
			{
				c, ok := s.Commits[k][0].(*messages.CommitChainMsg)
				if ok && !s.NoEntryYet(c.CommitChain.EntryHash, now) {
					delete(s.Commits, k)
					continue
				}
			}
			c, ok := s.Commits[k][0].(*messages.CommitEntryMsg)
			if ok && !s.NoEntryYet(c.CommitEntry.EntryHash, now) {
				delete(s.Commits, k)
				continue
			}
		}

		for _, v := range commits {
			_, ok := s.Replay.Valid(constants.TIME_TEST, v.GetRepeatHash().Fixed(), v.GetTimestamp(), now)
			if ok {
				keep = append(keep, v)
			}
		}
		if len(keep) > 0 {
			s.Commits[k] = keep
		} else {
			delete(s.Commits, k)
		}
	}

	return
}

func (list *DBStateList) SaveDBStateToDB(d *DBState) (progress bool) {
	// Take the height, and some function of the identity chain, and use that to decide to trim.  That
	// way, not all nodes in a simulation Trim() at the same time.
	v := int(d.DirectoryBlock.GetHeader().GetDBHeight()) + int(list.State.IdentityChainID.Bytes()[0])
	if v%4 == 0 {
		list.State.DB.Trim()
	}

	if !d.Locked || !d.ReadyToSave {
		return
	}

	if d.Saved {
		dblk, _ := list.State.DB.FetchDBKeyMRByHash(d.DirectoryBlock.GetKeyMR())
		if dblk == nil {
			panic(fmt.Sprintf("Claimed to be found on %s DBHeight %d Hash %x",
				list.State.FactomNodeName,
				d.DirectoryBlock.GetHeader().GetDBHeight(),
				d.DirectoryBlock.GetKeyMR().Bytes()))
		}
		return
	}

	head, _ := list.State.DB.FetchDirectoryBlockHead()

	list.State.DB.StartMultiBatch()

	if err := list.State.DB.ProcessABlockMultiBatch(d.AdminBlock); err != nil {
		panic(err.Error())
	}

	if err := list.State.DB.ProcessFBlockMultiBatch(d.FactoidBlock); err != nil {
		panic(err.Error())
	}

	if err := list.State.DB.ProcessECBlockMultiBatch(d.EntryCreditBlock, false); err != nil {
		panic(err.Error())
	}

	if len(d.EntryBlocks) > 0 {
		for _, eb := range d.EntryBlocks {
			if err := list.State.DB.ProcessEBlockMultiBatch(eb, true); err != nil {
				panic(err.Error())
			}
		}
		for _, e := range d.Entries {
			if err := list.State.DB.InsertEntryMultiBatch(e); err != nil {
				panic(err.Error())
			}
		}
	} else {
		pl := list.State.ProcessLists.Get(d.DirectoryBlock.GetHeader().GetDBHeight())
		if pl != nil {
			for _, eb := range pl.NewEBlocks {
				if err := list.State.DB.ProcessEBlockMultiBatch(eb, true); err != nil {
					panic(err.Error())
				}

				for _, e := range eb.GetBody().GetEBEntries() {
					if err := list.State.DB.InsertEntryMultiBatch(pl.GetNewEntry(e.Fixed())); err != nil {
						panic(err.Error())
					}
				}
			}
		}
	}

	if err := list.State.DB.ProcessDBlockMultiBatch(d.DirectoryBlock); err != nil {
		panic(err.Error())
	}

	if err := list.State.DB.ExecuteMultiBatch(); err != nil {
		panic(err.Error())
	}

	if d.DirectoryBlock.GetHeader().GetDBHeight() > 0 && d.DirectoryBlock.GetHeader().GetDBHeight() < head.GetHeader().GetDBHeight() {
		list.State.DB.SaveDirectoryBlockHead(head)
	}

	progress = true
	d.ReadyToSave = false
	d.Saved = true
	return
}

func (list *DBStateList) UpdateState() (progress bool) {

	list.Catchup()

	saved := 0
	for i, d := range list.DBStates {

		//fmt.Printf("dddd %20s %10s --- %10s %10v %10s %10v \n", "DBStateList Update", list.State.FactomNodeName, "Looking at", i, "DBHeight", list.Base+uint32(i))

		// Must process blocks in sequence.  Missing a block says we must stop.
		if d == nil {
			return
		}

		if i > 0 {
			progress = list.FixupLinks(list.DBStates[i-1], d)
		}

		progress = list.ProcessBlocks(d) || progress

		progress = list.SaveDBStateToDB(d) || progress

		// Make sure we move forward the Adminblock state in the process lists
		list.State.ProcessLists.Get(d.DirectoryBlock.GetHeader().GetDBHeight() + 1)

		if d.Saved {
			saved = i
		}
		if i-saved > 1 {
			break
		}
	}
	return
}

func (list *DBStateList) Last() *DBState {
	last := (*DBState)(nil)
	for _, ds := range list.DBStates {
		if ds == nil || ds.DirectoryBlock == nil {
			return last
		}
		last = ds
	}
	return last
}

func (list *DBStateList) Highest() uint32 {
	high := list.Base + uint32(len(list.DBStates)) - 1
	if high == 0 && len(list.DBStates) == 1 {
		return 1
	}
	return high
}

// Return true if we actually added the dbstate to the list
func (list *DBStateList) Put(dbState *DBState) bool {

	dblk := dbState.DirectoryBlock
	dbheight := dblk.GetHeader().GetDBHeight()

	// Count completed states, starting from the beginning (since base starts at
	// zero.
	cnt := 0
searchLoop:
	for i, v := range list.DBStates {
		if v == nil || v.DirectoryBlock == nil || !v.Locked {
			if v != nil && v.DirectoryBlock == nil {
				list.DBStates[i] = nil
			}
			break searchLoop
		}
		cnt++
	}

	keep := uint32(5) // How many states to keep around; debugging helps with more.
	if uint32(cnt) > keep {
		var dbstates []*DBState
		dbstates = append(dbstates, list.DBStates[cnt-int(keep):]...)
		list.DBStates = dbstates
		list.Base = list.Base + uint32(cnt) - keep
		list.Complete = list.Complete - uint32(cnt) + keep
	}

	index := int(dbheight) - int(list.Base)

	// If we have already processed this State, ignore it.
	if index < int(list.Complete) {
		return false
	}

	// make room for this entry.
	for len(list.DBStates) <= index {
		list.DBStates = append(list.DBStates, nil)
	}
	if list.DBStates[index] == nil {
		list.DBStates[index] = dbState
	}

	return true
}

func (list *DBStateList) Get(height int) *DBState {
	if height < 0 {
		return nil
	}
	i := height - int(list.Base)
	if i < 0 || i >= len(list.DBStates) {
		return nil
	}
	return list.DBStates[i]
}

func (list *DBStateList) NewDBState(isNew bool,
	directoryBlock interfaces.IDirectoryBlock,
	adminBlock interfaces.IAdminBlock,
	factoidBlock interfaces.IFBlock,
	entryCreditBlock interfaces.IEntryCreditBlock,
	eBlocks []interfaces.IEntryBlock,
	entries []interfaces.IEBEntry) *DBState {

	dbState := new(DBState)

	dbState.DBHash = directoryBlock.DatabasePrimaryIndex()
	dbState.ABHash = adminBlock.DatabasePrimaryIndex()
	dbState.FBHash = factoidBlock.DatabasePrimaryIndex()
	dbState.ECHash = entryCreditBlock.DatabasePrimaryIndex()

	dbState.isNew = isNew
	dbState.DirectoryBlock = directoryBlock
	dbState.AdminBlock = adminBlock
	dbState.FactoidBlock = factoidBlock
	dbState.EntryCreditBlock = entryCreditBlock
	dbState.EntryBlocks = eBlocks
	dbState.Entries = entries

	// If we actually add this to the list, return the dbstate.
	if list.Put(dbState) {
		return dbState
	} else {
		ht := dbState.DirectoryBlock.GetHeader().GetDBHeight()
		if ht == list.State.GetHighestSavedBlock() {
			index := int(ht) - int(list.State.DBStates.Base)
			if index > 0 {
				list.State.DBStates.DBStates[index] = dbState
				pdbs := list.State.DBStates.Get(int(ht - 1))
				if pdbs != nil {
					pdbs.SaveStruct.TrimBack(list.State, dbState)
				}
			}
		}
	}

	// Failed, so return nil
	return nil
}
