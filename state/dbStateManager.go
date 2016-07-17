// Copyright 2015 FactomProject Authors. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package state

import (
	"encoding/hex"
	"fmt"
	"time"

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

	DBHash interfaces.IHash
	ABHash interfaces.IHash
	FBHash interfaces.IHash
	ECHash interfaces.IHash

	DirectoryBlock   interfaces.IDirectoryBlock
	AdminBlock       interfaces.IAdminBlock
	FactoidBlock     interfaces.IFBlock
	EntryCreditBlock interfaces.IEntryCreditBlock
	Locked           bool
	ReadyToSave      bool
	Saved            bool
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

const SecondsBetweenTests = 20 // Default

func (list *DBStateList) String() string {
	str := "\nDBStates\n"
	str = fmt.Sprintf("%s  Base      = %d\n", str, list.Base)
	ts := "-nil-"
	if list.LastTime != nil {
		ts = list.LastTime.String()
	}
	str = fmt.Sprintf("%s  timestamp = %s\n", str, ts)
	str = fmt.Sprintf("%s  Complete  = %d\n", str, list.Complete)
	rec := "M"
	last := ""
	for i, ds := range list.DBStates {
		rec = "?"
		if ds != nil {
			rec = "nil"
			if ds.DirectoryBlock != nil {
				rec = "x"

				dblk, _ := list.State.DB.FetchDBlock(ds.DirectoryBlock.GetKeyMR())
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
		str = fmt.Sprintf("%s  %1s-DState ?-(DState nil) x-(Not in DB) s-(In DB) L-(Locked) R-(Ready to Save) S-(Saved)\n   DState Height: %d\n%s", str, rec, list.Base+uint32(i), ds.String())
		if rec == "?" && last == "" {
			last = str
		}
	}

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

func (list *DBStateList) GetHighestRecordedBlock() uint32 {
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

// Once a second at most, we check to see if we need to pull down some blocks to catch up.
func (list *DBStateList) Catchup() {

	now := list.State.GetTimestamp()

	dbsHeight := list.GetHighestRecordedBlock()

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

		if plHeight >= dbsHeight && plHeight-dbsHeight > 1 {
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

	list.Lastreq = begin

	end2 := begin + 400
	if end < end2 {
		end2 = end
	}

	if list.LastTime != nil && now.GetTimeSeconds()-list.LastTime.GetTimeSeconds() < SecondsBetweenTests {
		return
	}

	list.LastTime = now

	msg := messages.NewDBStateMissing(list.State, uint32(begin), uint32(end2))

	if msg != nil {
		list.State.RunLeader = false
		list.State.StartDelay = list.State.GetTimestamp().GetTimeMilli()
		list.State.NetworkOutMsgQueue() <- msg
		list.LastTime = now
	}

}

func (list *DBStateList) FixupLinks(p *DBState, d *DBState) (progress bool) {

	// If this block is new, then make sure all hashes are fully computed.
	if !d.isNew {
		return
	}

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
	d.EntryCreditBlock.GetHeader().SetDBHeight(d.DirectoryBlock.GetHeader().GetDBHeight())

	hash, err = p.AdminBlock.BackReferenceHash()
	if err != nil {
		panic(err.Error())
	}

	d.AdminBlock.GetHeader().SetPrevBackRefHash(hash)

	p.FactoidBlock.SetDBHeight(p.DirectoryBlock.GetHeader().GetDBHeight())
	d.FactoidBlock.SetDBHeight(d.DirectoryBlock.GetHeader().GetDBHeight())
	d.FactoidBlock.SetPrevKeyMR(p.FactoidBlock.GetKeyMR())
	d.FactoidBlock.SetPrevLedgerKeyMR(p.FactoidBlock.GetLedgerMR())

	d.DirectoryBlock.GetHeader().SetPrevFullHash(p.DirectoryBlock.GetFullHash())
	d.DirectoryBlock.GetHeader().SetPrevKeyMR(p.DirectoryBlock.GetKeyMR())
	d.DirectoryBlock.GetHeader().SetTimestamp(list.State.GetLeaderTimestamp())

	d.DirectoryBlock.SetABlockHash(d.AdminBlock)
	d.DirectoryBlock.SetECBlockHash(d.EntryCreditBlock)
	d.DirectoryBlock.SetFBlockHash(d.FactoidBlock)

	pl := list.State.ProcessLists.Get(d.DirectoryBlock.GetHeader().GetDBHeight())

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
	if d.Locked {
		return
	}

	list.LastTime = nil // If I saved or processed stuff, I'm good for a while

	// Any updates required to the state as established by the AdminBlock are applied here.
	d.AdminBlock.UpdateState(list.State)

	// Process the Factoid End of Block
	fs := list.State.GetFactoidState()
	fs.AddTransactionBlock(d.FactoidBlock)
	fs.AddECBlock(d.EntryCreditBlock)
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

	return
}

func (list *DBStateList) SaveDBStateToDB(d *DBState) (progress bool) {

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

	// Take the height, and some function of the identity chain, and use that to decide to trim.  That
	// way, not all nodes in a simulation Trim() at the same time.
	v := int(d.DirectoryBlock.GetHeader().GetDBHeight()) + int(list.State.IdentityChainID.Bytes()[0])
	if v%4 == 0 {
		list.State.DB.Trim()
	}

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
	pl := list.State.ProcessLists.Get(d.DirectoryBlock.GetHeader().GetDBHeight())
	if pl != nil {
		for _, eb := range pl.NewEBlocks {
			if err := list.State.DB.ProcessEBlockMultiBatch(eb, false); err != nil {
				panic(err.Error())
			}

			for _, e := range eb.GetBody().GetEBEntries() {
				if err := list.State.DB.InsertEntry(pl.NewEntries[e.Fixed()]); err != nil {
					panic(err.Error())
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
	progress = true
	d.ReadyToSave = false
	d.Saved = true
	return
}

func (list *DBStateList) UpdateState() (progress bool) {

	list.Catchup()

	for i, d := range list.DBStates {

		//fmt.Printf("dddd %20s %10s --- %10s %10v %10s %10v \n", "DBStateList Update", list.State.FactomNodeName, "Looking at", i, "DBHeight", list.Base+uint32(i))

		// Must process blocks in sequence.  Missing a block says we must stop.
		if d == nil {
			return
		}

		if d.Saved {
			continue
		}

		if i > 0 {
			progress = list.FixupLinks(list.DBStates[i-1], d)
		}
		progress = list.ProcessBlocks(d) || progress

		progress = list.SaveDBStateToDB(d) || progress

		// Make sure we move forward the Adminblock state in the process lists
		list.State.ProcessLists.Get(d.DirectoryBlock.GetHeader().GetDBHeight() + 1)

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

func (list *DBStateList) Put(dbState *DBState) {

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

	keep := uint32(2) // How many states to keep around; debugging helps with more.
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
		return
	}

	// make room for this entry.
	for len(list.DBStates) <= index {
		list.DBStates = append(list.DBStates, nil)
	}
	if list.DBStates[index] == nil {
		list.DBStates[index] = dbState
	}
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
	entryCreditBlock interfaces.IEntryCreditBlock) *DBState {

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

	list.Put(dbState)

	return dbState
}
