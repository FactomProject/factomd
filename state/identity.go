// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"bytes"
	"errors"
	"strings"

	"github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/identity"
	. "github.com/FactomProject/factomd/common/identityEntries"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"

	log "github.com/sirupsen/logrus"
)

// identLogger is the general logger for all identity related logs. You can add additional fields,
// or create more context loggers off of this
var identLogger = packageLogger.WithFields(log.Fields{"subpack": "identity"})

var (
	TWELVE_HOURS_S uint64 = 12 * 60 * 60
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

	for _, identity := range st.Identities {
		if identity.IdentityChainID.IsSameAs(id) {
			return identity.SigningKey, getReturnStatInt(identity.Status)
		}
	}
	return nil, -1
}

func (st *State) GetNetworkSkeletonKey() interfaces.IHash {
	i := st.isIdentityChain(st.GetNetworkSkeletonIdentity())
	if i == -1 { // There should always be a skeleton identity. It cannot be removed
		return nil
	}

	key := st.Identities[i].SigningKey
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
	// This adds the status
	st.CreateBlankFactomIdentity(skel)
	// This populates the identity with keys found
	err := st.AddIdentityFromChainID(skel)
	if err != nil {
		return err
	}

	return nil
}

func (st *State) AddIdentityFromChainID(cid interfaces.IHash) error {
	if cid.String() == st.GetNetworkBootStrapIdentity().String() { // Ignore Bootstrap Identity, as it is invalid
		return nil
	}

	index := st.isIdentityChain(cid)
	if index == -1 {
		index = st.CreateBlankFactomIdentity(cid)
	}

	managementChain, _ := primitives.HexToHash(MAIN_FACTOM_IDENTITY_LIST)
	dbase := st.GetDB()
	ents, err := dbase.FetchAllEntriesByChainID(managementChain)

	if err != nil {
		return err
	}
	if len(ents) == 0 {
		st.removeIdentity(index)
		return errors.New("Identity Error: No main Main Factom Identity Chain chain created")
	}

	// Check Identity chain
	eblkStackRoot := make([]interfaces.IEntryBlock, 0)
	mr, err := st.DB.FetchHeadIndexByChainID(cid)
	if err != nil {
		return err
	} else if mr == nil {
		st.removeIdentity(index)
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

	mr, err = st.DB.FetchHeadIndexByChainID(managementChain)
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
		for _, eHash := range entries {
			hs := eHash.String()
			if hs[0:10] != "0000000000" { //ignore minute markers
				ent, err := st.DB.FetchEntry(eHash)
				if err != nil || ent == nil {
					continue
				}
				if len(ent.ExternalIDs()) > 3 {
					// This is the Register Factom Identity Message
					if len(ent.ExternalIDs()[2]) == 32 {
						idChain := primitives.NewHash(ent.ExternalIDs()[2][:32])
						if string(ent.ExternalIDs()[1]) == "Register Factom Identity" && cid.IsSameAs(idChain) {
							RegisterFactomIdentity(ent, cid, height, st)
							break // Found the registration
						}
					}
				}
			}
		}
		mr = eblk.GetHeader().GetPrevKeyMR()
	}

	eblkStackSub := make([]interfaces.IEntryBlock, 0)
	if st.Identities[index].ManagementChainID == nil {
		st.removeIdentity(index)
		return errors.New("Identity Error: No management chain found")
	}
	mr, err = st.DB.FetchHeadIndexByChainID(st.Identities[index].ManagementChainID)
	if err != nil {
		return err
	} else if mr == nil {
		st.removeIdentity(index)
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
	err = checkIdentityForFull(index, st)
	if err != nil {
		st.removeIdentity(index)
		return errors.New("Error: Identity not full - " + err.Error())
	}

	return nil
}

func (st *State) RemoveIdentity(chainID interfaces.IHash) {
	index := st.isIdentityChain(chainID)
	st.removeIdentity(index)
}

func (st *State) removeIdentity(i int) {
	if st.Identities[i].Status == constants.IDENTITY_SKELETON {
		return // Do not remove skeleton identity
	}
	st.Identities = append(st.Identities[:i], st.Identities[i+1:]...)
}

func (st *State) isIdentityChain(cid interfaces.IHash) int {
	// is this an identity chain
	for i, identityChain := range st.Identities {
		if identityChain.IdentityChainID.IsSameAs(cid) {
			return i
		}
	}

	// or is it an identity management subchain
	for i, identityChain := range st.Identities {
		if identityChain.ManagementChainID != nil {
			if identityChain.ManagementChainID.IsSameAs(cid) {
				return i
			}
		}
	}
	return -1
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
	if index := st.isIdentityChain(cid); index != -1 {
		entryHashes := eblk.GetEntryHashes()
		for _, eHash := range entryHashes {
			entry, err := st.DB.FetchEntry(eHash)
			if err != nil {
				continue
			}
			LoadIdentityByEntry(entry, st, eblk.GetDatabaseHeight(), true)
		}
	}
}

func LoadIdentityByEntry(ent interfaces.IEBEntry, st *State, height uint32, initial bool) {
	flog := identLogger.WithFields(st.Logger.Data).WithField("func", "LoadIdentityByEntry")
	if ent == nil {
		return
	}
	hs := ent.GetChainID().String()
	cid := ent.GetChainID()
	if st.isIdentityChain(cid) == -1 {
		if st.isAuthorityChain(cid) != -1 {
			// The authority exists, but the Identity does not. This could be an issue if a
			// server changes their key as we would not notice the change
		}
		return
	}
	if hs[0:60] != "000000000000000000000000000000000000000000000000000000000000" { //ignore minute markers
		if len(ent.ExternalIDs()) > 1 {
			if string(ent.ExternalIDs()[1]) == "Register Server Management" {
				registerIdentityAsServer(ent, height, st)
			} else if string(ent.ExternalIDs()[1]) == "New Block Signing Key" {
				if len(ent.ExternalIDs()) == 7 {
					err := RegisterBlockSigningKey(ent, initial, height, st)
					if err != nil {
						flog.Warningf("RegisterBlkSigKey - %s", err.Error())
					}
				}
			} else if string(ent.ExternalIDs()[1]) == "New Bitcoin Key" {
				if len(ent.ExternalIDs()) == 9 {
					err := RegisterAnchorSigningKey(ent, initial, height, st, "BTC")
					if err != nil {
						flog.Warningf("RegisterAnchorKey - %s", err.Error())
					}
				}
			} else if string(ent.ExternalIDs()[1]) == "New Matryoshka Hash" {
				if len(ent.ExternalIDs()) == 7 {
					err := UpdateMatryoshkaHash(ent, initial, height, st)
					if err != nil {
						flog.Warningf("UpdateMatryoshka - %s", err.Error())
					}
				}
			} else if len(ent.ExternalIDs()) > 1 && string(ent.ExternalIDs()[1]) == "Identity Chain" {
				addIdentity(ent, height, st)
			} else if len(ent.ExternalIDs()) > 1 && string(ent.ExternalIDs()[1]) == "Server Management" {
				if len(ent.ExternalIDs()) == 4 {
					err := UpdateManagementKey(ent, height, st)
					if err != nil {
						flog.Warningf("ManageKey - %s", err.Error())
					}
				}
			}
		}
	}
}

// Creates a blank identity
func (st *State) CreateBlankFactomIdentity(chainID interfaces.IHash) int {
	if index := st.isIdentityChain(chainID); index != -1 {
		return index
	}
	var idnew []*Identity
	idnew = make([]*Identity, len(st.Identities)+1)

	var oneID Identity

	for i := 0; i < len(st.Identities); i++ {
		idnew[i] = st.Identities[i]
	}
	oneID.IdentityChainID = chainID

	oneID.Status = constants.IDENTITY_UNASSIGNED
	if chainID.IsSameAs(st.GetNetworkSkeletonIdentity()) {
		oneID.Status = constants.IDENTITY_SKELETON
	}
	oneID.IdentityRegistered = 0
	oneID.IdentityCreated = 0
	oneID.ManagementRegistered = 0
	oneID.ManagementCreated = 0

	oneID.ManagementChainID = primitives.NewZeroHash()
	oneID.Key1 = primitives.NewZeroHash()
	oneID.Key2 = primitives.NewZeroHash()
	oneID.Key3 = primitives.NewZeroHash()
	oneID.Key4 = primitives.NewZeroHash()
	oneID.MatryoshkaHash = primitives.NewZeroHash()
	oneID.SigningKey = primitives.NewZeroHash()

	idnew[len(st.Identities)] = &oneID

	st.Identities = idnew
	return len(st.Identities) - 1
}

func RegisterFactomIdentity(entry interfaces.IEBEntry, chainID interfaces.IHash, height uint32, st *State) error {
	extIDs := entry.ExternalIDs()
	if len(extIDs) == 0 {
		return errors.New("Identity Error Register Identity: Invalid external ID length")
	}
	if bytes.Compare([]byte{0x00}, extIDs[0]) != 0 || // Version
		!CheckExternalIDsLength(extIDs, []int{1, 24, 32, 33, 64}) { // Signiture
		return errors.New("Identity Error Register Identity: Invalid external ID length")
	}

	// find the Identity index from the chain id in the external id.  add this chainID as the management id
	idChain := primitives.NewHash(extIDs[2])
	IdentityIndex := st.isIdentityChain(idChain)
	if IdentityIndex == -1 {
		IdentityIndex = st.CreateBlankFactomIdentity(idChain)
	}

	sigmsg, err := AppendExtIDs(extIDs, 0, 2)
	if err != nil {
		return err
	} else {
		// Verify Signature
		idKey := st.Identities[IdentityIndex].Key1
		if CheckSig(idKey, extIDs[3][1:33], sigmsg, extIDs[4]) {
			st.Identities[IdentityIndex].ManagementRegistered = height
		} else {
			return errors.New("New Management Chain Register for identity [" + chainID.String()[:10] + "] is invalid. Bad signiture")
		}

	}
	st.Identities[IdentityIndex].IdentityRegistered = height

	return nil
}

func addIdentity(entry interfaces.IEBEntry, height uint32, st *State) error {
	extIDs := entry.ExternalIDs()
	// This check is here to prevent possible index out of bounds with extIDs[:6]
	if len(extIDs) != 7 {
		return errors.New("Identity Error Create Management: Invalid external ID length")
	}
	if bytes.Compare([]byte{0x00}, extIDs[0]) != 0 || // Version
		!CheckExternalIDsLength(extIDs[:6], []int{1, 14, 32, 32, 32, 32}) { // Nonce
		return errors.New("Identity Error Create Identity Chain: Invalid external ID length")
	}

	chainID := entry.GetChainID()

	IdentityIndex := st.isIdentityChain(chainID)

	if IdentityIndex == -1 {
		IdentityIndex = st.CreateBlankFactomIdentity(chainID)
	}
	h := primitives.NewHash(extIDs[2])
	st.Identities[IdentityIndex].Key1 = h
	h = primitives.NewHash(extIDs[3])
	st.Identities[IdentityIndex].Key2 = h
	h = primitives.NewHash(extIDs[4])
	st.Identities[IdentityIndex].Key3 = h
	h = primitives.NewHash(extIDs[5])
	st.Identities[IdentityIndex].Key4 = h
	st.Identities[IdentityIndex].IdentityCreated = height
	return nil
}

// This is used when adding an identity. If it is not full, the identity is not added, unless it
// is a federated or audit server. Returning an err makes the identity be removed, so return nil
// if we don't want it removed
func checkIdentityForFull(identityIndex int, st *State) error {
	status := st.Identities[identityIndex].Status
	if statusIsFedOrAudit(st.Identities[identityIndex].Status) || status == constants.IDENTITY_PENDING_FULL || status == constants.IDENTITY_SKELETON {
		return nil // If already full, we don't need to check. If it is fed or audit, we do not need to check
	}

	id := st.Identities[identityIndex]
	dif := id.IdentityCreated - id.IdentityRegistered
	if id.IdentityRegistered > id.IdentityCreated {
		dif = id.IdentityRegistered - id.IdentityCreated
	}
	if dif > TIME_WINDOW {
		return errors.New("Time window of identity create and register invalid")
	}

	dif = id.ManagementCreated - id.ManagementRegistered
	if id.ManagementRegistered > id.ManagementCreated {
		dif = id.ManagementRegistered - id.ManagementCreated
	}
	if dif > TIME_WINDOW {
		return errors.New("Time window of management create and register invalid")
	}

	if id.IdentityChainID == nil {
		return errors.New("Identity Error: No identity chain found")
	}
	if id.ManagementChainID == nil {
		return errors.New("Identity Error: No management chain found")
	}
	if id.SigningKey == nil {
		return errors.New("Identity Error: No block signing key found")
	}
	if id.Key1 == nil || id.Key2 == nil || id.Key3 == nil || id.Key4 == nil {
		return errors.New("Identity Error: Missing an identity key")
	}
	return nil
}

func UpdateManagementKey(entry interfaces.IEBEntry, height uint32, st *State) error {
	extIDs := entry.ExternalIDs()
	// This check is here to prevent possible index out of bounds with extIDs[:3]
	if len(extIDs) != 4 {
		return errors.New("Identity Error Create Management: Invalid external ID length")
	}
	if bytes.Compare([]byte{0x00}, extIDs[0]) != 0 || // Version
		!CheckExternalIDsLength(extIDs[:3], []int{1, 17, 32}) { // Nonce
		return errors.New("Identity Error Create Management: Invalid external ID length")
	}
	chainID := entry.GetChainID()

	idChain := primitives.NewHash(extIDs[2])
	IdentityIndex := st.isIdentityChain(chainID)
	if IdentityIndex == -1 {
		IdentityIndex = st.CreateBlankFactomIdentity(idChain)
	}

	st.Identities[IdentityIndex].ManagementCreated = height
	return nil
}

func registerIdentityAsServer(entry interfaces.IEBEntry, height uint32, st *State) error {
	extIDs := entry.ExternalIDs()
	if len(extIDs) == 0 {
		return errors.New("Identity Error Register Identity: Invalid external ID length")
	}
	if bytes.Compare([]byte{0x00}, extIDs[0]) != 0 || // Version
		!CheckExternalIDsLength(extIDs, []int{1, 26, 32, 33, 64}) { // Signiture
		return errors.New("Identity Error Register Identity: Invalid external ID length")
	}
	chainID := entry.GetChainID()
	IdentityIndex := st.isIdentityChain(chainID)
	if IdentityIndex == -1 {
		IdentityIndex = st.CreateBlankFactomIdentity(chainID)
	}

	sigmsg, err := AppendExtIDs(extIDs, 0, 2)
	if err != nil {
		return err
	} else {
		// Verify Signature
		idKey := st.Identities[IdentityIndex].Key1
		if CheckSig(idKey, extIDs[3][1:33], sigmsg, extIDs[4]) {
			st.Identities[IdentityIndex].ManagementRegistered = height
			st.Identities[IdentityIndex].ManagementChainID = primitives.NewHash(extIDs[2][:32])
		} else {
			return errors.New("New Management Chain Register for identity [" + chainID.String()[:10] + "] is invalid. Bad signiture")
		}

	}
	return nil
}

func RegisterBlockSigningKey(entry interfaces.IEBEntry, initial bool, height uint32, st *State) error {
	extIDs := entry.ExternalIDs()
	if len(extIDs) == 0 {
		return errors.New("Identity Error Block Signing Key: Invalid external ID length")
	}
	if bytes.Compare([]byte{0x00}, extIDs[0]) != 0 || // Version
		!CheckExternalIDsLength(extIDs, []int{1, 21, 32, 32, 8, 33, 64}) {
		return errors.New("Identity Error Block Signing Key: Invalid external ID length")
	}

	subChainID := entry.GetChainID()
	chainID := new(primitives.Hash)
	chainID.SetBytes(extIDs[2][:32])

	IdentityIndex := st.isIdentityChain(chainID)
	if IdentityIndex == -1 {
		return errors.New("Identity Error: This cannot happen. New block signing key to nonexistent identity")
	}

	if !st.Identities[IdentityIndex].ManagementChainID.IsSameAs(subChainID) {
		return errors.New("Identity Error: Entry was not placed in the correct management chain")
	}

	sigmsg, err := AppendExtIDs(extIDs, 0, 4)
	if err != nil {
		return err
	} else {
		//verify Signature
		idKey := st.Identities[IdentityIndex].Key1
		if CheckSig(idKey, extIDs[5][1:33], sigmsg, extIDs[6]) {
			// Check block key length
			if len(extIDs[3]) != 32 {
				return errors.New("New Block Signing key for identity [" + chainID.String()[:10] + "] is invalid length")
			}

			dbase := st.GetDB()
			dblk, err := dbase.FetchDBlockByHeight(height)

			if err == nil && dblk != nil && dblk.GetHeader().GetTimestamp().GetTimeSeconds() != 0 {
				if !CheckTimestamp(extIDs[4], dblk.GetHeader().GetTimestamp().GetTimeSeconds()) {
					return errors.New("New Block Signing key for identity  [" + chainID.String()[:10] + "] timestamp is too old")
				}
			} else {
				if !CheckTimestamp(extIDs[4], st.GetTimestamp().GetTimeSeconds()) {
					return errors.New("New Block Signing key for identity  [" + chainID.String()[:10] + "] timestamp is too old")
				}
			}

			st.Identities[IdentityIndex].SigningKey = primitives.NewHash(extIDs[3])
			// Add to admin block if the following:
			//		Not the initial load
			//		A Federated or Audit server
			//		This node is charge of admin block
			status := st.Identities[IdentityIndex].Status
			if !initial && statusIsFedOrAudit(status) && st.GetLeaderVM() == st.ComputeVMIndex(entry.GetChainID().Bytes()) {
				key := primitives.NewHash(extIDs[3])
				msg := messages.NewChangeServerKeyMsg(st, chainID, constants.TYPE_ADD_FED_SERVER_KEY, 0, 0, key)
				err := msg.(*messages.ChangeServerKeyMsg).Sign(st.serverPrivKey)
				if err != nil {
					return errors.New("New Block Signing key for identity [" + chainID.String()[:10] + "] Error: cannot sign msg")
				}
				st.InMsgQueue().Enqueue(msg)
			}
		} else {
			return errors.New("New Block Signing key for identity [" + chainID.String()[:10] + "] is invalid. Bad signiture")
		}
	}
	return nil
}

func UpdateMatryoshkaHash(entry interfaces.IEBEntry, initial bool, height uint32, st *State) error {
	extIDs := entry.ExternalIDs()
	if len(extIDs) == 0 {
		return errors.New("Identity Error MHash: Invalid external ID length")
	}
	if bytes.Compare([]byte{0x00}, extIDs[0]) != 0 || // Version
		!CheckExternalIDsLength(extIDs, []int{1, 19, 32, 32, 8, 33, 64}) { // Signiture
		return errors.New("Identity Error MHash: Invalid external ID length")
	}
	chainID := new(primitives.Hash)
	chainID.SetBytes(extIDs[2][:32])
	subChainID := entry.GetChainID()

	IdentityIndex := st.isIdentityChain(chainID)
	if IdentityIndex == -1 {
		return errors.New("Identity Error: This cannot happen. New Matryoshka Hash to nonexistent identity")
	}

	if !st.Identities[IdentityIndex].ManagementChainID.IsSameAs(subChainID) {
		return errors.New("Identity Error: Entry was not placed in the correct management chain")
	}

	sigmsg, err := AppendExtIDs(extIDs, 0, 4)
	if err != nil {
		//log.Printfln("Identity Error:", err)
		return err
	} else {
		// Verify Signature
		idKey := st.Identities[IdentityIndex].Key1
		if CheckSig(idKey, extIDs[5][1:33], sigmsg, extIDs[6]) {
			// Check MHash length
			if len(extIDs[3]) != 32 {
				return errors.New("New Matryoshka Hash for identity [" + chainID.String()[:10] + "] is invalid length")
			}

			dbase := st.GetDB()
			dblk, err := dbase.FetchDBlockByHeight(height)

			if err == nil && dblk != nil && dblk.GetHeader().GetTimestamp().GetTimeSeconds() != 0 {
				if !CheckTimestamp(extIDs[4], dblk.GetHeader().GetTimestamp().GetTimeSeconds()) {
					return errors.New("New Matryoshka Hash for identity  [" + chainID.String()[:10] + "] timestamp is too old")
				}
			} else {
				if !CheckTimestamp(extIDs[4], st.GetTimestamp().GetTimeSeconds()) {
					return errors.New("New Matryoshka Hash for identity  [" + chainID.String()[:10] + "] timestamp is too old")
				}
			}

			mhash := primitives.NewHash(extIDs[3])
			st.Identities[IdentityIndex].MatryoshkaHash = mhash
			// Add to admin block
			status := st.Identities[IdentityIndex].Status
			if !initial && statusIsFedOrAudit(status) && st.GetLeaderVM() == st.ComputeVMIndex(entry.GetChainID().Bytes()) {
				//if st.LeaderPL.VMIndexFor(constants.ADMIN_CHAINID) == st.GetLeaderVM() {
				msg := messages.NewChangeServerKeyMsg(st, chainID, constants.TYPE_ADD_MATRYOSHKA, 0, 0, mhash)
				err := msg.(*messages.ChangeServerKeyMsg).Sign(st.serverPrivKey)
				if err != nil {
					return errors.New("New Block Signing key for identity [" + chainID.String()[:10] + "] Error: cannot sign msg")
				}
				//log.Printfln("DEBUG: MHash ChangeServer Message Sent")
				st.InMsgQueue().Enqueue(msg)
				//}
			}
		} else {
			return errors.New("New Matryoshka Hash for identity [" + chainID.String()[:10] + "] is invalid. Bad signiture")
		}

	}
	return nil
}

func RegisterAnchorSigningKey(entry interfaces.IEBEntry, initial bool, height uint32, st *State, BlockChain string) error {
	extIDs := entry.ExternalIDs()
	if bytes.Compare([]byte{0x00}, extIDs[0]) != 0 ||
		!CheckExternalIDsLength(extIDs, []int{1, 15, 32, 1, 1, 20, 8, 33, 64}) {
		return errors.New("Identity Error Anchor Key: Invalid external ID length")
	}

	subChainID := entry.GetChainID()
	chainID := new(primitives.Hash)
	chainID.SetBytes(extIDs[2][:32])

	IdentityIndex := st.isIdentityChain(chainID)
	if IdentityIndex == -1 {
		return errors.New("Identity Error: This cannot happen. New Bitcoin Key to nonexistent identity")
	}

	if !st.Identities[IdentityIndex].ManagementChainID.IsSameAs(subChainID) {
		return errors.New("Identity Error: Entry was not placed in the correct management chain")
	}

	var ask []AnchorSigningKey
	var newAsk []AnchorSigningKey
	var oneAsk AnchorSigningKey

	ask = st.Identities[IdentityIndex].AnchorKeys
	newAsk = make([]AnchorSigningKey, len(ask)+1)

	oneAsk.BlockChain = BlockChain
	oneAsk.KeyLevel = extIDs[3][0]
	oneAsk.KeyType = extIDs[4][0]
	copy(oneAsk.SigningKey[:], extIDs[5])

	contains := false
	for i := 0; i < len(ask); i++ {
		if ask[i].KeyLevel == oneAsk.KeyLevel &&
			strings.Compare(ask[i].BlockChain, oneAsk.BlockChain) == 0 {
			contains = true
			ask[i] = oneAsk
		} else {
			newAsk[i] = ask[i]
		}
	}

	newAsk[len(ask)] = oneAsk
	sigmsg, err := AppendExtIDs(extIDs, 0, 6)
	if err != nil {
		return err
	} else {
		// Verify Signature
		idKey := st.Identities[IdentityIndex].Key1
		if CheckSig(idKey, extIDs[7][1:33], sigmsg, extIDs[8]) {
			var key [20]byte
			if len(extIDs[5]) != 20 {
				return errors.New("New Anchor key for identity [" + chainID.String()[:10] + "] is invalid length")
			}

			dbase := st.GetDB()
			dblk, err := dbase.FetchDBlockByHeight(height)

			if err == nil && dblk != nil && dblk.GetHeader().GetTimestamp().GetTimeSeconds() != 0 {
				if !CheckTimestamp(extIDs[6], dblk.GetHeader().GetTimestamp().GetTimeSeconds()) {
					return errors.New("New Anchor key for identity [" + chainID.String()[:10] + "] timestamp is too old")
				}
			} else {
				if !CheckTimestamp(extIDs[6], st.GetTimestamp().GetTimeSeconds()) {
					return errors.New("New Anchor key for identity [" + chainID.String()[:10] + "] timestamp is too old")
				}
			}

			if contains {
				st.Identities[IdentityIndex].AnchorKeys = ask
			} else {
				st.Identities[IdentityIndex].AnchorKeys = newAsk
			}
			// Add to admin block
			status := st.Identities[IdentityIndex].Status
			if !initial && statusIsFedOrAudit(status) && st.GetLeaderVM() == st.ComputeVMIndex(entry.GetChainID().Bytes()) {
				copy(key[:20], extIDs[5][:20])
				extIDs[5] = append(extIDs[5], []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}...)
				key := primitives.NewHash(extIDs[5])
				msg := messages.NewChangeServerKeyMsg(st, chainID, constants.TYPE_ADD_BTC_ANCHOR_KEY, extIDs[3][0], extIDs[4][0], key)
				err := msg.(*messages.ChangeServerKeyMsg).Sign(st.serverPrivKey)
				if err != nil {
					return errors.New("New Block Signing key for identity [" + chainID.String()[:10] + "] Error: cannot sign msg")
				}
				st.InMsgQueue().Enqueue(msg)
			}
		} else {
			return errors.New("New Anchor key for identity [" + chainID.String()[:10] + "] is invalid. Bad signiture")
		}
	}
	return nil
}

// Called by AddServer Message
func ProcessIdentityToAdminBlock(st *State, chainID interfaces.IHash, servertype int) bool {
	flog := identLogger.WithFields(st.Logger.Data).WithField("func", "ProcessIdentityToAdminBlock")
	var matryoshkaHash interfaces.IHash
	var blockSigningKey [32]byte
	var btcKey [20]byte
	var btcKeyLevel byte
	var btcKeyType byte

	err := st.AddIdentityFromChainID(chainID)
	if err != nil {
		flog.Errorf("Failed to process AddServerMessage for %s : %s", chainID.String()[:10], err.Error())
		return true
	}

	index := st.isIdentityChain(chainID)

	if index != -1 {
		id := st.Identities[index]
		zero := primitives.NewZeroHash()

		if id.SigningKey == nil || id.SigningKey.IsSameAs(zero) {
			flog.Errorf("Failed to process AddServerMessage: %s", "New Fed/Audit server ["+chainID.String()[:10]+"] does not have an Block Signing Key associated to it")
			if !statusIsFedOrAudit(id.Status) {
				st.removeIdentity(index)
			}
			return true
		} else {
			copy(blockSigningKey[:32], id.SigningKey.Bytes()[:32])
		}

		if id.AnchorKeys == nil {
			flog.Errorf("Failed to process AddServerMessage: %s", "New Fed/Audit server ["+chainID.String()[:10]+"] does not have an BTC Anchor Key associated to it")
			if !statusIsFedOrAudit(id.Status) {
				st.removeIdentity(index)
			}
			return true
		} else {
			for _, aKey := range id.AnchorKeys {
				if strings.Compare(aKey.BlockChain, "BTC") == 0 {
					copy(btcKey[:20], aKey.SigningKey[:20])
				}
			}
		}

		if id.MatryoshkaHash == nil || id.MatryoshkaHash.IsSameAs(zero) {
			flog.Errorf("Failed to process AddServerMessage: %s", "New Fed/Audit server ["+chainID.String()[:10]+"] does not have an Matryoshka Key associated to it")
			if !statusIsFedOrAudit(id.Status) {
				st.removeIdentity(index)
			}
			return true
		}
		matryoshkaHash = id.MatryoshkaHash

		if servertype == 0 {
			id.Status = constants.IDENTITY_PENDING_FEDERATED_SERVER
		} else if servertype == 1 {
			id.Status = constants.IDENTITY_PENDING_AUDIT_SERVER
		}
		st.Identities[index] = id
	} else {
		flog.Errorf("Failed to process AddServerMessage: %s", "New Fed/Audit server ["+chainID.String()[:10]+"] does not have an identity associated to it")
		return true
	}

	// Add to admin block
	if servertype == 0 {
		st.LeaderPL.AdminBlock.AddFedServer(chainID)
		st.Identities[index].Status = constants.IDENTITY_PENDING_FEDERATED_SERVER
	} else if servertype == 1 {
		st.LeaderPL.AdminBlock.AddAuditServer(chainID)
		st.Identities[index].Status = constants.IDENTITY_PENDING_AUDIT_SERVER
	}
	st.LeaderPL.AdminBlock.AddFederatedServerSigningKey(chainID, blockSigningKey)
	st.LeaderPL.AdminBlock.AddMatryoshkaHash(chainID, matryoshkaHash)
	st.LeaderPL.AdminBlock.AddFederatedServerBitcoinAnchorKey(chainID, btcKeyLevel, btcKeyType, btcKey)
	return true
}

// Verifies if is authority
func (st *State) VerifyIsAuthority(cid interfaces.IHash) bool {
	if st.isAuthorityChain(cid) != -1 {
		return true
	}
	return false
}

func UpdateIdentityStatus(ChainID interfaces.IHash, StatusTo uint8, st *State) {
	IdentityIndex := st.isIdentityChain(ChainID)
	if IdentityIndex == -1 {
		return
	}
	st.Identities[IdentityIndex].Status = StatusTo
}
