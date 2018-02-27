// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"bytes"
	"fmt"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/identity"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/log"
)

// Checks the signature of a message. Returns an int based on who signed it:
// 			1  -> Federated Signature
//			0  -> Audit Signature
//			-1 -> Neither Fed or Audit Signature
func (st *State) VerifyAuthoritySignature(msg []byte, sig *[constants.SIGNATURE_LENGTH]byte, dbheight uint32) (int, error) {
	feds := st.GetFedServers(dbheight)
	if feds == nil {
		return -1, fmt.Errorf("Federated Servers are unknown at directory block height %d", dbheight)
	}

	for _, fed := range feds {
		auth, _ := st.GetAuthority(fed.GetChainID())
		if auth == nil {
			continue
		}
		valid, err := auth.VerifySignature(msg, sig)
		if err == nil && valid {
			return 1, nil
		}
	}

	auds := st.GetAuditServers(dbheight)
	if auds == nil {
		return -1, fmt.Errorf("Audit Servers are unknown at directory block height %d", dbheight)
	}
	for _, aud := range auds {
		auth, _ := st.GetAuthority(aud.GetChainID())
		if auth == nil {
			continue
		}
		valid, err := auth.VerifySignature(msg, sig)
		if err == nil && valid {
			return 0, nil
		}
	}
	//fmt.Println("WARNING: A signature failed to validate.")

	return -1, fmt.Errorf("%s", "Signature Key Invalid or not Federated Server Key")
}

// Checks the signature of a message. Returns an int based on who signed it:
// 			1  -> Federated Signature
//			0  -> Audit Signature
//			-1 -> Neither Fed or Audit Signature
func (st *State) FastVerifyAuthoritySignature(msg []byte, sig interfaces.IFullSignature, dbheight uint32) (int, error) {
	feds := st.GetFedServers(dbheight)
	if feds == nil {
		return 0, fmt.Errorf("Federated Servers are unknown at directory block height %d", dbheight)
	}
	auds := st.GetAuditServers(dbheight)

	for _, fed := range feds {
		auth, _ := st.GetAuthority(fed.GetChainID())
		if auth == nil {
			continue
		}
		compareKey, err := auth.SigningKey.MarshalBinary()
		if err == nil {
			if pkEq(sig.GetKey(), compareKey) {
				valid, err := auth.VerifySignature(msg, sig.GetSignature())
				if err == nil && valid {
					return 1, nil
				}
			}
		}
	}

	for _, aud := range auds {
		auth, _ := st.GetAuthority(aud.GetChainID())
		if auth == nil {
			continue
		}
		compareKey, err := auth.SigningKey.MarshalBinary()
		if err == nil {
			if pkEq(sig.GetKey(), compareKey) {
				valid, err := auth.VerifySignature(msg, sig.GetSignature())
				if err == nil && valid {
					return 0, nil
				}
			}
		}
	}
	//fmt.Println("WARNING: A signature failed to validate.")

	return -1, fmt.Errorf("%s", "Signature Key Invalid or not Federated Server Key")
}

func pkEq(a, b []byte) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// Gets the authority matching the identity ChainID.
// Returns the authority and the int of its type:
//		1  ->  Federated
//		0  ->  Audit
// 		-1 ->  Not fed or audit
//		-2 -> Not found
func (st *State) GetAuthority(serverID interfaces.IHash) (*Authority, int) {
	for _, auth := range st.Authorities {
		if serverID.IsSameAs(auth.AuthorityChainID) {
			return auth, auth.Type()
		}
	}
	return nil, -2
}

// We keep a 1 block history of their keys, this is so if we change their
func (st *State) UpdateAuthSigningKeys(height uint32) {
	/*for index, auth := range st.Authorities {
		for _, key := range auth.KeyHistory {
			if key.ActiveDBHeight <= height {
				if len(st.Authorities[index].KeyHistory) == 1 {
					st.Authorities[index].KeyHistory = nil
				} else {
					st.Authorities[index].KeyHistory = st.Authorities[index].KeyHistory[1:]
				}
			}
		}
	}*/
	st.RepairAuthorities()
}

func (st *State) UpdateAuthorityFromABEntry(entry interfaces.IABEntry) error {
	var AuthorityIndex int
	data, err := entry.MarshalBinary()
	if err != nil {
		return err
	}
	switch entry.Type() {
	case constants.TYPE_REVEAL_MATRYOSHKA:
		r := new(adminBlock.RevealMatryoshkaHash)
		err := r.UnmarshalBinary(data)
		if err != nil {
			return err
		}
		// Does nothing for authority right now
	case constants.TYPE_ADD_MATRYOSHKA:
		m := new(adminBlock.AddReplaceMatryoshkaHash)
		err := m.UnmarshalBinary(data)
		if err != nil {
			return err
		}
		AuthorityIndex = st.AddAuthorityFromChainID(m.IdentityChainID)
		st.Authorities[AuthorityIndex].MatryoshkaHash = m.MHash
	case constants.TYPE_ADD_SERVER_COUNT:
		s := new(adminBlock.IncreaseServerCount)
		err := s.UnmarshalBinary(data)
		if err != nil {
			return err
		}

		st.AuthorityServerCount = st.AuthorityServerCount + int(s.Amount)
	case constants.TYPE_ADD_FED_SERVER:
		f := new(adminBlock.AddFederatedServer)
		err := f.UnmarshalBinary(data)
		if err != nil {
			return err
		}
		err = st.AddIdentityFromChainID(f.IdentityChainID)
		if err != nil {
			//fmt.Println("Error when Making Identity,", err)
		}
		AuthorityIndex = st.AddAuthorityFromChainID(f.IdentityChainID)
		st.Authorities[AuthorityIndex].Status.Store(constants.IDENTITY_FEDERATED_SERVER)
		// check Identity status
		UpdateIdentityStatus(f.IdentityChainID, constants.IDENTITY_FEDERATED_SERVER, st)
	case constants.TYPE_ADD_AUDIT_SERVER:
		a := new(adminBlock.AddAuditServer)
		err := a.UnmarshalBinary(data)
		if err != nil {
			return err
		}
		err = st.AddIdentityFromChainID(a.IdentityChainID)
		if err != nil {
			//fmt.Println("Error when Making Identity,", err)
		}
		AuthorityIndex = st.AddAuthorityFromChainID(a.IdentityChainID)
		st.Authorities[AuthorityIndex].Status.Store(constants.IDENTITY_AUDIT_SERVER)
		// check Identity status
		UpdateIdentityStatus(a.IdentityChainID, constants.IDENTITY_AUDIT_SERVER, st)
	case constants.TYPE_REMOVE_FED_SERVER:
		f := new(adminBlock.RemoveFederatedServer)
		err := f.UnmarshalBinary(data)
		if err != nil {
			return err
		}
		AuthorityIndex = st.isAuthorityChain(f.IdentityChainID)
		if AuthorityIndex == -1 {
			log.Println(f.IdentityChainID.String() + " Cannot be removed.  Not in Authorities List.")
		} else {
			st.RemoveAuthority(f.IdentityChainID)
			IdentityIndex := st.isIdentityChain(f.IdentityChainID)
			if IdentityIndex != -1 && IdentityIndex < len(st.Identities) {
				if st.Identities[IdentityIndex].IdentityChainID.IsSameAs(st.GetNetworkSkeletonIdentity()) {
					st.Identities[IdentityIndex].Status.Store(constants.IDENTITY_SKELETON)
				} else {
					st.removeIdentity(IdentityIndex)
				}
			}
		}
	case constants.TYPE_ADD_FED_SERVER_KEY:
		f := new(adminBlock.AddFederatedServerSigningKey)
		err := f.UnmarshalBinary(data)
		if err != nil {
			return err
		}
		keyBytes, err := f.PublicKey.MarshalBinary()
		if err != nil {
			return err
		}
		key := new(primitives.Hash)
		err = key.SetBytes(keyBytes)
		if err != nil {
			return err
		}
		addServerSigningKey(f.IdentityChainID, key, f.DBHeight, st)
	case constants.TYPE_ADD_BTC_ANCHOR_KEY:
		b := new(adminBlock.AddFederatedServerBitcoinAnchorKey)
		err := b.UnmarshalBinary(data)
		if err != nil {
			return err
		}
		pubKey, err := b.ECDSAPublicKey.MarshalBinary()
		if err != nil {
			return err
		}
		registerAuthAnchor(b.IdentityChainID, pubKey, b.KeyType, b.KeyPriority, st, "BTC")
	}
	return nil
}

func (st *State) GetAuthorityServerType(chainID interfaces.IHash) int { // 0 = Federated, 1 = Audit
	index := st.isAuthorityChain(chainID)
	if index == -1 {
		return -1
	}
	status := st.Authorities[index].Status.Load()
	if status == constants.IDENTITY_FEDERATED_SERVER ||
		status == constants.IDENTITY_PENDING_FEDERATED_SERVER {
		return 0
	}
	if status == constants.IDENTITY_AUDIT_SERVER ||
		status == constants.IDENTITY_PENDING_AUDIT_SERVER {
		return 1
	}
	return -1
}

func (st *State) AddAuthorityFromChainID(chainID interfaces.IHash) int {
	IdentityIndex := st.isIdentityChain(chainID)
	if IdentityIndex == -1 {
		st.AddIdentityFromChainID(chainID)
	}
	AuthorityIndex := st.isAuthorityChain(chainID)
	if AuthorityIndex == -1 {
		AuthorityIndex = st.createAuthority(chainID)
	}
	return AuthorityIndex
}

func (st *State) RemoveAuthority(chainID interfaces.IHash) bool {
	i := st.isAuthorityChain(chainID)
	if i == -1 {
		return false
	}
	if len(st.Authorities) > i+1 {
		st.Authorities = append(st.Authorities[:i], st.Authorities[i+1:]...)
	} else {
		st.Authorities = st.Authorities[:i]
	}
	return true
}

func (st *State) isAuthorityChain(cid interfaces.IHash) int {
	for i, authorityChain := range st.Authorities {
		if authorityChain.AuthorityChainID.IsSameAs(cid) {
			return i
		}
	}
	return -1
}

func (st *State) createAuthority(chainID interfaces.IHash) int {
	newAuth := new(Authority)
	newAuth.AuthorityChainID = chainID

	idIndex := st.isIdentityChain(chainID)
	if idIndex != -1 && st.Identities[idIndex].ManagementChainID != nil {
		newAuth.ManagementChainID = st.Identities[idIndex].ManagementChainID
	}
	newAuth.Status.Store(constants.IDENTITY_PENDING_FULL)

	st.Authorities = append(st.Authorities, newAuth)
	return len(st.Authorities) - 1
}

// If the Identity failed to create, it will be fixed here
func (s *State) RepairAuthorities() {
	// Fix any missing management chains
	for i, auth := range s.Authorities {
		if s.Authorities[i].ManagementChainID == nil {
			idIndex := s.isIdentityChain(s.Authorities[i].AuthorityChainID)
			if idIndex == -1 {
				err := s.AddIdentityFromChainID(auth.AuthorityChainID)
				if err != nil {
					continue
				}
				idIndex = s.isIdentityChain(s.Authorities[i].AuthorityChainID)
			}
			if idIndex != -1 {
				s.Authorities[i].ManagementChainID = s.Identities[idIndex].ManagementChainID
				s.Identities[idIndex].Status.Store(s.Authorities[i].Status.Load())
			}
		}
	}

	// Fix any missing keys
	for _, id := range s.Identities {
		if !id.IsFull() {
			s.FixMissingKeys(id)
		}
	}
}

func registerAuthAnchor(chainID interfaces.IHash, signingKey []byte, keyType byte, keyLevel byte, st *State, BlockChain string) {
	AuthorityIndex := st.AddAuthorityFromChainID(chainID)
	var oneASK AnchorSigningKey

	ask := st.Authorities[AuthorityIndex].AnchorKeys
	newASK := make([]AnchorSigningKey, len(ask)+1)

	for i := 0; i < len(ask); i++ {
		newASK[i] = ask[i]
	}

	oneASK.BlockChain = BlockChain
	oneASK.KeyLevel = keyLevel
	oneASK.KeyType = keyType
	copy(oneASK.SigningKey[:], signingKey)

	newASK[len(ask)] = oneASK
	st.Authorities[AuthorityIndex].AnchorKeys = newASK
}

func addServerSigningKey(chainID interfaces.IHash, key interfaces.IHash, height uint32, st *State) {
	AuthorityIndex := st.AddAuthorityFromChainID(chainID)
	if st.IdentityChainID.IsSameAs(chainID) && len(st.serverPendingPrivKeys) > 0 {
		for i, pubKey := range st.serverPendingPubKeys {
			pubData, err := pubKey.MarshalBinary()
			if err != nil {
				break
			}
			if bytes.Compare(pubData, key.Bytes()) == 0 {
				st.serverPrivKey = st.serverPendingPrivKeys[i]
				st.serverPubKey = st.serverPendingPubKeys[i]
				if len(st.serverPendingPrivKeys) > i+1 {
					st.serverPendingPrivKeys = append(st.serverPendingPrivKeys[:i], st.serverPendingPrivKeys[i+1:]...)
					st.serverPendingPubKeys = append(st.serverPendingPubKeys[:i], st.serverPendingPubKeys[i+1:]...)
				} else {
					st.serverPendingPrivKeys = st.serverPendingPrivKeys[:i]
					st.serverPendingPubKeys = st.serverPendingPubKeys[:i]
				}
				break
			}
		}
	}
	// Add Key History
	st.Authorities[AuthorityIndex].KeyHistory = append(st.Authorities[AuthorityIndex].KeyHistory, struct {
		ActiveDBHeight uint32
		SigningKey     primitives.PublicKey
	}{height, st.Authorities[AuthorityIndex].SigningKey})
	// Replace Active Key
	st.Authorities[AuthorityIndex].SigningKey = primitives.PubKeyFromString(key.String())
}
