// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/identity"
	"github.com/FactomProject/factomd/common/interfaces"
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
	// NEW
	auth := st.IdentityControl.GetAuthority(serverID)
	if auth == nil {
		return nil, -2
	}

	return auth, auth.Type()
}

// We keep a 2 block history of their keys, this is so if we change their
func (st *State) UpdateAuthSigningKeys(height uint32) {
	// NEW
	for key, auth := range st.IdentityControl.Authorities {
		chopOffIndex := 0 // Index of the keys we should chop off
		for i, key := range auth.KeyHistory {
			// Keeping 2 heights worth.
			if key.ActiveDBHeight <= height-2 {
				chopOffIndex = i
			}
		}

		if chopOffIndex > 0 {
			if len(st.IdentityControl.Authorities[key].KeyHistory) == chopOffIndex+1 {
				st.IdentityControl.Authorities[key].KeyHistory = nil
			} else {
				// This could be a memory leak if the authority keeps updating his keys every block,
				// but the line above sets to nil if there is only 1 item left, so it will eventually
				// garbage collect the whole slice
				st.IdentityControl.Authorities[key].KeyHistory = st.IdentityControl.Authorities[auth.AuthorityChainID.Fixed()].KeyHistory[chopOffIndex+1:]
			}

		}
	}

	st.RepairAuthorities()
}

func (st *State) UpdateAuthorityFromABEntry(entry interfaces.IABEntry) error {
	err := st.IdentityControl.ProcessABlockEntry(entry, st)
	if err != nil {
		return err
	}

	//switch entry.Type() {
	//case constants.TYPE_REVEAL_MATRYOSHKA:
	//	r := new(adminBlock.RevealMatryoshkaHash)
	//	err := r.UnmarshalBinary(data)
	//	if err != nil {
	//		return err
	//	}
	//	// Does nothing for authority right now
	//case constants.TYPE_ADD_MATRYOSHKA:
	//	m := new(adminBlock.AddReplaceMatryoshkaHash)
	//	err := m.UnmarshalBinary(data)
	//	if err != nil {
	//		return err
	//	}
	//	AuthorityIndex = st.AddAuthorityFromChainID(m.IdentityChainID)
	//	st.Authorities[AuthorityIndex].MatryoshkaHash = m.MHash
	//case constants.TYPE_ADD_SERVER_COUNT:
	//	s := new(adminBlock.IncreaseServerCount)
	//	err := s.UnmarshalBinary(data)
	//	if err != nil {
	//		return err
	//	}
	//
	//	st.AuthorityServerCount = st.AuthorityServerCount + int(s.Amount)
	//case constants.TYPE_ADD_FED_SERVER:
	//	f := new(adminBlock.AddFederatedServer)
	//	err := f.UnmarshalBinary(data)
	//	if err != nil {
	//		return err
	//	}
	//	err = st.AddIdentityFromChainID(f.IdentityChainID)
	//	if err != nil {
	//		//fmt.Println("Error when Making Identity,", err)
	//	}
	//	AuthorityIndex = st.AddAuthorityFromChainID(f.IdentityChainID)
	//	st.Authorities[AuthorityIndex].Status = constants.IDENTITY_FEDERATED_SERVER
	//	// check Identity status
	//	UpdateIdentityStatus(f.IdentityChainID, constants.IDENTITY_FEDERATED_SERVER, st)
	//case constants.TYPE_ADD_AUDIT_SERVER:
	//	a := new(adminBlock.AddAuditServer)
	//	err := a.UnmarshalBinary(data)
	//	if err != nil {
	//		return err
	//	}
	//	err = st.AddIdentityFromChainID(a.IdentityChainID)
	//	if err != nil {
	//		//fmt.Println("Error when Making Identity,", err)
	//	}
	//	AuthorityIndex = st.AddAuthorityFromChainID(a.IdentityChainID)
	//	st.Authorities[AuthorityIndex].Status = constants.IDENTITY_AUDIT_SERVER
	//	// check Identity status
	//	UpdateIdentityStatus(a.IdentityChainID, constants.IDENTITY_AUDIT_SERVER, st)
	//case constants.TYPE_REMOVE_FED_SERVER:
	//	f := new(adminBlock.RemoveFederatedServer)
	//	err := f.UnmarshalBinary(data)
	//	if err != nil {
	//		return err
	//	}
	//	AuthorityIndex = st.isAuthorityChain(f.IdentityChainID)
	//	if AuthorityIndex == -1 {
	//		log.Println(f.IdentityChainID.String() + " Cannot be removed.  Not in Authorities List.")
	//	} else {
	//		st.RemoveAuthority(f.IdentityChainID)
	//		IdentityIndex := st.isIdentityChain(f.IdentityChainID)
	//		if IdentityIndex != -1 && IdentityIndex < len(st.Identities) {
	//			if st.Identities[IdentityIndex].IdentityChainID.IsSameAs(st.GetNetworkSkeletonIdentity()) {
	//				st.Identities[IdentityIndex].Status = constants.IDENTITY_SKELETON
	//			} else {
	//				st.removeIdentity(IdentityIndex)
	//			}
	//		}
	//	}
	//case constants.TYPE_ADD_FED_SERVER_KEY:
	//	f := new(adminBlock.AddFederatedServerSigningKey)
	//	err := f.UnmarshalBinary(data)
	//	if err != nil {
	//		return err
	//	}
	//	keyBytes, err := f.PublicKey.MarshalBinary()
	//	if err != nil {
	//		return err
	//	}
	//	key := new(primitives.Hash)
	//	err = key.SetBytes(keyBytes)
	//	if err != nil {
	//		return err
	//	}
	//	addServerSigningKey(f.IdentityChainID, key, f.DBHeight, st)
	//case constants.TYPE_ADD_BTC_ANCHOR_KEY:
	//	b := new(adminBlock.AddFederatedServerBitcoinAnchorKey)
	//	err := b.UnmarshalBinary(data)
	//	if err != nil {
	//		return err
	//	}
	//	pubKey, err := b.ECDSAPublicKey.MarshalBinary()
	//	if err != nil {
	//		return err
	//	}
	//	registerAuthAnchor(b.IdentityChainID, pubKey, b.KeyType, b.KeyPriority, st, "BTC")
	//}

	return nil
}

func (st *State) GetAuthorityServerType(chainID interfaces.IHash) int { // 0 = Federated, 1 = Audit
	auth := st.IdentityControl.GetAuthority(chainID)
	if auth == nil {
		return -1
	}

	status := auth.Status
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

// If the Identity failed to create, it will be fixed here
func (s *State) RepairAuthorities() {
	// Fix any missing management chains
	for _, iAuth := range s.IdentityControl.GetAuthorities() {
		auth := iAuth.(*Authority)
		if auth.ManagementChainID == nil || auth.ManagementChainID.IsZero() {
			id := s.IdentityControl.GetIdentity(auth.AuthorityChainID)
			if id == nil {
				err := s.AddIdentityFromChainID(auth.AuthorityChainID)
				if err != nil {
					continue
				}
				id = s.IdentityControl.GetIdentity(auth.AuthorityChainID)
			}
			if id != nil {
				auth.ManagementChainID = id.ManagementChainID
				id.Status = auth.Status
				s.IdentityControl.SetAuthority(auth.AuthorityChainID, auth)
				s.IdentityControl.SetIdentity(auth.AuthorityChainID, id)
			}
		}
	}

	// Fix any missing keys
	for _, id := range s.IdentityControl.GetIdentities() {
		if !id.IsFull() {
			s.FixMissingKeys(id)
		}
	}
}
