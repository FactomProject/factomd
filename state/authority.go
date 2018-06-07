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
func (st *State) VerifyAuthoritySignature(msg []byte, sig *[constants.SIGNATURE_LENGTH]byte, dbheight uint32) (rval int, err error) {

	//defer func() { // debug code
	//	//st.LogMessage("executeMsg", "Signature Fail", msg)
	//	if rval <= 0 {
	//		m, err := messages.General.UnmarshalMessage(append(msg[:], make([]byte, 256)[:]...))
	//		if err != nil {
	//			st.LogPrintf("executeMsg", "Unable to unmarshal message")
	//		} else {
	//			st.LogMessage("executeMsg", "VerifyAuthoritySignature", m)
	//		}
	//		st.LogPrintf("executeMsg", "VerifyAuthoritySignature failed signature")
	//
	//		feds := st.GetFedServers(dbheight)
	//		for _, fed := range feds {
	//			st.LogPrintf("executeMsg", "L %s:%s", messages.LookupName(fed.GetChainID().String()), fed.GetChainID().String()[6:12])
	//		}
	//
	//		auds := st.GetAuditServers(dbheight)
	//		if auds == nil {
	//			st.LogPrintf("executeMsg", "Audit Servers are unknown at directory block height %d", dbheight)
	//		}
	//		for _, aud := range auds {
	//			st.LogPrintf("executeMsg", "A %s:%s", messages.LookupName(aud.GetChainID().String()), aud.GetChainID().String()[6:12])
	//		}
	//
	//		st.LogPrintf("executeMsg", "auth.VerifySignature(msg:%x, sig:%x)", msg, sig)
	//		for _, s := range feds {
	//			auth, _ := st.GetAuthority(s.GetChainID())
	//			valid, err := auth.VerifySignature2(msg, sig)
	//			st.LogPrintf("executeMsg", "L-%x valid:%v, err:%v", s.GetChainID().Bytes()[3:6], valid, err)
	//		}
	//		for _, s := range auds {
	//			auth, _ := st.GetAuthority(s.GetChainID())
	//			valid, err := auth.VerifySignature2(msg, sig)
	//			st.LogPrintf("executeMsg", "A-%x valid:%v, err:%v", s.GetChainID().Bytes()[3:6], valid, err)
	//		}
	//
	//	}
	//}() // end debug code
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
	if st.CurrentMinute == 0 {
		// Also allow leaders who were demoted if we are in minute 0
		feds := st.LeaderPL.StartingFedServers
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

	// The checking pl for nil happens for unit testing
	if st.CurrentMinute == 0 && st.LeaderPL != nil {
		// Also allow leaders who were demoted if we are in minute 0
		feds := st.LeaderPL.StartingFedServers
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
	}
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
	auth := st.IdentityControl.GetAuthority(serverID)
	if auth == nil {
		return nil, -2
	}

	return auth, auth.Type()
}

// We keep a 2 block history of their keys, this is so if we change their key and need to verify
// a message from 1 block ago, we still can. This function garbage collects old keys
func (st *State) UpdateAuthSigningKeys(height uint32) {
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

// RepairAuthorities will put the management chain of an identity in the authority if it
// is missing.
func (s *State) RepairAuthorities() {
	// Fix any missing management chains
	for _, iAuth := range s.IdentityControl.GetAuthorities() {
		auth := iAuth.(*Authority)
		if auth.ManagementChainID == nil || auth.ManagementChainID.IsZero() {
			id := s.IdentityControl.GetIdentity(auth.AuthorityChainID)
			if id != nil {
				auth.ManagementChainID = id.ManagementChainID
				id.Status = auth.Status
				s.IdentityControl.SetAuthority(auth.AuthorityChainID, auth)
				s.IdentityControl.SetIdentity(auth.AuthorityChainID, id)
			}
		}
	}
}
