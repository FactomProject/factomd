// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"bytes"
	"errors"

	"github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/identity"
	. "github.com/FactomProject/factomd/common/identityEntries"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"

	"sort"

	"fmt"

	log "github.com/sirupsen/logrus"
)

var _ = DecodeIdentityChainStructureFromExtIDs

// identLogger is the general logger for all identity related logs. You can add additional fields,
// or create more context loggers off of this
var identLogger = packageLogger.WithFields(log.Fields{"subpack": "identity"})

var (
	// Time window for identity to require registration: 24hours = 144 blocks
	TIME_WINDOW uint32 = 144

	// Where all Identities register
	MAIN_FACTOM_IDENTITY_LIST = "888888001750ede0eff4b05f0c3f557890b256450cabbb84cada937f9c258327"
)

// GetSigningKey will return the signing key of the identity, and it's type
//		Returns:
//			-1	--> Follower
//			0 	--> Audit Server
//			1	--> Federated
func (st *State) GetSigningKey(id interfaces.IHash) (interfaces.IHash, int) {
	getReturnStatInt := func(stat uint8) int {
		if stat == constants.IDENTITY_PENDING_FEDERATED_SERVER || stat == constants.IDENTITY_FEDERATED_SERVER {
			return 1
		}
		if stat == constants.IDENTITY_AUDIT_SERVER || stat == constants.IDENTITY_PENDING_AUDIT_SERVER {
			return 0
		}
		return -1
	}

	auth := st.IdentityControl.GetAuthority(id)
	if auth != nil {
		key := auth.SigningKey.Fixed()
		hash, _ := primitives.NewShaHash(key[:])
		if !(hash == nil || hash.IsZero()) {
			return hash, getReturnStatInt(auth.Status)
		}
	}

	identity := st.IdentityControl.GetIdentity(id)
	if identity != nil {
		return identity.SigningKey, getReturnStatInt(identity.Status)
	}
	return nil, -1
}

func (st *State) GetNetworkSkeletonKey() interfaces.IHash {
	id := st.IdentityControl.GetIdentity(st.GetNetworkSkeletonIdentity())
	if id == nil {
		// There should always be a skeleton identity. It cannot be removed
		// If there is one, just use the bootstrap key
		return st.GetNetworkBootStrapKey()
	}

	key := id.SigningKey
	// If the key is all 0s, use the bootstrap key
	if key.IsSameAs(primitives.NewZeroHash()) || key == nil {
		return st.GetNetworkBootStrapKey()
	} else {
		return key
	}
}

// Add the skeleton identity and try to build it
func (st *State) IntiateNetworkSkeletonIdentity() error {
	skel := st.GetNetworkSkeletonIdentity()
	st.IdentityControl.SetSkeletonIdentity(skel)

	// This populates the identity with keys found
	err := st.AddIdentityFromChainID(skel)
	if err != nil {
		return err
	}

	return nil
}

func (st *State) InitiateNetworkIdentityRegistration() error {
	reg := st.GetNetworkIdentityRegistrationChain()
	st.IdentityControl.SetIdentityRegistration(reg)

	// This populates the identity with keys found
	err := st.AddIdentityFromChainID(reg)
	if err != nil {
		return err
	}

	return nil
}

type IdentityEntry struct {
	Entry       interfaces.IEBEntry
	Timestamp   interfaces.Timestamp
	Blockheight uint32
}

// LoadIdentityByEntry is only useful when initial is set to false. If initial is false, it will track changes
// in an identity that corresponds to an authority. If initial is true, then calling ProcessIdentityEntry directly will
// have the same result.
func (st *State) LoadIdentityByEntry(ent interfaces.IEBEntry, height uint32, dblockTimestamp interfaces.Timestamp, d *DBState) {
	flog := identLogger.WithFields(st.Logger.Data).WithField("func", "LoadIdentityByEntry")
	if ent == nil {
		return
	}

	affectAblock := d != nil && d.DirectoryBlock.GetDatabaseHeight() == height+1
	if !affectAblock {
		d = nil
	}

	// Not an entry identities care about
	if bytes.Compare(ent.GetChainID().Bytes()[:3], []byte{0x88, 0x88, 0x88}) != 0 {
		return
	}

	var a interfaces.IAdminBlock
	if d != nil {
		a = d.AdminBlock
	} else {
		a = nil
	}

	_, err := st.IdentityControl.ProcessIdentityEntryWithABlockUpdate(ent, height, dblockTimestamp, a, true)
	if err != nil {
		flog.Errorf(err.Error())
	}
}

// Called by AddServer Message
func ProcessIdentityToAdminBlock(st *State, chainID interfaces.IHash, servertype int) bool {
	flog := identLogger.WithFields(st.Logger.Data).WithField("func", "ProcessIdentityToAdminBlock")

	err := st.AddIdentityFromChainID(chainID)
	if err != nil {
		flog.Errorf("Failed to process AddServerMessage for %s : %s", chainID.String()[:10], err.Error())
		return true
	}

	id := st.IdentityControl.GetIdentity(chainID)

	if id != nil {
		if ok, err := id.IsPromteable(); !ok {
			flog.Errorf("Failed to process AddServerMessage for %s : %s", chainID.String()[:10], err.Error())
			return true
		}

	} else {
		flog.Errorf("Failed to process AddServerMessage: %s", "New Fed/Audit server ["+chainID.String()[:10]+"] does not have an identity associated to it")
		return true
	}

	// Add to admin block
	if servertype == 0 {
		id.Status = constants.IDENTITY_PENDING_FEDERATED_SERVER
		st.LeaderPL.AdminBlock.AddFedServer(chainID)
	} else if servertype == 1 {
		id.Status = constants.IDENTITY_PENDING_AUDIT_SERVER
		st.LeaderPL.AdminBlock.AddAuditServer(chainID)
	}

	st.IdentityControl.SetIdentity(chainID, id)
	st.LeaderPL.AdminBlock.AddFederatedServerSigningKey(chainID, id.SigningKey.Fixed())
	st.LeaderPL.AdminBlock.AddMatryoshkaHash(chainID, id.MatryoshkaHash)
	for _, a := range id.AnchorKeys {
		st.LeaderPL.AdminBlock.AddFederatedServerBitcoinAnchorKey(chainID, a.KeyLevel, a.KeyType, a.SigningKey)
	}
	if !id.CoinbaseAddress.IsZero() {
		st.LeaderPL.AdminBlock.AddCoinbaseAddress(chainID, id.CoinbaseAddress)
	}
	st.LeaderPL.AdminBlock.AddEfficiency(chainID, id.Efficiency)

	return true
}

// Verifies if is authority
//		Return true if authority, false if not
func (st *State) VerifyIsAuthority(cid interfaces.IHash) bool {
	return st.IdentityControl.GetAuthority(cid) != nil
}

// AddIdentityFromChainID will add an identity to our list to watch and sync it.
func (st *State) AddIdentityFromChainID(cid interfaces.IHash) error {
	id := st.IdentityControl.GetIdentity(cid)
	if id == nil {
		id = NewIdentity()
		id.IdentityChainID = cid
		st.IdentityControl.SetIdentity(cid, id)
	}

	err := st.AddIdentityEblocks(cid, true)
	if err != nil {
		return err
	}

	// This is the initial call, so the current height should not trigger a keychange in the admin block
	st.SyncIdentities(nil)
	return nil
}

// SyncIdentities will run through the identities and sync any identities that need to be updated.
// If the current height is equal to the eblock+1, then this entry can trigger a change in key to the admin block
func (st *State) SyncIdentities(d *DBState) {
	// This will search an eblock, and sync it's entries if found.
	getAndSyncEntries := func(ieb EntryBlockMarker) error {
		eblock, err := st.DB.FetchEBlock(ieb.KeyMr)
		if err != nil {
			return fmt.Errorf("eblock missing")
		}

		entries := make([]IdentityEntry, len(eblock.GetEntryHashes()))
		for i, e := range eblock.GetEntryHashes() {
			if e.IsMinuteMarker() {
				continue
			}
			entry, err := st.DB.FetchEntry(e)
			if err != nil || entry == nil {
				return fmt.Errorf("eblock missing entries")
			}

			entries[i] = IdentityEntry{entry, ieb.DblockTimestamp, ieb.DBHeight}
		}

		// We have all the entries, so now we can process each entry and advance our eblock syncing
		for _, ie := range entries {
			st.LoadIdentityByEntry(ie.Entry, ieb.DBHeight, ieb.DblockTimestamp, d)
		}
		st.IdentityControl.ProcessOldEntries()
		return nil
	}

SyncIdentitiesLoop:
	for _, id := range st.IdentityControl.Identities {
		change := false
		hasManage := !(id.ManagementChainID == nil || id.ManagementChainID.IsZero())
		for !id.IdentityChainSync.Synced() {
			// If we find a management chain id, we need to add all of it's Eblocks to it's list to be synced before we
			// synced it's management chain
			eb := id.IdentityChainSync.NextEBlock()
			if eb == nil {
				panic(fmt.Sprintf("NextEblock was nil, but the identity chain was not fully synced. ID: %s", id.IdentityChainID.String()))
				continue SyncIdentitiesLoop
			}

			// Fetch the eblock and process the entries
			err := getAndSyncEntries(*eb)
			if err != nil {
				break
			}

			// No error means it was synced, so update our eblock sync list
			id.IdentityChainSync.BlockParsed(*eb)
			change = true
		}
		if !hasManage {
			// Found a manage chain
			if id.ManagementChainID != nil && !id.ManagementChainID.IsZero() {
				err := st.AddIdentityEblocks(id.ManagementChainID, false)
				if err != nil {
					panic(fmt.Sprintf("Could not add managment eblocks: %s", err.Error()))
				}
			}
		}

		if change {
			st.IdentityControl.SetIdentity(id.IdentityChainID, id)
		}

		change = false
		for !id.ManagementChainSync.Synced() {
			// If we find a management chain id, we need to add all of it's Eblocks to it's list to be synced before we
			// synced it's management chain
			eb := id.ManagementChainSync.NextEBlock()
			if eb == nil {
				break // Manage chain might not have any eblocks to be processed
			}

			// If the root chain is not synced, we cannot sync past it
			if !id.IdentityChainSync.Synced() && eb.DBHeight > id.IdentityChainSync.Current.DBHeight {
				break
			}

			// Fetch the eblock and process the entries
			err := getAndSyncEntries(*eb)
			if err != nil {
				break
			}

			// No error means it was synced, so update our eblock sync list
			id.ManagementChainSync.BlockParsed(*eb)
			change = true
		}

		if change {
			st.IdentityControl.SetIdentity(id.IdentityChainID, id)
		}
	}
}

// AddIdentityEblocks will find all eblocks for a root/management chain and add them to the sync list
func (st *State) AddIdentityEblocks(cid interfaces.IHash, rootChain bool) error {
	id := st.IdentityControl.GetIdentity(cid)
	if id == nil {
		return fmt.Errorf("[%s] identity not found", cid.String()[:10])
	}

	// A management chain was found. We need to add eblocks to it's synclist
	eblocks, err := st.DB.FetchAllEBlocksByChain(cid)
	if err != nil {
		return fmt.Errorf("This is a problem. Eblocks were not able to be fetched for %s", cid.String()[:10])
	}
	markers := make([]EntryBlockMarker, len(eblocks))
	for i, eb := range eblocks {
		keymr, err := eb.KeyMR()
		if err != nil {
			return fmt.Errorf("Keymr of eblock was unable to be computed")
		}
		dblock, err := st.DB.FetchDBlockByHeight(eb.GetDatabaseHeight())
		if err != nil {
			return fmt.Errorf("DBlock at %d not found on disk", eb.GetDatabaseHeight())
		}
		markers[i] = EntryBlockMarker{keymr, eb.GetHeader().GetEBSequence(), eb.GetDatabaseHeight(), dblock.GetTimestamp()}
	}

	sort.Sort(EntryBlockMarkerList(markers))
	for _, m := range markers {
		if rootChain {
			id.IdentityChainSync.AddNewHeadMarker(m)
		} else {
			id.ManagementChainSync.AddNewHeadMarker(m)
		}
	}

	return nil
}

// AddNewIdentityEblocks will scan the new eblock list and identify any eblocks of interest.
// If an eblock belongs to a current identity, we add it to the eblock lists of the identity
// to be synced.
func (st *State) AddNewIdentityEblocks(eblocks []interfaces.IEntryBlock, dblockTimestamp interfaces.Timestamp) {
	for _, eb := range eblocks {
		if bytes.Compare(eb.GetChainID().Bytes()[:3], []byte{0x88, 0x88, 0x88}) == 0 {
			// Can belong to an identity chain id
			id := st.IdentityControl.GetIdentity(eb.GetChainID())
			if id != nil {
				keymr, err := eb.KeyMR()
				if err != nil {
					continue // This shouldn't ever happen
				}
				// It belongs to an identity that we are watching
				if id.IdentityChainID.IsSameAs(eb.GetChainID()) {
					// Identity chain eblock
					id.IdentityChainSync.AddNewHead(keymr, eb.GetHeader().GetEBSequence(), eb.GetDatabaseHeight(), dblockTimestamp)
				} else if id.ManagementChainID.IsSameAs(eb.GetChainID()) {
					// Manage chain eblock
					id.ManagementChainSync.AddNewHead(keymr, eb.GetHeader().GetEBSequence(), eb.GetDatabaseHeight(), dblockTimestamp)
				}
				st.IdentityControl.SetIdentity(id.IdentityChainID, id)
			}
		}
	}
}

/*************************
	Old methods used to
	lookup an identity
	in blockchain
 *************************/

// FetchIdentityChainEntriesInCreateOrder will grab all entries in a chain for an identity in the order they were created.
func (s *State) FetchIdentityChainEntriesInCreateOrder(chainid interfaces.IHash) ([]IdentityEntry, error) {
	head, err := s.DB.FetchHeadIndexByChainID(chainid)
	if err != nil {
		return nil, err
	}

	if head == nil {
		return nil, fmt.Errorf("chain %x does not exist", chainid.Fixed())
	}

	// Get Eblocks
	var blocks []interfaces.IEntryBlock
	next := head
	for {
		if next.IsZero() {
			break
		}

		// Get the EBlock, and add to list to parse
		block, err := s.DB.FetchEBlock(next)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)

		next = block.GetHeader().GetPrevKeyMR()
	}

	var entries []IdentityEntry
	// Walk through eblocks in reverse order to get entries
	for i := len(blocks) - 1; i >= 0; i-- {
		eb := blocks[i]

		height := eb.GetDatabaseHeight()
		// Get the timestamp
		dblock, err := s.DB.FetchDBlockByHeight(height)
		if err != nil {
			return nil, err
		}
		ts := dblock.GetTimestamp()

		ehashes := eb.GetEntryHashes()
		for _, e := range ehashes {
			if e.IsMinuteMarker() {
				continue
			}
			entry, err := s.DB.FetchEntry(e)
			if err != nil {
				return nil, err
			}

			entries = append(entries, IdentityEntry{entry, ts, height})
		}
	}

	return entries, nil
}

func (st *State) LookupIdentityInBlockchainByChainID(cid interfaces.IHash) error {
	identityRegisterChain, _ := primitives.HexToHash(MAIN_FACTOM_IDENTITY_LIST)

	// No root entries, means no identity
	rootEntries, err := st.FetchIdentityChainEntriesInCreateOrder(cid)
	if err != nil {
		st.IdentityControl.RemoveIdentity(cid)
		return err
	}

	// ** Step 1 **
	// First we need to determine if the identity is registered. We will have to parse the entire
	// register chain (TODO: This should probably be optimized)
	regEntries, err := st.FetchIdentityChainEntriesInCreateOrder(identityRegisterChain)
	if err != nil {
		st.IdentityControl.RemoveIdentity(cid)
		return err
	}

	parseEntryList := func(list []IdentityEntry) {
		for _, e := range list {
			// Instead of calling LoadIdentityByEntry, we can call process directly, as this is initializing
			// an identity.
			st.IdentityControl.ProcessIdentityEntry(e.Entry, e.Blockheight, e.Timestamp, true)
		}
		st.IdentityControl.ProcessOldEntries()
	}

	parseEntryList(regEntries)

	// ** Step 2 **
	// Parse the identity's chain id, which will give us the management chain ID
	parseEntryList(rootEntries)

	// ** Step 3 **
	// Parse the entries contained in the management chain (if exists!)
	id := st.IdentityControl.GetIdentity(cid)
	if id == nil {
		st.IdentityControl.RemoveIdentity(cid)
		return fmt.Errorf("Identity was not found")
	}

	// The id stops here
	if id.ManagementChainID.IsZero() {
		st.IdentityControl.RemoveIdentity(cid)
		return fmt.Errorf("No management chain found for identity")
	}

	manageEntries, err := st.FetchIdentityChainEntriesInCreateOrder(id.ManagementChainID)
	if err != nil {
		st.IdentityControl.RemoveIdentity(cid)
		return err
	}

	parseEntryList(manageEntries)

	// ** Step 4 **
	// Check if it is promotable
	id = st.IdentityControl.GetIdentity(cid)
	if ok, err := id.IsPromteable(); !ok {
		st.IdentityControl.RemoveIdentity(cid)
		return errors.New("Error: Identity not full - " + err.Error())
	}
	return nil
}
