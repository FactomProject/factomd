// Copyright 2015 FactomProject Authors. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package state

import (
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/log"
	"time"
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

	dbstring         string
	DirectoryBlock   interfaces.IDirectoryBlock
	AdminBlock       interfaces.IAdminBlock
	FactoidBlock     interfaces.IFBlock
	EntryCreditBlock interfaces.IEntryCreditBlock
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

const SecondsBetweenTests = 1 // Default

func (list *DBStateList) String() string {
	str := "\nDBStates\n"
	str = fmt.Sprintf("%s  Base      = %d\n", str, list.Base)
	str = fmt.Sprintf("%s  timestamp = %s\n", str, list.LastTime.String())
	str = fmt.Sprintf("%s  Complete  = %d\n", str, list.Complete)
	rec := "M"
	last := ""
	for i, ds := range list.DBStates {
		rec = "M"
		if ds != nil && ds.DirectoryBlock != nil {
			dblk, _ := list.State.DB.FetchDBlockByHash(ds.DirectoryBlock.GetKeyMR())
			if dblk != nil {
				rec = "R"
			}
			if ds.Saved {
				rec = "S"
			}
		}
		if last != "" {
			str = last
		}
		str = fmt.Sprintf("%s  %1s-DState\n   DState Height: %d\n%s", str, rec, list.Base+uint32(i), ds.String())
		if rec == "M" && last == "" {
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
	ht := uint32(0)
	for i, dbstate := range list.DBStates {
		if dbstate != nil && dbstate.Saved {
			ht = list.Base + uint32(i)
		}
	}

	return ht
}

// Once a second at most, we check to see if we need to pull down some blocks to catch up.
func (list *DBStateList) Catchup() {

	now := list.State.GetTimestamp()

	dbsHeight := list.GetHighestRecordedBlock()

	// We only check if we need updates once every so often.
	if int(now)/1000-int(list.LastTime)/1000 < SecondsBetweenTests {
		return
	}
	list.LastTime = now

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
			return
		}

		if plHeight > dbsHeight && plHeight-dbsHeight > 1 {
			begin = int(dbsHeight + 1)
			end = int(plHeight - 1)
		} else {
			return
		}
	}

	list.Lastreq = begin

	end2 := begin + 400
	if end < end2 {
		end2 = end
	}

	msg := messages.NewDBStateMissing(list.State, uint32(begin), uint32(end2))

	if msg != nil {
		//list.State.ProcessLists = NewProcessLists(list.State)
		list.State.EOM = 0
		list.State.GreenFlg = false
		list.State.NetworkOutMsgQueue() <- msg
	}

}

func (list *DBStateList) FixupLinks(i int, d *DBState) {
	p := list.DBStates[i-1]

	// If this block is new, then make sure all hashes are fully computed.
	if d.isNew {

		hash, _ := p.EntryCreditBlock.HeaderHash()
		d.EntryCreditBlock.GetHeader().SetPrevHeaderHash(hash)

		hash, _ = p.EntryCreditBlock.GetFullHash()
		d.EntryCreditBlock.GetHeader().SetPrevFullHash(hash)

		d.AdminBlock.GetHeader().SetPrevFullHash(hash)

		p.FactoidBlock.SetDBHeight(p.DirectoryBlock.GetHeader().GetDBHeight())
		d.FactoidBlock.SetDBHeight(d.DirectoryBlock.GetHeader().GetDBHeight())
		d.FactoidBlock.SetPrevKeyMR(p.FactoidBlock.GetKeyMR().Bytes())
		d.FactoidBlock.SetPrevFullHash(p.FactoidBlock.GetPrevFullHash().Bytes())

		d.DirectoryBlock.GetHeader().SetPrevFullHash(p.DirectoryBlock.GetFullHash())
		d.DirectoryBlock.GetHeader().SetPrevKeyMR(p.DirectoryBlock.GetKeyMR())
		d.DirectoryBlock.GetHeader().SetTimestamp(0)

		d.DirectoryBlock.GetDBEntries()[0].SetKeyMR(d.AdminBlock.GetHash())
		d.DirectoryBlock.GetDBEntries()[1].SetKeyMR(d.EntryCreditBlock.GetHash())
		d.DirectoryBlock.GetDBEntries()[2].SetKeyMR(d.FactoidBlock.GetHash())

		pl := list.State.ProcessLists.Get(d.DirectoryBlock.GetHeader().GetDBHeight())

		for _, eb := range pl.NewEBlocks {
			key, err := eb.KeyMR()
			if err != nil {
				panic(err.Error())
			}
			d.DirectoryBlock.AddEntry(eb.GetChainID(), key)
		}
		d.DirectoryBlock.BuildBodyMR()

		//d.DirectoryBlock.GetKeyMR()
		//_, err := d.DirectoryBlock.BuildBodyMR()
		//if err != nil {
		//	panic(err.Error())
		//}
	}
}

func (list *DBStateList) UpdateState() (progress bool) {

	list.Catchup()

	for i, d := range list.DBStates {

		// Must process blocks in sequence.  Missing a block says we must stop.
		if d == nil {
			return
		}

		if d.Saved {
			continue
		}

		// Make sure the directory block is properly synced up with the prior block, if there
		// is one.

		dblk, _ := list.State.DB.FetchDBlockByKeyMR(d.DirectoryBlock.GetKeyMR())
		if dblk == nil {
			if i > 0 {
				p := list.DBStates[i-1]
				if !p.Saved {
					break
				}
			}

			//fmt.Println("Saving DBHeight ", d.DirectoryBlock.GetHeader().GetDBHeight(), " on ", list.State.GetFactomNodeName())

			// If we have previous blocks, update blocks that this follower potentially constructed.  We can optimize and skip
			// this step if we got the block from a peer.  TODO we must however check the sigantures on the
			// block before we write it to disk.
			if i > 0 {
				list.FixupLinks(i, d)
			}
			d.DirectoryBlock.MarshalBinary()
			d.dbstring = d.DirectoryBlock.String()

			list.State.DB.StartMultiBatch()

			if err := list.State.DB.ProcessDBlockMultiBatch(d.DirectoryBlock); err != nil {
				panic(err.Error())
			}

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

			if err := list.State.DB.ExecuteMultiBatch(); err != nil {
				panic(err.Error())
			}

		}

		/*
		dblk2, _ := list.State.DB.FetchDBlockByKeyMR(d.DirectoryBlock.GetKeyMR())
		if dblk2 == nil {
			fmt.Printf("Failed to save the Directory Block %d %x\n",
				d.DirectoryBlock.GetHeader().GetDBHeight(),
				d.DirectoryBlock.GetKeyMR().Bytes()[:3])
			panic("Failed to save Directory Block")
		}
		keyMR2 := dblk2.GetKeyMR()
		if !d.DirectoryBlock.GetKeyMR().IsSameAs(keyMR2) {
			fmt.Println(dblk == nil)
			fmt.Printf("Keys differ %x and %x", d.DirectoryBlock.GetKeyMR().Bytes()[:3], keyMR2.Bytes()[:3])
			panic("KeyMR failure")
		}
		if i > 0 {
			dbprev, _ := list.State.DB.FetchDBlockByKeyMR(d.DirectoryBlock.GetHeader().GetPrevKeyMR())
			if dbprev == nil {
				fmt.Println(list.DBStates[i-1].dbstring)
				fmt.Println(list.DBStates[i-1].DirectoryBlock.String())
				fmt.Println(d.DirectoryBlock.String())
				panic(fmt.Sprintf("%s Hashes have been altered for Directory Blocks", list.State.FactomNodeName))
			}
		}
	  */
		list.LastTime = list.State.GetTimestamp() // If I saved or processed stuff, I'm good for a while
		d.Saved = true                            // Only after all is done will I admit this state has been saved.

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
		if uint32(i) > list.Complete {
			list.Complete = uint32(i)
		}
		progress = true
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

	// Hold off on any requests if I'm actually processing...
	list.LastTime = list.State.GetTimestamp()

	dblk := dbState.DirectoryBlock
	dbheight := dblk.GetHeader().GetDBHeight()

	// Count completed states, starting from the beginning (since base starts at
	// zero.
	cnt := 0
	for i, v := range list.DBStates {
		if v == nil || v.DirectoryBlock == nil || !v.Saved {
			if v != nil && v.DirectoryBlock == nil {
				list.DBStates[i] = nil
			}
			break
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
		if list.State.GetOut() {
			list.State.Println("Ignoring!  Block vs Base: ", dbheight, "/", list.Base)
		}
		return
	}

	// make room for this entry.
	for len(list.DBStates) <= index {
		list.DBStates = append(list.DBStates, nil)
	}
	if list.DBStates[index] == nil {
		list.DBStates[index] = dbState
	}

	hash, err := dbState.AdminBlock.GetKeyMR()
	if err != nil {
		panic(err)
	}
	dbState.DirectoryBlock.GetDBEntries()[0].SetKeyMR(hash)
	hash, err = dbState.EntryCreditBlock.GetFullHash()
	if err != nil {
		panic(err)
	}
	dbState.DirectoryBlock.GetDBEntries()[1].SetKeyMR(hash)
	hash = dbState.FactoidBlock.GetHash()
	dbState.DirectoryBlock.GetDBEntries()[2].SetKeyMR(hash)

}

func (list *DBStateList) Get(height uint32) *DBState {
	i := int(height) - int(list.Base)
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

	dbState.DBHash = directoryBlock.GetHash()
	dbState.ABHash = adminBlock.GetHash()
	dbState.FBHash = factoidBlock.GetHash()
	dbState.ECHash = entryCreditBlock.GetHash()

	dbState.isNew = isNew
	dbState.DirectoryBlock = directoryBlock
	dbState.AdminBlock = adminBlock
	dbState.FactoidBlock = factoidBlock
	dbState.EntryCreditBlock = entryCreditBlock

	list.Put(dbState)
	return dbState
}
