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
	"github.com/FactomProject/factomd/common/messages"
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
	if id == nil { // There should always be a skeleton identity. It cannot be removed
		return nil
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

type IdentityEntry struct {
	Entry       interfaces.IEBEntry
	Timestamp   interfaces.Timestamp
	Blockheight uint32
}

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

func (st *State) AddIdentityFromChainID(cid interfaces.IHash) error {
	identityRegisterChain, _ := primitives.HexToHash(MAIN_FACTOM_IDENTITY_LIST)

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
			st.IdentityControl.ProcessIdentityEntry(e.Entry, e.Blockheight, e.Timestamp, true)
		}
		st.IdentityControl.ProcessOldEntries()
	}

	parseEntryList(regEntries)

	// ** Step 2 **
	// Parse the identity's chain id, which will give us the management chain ID
	rootEntries, err := st.FetchIdentityChainEntriesInCreateOrder(cid)
	if err != nil {
		st.IdentityControl.RemoveIdentity(cid)
		return err
	}

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

func (st *State) OLDAddIdentityFromChainID(cid interfaces.IHash) error {
	if cid.String() == st.GetNetworkBootStrapIdentity().String() || cid.String() == st.GetNetworkSkeletonIdentity().String() { // Ignore Bootstrap Identity, as it is invalid
		return nil
	}

	id := st.IdentityControl.GetIdentity(cid)
	if id == nil {
		id = NewIdentity()
		st.IdentityControl.SetIdentity(cid, id)
	}

	RegIdentityChain, _ := primitives.HexToHash(MAIN_FACTOM_IDENTITY_LIST)
	dbase := st.GetDB()
	ents, err := dbase.FetchAllEntriesByChainID(RegIdentityChain)

	if err != nil {
		return err
	}
	if len(ents) == 0 {
		st.IdentityControl.RemoveIdentity(cid)
		return errors.New("Identity Error: No main Main Factom Identity Chain chain created")
	}

	// Check Identity chain
	eblkStackRoot := make([]interfaces.IEntryBlock, 0)
	mr, err := st.DB.FetchHeadIndexByChainID(cid)
	if err != nil {
		return err
	} else if mr == nil {
		st.IdentityControl.RemoveIdentity(cid)
		return errors.New("Identity Error: Identity Chain not found")
	}
	for !mr.IsSameAs(primitives.NewZeroHash()) {
		eblk, err := st.DB.FetchEBlock(mr)
		if err != nil || eblk == nil {
			break
		}
		eblkStackRoot = append(eblkStackRoot, eblk)
		mr = eblk.GetHeader().GetPrevKeyMR()
	}

	for i := len(eblkStackRoot) - 1; i >= 0; i-- {
		LoadIdentityByEntryBlock(eblkStackRoot[i], st)
	}

	mr, err = st.DB.FetchHeadIndexByChainID(RegIdentityChain)
	if err != nil {
		return err
	}
	// Check Factom Main Identity List
	for !mr.IsSameAs(primitives.NewZeroHash()) {
		eblk, err := st.DB.FetchEBlock(mr)
		if err != nil {
			return err
		}
		if eblk == nil {
			break
		}
		entries := eblk.GetEntryHashes()
		height := eblk.GetDatabaseHeight()
		dblock, err := st.DB.FetchDBlockByHeight(height)
		if err != nil {
			return err
		}
		for _, eHash := range entries {
			if eHash.IsMinuteMarker() { //ignore minute markers
				ent, err := st.DB.FetchEntry(eHash)
				if err != nil || ent == nil {
					continue
				}
				if len(ent.ExternalIDs()) > 3 {
					// This is the Register Factom Identity Message
					if len(ent.ExternalIDs()[2]) == 32 {
						idChain := primitives.NewHash(ent.ExternalIDs()[2][:32])
						if string(ent.ExternalIDs()[1]) == "Register Factom Identity" && cid.IsSameAs(idChain) {
							st.IdentityControl.ProcessIdentityEntry(ent, height, dblock.GetTimestamp(), true)
							//RegisterFactomIdentity(ent, cid, height, st)
							break // Found the registration
						}
					}
				}
			}
		}
		mr = eblk.GetHeader().GetPrevKeyMR()
	}

	id = st.IdentityControl.GetIdentity(cid)

	x := cid.String()
	fmt.Println(x)
	eblkStackSub := make([]interfaces.IEntryBlock, 0)
	if id == nil || id.ManagementChainID == nil || id.ManagementChainID.IsZero() {
		st.IdentityControl.RemoveIdentity(cid)
		return errors.New("Identity Error: No management chain found")
	}
	mr, err = st.DB.FetchHeadIndexByChainID(id.ManagementChainID)
	if err != nil {
		return err
	} else if mr == nil {
		st.IdentityControl.RemoveIdentity(cid)
		return nil
	}
	for !mr.IsSameAs(primitives.NewZeroHash()) {
		eblk, err := st.DB.FetchEBlock(mr)
		if err != nil {
			break
		}
		eblkStackSub = append(eblkStackSub, eblk)
		mr = eblk.GetHeader().GetPrevKeyMR()
	}
	for i := len(eblkStackSub) - 1; i >= 0; i-- {
		LoadIdentityByEntryBlock(eblkStackSub[i], st)
	}

	st.IdentityControl.ProcessOldEntries()

	id = st.IdentityControl.GetIdentity(cid)
	if ok, err := id.IsPromteable(); !ok {
		st.IdentityControl.RemoveIdentity(cid)
		return errors.New("Error: Identity not full - " + err.Error())
	}

	return nil
}

// Should only be called if the Identity is being initialized.
// Using this will not send any message out if a key is changed.
// Eg. Only call from addserver or you don't want any messages being sent.
func LoadIdentityByEntryBlock(eblk interfaces.IEntryBlock, st *State) {
	if eblk == nil {
		identLogger.WithFields(st.Logger.Data).WithField("func", "LoadIdentityByEntryBlock").Info("Initializing identity failed as eblock is nil")
		return
	}
	cid := eblk.GetChainID()
	if cid == nil {
		return
	}

	id := st.IdentityControl.GetIdentity(cid)
	// New parsing
	if id != nil {
		dblock, err := st.DB.FetchDBlockByHeight(eblk.GetDatabaseHeight())
		if err != nil {
			// TODO: Should we panic here? It's a problem because we cannot parse the identity
			panic("")
		}
		entryHashes := eblk.GetEntryHashes()
		for _, eHash := range entryHashes {
			entry, err := st.DB.FetchEntry(eHash)
			if err != nil {
				continue
			}
			LoadIdentityByEntry(entry, st, eblk.GetDatabaseHeight(), dblock.GetTimestamp(), true)
		}
		// Ordering
		st.IdentityControl.ProcessOldEntries()
	}
}

func LoadIdentityByEntry(ent interfaces.IEBEntry, st *State, height uint32, dblockTimestamp interfaces.Timestamp, initial bool) {
	flog := identLogger.WithFields(st.Logger.Data).WithField("func", "LoadIdentityByEntry")
	if ent == nil {
		return
	}

	// Not an entry identities care about
	if bytes.Compare(ent.GetChainID().Bytes()[:3], []byte{0x88, 0x88, 0x88}) != 0 {
		return
	}

	var orig *Identity
	// Not initial means we need to keep track of key changes
	if !initial {
		id := st.IdentityControl.GetIdentity(ent.GetChainID())
		if id != nil {
			orig = id.Clone()
		}
	}

	change, err := st.IdentityControl.ProcessIdentityEntry(ent, height, dblockTimestamp, true)
	if err != nil {
		flog.Errorf(err.Error())
	}

	// TODO: This is inefficient per entry. It is only called for identity entries
	if !initial && orig != nil && change {
		// Can do changing of keys here
		id := st.IdentityControl.GetIdentity(ent.GetChainID())
		if statusIsFedOrAudit(id.Status) {
			// Is this a change in signing key?
			if !orig.SigningKey.IsSameAs(id.SigningKey) {
				key := id.SigningKey
				msg := messages.NewChangeServerKeyMsg(st, id.IdentityChainID, constants.TYPE_ADD_FED_SERVER_KEY, 0, 0, key)
				err := msg.(*messages.ChangeServerKeyMsg).Sign(st)
				if err == nil {
					st.InMsgQueue().Enqueue(msg)
				}
			}

			// Is this a change in MHash?
			if !orig.MatryoshkaHash.IsSameAs(id.MatryoshkaHash) {
				msg := messages.NewChangeServerKeyMsg(st, id.IdentityChainID, constants.TYPE_ADD_MATRYOSHKA, 0, 0, id.MatryoshkaHash)
				err := msg.(*messages.ChangeServerKeyMsg).Sign(st)
				if err == nil {
					st.InMsgQueue().Enqueue(msg)
				}
			}

			var newKey *AnchorSigningKey = nil
			// Is this a change in Anchor?
			// 	Need to find if the set has changed
			if len(orig.AnchorKeys) != len(id.AnchorKeys) {
				// New key to the set, always appended
				newKey = &id.AnchorKeys[len(id.AnchorKeys)-1]
			} else {
				// An existing key could have been changed
				sort.Sort(AnchorSigningKeySort(id.AnchorKeys))
				sort.Sort(AnchorSigningKeySort(orig.AnchorKeys))
				for i := range id.AnchorKeys {
					if !id.AnchorKeys[i].IsSameAs(&orig.AnchorKeys[i]) {
						newKey = &id.AnchorKeys[i]
						break
					}
				}
			}

			if newKey != nil {
				hashLengthKey := append(newKey.SigningKey[:], []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}...)
				key := primitives.NewHash(hashLengthKey)
				msg := messages.NewChangeServerKeyMsg(st, id.IdentityChainID, constants.TYPE_ADD_BTC_ANCHOR_KEY, newKey.KeyLevel, newKey.KeyType, key)
				err := msg.(*messages.ChangeServerKeyMsg).Sign(st.serverPrivKey)
				if err == nil {
					st.InMsgQueue().Enqueue(msg)
				}
			}
		}
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
	return true
}

// Verifies if is authority
//		Return true if authority, false if not
func (st *State) VerifyIsAuthority(cid interfaces.IHash) bool {
	return st.IdentityControl.GetAuthority(cid) != nil
}
