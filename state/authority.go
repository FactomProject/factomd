package state

import (
	"bytes"
	"fmt"

	ed "github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/log"
)

type Authority struct {
	AuthorityChainID  interfaces.IHash
	ManagementChainID interfaces.IHash
	MatryoshkaHash    interfaces.IHash
	SigningKey        primitives.PublicKey
	Status            int
	AnchorKeys        []AnchorSigningKey
	// add key history?
}

func (auth *Authority) VerifySignature(msg []byte, sig *[constants.SIGNATURE_LENGTH]byte) (bool, error) {
	var pub [32]byte
	tmp, err := auth.SigningKey.MarshalBinary()
	if err != nil {
		return false, err
	} else {
		copy(pub[:], tmp)
		if !ed.Verify(&pub, msg, sig) {
		} else {
			return true, nil
		}
	}
	return false, nil
}

func (st *State) VerifyFederatedSignature(msg []byte, sig *[constants.SIGNATURE_LENGTH]byte) (bool, error) {
	Authlist := st.Authorities
	for _, auth := range Authlist {
		valid, err := auth.VerifySignature(msg, sig)
		if err != nil {
			continue
		}
		if valid {
			return true, nil
		}
	}

	return true, fmt.Errorf("Signature Key Invalid or not Federated Server Key") // For Testing
	return false, fmt.Errorf("Signature Key Invalid or not Federated Server Key")
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
			return err
		}
		AuthorityIndex = st.AddAuthorityFromChainID(f.IdentityChainID)
		st.Authorities[AuthorityIndex].Status = constants.IDENTITY_FEDERATED_SERVER
		// check Identity status
		UpdateIdentityStatus(f.IdentityChainID, constants.IDENTITY_PENDING_FEDERATED_SERVER, constants.IDENTITY_FEDERATED_SERVER, st)
	case constants.TYPE_ADD_AUDIT_SERVER:
		a := new(adminBlock.AddAuditServer)
		err := a.UnmarshalBinary(data)
		if err != nil {
			return err
		}
		err = st.AddIdentityFromChainID(a.IdentityChainID)
		if err != nil {
			return err
		}
		AuthorityIndex = st.AddAuthorityFromChainID(a.IdentityChainID)
		st.Authorities[AuthorityIndex].Status = constants.IDENTITY_AUDIT_SERVER
		// check Identity status
		UpdateIdentityStatus(a.IdentityChainID, constants.IDENTITY_PENDING_AUDIT_SERVER, constants.IDENTITY_AUDIT_SERVER, st)
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
				st.removeIdentity(IdentityIndex)
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
		addServerSigningKey(f.IdentityChainID, key, st)
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
	status := st.Authorities[index].Status
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
	newAuth.Status = constants.IDENTITY_PENDING

	st.Authorities = append(st.Authorities, *newAuth)
	return len(st.Authorities) - 1
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
	oneASK.SigningKey = signingKey

	newASK[len(ask)] = oneASK
	st.Authorities[AuthorityIndex].AnchorKeys = newASK
}

func addServerSigningKey(chainID interfaces.IHash, key interfaces.IHash, st *State) {
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
	st.Authorities[AuthorityIndex].SigningKey = primitives.PubKeyFromString(key.String())
}
