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
