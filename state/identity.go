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

	for _, e := range regEntries {
		// Instead of calling LoadIdentityByEntry, we can call process directly, as this is initializing
		// an identity.
		if e.Entry == nil {
			continue
		}

		// We only care about the identity passed, so ignore all other entries
		if len(e.Entry.ExternalIDs()) == 5 && bytes.Compare(e.Entry.ExternalIDs()[2], cid.Bytes()) == 0 {
			st.IdentityControl.ProcessIdentityEntry(e.Entry, e.Blockheight, e.Timestamp, true)
		}
	}
	st.IdentityControl.ProcessOldEntries()

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

// LoadIdentityByEntry is only useful when initial is set to false. If initial is false, it will track changes
// in an identity that corresponds to an authority. If initial is true, then calling ProcessIdentityEntry directly will
// have the same result.
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
				// Add to admin block
				if st.VerifyIsAuthority(id.IdentityChainID) {
					err := st.LeaderPL.AdminBlock.AddFederatedServerSigningKey(id.IdentityChainID, key.Fixed())
					if err != nil {
						flog.Errorf(err.Error())
					}
				}
			}

			// Is this a change in MHash?
			if !orig.MatryoshkaHash.IsSameAs(id.MatryoshkaHash) {
				if st.VerifyIsAuthority(id.IdentityChainID) {
					err := st.LeaderPL.AdminBlock.AddMatryoshkaHash(id.IdentityChainID, id.MatryoshkaHash)
					if err != nil {
						flog.Errorf(err.Error())
					}
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
				if st.VerifyIsAuthority(id.IdentityChainID) {
					err := st.LeaderPL.AdminBlock.AddFederatedServerBitcoinAnchorKey(id.IdentityChainID, newKey.KeyLevel, newKey.KeyType, newKey.SigningKey)
					if err != nil {
						flog.Errorf(err.Error())
					}
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
