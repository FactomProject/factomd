// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockchainState

import (
	"fmt"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

func (bs *BlockchainState) ProcessABlock(aBlock interfaces.IAdminBlock, dBlock interfaces.IDirectoryBlock, prevHeader []byte) error {
	bs.Init()

	if bs.ABlockHeadRefHash.String() != aBlock.GetHeader().GetPrevBackRefHash().String() {
		return fmt.Errorf("Invalid ABlock %v previous KeyMR - expected %v, got %v\n", aBlock.GetHash(), bs.ABlockHeadRefHash.String(), aBlock.GetHeader().GetPrevBackRefHash().String())
	}
	bs.ABlockHeadRefHash = aBlock.DatabaseSecondaryIndex().(*primitives.Hash)

	if bs.DBlockHeight != aBlock.GetDatabaseHeight() {
		return fmt.Errorf("Invalid ABlock height - expected %v, got %v", bs.DBlockHeight, aBlock.GetDatabaseHeight())
	}

	err := CheckABlockMinuteNumbers(aBlock)
	if err != nil {
		return err
	}

	err = bs.CheckDBSignatureEntries(aBlock, dBlock, prevHeader)
	if err != nil {
		return err
	}

	for _, v := range aBlock.GetABEntries() {
		err = bs.ProcessABlockEntry(v)
		if err != nil {
			return err
		}
	}

	return nil
}

func CheckABlockMinuteNumbers(aBlock interfaces.IAdminBlock) error {
	//Check whether MinuteNumbers are increasing
	entries := aBlock.GetABEntries()

	var lastMinute uint8 = 0
	for i, v := range entries {
		if v.Type() == constants.TYPE_MINUTE_NUM {
			minute := v.(*adminBlock.EndOfMinuteEntry).MinuteNumber
			if minute < 1 || minute > 10 {
				return fmt.Errorf("ABlock Invalid minute number at position %v", i)
			}
			if minute <= lastMinute {
				return fmt.Errorf("ABlock Invalid minute number at position %v", i)
			}
			lastMinute = minute
		}
	}

	return nil
}

func (bs *BlockchainState) CheckDBSignatureEntries(aBlock interfaces.IAdminBlock, dBlock interfaces.IDirectoryBlock, prevHeader []byte) error {
	if dBlock.GetDatabaseHeight() == 0 {
		return nil
	}

	entries := aBlock.GetABEntries()

	foundSigs := map[string]string{}

	for _, v := range entries {
		if v.Type() == constants.TYPE_DB_SIGNATURE {
			dbs := v.(*adminBlock.DBSignatureEntry)
			if foundSigs[dbs.IdentityAdminChainID.String()] != "" {
				return fmt.Errorf("Found duplicate entry for ChainID %v", dbs.IdentityAdminChainID.String())
			}
			foundSigs[dbs.IdentityAdminChainID.String()] = "ok"
			pub := dbs.PrevDBSig.Pub

			if bs.IdentityChains[dbs.IdentityAdminChainID.String()] != pub.String() {
				return fmt.Errorf("Invalid Public Key in DBSignatureEntry %v - expected %v, got %v", v.Hash().String(), bs.IdentityChains[dbs.IdentityAdminChainID.String()], pub.String())
			}

			if dbs.PrevDBSig.Verify(prevHeader) == false {
				return fmt.Errorf("Invalid signature in DBSignatureEntry %v", v.Hash().String())
			}
		}
	}
	if len(foundSigs) != len(bs.IdentityChains) {
		return fmt.Errorf("Invalid number of DBSignatureEntries found in aBlock %v", aBlock.DatabasePrimaryIndex().String())
	}
	return nil
}

func (bs *BlockchainState) ProcessABlockEntry(entry interfaces.IABEntry) error {
	switch entry.Type() {
	case constants.TYPE_REVEAL_MATRYOSHKA:
		return bs.RevealMatryoshkaHash(entry)
	case constants.TYPE_ADD_MATRYOSHKA:
		return bs.AddReplaceMatryoshkaHash(entry)
	case constants.TYPE_ADD_SERVER_COUNT:
		return bs.IncreaseServerCount(entry)
	case constants.TYPE_ADD_FED_SERVER:
		//return bs.AddFederatedServer(entry)
	case constants.TYPE_ADD_AUDIT_SERVER:
		//return bs.AddAuditServer(entry)
	case constants.TYPE_REMOVE_FED_SERVER:
		//return bs.RemoveFederatedServer(entry)
	case constants.TYPE_ADD_FED_SERVER_KEY:
		//return bs.AddFederatedServerSigningKey(entry)
	case constants.TYPE_ADD_BTC_ANCHOR_KEY:
		//return bs.AddFederatedServerBitcoinAnchorKey(entry)
	}
	return nil
}

func (bs *BlockchainState) RevealMatryoshkaHash(entry interfaces.IABEntry) error {
	//e:=entry.(*adminBlock.RevealMatryoshkaHash)
	// Does nothing for authority right now
	return nil
}

func (bs *BlockchainState) AddReplaceMatryoshkaHash(entry interfaces.IABEntry) error {
	e := entry.(*adminBlock.AddReplaceMatryoshkaHash)

	auth := bs.IdentityManager.GetAuthority(e.IdentityChainID)
	if auth == nil {
		return fmt.Errorf("Authority %v not found", e.IdentityChainID.String())
	}
	auth.MatryoshkaHash = e.MHash
	bs.IdentityManager.SetAuthority(e.IdentityChainID, auth)

	return nil
}

func (bs *BlockchainState) IncreaseServerCount(entry interfaces.IABEntry) error {
	e := entry.(*adminBlock.IncreaseServerCount)
	bs.IdentityManager.AuthorityServerCount = bs.IdentityManager.AuthorityServerCount + int(e.Amount)
	return nil
}

/*
func (bs *BlockchainState) AddFederatedServer(entry interfaces.IABEntry) error {
	e := entry.(*adminBlock.AddFederatedServer)

	err := bs.AddIdentityFromChainID(e.IdentityChainID)
	if err != nil {
		//fmt.Println("Error when Making Identity,", err)
	}
	bs.Authorities[e.IdentityChainID.String()].Status = constants.IDENTITY_FEDERATED_SERVER
	// check Identity status
	bs.UpdateIdentityStatus(e.IdentityChainID, constants.IDENTITY_FEDERATED_SERVER)
	return nil
}
func (bs *BlockchainState) AddAuditServer(entry interfaces.IABEntry) error {
	e := entry.(*adminBlock.AddAuditServer)

	err := bs.AddIdentityFromChainID(e.IdentityChainID)
	if err != nil {
		//fmt.Println("Error when Making Identity,", err)
	}
	bs.Authorities[e.IdentityChainID.String()].Status = constants.IDENTITY_AUDIT_SERVER
	// check Identity status
	bs.UpdateIdentityStatus(e.IdentityChainID, constants.IDENTITY_AUDIT_SERVER)
	return nil
}
func (bs *BlockchainState) RemoveFederatedServer(entry interfaces.IABEntry) error {
	e := entry.(*adminBlock.RemoveFederatedServer)

	AuthorityIndex = bs.isAuthorityChain(e.IdentityChainID)
	if AuthorityIndex == -1 {
		log.Println(e.IdentityChainID.String() + " Cannot be removed.  Not in Authorities Libs.")
	} else {
		bs.RemoveAuthority(e.IdentityChainID)
		IdentityIndex := bs.isIdentityChain(e.IdentityChainID)
		if IdentityIndex != -1 && IdentityIndex < len(bs.Identities) {
			if bs.Identities[IdentityIndex].IdentityChainID.IsSameAs(bs.GetNetworkSkeletonIdentity()) {
				bs.Identities[IdentityIndex].Status = constants.IDENTITY_SKELETON
			} else {
				bs.removeIdentity(IdentityIndex)
			}
		}
	}
	return nil
}
func (bs *BlockchainState) AddFederatedServerSigningKey(entry interfaces.IABEntry) error {
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
func (bs *BlockchainState) AddFederatedServerBitcoinAnchorKey(entry interfaces.IABEntry) error {
	e := entry.(*adminBlock.AddFederatedServerBitcoinAnchorKey)

	pubKey, err := e.ECDSAPublicKey.MarshalBinary()
	if err != nil {
		return err
	}
	registerAuthAnchor(e.IdentityChainID, pubKey, e.KeyType, e.KeyPriority, am, "BTC")
	return nil
}

/*
func (bs *BlockchainState) GetAuthorityServerType(chainID interfaces.IHash) int { // 0 = Federated, 1 = Audit
	auth := bs.Authorities[chainID.String()]
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
*/
/*
func (bs *BlockchainState) AddAuthorityFromChainID(chainID interfaces.IHash) int {
	IdentityIndex := bs.isIdentityChain(chainID)
	if IdentityIndex == -1 {
		bs.AddIdentityFromChainID(chainID)
	}
	AuthorityIndex := bs.isAuthorityChain(chainID)
	if AuthorityIndex == -1 {
		AuthorityIndex = bs.createAuthority(chainID)
	}
	return AuthorityIndex
}

func (bs *BlockchainState) RemoveAuthority(chainID interfaces.IHash) bool {
	_, ok := bs.Authorities[chainID.String()]
	if ok == false {
		return false
	}
	delete(bs.Authorities, chainID.String())
	return true
}
*/

/*
// Checks the signature of a message. Returns an int based on who signed it:
// 			1  -> Federated Signature
//			0  -> Audit Signature
//			-1 -> Neither Fed or Audit Signature
func (bs *BlockchainState) VerifyAuthoritySignature(msg []byte, sig *[constants.SIGNATURE_LENGTH]byte, dbheight uint32) (int, error) {
	feds := bs.GetFedServers(dbheight)
	if feds == nil {
		return 0, fmt.Errorf("Federated Servers are unknown at directory block hieght %d", dbheight)
	}
	auds := bs.GetAuditServers(dbheight)

	for _, fed := range feds {
		auth, _ := bs.GetAuthority(fed.GetChainID())
		if auth == nil {
			continue
		}
		valid, err := auth.VerifySignature(msg, sig)
		if err == nil && valid {
			return 1, nil
		}
	}

	for _, aud := range auds {
		auth, _ := bs.GetAuthority(aud.GetChainID())
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
func (bs *BlockchainState) FastVerifyAuthoritySignature(msg []byte, sig interfaces.IFullSignature, dbheight uint32) (int, error) {
	feds := bs.GetFedServers(dbheight)
	if feds == nil {
		return 0, fmt.Errorf("Federated Servers are unknown at directory block hieght %d", dbheight)
	}
	auds := bs.GetAuditServers(dbheight)

	for _, fed := range feds {
		auth, _ := bs.GetAuthority(fed.GetChainID())
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
		auth, _ := bs.GetAuthority(aud.GetChainID())
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
func (bs *BlockchainState) GetAuthority(serverID interfaces.IHash) (*Authority, int) {
	auth, ok := bs.Authorities[serverID.String()]
	if ok == false {
		return nil, -2
	}
	return auth, auth.Type()
}

// We keep a 1 block history of their keys, this is so if we change their
func (bs *BlockchainState) UpdateAuthSigningKeys(height uint32) {
	/*for index, auth := range bs.Authorities {
		for _, key := range auth.KeyHistory {
			if key.ActiveDBHeight <= height {
				if len(bs.Authorities[index].KeyHistory) == 1 {
					bs.Authorities[index].KeyHistory = nil
				} else {
					bs.Authorities[index].KeyHistory = bs.Authorities[index].KeyHistory[1:]
				}
			}
		}
	}*/ /*
	bs.RepairAuthorities()
}

/*
// If the Identity failed to create, it will be fixed here
func (bs *BlockchainState) RepairAuthorities() {
	// Fix any missing management chains
	for k, auth := range bs.Authorities {
		if bs.Authorities[k].ManagementChainID == nil {
			idID := bs.Authorities[k].AuthorityChainID.String()
			identity := bs.Identities[idID]
			if identity == nil {
				err := bs.AddIdentityFromChainID(auth.AuthorityChainID)
				if err != nil {
					continue
				}
			}
			bs.Authorities[k].ManagementChainID = bs.Identities[idID].ManagementChainID
			bs.Identities[idID].Status = bs.Authorities[k].Status
		}
	}

	// Fix any missing keys
	for _, id := range bs.Identities {
		if !id.IsFull() {
			bs.FixMissingKeys(id)
		}
	}
}

/*
func registerAuthAnchor(chainID interfaces.IHash, signingKey []byte, keyType byte, keyLevel byte, st *State, BlockChain string) {
	AuthorityIndex := bs.AddAuthorityFromChainID(chainID)
	var oneASK AnchorSigningKey

	ask := bs.Authorities[AuthorityIndex].AnchorKeys
	newASK := make([]AnchorSigningKey, len(ask)+1)

	for i := 0; i < len(ask); i++ {
		newASK[i] = ask[i]
	}

	oneASK.BlockChain = BlockChain
	oneASK.KeyLevel = keyLevel
	oneASK.KeyType = keyType
	oneASK.SigningKey = signingKey

	newASK[len(ask)] = oneASK
	bs.Authorities[AuthorityIndex].AnchorKeys = newASK
}

func addServerSigningKey(chainID interfaces.IHash, key interfaces.IHash, height uint32, st *State) {
	AuthorityIndex := bs.AddAuthorityFromChainID(chainID)
	if bs.IdentityChainID.IsSameAs(chainID) && len(bs.serverPendingPrivKeys) > 0 {
		for i, pubKey := range bs.serverPendingPubKeys {
			pubData, err := pubKey.MarshalBinary()
			if err != nil {
				break
			}
			if bytes.Compare(pubData, key.Bytes()) == 0 {
				bs.serverPrivKey = bs.serverPendingPrivKeys[i]
				bs.serverPubKey = bs.serverPendingPubKeys[i]
				if len(bs.serverPendingPrivKeys) > i+1 {
					bs.serverPendingPrivKeys = append(bs.serverPendingPrivKeys[:i], bs.serverPendingPrivKeys[i+1:]...)
					bs.serverPendingPubKeys = append(bs.serverPendingPubKeys[:i], bs.serverPendingPubKeys[i+1:]...)
				} else {
					bs.serverPendingPrivKeys = bs.serverPendingPrivKeys[:i]
					bs.serverPendingPubKeys = bs.serverPendingPubKeys[:i]
				}
				break
			}
		}
	}
	// Add Key History
	bs.Authorities[AuthorityIndex].KeyHistory = append(bs.Authorities[AuthorityIndex].KeyHistory, struct {
		ActiveDBHeight uint32
		SigningKey     primitives.PublicKey
	}{height, bs.Authorities[AuthorityIndex].SigningKey})
	// Replace Active Key
	bs.Authorities[AuthorityIndex].SigningKey = primitives.PubKeyFromString(key.String())
}
*/
/*
func (bs *BlockchainState) UpdateIdentityStatus(chainID interfaces.IHash, statusTo int) {
	_, ok := bs.Identities[chainID.String()]
	if ok == false {
		return
	}
	bs.Identities[chainID.String()].Status = statusTo
}

func (bs *BlockchainState) RemoveIdentity(chainID interfaces.IHash) {
	index := st.isIdentityChain(chainID)
	st.removeIdentity(index)
}

func (bs *BlockchainState) removeIdentity(i int) {
	if st.Identities[i].Status == constants.IDENTITY_SKELETON {
		return // Do not remove skeleton identity
	}
	st.Identities = append(st.Identities[:i], st.Identities[i+1:]...)
}*/
/*
func (bs *BlockchainState) isIdentityChain(cid interfaces.IHash) int {
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
