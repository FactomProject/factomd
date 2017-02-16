// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package meta

/*
import (
	"fmt"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
)

type IdentityManager struct {
	Authorities          map[string]*Authority
	Identities           map[string]*Identity
	AuthorityServerCount int
}

func (im *IdentityManager) Init() {
	if im.Authorities == nil {
		im.Authorities = map[string]*Authority{}
	}
	if im.Identities == nil {
		im.Identities = map[string]*Identity{}
	}
}

/*
func (im *IdentityManager) AddIdentity(chainID interfaces.IHash, id *Identity) {

}

func (im *IdentityManager) RemoveIdentity(chainID interfaces.IHash) {

}
*/
/*
func (im *IdentityManager) GetIdentity(chainID interfaces.IHash) *Identity {
	return im.Identities[chainID.String()]
}

func (im *IdentityManager) SetAuthority(chainID interfaces.IHash, auth *Authority) {
	im.Authorities[chainID.String()] = auth
}

/*
func (im *IdentityManager) RemoveAuthority(chainID interfaces.IHash) {

}

func (im *IdentityManager) GetAuthority(chainID interfaces.IHash) *Authority {
	return im.Authorities[chainID.String()]
}
*/
/*
func (im *IdentityManager) CreateAuthority(chainID interfaces.IHash) {
	newAuth := new(Authority)
	newAuth.AuthorityChainID = chainID

	identity := im.GetIdentity(chainID)
	if identity != nil {
		if identity.ManagementChainID != nil {
			newAuth.ManagementChainID = identity.ManagementChainID
		}
	}
	newAuth.Status = constants.IDENTITY_PENDING_FULL

	im.AddAuthority(chainID, newAuth)
}

/*
// Checks the signature of a message. Returns an int based on who signed it:
// 			1  -> Federated Signature
//			0  -> Audit Signature
//			-1 -> Neither Fed or Audit Signature
func (im *IdentityManager) VerifyAuthoritySignature(msg []byte, sig *[constants.SIGNATURE_LENGTH]byte, dbheight uint32) (int, error) {
	feds := im.GetFedServers(dbheight)
	if feds == nil {
		return 0, fmt.Errorf("Federated Servers are unknown at directory block hieght %d", dbheight)
	}
	auds := im.GetAuditServers(dbheight)

	for _, fed := range feds {
		auth, _ := im.GetAuthority(fed.GetChainID())
		if auth == nil {
			continue
		}
		valid, err := auth.VerifySignature(msg, sig)
		if err == nil && valid {
			return 1, nil
		}
	}

	for _, aud := range auds {
		auth, _ := im.GetAuthority(aud.GetChainID())
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
*/
/*
// Checks the signature of a message. Returns an int based on who signed it:
// 			1  -> Federated Signature
//			0  -> Audit Signature
//			-1 -> Neither Fed or Audit Signature
func (im *IdentityManager) FastVerifyAuthoritySignature(msg []byte, sig interfaces.IFullSignature, dbheight uint32) (int, error) {
	feds := im.GetFedServers(dbheight)
	if feds == nil {
		return 0, fmt.Errorf("Federated Servers are unknown at directory block hieght %d", dbheight)
	}
	auds := im.GetAuditServers(dbheight)

	for _, fed := range feds {
		auth, _ := im.GetAuthority(fed.GetChainID())
		if auth == nil {
			continue
		}
		compareKey, err := auth.SigningKey.MarshalBinary()
		if err == nil {
			if primitives.AreBytesEqual(sig.GetKey(), compareKey) {
				valid, err := auth.VerifySignature(msg, sig.GetSignature())
				if err == nil && valid {
					return 1, nil
				}
			}
		}
	}

	for _, aud := range auds {
		auth, _ := im.GetAuthority(aud.GetChainID())
		if auth == nil {
			continue
		}
		compareKey, err := auth.SigningKey.MarshalBinary()
		if err == nil {
			if primitives.AreBytesEqual(sig.GetKey(), compareKey) {
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
*/
/*
// Gets the authority matching the identity ChainID.
// Returns the authority and the int of its type:
//		1  ->  Federated
//		0  ->  Audit
// 		-1 ->  Not fed or audit
//		-2 -> Not found
func (im *IdentityManager) GetAuthority(serverID interfaces.IHash) (*Authority, int) {
	auth, ok := im.Authorities[serverID.String()]
	if ok == false {
		return nil, -2
	}
	return auth, auth.Type()
}

// We keep a 1 block history of their keys, this is so if we change their
func (im *IdentityManager) UpdateAuthSigningKeys(height uint32) {
	/*for index, auth := range im.Authorities {
		for _, key := range auth.KeyHistory {
			if key.ActiveDBHeight <= height {
				if len(im.Authorities[index].KeyHistory) == 1 {
					im.Authorities[index].KeyHistory = nil
				} else {
					im.Authorities[index].KeyHistory = im.Authorities[index].KeyHistory[1:]
				}
			}
		}
	}*/ /*
	im.RepairAuthorities()
}

func (im *IdentityManager) UpdateAuthorityFromABEntry(entry interfaces.IABEntry) error {
	switch entry.Type() {
	case constants.TYPE_REVEAL_MATRYOSHKA:
		return im.RevealMatryoshkaHash(e)
	case constants.TYPE_ADD_MATRYOSHKA:
		return im.AddReplaceMatryoshkaHash(e)
	case constants.TYPE_ADD_SERVER_COUNT:
		return im.IncreaseServerCount(e)
	case constants.TYPE_ADD_FED_SERVER:
		return im.AddFederatedServer(e)
	case constants.TYPE_ADD_AUDIT_SERVER:
		return im.AddAuditServer(e)
	case constants.TYPE_REMOVE_FED_SERVER:
		return im.RemoveFederatedServer(e)
	case constants.TYPE_ADD_FED_SERVER_KEY:
		return im.AddFederatedServerSigningKey(e)
	case constants.TYPE_ADD_BTC_ANCHOR_KEY:
		return im.AddFederatedServerBitcoinAnchorKey(e)
	}
	return nil
}

func (im *IdentityManager) RevealMatryoshkaHash(entry interfaces.IABEntry) error {
	//e:=entry.(*adminBlock.RevealMatryoshkaHash)
	// Does nothing for authority right now
	return nil
}

func (im *IdentityManager) AddReplaceMatryoshkaHash(entry interfaces.IABEntry) error {
	e := entry.(*adminBlock.AddReplaceMatryoshkaHash)

	auth, _ := im.GetAuthority(e.IdentityChainID)
	if auth == nil {
		return fmt.Errorf("Authority %v not found", e.IdentityChainID.String())
	}
	auth.MatryoshkaHash = e.MHash
	im.SetAuthority(e.IdentityChainID, auth)

	return nil
}

func (im *IdentityManager) IncreaseServerCount(entry interfaces.IABEntry) error {
	e := entry.(*adminBlock.IncreaseServerCount)
	im.AuthorityServerCount = im.AuthorityServerCount + int(e.Amount)
	return nil
}

func (im *IdentityManager) AddFederatedServer(entry interfaces.IABEntry) error {
	e := entry.(*adminBlock.AddFederatedServer)

	err := im.AddIdentityFromChainID(e.IdentityChainID)
	if err != nil {
		//fmt.Println("Error when Making Identity,", err)
	}
	im.Authorities[e.IdentityChainID.String()].Status = constants.IDENTITY_FEDERATED_SERVER
	// check Identity status
	im.UpdateIdentityStatus(e.IdentityChainID, constants.IDENTITY_FEDERATED_SERVER)
	return nil
}
func (im *IdentityManager) AddAuditServer(entry interfaces.IABEntry) error {
	e := entry.(*adminBlock.AddAuditServer)

	err := im.AddIdentityFromChainID(e.IdentityChainID)
	if err != nil {
		//fmt.Println("Error when Making Identity,", err)
	}
	im.Authorities[e.IdentityChainID.String()].Status = constants.IDENTITY_AUDIT_SERVER
	// check Identity status
	im.UpdateIdentityStatus(e.IdentityChainID, constants.IDENTITY_AUDIT_SERVER)
	return nil
}
func (im *IdentityManager) RemoveFederatedServer(entry interfaces.IABEntry) error {
	e := entry.(*adminBlock.RemoveFederatedServer)

	AuthorityIndex = im.isAuthorityChain(e.IdentityChainID)
	if AuthorityIndex == -1 {
		log.Println(e.IdentityChainID.String() + " Cannot be removed.  Not in Authorities Liim.")
	} else {
		im.RemoveAuthority(e.IdentityChainID)
		IdentityIndex := im.isIdentityChain(e.IdentityChainID)
		if IdentityIndex != -1 && IdentityIndex < len(im.Identities) {
			if im.Identities[IdentityIndex].IdentityChainID.IsSameAs(im.GetNetworkSkeletonIdentity()) {
				im.Identities[IdentityIndex].Status = constants.IDENTITY_SKELETON
			} else {
				im.removeIdentity(IdentityIndex)
			}
		}
	}
	return nil
}
func (im *IdentityManager) AddFederatedServerSigningKey(entry interfaces.IABEntry) error {
	e := entry.(*adminBlock.AddFederatedServerSigningKey)

	keyBytes, err := e.PublicKey.MarshalBinary()
	if err != nil {
		return err
	}
	key := new(primitives.Hash)
	err = key.SetBytes(keyBytes)
	if err != nil {
		return err
	}
	addServerSigningKey(e.IdentityChainID, key, e.DBHeight, am)
	return nil
}
func (im *IdentityManager) AddFederatedServerBitcoinAnchorKey(entry interfaces.IABEntry) error {
	e := entry.(*adminBlock.AddFederatedServerBitcoinAnchorKey)

	pubKey, err := e.ECDSAPublicKey.MarshalBinary()
	if err != nil {
		return err
	}
	registerAuthAnchor(e.IdentityChainID, pubKey, e.KeyType, e.KeyPriority, am, "BTC")
	return nil
}

func (im *IdentityManager) GetAuthorityServerType(chainID interfaces.IHash) int { // 0 = Federated, 1 = Audit
	auth := im.Authorities[chainID.String()]
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

/*
func (im *IdentityManager) AddAuthorityFromChainID(chainID interfaces.IHash) int {
	IdentityIndex := im.isIdentityChain(chainID)
	if IdentityIndex == -1 {
		im.AddIdentityFromChainID(chainID)
	}
	AuthorityIndex := im.isAuthorityChain(chainID)
	if AuthorityIndex == -1 {
		AuthorityIndex = im.createAuthority(chainID)
	}
	return AuthorityIndex
}

func (im *IdentityManager) RemoveAuthority(chainID interfaces.IHash) bool {
	_, ok := im.Authorities[chainID.String()]
	if ok == false {
		return false
	}
	delete(im.Authorities, chainID.String())
	return true
}
*/
/*
// If the Identity failed to create, it will be fixed here
func (im *IdentityManager) RepairAuthorities() {
	// Fix any missing management chains
	for k, auth := range im.Authorities {
		if im.Authorities[k].ManagementChainID == nil {
			idID := im.Authorities[k].AuthorityChainID.String()
			identity := im.Identities[idID]
			if identity == nil {
				err := im.AddIdentityFromChainID(auth.AuthorityChainID)
				if err != nil {
					continue
				}
			}
			im.Authorities[k].ManagementChainID = im.Identities[idID].ManagementChainID
			im.Identities[idID].Status = im.Authorities[k].Status
		}
	}

	// Fix any missing keys
	for _, id := range im.Identities {
		if !id.IsFull() {
			im.FixMissingKeys(id)
		}
	}
}

/*
func registerAuthAnchor(chainID interfaces.IHash, signingKey []byte, keyType byte, keyLevel byte, st *State, BlockChain string) {
	AuthorityIndex := im.AddAuthorityFromChainID(chainID)
	var oneASK AnchorSigningKey

	ask := im.Authorities[AuthorityIndex].AnchorKeys
	newASK := make([]AnchorSigningKey, len(ask)+1)

	for i := 0; i < len(ask); i++ {
		newASK[i] = ask[i]
	}

	oneASK.BlockChain = BlockChain
	oneASK.KeyLevel = keyLevel
	oneASK.KeyType = keyType
	oneASK.SigningKey = signingKey

	newASK[len(ask)] = oneASK
	im.Authorities[AuthorityIndex].AnchorKeys = newASK
}

func addServerSigningKey(chainID interfaces.IHash, key interfaces.IHash, height uint32, st *State) {
	AuthorityIndex := im.AddAuthorityFromChainID(chainID)
	if im.IdentityChainID.IsSameAs(chainID) && len(im.serverPendingPrivKeys) > 0 {
		for i, pubKey := range im.serverPendingPubKeys {
			pubData, err := pubKey.MarshalBinary()
			if err != nil {
				break
			}
			if bytes.Compare(pubData, key.Bytes()) == 0 {
				im.serverPrivKey = im.serverPendingPrivKeys[i]
				im.serverPubKey = im.serverPendingPubKeys[i]
				if len(im.serverPendingPrivKeys) > i+1 {
					im.serverPendingPrivKeys = append(im.serverPendingPrivKeys[:i], im.serverPendingPrivKeys[i+1:]...)
					im.serverPendingPubKeys = append(im.serverPendingPubKeys[:i], im.serverPendingPubKeys[i+1:]...)
				} else {
					im.serverPendingPrivKeys = im.serverPendingPrivKeys[:i]
					im.serverPendingPubKeys = im.serverPendingPubKeys[:i]
				}
				break
			}
		}
	}
	// Add Key History
	im.Authorities[AuthorityIndex].KeyHistory = append(im.Authorities[AuthorityIndex].KeyHistory, struct {
		ActiveDBHeight uint32
		SigningKey     primitives.PublicKey
	}{height, im.Authorities[AuthorityIndex].SigningKey})
	// Replace Active Key
	im.Authorities[AuthorityIndex].SigningKey = primitives.PubKeyFromString(key.String())
}
*/
/*
func (im *IdentityManager) UpdateIdentityStatus(chainID interfaces.IHash, statusTo int) {
	_, ok := im.Identities[chainID.String()]
	if ok == false {
		return
	}
	im.Identities[chainID.String()].Status = statusTo
}

func (im *IdentityManager) RemoveIdentity(chainID interfaces.IHash) {
	index := st.isIdentityChain(chainID)
	st.removeIdentity(index)
}

func (im *IdentityManager) removeIdentity(i int) {
	if st.Identities[i].Status == constants.IDENTITY_SKELETON {
		return // Do not remove skeleton identity
	}
	st.Identities = append(st.Identities[:i], st.Identities[i+1:]...)
}*/
/*
func (im *IdentityManager) isIdentityChain(cid interfaces.IHash) int {
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
*/
