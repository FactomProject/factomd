// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package state

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/util/atomic"

	// "github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/identity"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/log"
)

var _ = hex.EncodeToString
var _ = fmt.Print
var _ = time.Now()
var _ = log.Print

type DBState struct {
	IsNew bool

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

	Repeat      bool
	ReadyToSave bool
	Locked      bool
	Signed      bool
	Saved       bool

	Added interfaces.Timestamp

	FinalExchangeRate uint64
	NextTimestamp     interfaces.Timestamp
}

var _ interfaces.BinaryMarshallable = (*DBState)(nil)

func (dbs *DBState) Init() {
	if dbs.SaveStruct == nil {
		dbs.SaveStruct = new(SaveState)
		dbs.SaveStruct.Init()
	}
	if dbs.SaveStruct.IdentityControl == nil {
		atomic.WhereAmIMsg("no identity manager")
	}

	if dbs.DBHash == nil {
		dbs.DBHash = primitives.NewZeroHash()
	}
	if dbs.ABHash == nil {
		dbs.ABHash = primitives.NewZeroHash()
	}
	if dbs.FBHash == nil {
		dbs.FBHash = primitives.NewZeroHash()
	}
	if dbs.ECHash == nil {
		dbs.ECHash = primitives.NewZeroHash()
	}

	if dbs.DirectoryBlock == nil {
		dbs.DirectoryBlock = directoryBlock.NewDirectoryBlock(nil)
	}
	if dbs.AdminBlock == nil {
		dbs.AdminBlock = adminBlock.NewAdminBlock(nil)
	}
	if dbs.FactoidBlock == nil {
		dbs.FactoidBlock = factoid.NewFBlock(nil)
	}
	if dbs.EntryCreditBlock == nil {
		dbs.EntryCreditBlock = entryCreditBlock.NewECBlock()
	}

	if dbs.Added == nil {
		dbs.Added = primitives.NewTimestampFromMilliseconds(0)
	}
	if dbs.NextTimestamp == nil {
		dbs.NextTimestamp = primitives.NewTimestampFromMilliseconds(0)
	}
}

func (a *DBState) IsSameAs(b *DBState) bool {
	if a == nil || b == nil {
		if a == nil && b == nil {
			return true
		}
		return false
	}

	if a.IsNew != b.IsNew {
		return false
	}

	if a.SaveStruct != nil {
		if a.SaveStruct.IsSameAs(b.SaveStruct) == false {
			return false
		}
	} else {
		if b.SaveStruct != nil {
			return false
		}
	}

	if a.DBHash.IsSameAs(b.DBHash) == false {
		return false
	}
	if a.ABHash.IsSameAs(b.ABHash) == false {
		return false
	}
	if a.FBHash.IsSameAs(b.FBHash) == false {
		return false
	}
	if a.ECHash.IsSameAs(b.ECHash) == false {
		return false
	}

	if a.DirectoryBlock.IsSameAs(b.DirectoryBlock) == false {
		return false
	}
	if a.AdminBlock.IsSameAs(b.AdminBlock) == false {
		return false
	}
	if a.FactoidBlock.IsSameAs(b.FactoidBlock) == false {
		return false
	}
	if a.EntryCreditBlock.IsSameAs(b.EntryCreditBlock) == false {
		return false
	}

	if len(a.EntryBlocks) != len(b.EntryBlocks) {
		return false
	}
	for i := range a.EntryBlocks {
		if a.EntryBlocks[i].IsSameAs(b.EntryBlocks[i]) == false {
			return false
		}
	}

	if len(a.Entries) != len(b.Entries) {
		return false
	}
	for i := range a.Entries {
		if a.Entries[i].IsSameAs(b.Entries[i]) == false {
			return false
		}
	}

	if a.Repeat != b.Repeat {
		return false
	}
	if a.ReadyToSave != b.ReadyToSave {
		return false
	}
	if a.Locked != b.Locked {
		return false
	}
	if a.Signed != b.Signed {
		return false
	}
	if a.Saved != b.Saved {
		return false
	}

	if a.Added.IsSameAs(b.Added) == false {
		return false
	}

	if a.FinalExchangeRate != b.FinalExchangeRate {
		return false
	}

	if a.NextTimestamp.IsSameAs(b.NextTimestamp) == false {
		return false
	}

	return true
}

func (dbs *DBState) MarshalBinary() (rval []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("DBState.MarshalBinary panic Error Marshalling a dbstate %v", r)
		}
	}()

	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "DBState.MarshalBinary err:%v", *pe)

		}
	}(&err)

	dbs.Init()
	b := primitives.NewBuffer(nil)

	err = b.PushBinaryMarshallable(dbs.SaveStruct)
	if err != nil {
		return nil, err
	}

	err = b.PushBinaryMarshallable(dbs.DBHash)
	if err != nil {
		return nil, err
	}
	err = b.PushBinaryMarshallable(dbs.ABHash)
	if err != nil {
		return nil, err
	}
	err = b.PushBinaryMarshallable(dbs.FBHash)
	if err != nil {
		return nil, err
	}
	err = b.PushBinaryMarshallable(dbs.ECHash)
	if err != nil {
		return nil, err
	}

	err = b.PushBinaryMarshallable(dbs.DirectoryBlock)
	if err != nil {
		return nil, err
	}
	err = b.PushBinaryMarshallable(dbs.AdminBlock)
	if err != nil {
		return nil, err
	}
	err = b.PushBinaryMarshallable(dbs.FactoidBlock)
	if err != nil {
		return nil, err
	}
	err = b.PushBinaryMarshallable(dbs.EntryCreditBlock)
	if err != nil {
		return nil, err
	}

	err = b.PushBinaryMarshallable(dbs.NextTimestamp)
	if err != nil {
		return nil, err
	}
	return b.DeepCopyBytes(), nil
}

func (dbs *DBState) UnmarshalBinaryData(p []byte) (newData []byte, err error) {
	dbs.Init()

	dbs.EntryBlocks = []interfaces.IEntryBlock{}
	dbs.Entries = []interfaces.IEBEntry{}

	newData = p
	b := primitives.NewBuffer(p)

	dbs.IsNew = false

	SaveStruct := new(SaveState)
	SaveStruct.Init()
	if SaveStruct.IdentityControl == nil {
		atomic.WhereAmIMsg("no identity manager")
	}

	err = b.PopBinaryMarshallable(SaveStruct)
	if err != nil {
		return
	}

	err = b.PopBinaryMarshallable(dbs.DBHash)
	if err != nil {
		return
	}
	err = b.PopBinaryMarshallable(dbs.ABHash)
	if err != nil {
		return
	}
	err = b.PopBinaryMarshallable(dbs.FBHash)
	if err != nil {
		return
	}
	err = b.PopBinaryMarshallable(dbs.ECHash)
	if err != nil {
		return
	}

	err = b.PopBinaryMarshallable(dbs.DirectoryBlock)
	if err != nil {
		return
	}
	err = b.PopBinaryMarshallable(dbs.AdminBlock)
	if err != nil {
		return
	}
	err = b.PopBinaryMarshallable(dbs.FactoidBlock)
	if err != nil {
		return
	}
	err = b.PopBinaryMarshallable(dbs.EntryCreditBlock)
	if err != nil {
		return
	}

	dbs.Repeat = false
	dbs.ReadyToSave = true
	dbs.Locked = true
	dbs.Signed = true
	dbs.Saved = true

	err = b.PopBinaryMarshallable(dbs.NextTimestamp)
	if err != nil {
		return
	}

	dbs.SaveStruct = SaveStruct // OK, this worked so keep the save struct

	if dbs.SaveStruct.IdentityControl == nil {
		atomic.WhereAmIMsg("no identity manager")
	}

	newData = b.DeepCopyBytes()
	return
}

func (dbs *DBState) UnmarshalBinary(p []byte) error {
	_, err := dbs.UnmarshalBinaryData(p)
	return err
}

type DBStateList struct {
	LastEnd       int
	LastBegin     int
	TimeToAsk     interfaces.Timestamp
	ProcessHeight uint32
	SavedHeight   uint32
	State         *State
	Base          uint32
	Complete      uint32
	DBStates      []*DBState
}

var _ interfaces.BinaryMarshallable = (*DBStateList)(nil)

func (dbsl *DBStateList) Init() {
	if dbsl.TimeToAsk == nil {
		dbsl.TimeToAsk = primitives.NewTimestampFromMilliseconds(0)
	}
}

func (a *DBStateList) IsSameAs(b *DBStateList) bool {
	if a == nil || b == nil {
		if a == nil && b == nil {
			return true
		}
		return false
	}

	if a.LastEnd != b.LastEnd {
		return false
	}
	if a.LastBegin != b.LastBegin {
		return false
	}
	if a.TimeToAsk.IsSameAs(b.TimeToAsk) == false {
		return false
	}
	if a.ProcessHeight != b.ProcessHeight {
		return false
	}
	if a.SavedHeight != b.SavedHeight {
		return false
	}

	//State    *State
	if a.Base != b.Base {
		return false
	}
	if a.Complete != b.Complete {
		return false
	}

	if len(a.DBStates) != len(b.DBStates) {
		return false
	}
	for i := range a.DBStates {
		if a.DBStates[i].IsSameAs(b.DBStates[i]) == false {
			return false
		}
	}

	return true
}

func (dbsl *DBStateList) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "DBStateList.MarshalBinary err:%v", *pe)
		}
	}(&err)
	dbsl.Init()
	buf := primitives.NewBuffer(nil)

	err = buf.PushUInt32(uint32(dbsl.LastEnd))
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(uint32(dbsl.LastBegin))
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(dbsl.ProcessHeight)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(dbsl.SavedHeight)
	if err != nil {
		return nil, err
	}
	//TODO: handle State
	err = buf.PushUInt32(dbsl.Base)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(dbsl.Complete)
	if err != nil {
		return nil, err
	}
	dlen := 0
	for _, v := range dbsl.DBStates {
		if v.Saved {
			dlen++
		} else {
			break
		}
	}

	err = buf.PushVarInt(uint64(dlen))
	if err != nil {
		return nil, err
	}
	for _, v := range dbsl.DBStates {
		if v.Saved && v.Locked {
			if !v.Locked {
				panic("unlocked save state")
			}

			err = buf.PushBinaryMarshallable(v)
			if err != nil {
				return nil, err
			}
		} else {
			break
		}
	}

	return buf.DeepCopyBytes(), nil
}

func (dbsl *DBStateList) UnmarshalBinaryData(p []byte) (newData []byte, err error) {
	dbsl.Init()
	dbsl.DBStates = []*DBState{}
	newData = p

	buf := primitives.NewBuffer(p)

	x, err := buf.PopUInt32()
	if err != nil {
		dbsl.State.LogPrintf("dbstateprocess", "DBStateList.UnmarshalBinaryData LastEnd err: %v", err)
		return
	}
	dbsl.LastEnd = int(x)
	x, err = buf.PopUInt32()
	if err != nil {
		dbsl.State.LogPrintf("dbstateprocess", "DBStateList.UnmarshalBinaryData LastBegin err: %v", err)
		return
	}
	dbsl.LastBegin = int(x)

	dbsl.ProcessHeight, err = buf.PopUInt32()
	if err != nil {
		dbsl.State.LogPrintf("dbstateprocess", "DBStateList.UnmarshalBinaryData ProcessHeight err: %v", err)
		return
	}
	dbsl.SavedHeight, err = buf.PopUInt32()
	if err != nil {
		dbsl.State.LogPrintf("dbstateprocess", "DBStateList.UnmarshalBinaryData SavedHeight err: %v", err)
		return
	}

	//TODO: handle State
	dbsl.Base, err = buf.PopUInt32()
	if err != nil {
		dbsl.State.LogPrintf("dbstateprocess", "DBStateList.UnmarshalBinaryData Base err: %v", err)
		return
	}
	dbsl.Complete, err = buf.PopUInt32()
	if err != nil {
		dbsl.State.LogPrintf("dbstateprocess", "DBStateList.UnmarshalBinaryData Complete err: %v", err)
		return
	}

	listLen, err := buf.PopVarInt()
	if err != nil {
		dbsl.State.LogPrintf("dbstateprocess", "DBStateList.UnmarshalBinaryData listLen err: %v", err)
		return
	}
	for i := 0; i < int(listLen); i++ {
		dbs := new(DBState)
		err = buf.PopBinaryMarshallable(dbs)
		if dbs.SaveStruct.IdentityControl == nil {
			atomic.WhereAmIMsg("no identity control")
		}
		if err != nil {
			dbsl.State.LogPrintf("dbstateprocess", "DBStateList.UnmarshalBinaryData (%d) err: %v", int(dbsl.Base)+i, err)
			return
		}
		dbsl.DBStates = append(dbsl.DBStates, dbs)

	}

	newData = buf.DeepCopyBytes()

	return
}

func (dbsl *DBStateList) UnmarshalBinary(p []byte) error {
	_, err := dbsl.UnmarshalBinaryData(p)
	return err
}

// Validate this directory block given the next Directory Block.  Need to check the
// signatures as being from the authority set, and valid. Also check that this DBState holds
// a previous KeyMR that matches the previous DBState KeyMR.
//
// Return a -1 on failure.
//
func (d *DBState) ValidNext(state *State, next *messages.DBStateMsg) int {
	s := state
	_ = s
	dirblk := next.DirectoryBlock
	dbheight := dirblk.GetHeader().GetDBHeight()
	// If we don't have the previous blocks processed yet, then let's wait on this one.
	highestSavedBlk := state.GetHighestSavedBlk()

	if dbheight == 0 && highestSavedBlk == 0 {
		//state.AddStatus(fmt.Sprintf("DBState.ValidNext: rtn 1 genesis block is valid dbht: %d", dbheight))
		// The genesis block is valid by definition.
		return 1
	}

	// Don't reload blocks!
	if dbheight <= highestSavedBlk && !next.IsInDB {
		state.LogPrintf("dbstateprocess", "Invalid DBState because dbheight %d < highestSavedBlk %d",
			dbheight, highestSavedBlk)
		return -1
	}

	if dbheight > highestSavedBlk+1 {
		state.LogPrintf("dbstateprocess", "Invalid DBState because dbheight %d > highestSavedBlk+1 %d",
			dbheight, highestSavedBlk+1)
		return 0
	}
	// This node cannot be validated until the previous node (d) has been saved to disk, which means processed
	// and signed as well.
	if d == nil {
		state.LogPrintf("dbstateprocess", "Cannot validate because DBState is nil",
			dbheight)
		return 0
	}
	if !d.Locked || !d.Signed || !d.Saved {
		state.LogPrintf("dbstateprocess", "Cannot validate dbstate at dbht %d because DBState is not locked %v, "+
			"not Signed %v, or not saved %v", dbheight, d.Locked, d.Signed, d.Saved)
		return 0
	}

	valid := next.ValidateSignatures(state)
	if !next.IsInDB && !next.IgnoreSigs && valid != 1 {
		state.LogPrintf("dbstateprocess", "cannot validate dbstate at dbht %d (return %v) because "+
			"!nextIsInDB %v && !next.IgnoreSigs %v", dbheight, valid, next.IsInDB, next.IgnoreSigs)
		return valid
	}

	// Get the keymr of the Previous DBState
	pkeymr := d.DirectoryBlock.GetKeyMR()
	// Get the Previous KeyMR pointer in the possible new Directory Block
	prevkeymr := dirblk.GetHeader().GetPrevKeyMR()
	if !pkeymr.IsSameAs(prevkeymr) {
		pdir, err := state.DB.FetchDBlockByHeight(dbheight - 1)
		if err != nil {
			state.LogPrintf("dbstateprocess", "Invalid dbstate at dbht %d because "+
				"we can find no previous directory block at %d err: %v",
				dbheight, dbheight-1, err)
			return -1
		}

		if pkeymr.Fixed() != pdir.GetKeyMR().Fixed() {
			state.LogPrintf("dbstateprocess", "At DBHeight %d the previous block in the database keymr %x does not match the dbstate %x ",
				dbheight, pdir.GetKeyMR().Bytes(), pkeymr.Bytes())
			return -1
		}
	}

	return 1

}

func (list *DBStateList) String() string {
	str := "\n========DBStates Start=======\nddddd DBStates\n"
	str = fmt.Sprintf("dddd %s  Base      = %d\n", str, list.Base)
	ts := "-nil-"
	if list.TimeToAsk != nil {
		ts = list.TimeToAsk.String()
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

				dblk := ds.DirectoryBlock
				//				dblk, _ := list.State.GetDB().FetchDBlock(ds.DirectoryBlock.GetKeyMR())

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
		str = fmt.Sprintf("%s      State: IsNew %5v ReadyToSave %5v Locked %5v Signed %5v Saved %5v\n", str, ds.IsNew, ds.ReadyToSave, ds.Locked, ds.Signed, ds.Saved)
		str = fmt.Sprintf("%s      DBlk Height   = %v \n", str, ds.DirectoryBlock.GetHeader().GetDBHeight())
		str = fmt.Sprintf("%s      DBlock        = %x \n", str, ds.DirectoryBlock.GetHash().Bytes()[:5])
		str = fmt.Sprintf("%s      ABlock        = %x \n", str, ds.AdminBlock.GetHash().Bytes()[:5])
		str = fmt.Sprintf("%s      FBlock        = %x \n", str, ds.FactoidBlock.GetHash().Bytes()[:5])
		str = fmt.Sprintf("%s      ECBlock       = %x \n", str, ds.EntryCreditBlock.GetHash().Bytes()[:5])
	}
	return str
}

func (list *DBStateList) GetHighestLockedSignedAndSavesBlk() uint32 {
	ht := list.Base
	for i, dbstate := range list.DBStates {
		if dbstate != nil && dbstate.Locked && dbstate.Signed && dbstate.Saved {
			ht = list.Base + uint32(i)
		} else {
			if dbstate == nil {
				break
			}
		}
	}
	return ht
}

func (list *DBStateList) GetHighestCompletedBlk() uint32 {
	ht := list.Base
	for i, dbstate := range list.DBStates {
		if dbstate != nil && dbstate.Locked {
			ht = list.Base + uint32(i)
		} else {
			if dbstate == nil {
				break
			}
		}
	}
	return ht
}

func (list *DBStateList) GetHighestSignedBlk() uint32 {
	ht := list.Base
	for i, dbstate := range list.DBStates {
		if dbstate != nil && dbstate.Signed {
			ht = list.Base + uint32(i)
		} else {
			if dbstate == nil {
				return ht
			}
		}
	}
	return ht
}

func (list *DBStateList) GetHighestSavedBlk() uint32 {
	ht := list.Base
	for i, dbstate := range list.DBStates {
		if dbstate != nil && dbstate.Saved {
			ht = list.Base + uint32(i)
		} else {
			if dbstate == nil {
				break
			}
		}
	}
	return ht
}

// a contains b, returns true
func containsServer(haystack []interfaces.IServer, needle interfaces.IServer) bool {
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
	if !d.IsNew || p == nil {
		return
	}

	//	list.State.LogPrintf("dbstateprocess", "FixupLinks(%d,%d)", p.DirectoryBlock.GetHeader().GetDBHeight(), d.DirectoryBlock.GetHeader().GetDBHeight())
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
	//previousPL := list.State.ProcessLists.Get(cur)
	currentPL := list.State.ProcessLists.Get(currentDBHeight)

	// Servers
	startingFeds := currentPL.StartingFedServers
	currentFeds := currentPL.FedServers
	currentAuds := currentPL.AuditServers

	// Set the Start servers for the next block

	// DB Sigs
	majority := (len(currentFeds) / 2) + 1
	lenDBSigs := len(list.State.ProcessLists.Get(currentDBHeight).DBSignatures)
	if lenDBSigs < majority {
		//list.State.AddStatus(fmt.Sprintf("FIXUPLINKS: return without processing: lenDBSigs)(%v) < majority(%d)",
		//	lenDBSigs,
		//	majority))

		return false
	}
	//list.State.AddStatus(fmt.Sprintf("FIXUPLINKS: Adding the first %d dbsigs",
	//	majority))

	for _, sig := range list.State.ProcessLists.Get(currentDBHeight).DBSignatures {
		d.AdminBlock.AddDBSig(sig.ChainID, sig.Signature)
	}

	//list.State.AddStatus("FIXUPLINKS: Adding the deltas to the Admin Block, if necessary")

	// Correcting Server Lists (Caused by Server Faults)
	// 	This will correct any deltas from the serverlists
	for _, cf := range currentFeds {
		if !containsServer(startingFeds, cf) {
			fmt.Printf("******* FUL: %12s %12s  Server %x\n", "Promote", list.State.FactomNodeName, cf.GetChainID().Bytes()[3:6])
			// Promote to federated
			addEntry := adminBlock.NewAddFederatedServer(cf.GetChainID(), currentDBHeight+1)
			d.AdminBlock.AddFirstABEntry(addEntry)
		}
	}

	for _, pf := range startingFeds {
		if !containsServer(currentFeds, pf) {
			// The fed is n
			if containsServer(currentAuds, pf) {
				demoteEntry := adminBlock.NewAddAuditServer(pf.GetChainID(), currentDBHeight+1)
				d.AdminBlock.AddFirstABEntry(demoteEntry)
				fmt.Printf("******* FUL: %12s %12s  Server %x, DBHeight: %d\n", "Demote", list.State.FactomNodeName, pf.GetChainID().Bytes()[3:6], d.DirectoryBlock.GetDatabaseHeight())
			}
			_ = currentAuds
		}
	}

	// Additional Admin block changed can be made from identity changes
	list.State.SyncIdentities(d)

	// every 25 blocks +0 we add grant payouts
	// If this is a coinbase descriptor block, add that now
	if currentDBHeight > constants.COINBASE_ACTIVATION && currentDBHeight%constants.COINBASE_PAYOUT_FREQUENCY == 0 {
		// Build outputs
		auths := list.State.IdentityControl.GetSortedAuthorities()
		outputs := make([]interfaces.ITransAddress, 0)
		for _, a := range auths {
			ia := a.(*identity.Authority)
			if ia.CoinbaseAddress.IsZero() {
				continue
			}
			amt := primitives.CalculateCoinbasePayout(ia.Efficiency)
			if amt == 0 {
				continue
			}

			o := factoid.NewOutAddress(ia.CoinbaseAddress, amt)
			outputs = append(outputs, o)
		}

		err = d.AdminBlock.AddCoinbaseDescriptor(outputs)
		if err != nil {
			panic(err)
		}
	}

	// every 25 blocks +1 we add grant payouts
	if currentDBHeight > constants.COINBASE_ACTIVATION && currentDBHeight%constants.COINBASE_PAYOUT_FREQUENCY == 1 {
		// Add the grants to the list
		grantPayouts := GetGrantPayoutsFor(currentDBHeight)
		if len(grantPayouts) > 0 {
			err := d.AdminBlock.AddCoinbaseDescriptor(grantPayouts)
			if err != nil {
				panic(err)
			}
		}
	}

	err = d.AdminBlock.InsertIdentityABEntries()
	if err != nil {
		fmt.Println(err)
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

	// These two lines are crucial. They init/sort the dblock
	d.DirectoryBlock.BuildBodyMR()
	d.DirectoryBlock.MarshalBinary()

	progress = true
	d.IsNew = false
	list.State.ResetTryCnt = 0

	//fmt.Println("Fixup", d.DirectoryBlock.GetHeader().GetDBHeight())
	//authlistMsg := list.State.EFactory.NewAuthorityListInternal(currentFeds, currentAuds, currentDBHeight)
	//list.State.ElectionsQueue().Enqueue(authlistMsg)

	return
}

func (list *DBStateList) ProcessBlocks(d *DBState) (progress bool) {
	dbht := d.DirectoryBlock.GetHeader().GetDBHeight()

	s := list.State

	// If we are locked, the block has already been processed.  If the block IsNew then it has not yet had
	// its links patched, so we can't process it.  But if this is a repeat block (we have already processed
	// at this height) then we simply return.
	if d.Locked || d.IsNew || d.Repeat {
		s.LogPrintf("dbstateprocess", "ProcessBlocks(%d) Skipping d.Locked(%v) || d.IsNew(%v) || d.Repeat(%v) : ", dbht, d.Locked, d.IsNew, d.Repeat)
		return false
	}

	// If we detect that we have processed at this height, flag the dbstate as a repeat, progress is good, and
	// go forward.

	if dbht > 0 && dbht < list.ProcessHeight {
		progress = true
		d.Repeat = true
		s.LogPrintf("dbstateprocess", "ProcessBlocks(%d) Skipping old(repeat) state", dbht)
		return false
	}

	// If we detect that we have processed at this height, flag the dbstate as a repeat, progress is good, and
	// go forward. If dbHeight == list.ProcessHeight and current minute is 0, we want don't want to mark as a repeat,
	// so we can avoid the Election in Minute 9 bug.
	if dbht > 0 && dbht == list.ProcessHeight && list.State.CurrentMinute > 0 {
		progress = true
		d.Repeat = true
		s.LogPrintf("dbstateprocess", "ProcessBlocks(%d) Skipping Repeated current block", dbht, d.Locked, d.IsNew, d.Repeat)
		return false
	}

	if dbht > 1 {
		pd := list.State.DBStates.Get(int(dbht - 1))

		if pd == nil {
			s.LogPrintf("dbstateprocess", "ProcessBlocks(%d) Skipping Prev Block Missing", dbht)
			return false // Can't process out of order
		}
		if !pd.Saved {
			s.LogPrintf("dbstateprocess", "ProcessBlocks(%d) Skipping Prev Block not saved", dbht)
			return false // can't process till the prev is saved
		}
	}

	// Bring the current federated servers and audit servers forward to the
	// next block.

	if list.State.DebugConsensus {
		PrintState(list.State)
	}

	pl := list.State.ProcessLists.Get(dbht)
	pln := list.State.ProcessLists.Get(dbht + 1)

	if pl == nil {

		s.LogPrintf("dbstateprocess", "ProcessBlocks(%d) Skipping No ProcessList", dbht)
		return false
	}

	s.LogPrintf("dbstateprocess", "ProcessBlocks(%d)", dbht)

	//
	// ***** Apply the AdminBlock changes to the next DBState
	//
	//list.State.AddStatus(fmt.Sprintf("PROCESSBLOCKS:  Processing Admin Block at dbht: %d", d.AdminBlock.GetDBHeight()))
	err := d.AdminBlock.UpdateState(list.State)

	s.LogPrintf("dbstateprocess", "ProcessBlocks(%d) after update auth %d/%d ", dbht, len(pl.FedServers), len(pl.AuditServers))

	if err != nil {
		panic(err)
	}
	err = d.EntryCreditBlock.UpdateState(list.State)
	if err != nil {
		panic(err)
	}

	pln2 := list.State.ProcessLists.Get(dbht + 2)
	pln2.FedServers = append(pln2.FedServers[:0], pln.FedServers...)
	pln2.AuditServers = append(pln2.AuditServers[:0], pln.AuditServers...)

	pln2.SortAuditServers()
	pln2.SortFedServers()

	pl.SortAuditServers()
	pl.SortFedServers()
	pln.SortAuditServers()
	pln.SortFedServers()

	// Now the authority lists are set, set the starting
	pln.SetStartingAuthoritySet()
	pln2.SetStartingAuthoritySet()

	// *******************
	// Factoid Block Processing
	// *******************
	fs := list.State.GetFactoidState()

	s.LogPrintf("dbstateprocess", "ProcessBlocks(%d) Process Factoids dbht %d factoid",
		dbht, fs.(*FactoidState).DBHeight)

	// get all the prior balances of the Factoid addresses that may have changed
	// in this block.  If you want the balance of the highest saved block, look to
	// list.State.FactoidBalancesPapi if it is not null.  If you have no entry there,
	// then look to list.State.FactoidBalancesP

	if s.RestoreFCT != nil {
		for k, v := range s.RestoreFCT {
			s.FactoidBalancesP[k] = v
		}
	}
	if s.RestoreEC != nil {
		for k, v := range s.RestoreEC {
			s.ECBalancesP[k] = v
		}
	}

	list.State.FactoidBalancesPMutex.Lock()
	list.State.FactoidBalancesPapi = make(map[[32]byte]int64, len(pl.FactoidBalancesT))
	list.State.RestoreFCT = make(map[[32]byte]int64, len(pl.FactoidBalancesT))
	list.State.RestoreEC = make(map[[32]byte]int64, len(pl.FactoidBalancesT))
	for k := range pl.FactoidBalancesT {
		list.State.FactoidBalancesPapi[k] = list.State.FactoidBalancesP[k] // Capture the previous block balances for the APIs
		list.State.RestoreFCT[k] = list.State.FactoidBalancesP[k]          // Capture the previous block balances to restore if we have to apply a dbstate over this block.
	}
	for k := range pl.ECBalancesT {
		list.State.RestoreEC[k] = list.State.ECBalancesP[k]
	}
	list.State.FactoidBalancesPMutex.Unlock()

	// get all the prior balances of the entry credit addresses that may have changed
	// in this block.  If you want the balance of the highest saved block, look to
	// list.State.ECBalancesPapi if it is not null.  If you have no entry there,
	// then look to list.State.ECBalancesP
	list.State.ECBalancesPMutex.Lock()
	list.State.ECBalancesPapi = make(map[[32]byte]int64, len(pl.ECBalancesT))
	for k := range pl.ECBalancesT {
		list.State.ECBalancesPapi[k] = list.State.ECBalancesP[k]
	}
	list.State.ECBalancesPMutex.Unlock()

	// Process the Factoid End of Block
	err = fs.AddTransactionBlock(d.FactoidBlock)
	if err != nil {
		panic(err)
	}
	err = fs.AddECBlock(d.EntryCreditBlock)
	if err != nil {
		panic(err)
	}

	if list.State.DBFinished {
		fs.(*FactoidState).DBHeight = dbht
		list.State.Balancehash = fs.GetBalanceHash(false)
	}

	// Make the current exchange rate whatever we had in the previous block.
	// UNLESS there was a FER entry processed during this block  change height will be left at 1 on a change block
	if list.State.FERChangeHeight == 1 {
		list.State.FERChangeHeight = 0
	} else {
		if list.State.FactoshisPerEC != d.FactoidBlock.GetExchRate() {
			//list.State.AddStatus(fmt.Sprint("PROCESSBLOCKS:  setting rate", list.State.FactoshisPerEC,
			//	" to ", d.FactoidBlock.GetExchRate(),
			//	" - Height ", d.DirectoryBlock.GetHeader().GetDBHeight()))
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
	list.ProcessHeight = dbht
	progress = true
	d.Locked = true // Only after all is done will I admit this state has been saved.

	pln.SortFedServers()
	pln.SortAuditServers()

	// Sync Identities
	// 	Do the sync first, which will sync any Eblocks added from the prior block
	//	Then add eblocks from this current block, they will be synced come the next block.
	//	The order is important as when we are in this function, we only know n-1 is saved to disk
	list.State.SyncIdentities(nil)                                                   // Sync n-1 eblocks
	list.State.AddNewIdentityEblocks(d.EntryBlocks, d.DirectoryBlock.GetTimestamp()) // Add eblocks to be synced
	list.State.UpdateAuthSigningKeys(d.DirectoryBlock.GetDatabaseHeight())           // Remove old keys from key history

	// Canceling Coinbase Descriptors
	list.State.IdentityControl.CancelManager.GC(d.DirectoryBlock.GetDatabaseHeight()) // garbage collect

	///////////////////////////////
	// Cleanup Tasks
	///////////////////////////////
	list.State.Commits.Cleanup(list.State)
	// This usually gets cleaned up when creating the coinbase. If syncing from disk or dbstates, this routine will clean
	// up any leftover valid cancels.
	if d.DirectoryBlock.GetDatabaseHeight() > constants.COINBASE_DECLARATION {
		_, ok := list.State.IdentityControl.CanceledCoinbaseOutputs[d.DirectoryBlock.GetDatabaseHeight()-constants.COINBASE_DECLARATION]
		if ok {
			// No longer need this
			delete(list.State.IdentityControl.CanceledCoinbaseOutputs, d.DirectoryBlock.GetDatabaseHeight()-constants.COINBASE_DECLARATION)
		}
	}

	// s := list.State
	// // Time out commits every now and again.
	// now := s.GetTimestamp()
	// for k, msg := range s.Commits {
	// 	{
	// 		c, ok := msg.(*messages.CommitChainMsg)
	// 		if ok && !s.NoEntryYet(c.CommitChain.EntryHash, now) {
	// 			delete(s.Commits, k)
	// 			continue
	// 		}
	// 	}
	// 	c, ok := msg.(*messages.CommitEntryMsg)
	// 	if ok && !s.NoEntryYet(c.CommitEntry.EntryHash, now) {
	// 		delete(s.Commits, k)
	// 		continue
	// 	}

	// 	_, ok = s.Replay.Valid(constants.TIME_TEST, msg.GetRepeatHash().Fixed(), msg.GetTimestamp(), now)
	// 	if !ok {
	// 		delete(s.Commits, k)
	// 	}
	// }

	// Writing the DBState to a debug file allows for later analyzing the last block not saved to the database.
	// Do not do this while loading from disk, as those blocks are already saved
	if list.State.DBFinished && globals.Params.WriteProcessedDBStates {
		list.WriteDBStateToDebugFile(d)
	}

	fs.(*FactoidState).DBHeight = list.State.GetDirectoryBlock().GetHeader().GetDBHeight()

	tbh := list.State.FactoidState.GetBalanceHash(true) // recompute temp balance hash here
	list.State.Balancehash = fs.GetBalanceHash(false)
	list.State.LogPrintf("dbstateprocess", "ProcessBlock(%d) BalanceHash P %x T %x", dbht, list.State.Balancehash.Bytes()[0:4], tbh.Bytes()[0:4])

	// We will only save blocks marked to be saved.  As such, this must follow
	// the "d.saved = true" above
	if list.State.StateSaverStruct.FastBoot && d.DirectoryBlock.GetHeader().GetDBHeight() != 0 {
		d.SaveStruct = SaveFactomdState(list.State, d)

		err := list.State.StateSaverStruct.SaveDBStateList(list.State, list.State.DBStates, list.State.Network)
		if err != nil {
			list.State.LogPrintf("dbstateprocess", "Error while saving Fastboot %v", err)
		}
	}

	// All done with this block move to the next height if we are loading by blocks
	if s.LLeaderHeight == dbht {
		// if we are following by blocks then this move us forward but if we are following by minutes the
		// code in ProcessEOM for minute 10 will have moved us forward
		s.SetLeaderTimestamp(d.DirectoryBlock.GetTimestamp())
		// todo: is there a reason not to do this in MoveStateToHeight?
		fs.(*FactoidState).DBHeight = dbht + 1
		s.MoveStateToHeight(dbht+1, 0)
	}

	// Note about dbsigs.... If we processed the previous minute, then we generate the DBSig for the next block.
	// But if we didn't process the previous block, like we start from scratch, or we had to reset the entire
	// network, then no dbsig exists.  This code doesn't execute, and so we have no dbsig.  In that case, on
	// the next EOM, we see the block hasn't been signed, and we sign the block (That is the call to SendDBSig()
	// above).
	pldbs := s.ProcessLists.Get(s.LLeaderHeight)
	if s.Leader && !pldbs.DBSigAlreadySent {
		s.SendDBSig(s.LLeaderHeight, s.LeaderVMIndex) // ProcessBlocks()
	}

	return
}

// We don't really do the signing here, but just check that we have all the signatures.
// If we do, we count that as progress.
func (list *DBStateList) SignDB(d *DBState) (process bool) {
	dbheight := d.DirectoryBlock.GetHeader().GetDBHeight()
	list.State.LogPrintf("dbstateprocess", "SignDB(%d)", dbheight)
	if d.Signed {
		//s := list.State
		//		s.MoveStateToHeight(dbheight + 1)
		list.State.LogPrintf("dbstateprocess", "SignDB(%d) done, already signed", dbheight)
		return false
	}

	// If we have the next dbstate in the list, then all the signatures for this dbstate
	// have been checked, so we can consider this guy signed.
	if dbheight == 0 || list.Get(int(dbheight+1)) != nil || d.Repeat == true {
		d.Signed = true
		//		s := list.State
		//		s.MoveStateToHeight(dbheight+1, 0)
		list.State.LogPrintf("dbstateprocess", "SignDB(%d) done, next block exists", dbheight)

		return true
	}

	pl := list.State.ProcessLists.Get(dbheight)
	if pl == nil {
		list.State.LogPrintf("dbstateprocess", "SignDB(%d) skip, no processlist!", d.DirectoryBlock.GetHeader().GetDBHeight())
		return false
	} else if !pl.Complete() {
		list.State.LogPrintf("dbstateprocess", "SignDB(%d) skip, processlist not complete!", d.DirectoryBlock.GetHeader().GetDBHeight())
		return false
	}

	// If we don't have the next dbstate yet, see if we have all the signatures.
	pl = list.State.ProcessLists.Get(dbheight + 1)
	if pl == nil {
		list.State.LogPrintf("dbstateprocess", "SignDB(%d) skip, missing next processlist!", d.DirectoryBlock.GetHeader().GetDBHeight())
		return false
	}

	//// Don't sign while negotiating the EOM 0
	////todo: Can this be !list.State.DBSigDone?
	//if list.State.EOM {
	//	list.State.LogPrintf("dbstateprocess", "SignDB(%d) negotiating the EOM!", d.DirectoryBlock.GetHeader().GetDBHeight())
	//	return false
	//}

	// Don't sign unless we are in minute 0
	if list.State.CurrentMinute != 0 {
		list.State.LogPrintf("dbstateprocess", "SignDB(%d) Waiting for minute 1!", dbheight)
		return false
	}

	s := list.State
	d.Signed = true
	// once we start following by minutes ProcessLists.UpdateState will have already advanced the time
	if s.GetLLeaderHeight() != dbheight+1 {
		s.MoveStateToHeight(dbheight+1, 0)
	}

	list.State.LogPrintf("dbstateprocess", "SignDB(%d) Signed! have sigs", dbheight)
	return
}

// WriteDBStateToDebugFile will write the marshaled dbstate to a file alongside the database.
// This can be written on the processblocks, so in the event the block does not get written to disk
// in the event of a stall, the dbstate can be analyzed. The written dbstate does NOT include entries.
func (list *DBStateList) WriteDBStateToDebugFile(d *DBState) {
	// Because DBStates include the Savestate, we cannot marshal it, as the savestate is not set
	// So change it to a full block

	f := NewWholeBlock()
	f.DBlock = d.DirectoryBlock
	f.ABlock = d.AdminBlock
	f.FBlock = d.FactoidBlock
	f.ECBlock = d.EntryCreditBlock
	f.EBlocks = d.EntryBlocks

	data, err := f.MarshalBinary()
	if err != nil {
		fmt.Printf("An error has occurred while writing the DBState to disk: %s\n", err.Error())
		return
	}

	filename := fmt.Sprintf("processed_dbstate_%d.block", d.DirectoryBlock.GetDatabaseHeight()%10)
	path := filepath.Join(list.State.LdbPath, list.State.Network, "dbstates", filename)

	//fmt.Printf("Saving DBH %d to %s\n", list.State.LLeaderHeight, path)

	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0775)
	if err != nil {
		fmt.Printf("An error has occurred while writing the DBState to disk: %s\n", err.Error())
		return
	}

	file.Write(data)
	file.Close()
}

func ReadDBStateFromDebugFile(filename string) (*WholeBlock, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0775)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	wb := NewWholeBlock()
	err = wb.UnmarshalBinary(data)
	return wb, err
}

func (list *DBStateList) SaveDBStateToDB(d *DBState) (progress bool) {
	dbheight := int(d.DirectoryBlock.GetHeader().GetDBHeight())
	// Take the height, and some function of the identity chain, and use that to decide to trim.  That
	// way, not all nodes in a simulation Trim() at the same time.

	if !d.Signed || !d.ReadyToSave || list.State.DB == nil {
		return
	}
	list.State.LogPrintf("dbstateprocess", "SaveDBStateToDB(%d) Balance hash %x", d.DirectoryBlock.GetHeader().GetDBHeight(), list.State.Balancehash.Bytes()[:4])

	// If this is a repeated block, and I have already saved at this height, then we can safely ignore
	// this dbstate.
	if d.Repeat == true && uint32(dbheight) <= list.SavedHeight {
		progress = true
		d.ReadyToSave = false
		d.Saved = true
	}

	if dbheight > 0 {
		dp := list.State.GetDBState(uint32(dbheight - 1))
		if dp == nil {
			list.State.LogPrintf("dbstateprocess", "SaveDBStateToDB(%d) no previous block!", d.DirectoryBlock.GetHeader().GetDBHeight())
			return
		} else if !dp.Saved {
			list.State.LogPrintf("dbstateprocess", "SaveDBStateToDB(%d) previous not saved!", d.DirectoryBlock.GetHeader().GetDBHeight())
			return
		}
	}

	list.State.RestoreFCT = nil
	list.State.RestoreEC = nil

	if d.Saved {
		Havedblk, err := list.State.DB.DoesKeyExist(databaseOverlay.DIRECTORYBLOCK, d.DirectoryBlock.GetKeyMR().Bytes())
		if err != nil || !Havedblk {
			panic(fmt.Sprintf("Claimed to be found on %s DBHeight %d Hash %x",
				list.State.FactomNodeName,
				d.DirectoryBlock.GetHeader().GetDBHeight(),
				d.DirectoryBlock.GetKeyMR().Bytes()))
		}

		//		list.State.LogPrintf("dbstateprocess", "SaveDBStateToDB(%d) Already saved, add to replay!", d.DirectoryBlock.GetHeader().GetDBHeight())
		// Set the Block Replay flag for all these transactions that are already in the database.
		for _, fct := range d.FactoidBlock.GetTransactions() {
			list.State.FReplay.IsTSValidAndUpdateState(
				constants.BLOCK_REPLAY,
				fct.GetSigHash().Fixed(),
				fct.GetTimestamp(),
				d.DirectoryBlock.GetHeader().GetTimestamp())
		}
		list.State.Saving = false
		return
	}

	// Past this point, we cannot Return without recording the transactions in the dbstate.  This is because we
	// have marked them all as saved to disk!  So we gotta save them to disk.  Or panic trying.

	//	list.State.LogPrintf("dbstateprocess", "SaveDBStateToDB(%d) %s\n", dbheight, d.String())
	// Only trim when we are really saving.
	v := dbheight + int(list.State.IdentityChainID.Bytes()[4])
	if v%4 == 0 {
		list.State.DB.Trim()
	}

	// Save
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

	for _, en := range d.EntryCreditBlock.GetEntries() {
		switch en.ECID() {
		case constants.ECIDChainCommit:
			list.State.NumNewChains++
		case constants.ECIDEntryCommit:
			list.State.NumNewEntries++
		}
	}

	pl := list.State.ProcessLists.Get(uint32(dbheight))

	allowedEBlocks := make(map[[32]byte]struct{})
	allowedEntries := make(map[[32]byte]struct{})

	// Eblocks from DBlock
	for _, eb := range d.DirectoryBlock.GetEBlockDBEntries() {
		allowedEBlocks[eb.GetKeyMR().Fixed()] = struct{}{}
	}

	// Go through eblocks to build allowed entry map
	for _, eb := range d.EntryBlocks {
		keymr, err := eb.KeyMR()
		if err != nil {
			panic(err)
		}
		// If its a good eblock, add it's entries to the allowed
		if _, ok := allowedEBlocks[keymr.Fixed()]; ok {
			for _, e := range eb.GetEntryHashes() {
				allowedEntries[e.Fixed()] = struct{}{}
			}
		} else {
			list.State.LogPrintf("dbstateprocess", "Error putting entries in allowedmap, as Eblock is not in Dblock")
		}
	}

	// Info from DBState
	for _, eb := range d.EntryBlocks {
		keymr, err := eb.KeyMR()
		if err != nil {
			panic(err)
		}
		// If it's in the DBlock
		if _, ok := allowedEBlocks[keymr.Fixed()]; ok {
			if err := list.State.DB.ProcessEBlockMultiBatch(eb, true); err != nil {
				panic(err.Error())
			}
		} else {
			list.State.LogPrintf("dbstateprocess", "Error saving eblock from dbstate, eblock not allowed")
		}
	}
	for _, e := range d.Entries {
		// If it's in the DBlock
		list.State.WriteEntry <- e
	}
	list.State.NumEntries += len(d.Entries)
	list.State.NumEntryBlocks += len(d.EntryBlocks)

	// Info from ProcessList
	if pl != nil {
		for _, eb := range pl.NewEBlocks {
			keymr, err := eb.KeyMR()
			if err != nil {
				panic(err)
			}
			if _, ok := allowedEBlocks[keymr.Fixed()]; ok {
				if err := list.State.DB.ProcessEBlockMultiBatch(eb, true); err != nil {
					panic(err.Error())
				}

			} else {
				list.State.LogPrintf("dbstateprocess", "Error saving eblock from process list, eblock not allowed")
			}
		}
	}

	if err := list.State.DB.ProcessDBlockMultiBatch(d.DirectoryBlock); err != nil {
		panic(err.Error())
	}

	if err := list.State.DB.ExecuteMultiBatch(); err != nil {
		panic(err.Error())
	}

	// Info from ProcessList
	if pl != nil {
		for _, eb := range pl.NewEBlocks {
			keymr, err := eb.KeyMR()
			if err != nil {
				panic(err)
			}
			if _, ok := allowedEBlocks[keymr.Fixed()]; ok {
				for _, e := range eb.GetBody().GetEBEntries() {
					pl.State.WriteEntry <- pl.GetNewEntry(e.Fixed())
				}
			} else {
				list.State.LogPrintf("dbstateprocess", "Error saving eblock from process list, eblock not allowed")
			}
		}
		pl.NewEBlocks = make(map[[32]byte]interfaces.IEntryBlock)
		pl.NewEntries = make(map[[32]byte]interfaces.IEntry)
	}

	d.EntryBlocks = make([]interfaces.IEntryBlock, 0)
	d.Entries = make([]interfaces.IEBEntry, 0)

	// Not activated.  Set to true if you want extra checking of the data saved to the database.
	if false {
		good := true
		mr, err := list.State.DB.FetchDBKeyMRByHeight(uint32(dbheight))
		if err != nil {
			list.State.LogPrintf("dbstateprocess", err.Error())
			panic(fmt.Sprintf("%20s At Directory Block Height %d", list.State.FactomNodeName, dbheight))
			return
		}
		if mr == nil {
			list.State.LogPrintf("dbstateprocess", "There is no mr returned by list.State.DB.FetchDBKeyMRByHeight() at %d\n", dbheight)
			mr = d.DirectoryBlock.GetKeyMR()
			good = false
		}

		td, err := list.State.DB.FetchDBlock(mr)
		if err != nil || td == nil {
			if err != nil {
				list.State.LogPrintf("dbstateprocess", err.Error())
			} else {
				list.State.LogPrintf("dbstateprocess", "Could not get directory block by primary key at Block Height %d\n", dbheight)
			}
			panic(fmt.Sprintf("%20s Error reading db by mr at Directory Block Height %d", list.State.FactomNodeName, dbheight))
			return
		}
		if td.GetKeyMR().Fixed() != mr.Fixed() {
			list.State.LogPrintf("dbstateprocess", "Key MR is wrong at Directory Block Height %d\n", dbheight)
			fmt.Fprintln(os.Stderr, d.DirectoryBlock.String(), "\n==============================================\n Should be:\n", td.String())
			panic(fmt.Sprintf("%20s KeyMR is wrong at Directory Block Height %d", list.State.FactomNodeName, dbheight))
			return
		}
		if !good {
			return
		}
	}

	// Set the Block Replay flag for all these transactions we are saving to the database.
	for _, fct := range d.FactoidBlock.GetTransactions() {
		list.State.FReplay.IsTSValidAndUpdateState(
			constants.BLOCK_REPLAY,
			fct.GetSigHash().Fixed(),
			fct.GetTimestamp(),
			d.DirectoryBlock.GetHeader().GetTimestamp())
	}

	list.State.NumFCTTrans += len(d.FactoidBlock.GetTransactions()) - 1

	list.SavedHeight = uint32(dbheight)
	list.State.Saving = false
	progress = true
	d.ReadyToSave = false
	d.Saved = true

	// Now that we have saved the perm balances, we can clear the api hashmaps that held the differences
	// between the actual saved block prior, and this saved block.  If you are looking for balances of
	// the highest saved block, you first look to see that one of the "<fct or ec>Papi" maps exist, then
	// if that map has a value for your address.  If it doesn't exist, or doesn't have a value, then look
	// in the "<fct or ec>P" map.
	list.State.FactoidBalancesPMutex.Lock()
	list.State.FactoidBalancesPapi = nil
	list.State.FactoidBalancesPMutex.Unlock()

	list.State.ECBalancesPMutex.Lock()
	list.State.ECBalancesPapi = nil
	list.State.ECBalancesPMutex.Unlock()

	return
}

func (list *DBStateList) UpdateState() (progress bool) {

	s := list.State
	_ = s
	if len(list.DBStates) != 0 {
		l := "["
		for _, d := range list.DBStates {
			if d == nil {
				l += "nil "
			} else {
				status := []byte("______")
				if d.Locked {
					status[0] = 'L'
				}
				if d.ReadyToSave {
					status[1] = 'R'
				}
				if d.Signed {
					status[2] = 'S'
				}
				if d.Saved {
					status[3] = 'V'
				}
				if d.Repeat {
					status[4] = 'D'
				}
				if d.SaveStruct != nil && d.SaveStruct.IdentityControl != nil {
					status[5] = '!'
				}
				l += fmt.Sprintf("%d%s, ", d.DirectoryBlock.GetHeader().GetDBHeight(), string(status))
			}
		}
		l += "]"
		s.LogPrintf("dbstateprocess", "updateState() %d %s", list.Base, l)
	}
	saved := 0
	for i, d := range list.DBStates {
		// loop only thru this and future blocks
		//for i := int(list.State.LLeaderHeight); i < int(list.Base)+len(list.DBStates); i++ {
		//	d := list.Get(i)
		//	//fmt.Printf("dddd %20s %10s --- %10s %10v %10s %10v \n", "DBStateList Update", list.State.FactomNodeName, "Looking at", i, "DBHeight", list.Base+uint32(i))

		// Must process blocks in sequence.  Missing a block says we must stop.
		if d == nil {
			return
		}
		if d.Locked && d.Signed && d.Saved {
			continue
		}
		//todo: Make the for start here and move forward
		dbHeight := d.DirectoryBlock.GetHeader().GetDBHeight()
		highestLockedSignedAndSavedBlk := list.State.GetHighestLockedSignedAndSavesBlk()
		if dbHeight != 0 && dbHeight <= highestLockedSignedAndSavedBlk {
			//			s.LogPrintf("dbstateprocess", "skip reprocessing %d", dbHeight)
			continue // don't reprocess old blocks
		}
		var p bool = false
		// if this is not the first block then fixup the links
		if i > 0 {
			p = list.FixupLinks(list.DBStates[i-1], d)
			progress = p || progress
		}

		p = list.ProcessBlocks(d)
		progress = p || progress

		p = list.SignDB(d)
		progress = p || progress

		p = list.SaveDBStateToDB(d)
		progress = p || progress

		// Make sure we move forward the Adminblock state in the process lists
		list.State.ProcessLists.Get(dbHeight + 1)

		// remember the last saved block
		if d.Saved {
			saved = i
		}

		// only process one block past the last saved block
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

	list.State.LogPrintf("dbstateprocess", "DBStateList put dbstate dbht %d locked %v signed %v saved %v",
		dbheight, dbState.Locked, dbState.Signed, dbState.Saved)

	// Count completed, done, don't have to do anything more to states,
	// starting from the beginning (since base starts at zero).
	cnt := 0
searchLoop:
	for _, v := range list.DBStates {
		if dbheight > 0 && (v == nil || v.DirectoryBlock == nil || !(v.Saved && v.Locked && v.Signed)) {
			break searchLoop
		}
		cnt++
	}

	keep := uint32(3) // How many states to keep around; debugging helps with more.
	if uint32(cnt) > keep && int(list.Complete)-cnt+int(keep) > 0 {
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
	list.DBStates[index] = dbState
	if dbState.SaveStruct.IdentityControl == nil {
		atomic.WhereAmIMsg("no identity control")
	} else {

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
	dbState.Init() // Creat all the sub structor...

	dbState.DBHash = directoryBlock.DatabasePrimaryIndex()
	dbState.ABHash = adminBlock.DatabasePrimaryIndex()
	dbState.FBHash = factoidBlock.DatabasePrimaryIndex()
	dbState.ECHash = entryCreditBlock.DatabasePrimaryIndex()

	dbState.IsNew = isNew
	dbState.DirectoryBlock = directoryBlock
	dbState.AdminBlock = adminBlock
	dbState.FactoidBlock = factoidBlock
	dbState.EntryCreditBlock = entryCreditBlock
	dbState.EntryBlocks = eBlocks
	dbState.Entries = entries

	dbState.Added = list.State.GetTimestamp()

	// If we actually add this to the list, return the dbstate.
	if list.Put(dbState) {
		if dbState.SaveStruct.IdentityControl == nil {
			atomic.WhereAmIMsg("no identity control")
		}

		return dbState
	} else {
		ht := dbState.DirectoryBlock.GetHeader().GetDBHeight()
		if ht == list.State.GetHighestCompletedBlk() {
			index := int(ht) - int(list.State.DBStates.Base)
			if index > 0 {
				list.State.DBStates.DBStates[index] = dbState
				if dbState.SaveStruct.IdentityControl == nil {
					atomic.WhereAmIMsg("no identity control")
				}
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
