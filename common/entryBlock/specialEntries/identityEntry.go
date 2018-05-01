// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package specialEntries

import (
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
)

//https://github.com/FactomProject/FactomDocs/blob/master/Identity.md

const (
	TypeRegisterServerManagement = "Register Server Management"
	TypeNewBlockSigningKey       = "New Block Signing Key"
	TypeNewBitcoinKey            = "New Bitcoin Key"
	TypeNewMatryoshkaHash        = "New Matryoshka Hash"
	TypeIdentityChain            = "Identity Chain"
	TypeServerManagement         = "Server Management"

	IdentityChainID = "888888001750ede0eff4b05f0c3f557890b256450cabbb84cada937f9c258327"
)

func GetIdentityType(ent interfaces.IEBEntry) string {
	if ent == nil {
		return ""
	}
	extIDs := ent.ExternalIDs()
	if len(extIDs) < 2 {
		return ""
	}
	switch string(extIDs[1]) {
	case TypeRegisterServerManagement:
		return TypeRegisterServerManagement
	case TypeNewBlockSigningKey:
		return TypeNewBlockSigningKey
	case TypeNewBitcoinKey:
		return TypeNewBitcoinKey
	case TypeNewMatryoshkaHash:
		return TypeNewMatryoshkaHash
	case TypeIdentityChain:
		return TypeIdentityChain
	case TypeServerManagement:
		return TypeServerManagement
	}
	return ""
}

func ValidateIdentityEntry(ent interfaces.IEBEntry) error {
	if ent == nil {
		return fmt.Errorf("No entity provided")
	}
	if ent.GetChainID().String() != IdentityChainID {
		return fmt.Errorf("Wrong Entry ChainID - expected %v, got %v", IdentityChainID, ent.GetChainID().String())
	}
	extIDs := ent.ExternalIDs()
	if len(extIDs) < 2 {
		return fmt.Errorf("Not enough ExternalIDs - %v", len(extIDs))
	}
	if len(extIDs[0]) != 1 {
		return fmt.Errorf("Invalid extIDs[0] - %x", extIDs[0])
	}
	if extIDs[0][0] != 0 {
		return fmt.Errorf("Invalid extIDs[0] - %x", extIDs[0])
	}
	switch string(extIDs[1]) {
	case TypeRegisterServerManagement:
		return ValidateRegisterServerManagement(ent)
	case TypeNewBlockSigningKey:
		return ValidateNewBlockSigningKey(ent)
	case TypeNewBitcoinKey:
		return ValidateNewBitcoinKey(ent)
	case TypeNewMatryoshkaHash:
		return ValidateNewMatryoshkaHash(ent)
	case TypeIdentityChain:
		return ValidateIdentityChain(ent)
	case TypeServerManagement:
		return ValidateServerManagement(ent)
	}
	return fmt.Errorf("Invalid IdentityEntry type - %v", string(extIDs[1]))
}

func ValidateRegisterServerManagement(ent interfaces.IEBEntry) error {
	if len(ent.ExternalIDs()) != 9 {
		return fmt.Errorf("Wrong number of ExtIDs - expected 7, got %v", len(ent.ExternalIDs()))
	}
	if CheckExternalIDsLength(ent.ExternalIDs(), []int{1, 15, 32, 1, 1, 20, 8, 33, 64}) == false {
		return fmt.Errorf("Wrong lengths of ExtIDs")
	}
	return nil
}
func ValidateNewBlockSigningKey(ent interfaces.IEBEntry) error {
	if len(ent.ExternalIDs()) != 7 {
		return fmt.Errorf("Wrong number of ExtIDs - expected 7, got %v", len(ent.ExternalIDs()))
	}
	if CheckExternalIDsLength(ent.ExternalIDs(), []int{1, 21, 32, 32, 8, 33, 64}) == false {
		return fmt.Errorf("Wrong lengths of ExtIDs")
	}
	return nil
}
func ValidateNewBitcoinKey(ent interfaces.IEBEntry) error {
	if len(ent.ExternalIDs()) != 9 {
		return fmt.Errorf("Wrong number of ExtIDs - expected 9, got %v", len(ent.ExternalIDs()))
	}
	if CheckExternalIDsLength(ent.ExternalIDs(), []int{1, 15, 32, 1, 1, 20, 8, 33, 64}) == false {
		return fmt.Errorf("Wrong lengths of ExtIDs")
	}

	return nil
}
func ValidateNewMatryoshkaHash(ent interfaces.IEBEntry) error {
	if len(ent.ExternalIDs()) != 7 {
		return fmt.Errorf("Wrong number of ExtIDs - expected 7, got %v", len(ent.ExternalIDs()))
	}
	if CheckExternalIDsLength(ent.ExternalIDs(), []int{1, 19, 32, 32, 8, 33, 64}) == false {
		return fmt.Errorf("Wrong lengths of ExtIDs")
	}
	return nil
}
func ValidateIdentityChain(ent interfaces.IEBEntry) error {
	if len(ent.ExternalIDs()) != 6 {
		return fmt.Errorf("Wrong number of ExtIDs - expected 6, got %v", len(ent.ExternalIDs()))
	}
	if CheckExternalIDsLength(ent.ExternalIDs(), []int{1, 14, 32, 32, 32, 32}) == false {
		return fmt.Errorf("Wrong lengths of ExtIDs")
	}

	return nil
}
func ValidateServerManagement(ent interfaces.IEBEntry) error {
	if len(ent.ExternalIDs()) != 4 {
		return fmt.Errorf("Wrong number of ExtIDs - expected 4, got %v", len(ent.ExternalIDs()))
	}
	if CheckExternalIDsLength(ent.ExternalIDs()[:3], []int{1, 17, 32}) == false {
		return fmt.Errorf("Wrong lengths of ExtIDs")
	}

	return nil
}

// Checking the external ids if they match the needed lengths
func CheckExternalIDsLength(extIDs [][]byte, lengths []int) bool {
	if len(extIDs) != len(lengths) {
		return false
	}
	for i := range extIDs {
		if lengths[i] != len(extIDs[i]) {
			return false
		}
	}
	return true
}

/*
		if len(ent.ExternalIDs()) > 1 {
			if string(ent.ExternalIDs()[1]) == "Register Server Management" {
				registerIdentityAsServer(ent, height, st)
			} else if string(ent.ExternalIDs()[1]) == "New Block Signing Key" {
				if len(ent.ExternalIDs()) == 7 {
					RegisterBlockSigningKey(ent, initial, height, st)
				}
			} else if string(ent.ExternalIDs()[1]) == "New Bitcoin Key" {
				if len(ent.ExternalIDs()) == 9 {
					RegisterAnchorSigningKey(ent, initial, height, st, "BTC")
				}
			} else if string(ent.ExternalIDs()[1]) == "New Matryoshka Hash" {
				if len(ent.ExternalIDs()) == 7 {
					UpdateMatryoshkaHash(ent, initial, height, st)
				}
			} else if len(ent.ExternalIDs()) > 1 && string(ent.ExternalIDs()[1]) == "Identity Chain" {
				addIdentity(ent, height, st)
			} else if len(ent.ExternalIDs()) > 1 && string(ent.ExternalIDs()[1]) == "Server Management" {
				if len(ent.ExternalIDs()) == 4 {
					UpdateManagementKey(ent, height, st)
				}
			}
		}


func RegisterFactomIdentity(entry interfaces.IEBEntry, chainID interfaces.IHash, height uint32, st *State) error {
	extIDs := entry.ExternalIDs()
	if len(extIDs) == 0 {
		return errors.New("Identity Error Register Identity: Invalid external ID length")
	}
	if bytes.Compare([]byte{0x00}, extIDs[0]) != 0 || // Version
		!CheckExternalIDsLength(extIDs, []int{1, 24, 32, 33, 64}) { // Signature
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
			return errors.New("New Management Chain Register for identity [" + chainID.String()[:10] + "] is invalid. Bad signature")
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
		!CheckExternalIDsLength(extIDs, []int{1, 26, 32, 33, 64}) { // Signature
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
			return errors.New("New Management Chain Register for identity [" + chainID.String()[:10] + "] is invalid. Bad signature")
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
				st.InMsgQueue() <- msg
			}
		} else {
			return errors.New("New Block Signing key for identity [" + chainID.String()[:10] + "] is invalid. Bad signature")
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
		!CheckExternalIDsLength(extIDs, []int{1, 19, 32, 32, 8, 33, 64}) { // Signature
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
		return nil
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
				st.InMsgQueue() <- msg
				//}
			}
		} else {
			return errors.New("New Matryoshka Hash for identity [" + chainID.String()[:10] + "] is invalid. Bad signature")
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
	oneAsk.SigningKey = extIDs[5]

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
				st.InMsgQueue() <- msg
			}
		} else {
			return errors.New("New Anchor key for identity [" + chainID.String()[:10] + "] is invalid. Bad signature")
		}
	}
	return nil
}

*/
