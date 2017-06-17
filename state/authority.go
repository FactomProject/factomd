// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"bytes"
	"encoding/json"
	"fmt"

	ed "github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
	"github.com/FactomProject/factomd/log"
)

type HistoricKey struct {
	ActiveDBHeight uint32
	SigningKey     primitives.PublicKey
}

var _ interfaces.BinaryMarshallable = (*HistoricKey)(nil)

func RandomHistoricKey() *HistoricKey {
	hk := new(HistoricKey)

	hk.ActiveDBHeight = random.RandUInt32()
	hk.SigningKey = *primitives.RandomPrivateKey().Pub

	return hk
}

func (e *HistoricKey) IsSameAs(b *HistoricKey) bool {
	if e.ActiveDBHeight != b.ActiveDBHeight {
		return false
	}
	if e.SigningKey.IsSameAs(&b.SigningKey) == false {
		return false
	}

	return true
}

func (e *HistoricKey) MarshalBinary() ([]byte, error) {
	buf := primitives.NewBuffer(nil)

	err := buf.PushUInt32(e.ActiveDBHeight)
	if err != nil {
		return nil, err
	}

	err = buf.PushBinaryMarshallable(&e.SigningKey)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (e *HistoricKey) UnmarshalBinaryData(p []byte) (newData []byte, err error) {
	newData = p
	buf := primitives.NewBuffer(p)

	e.ActiveDBHeight, err = buf.PopUInt32()
	if err != nil {
		return
	}

	err = buf.PopBinaryMarshallable(&e.SigningKey)
	if err != nil {
		return
	}

	newData = buf.DeepCopyBytes()
	return
}

func (e *HistoricKey) UnmarshalBinary(p []byte) error {
	_, err := e.UnmarshalBinaryData(p)
	return err
}

type Authority struct {
	AuthorityChainID  interfaces.IHash
	ManagementChainID interfaces.IHash
	MatryoshkaHash    interfaces.IHash
	SigningKey        primitives.PublicKey
	Status            uint8
	AnchorKeys        []AnchorSigningKey

	KeyHistory []HistoricKey
}

var _ interfaces.BinaryMarshallable = (*Authority)(nil)

func RandomAuthority() *Authority {
	a := new(Authority)

	a.AuthorityChainID = primitives.RandomHash()
	a.ManagementChainID = primitives.RandomHash()
	a.MatryoshkaHash = primitives.RandomHash()

	a.SigningKey = *primitives.RandomPrivateKey().Pub
	a.Status = random.RandUInt8()

	l := random.RandIntBetween(1, 10)
	for i := 0; i < l; i++ {
		a.AnchorKeys = append(a.AnchorKeys, *RandomAnchorSigningKey())
	}

	l = random.RandIntBetween(1, 10)
	for i := 0; i < l; i++ {
		a.KeyHistory = append(a.KeyHistory, *RandomHistoricKey())
	}

	return a
}

func (e *Authority) IsSameAs(b *Authority) bool {
	if e.AuthorityChainID.IsSameAs(b.AuthorityChainID) == false {
		return false
	}
	if e.ManagementChainID.IsSameAs(b.ManagementChainID) == false {
		return false
	}
	if e.MatryoshkaHash.IsSameAs(b.MatryoshkaHash) == false {
		return false
	}
	if e.SigningKey.IsSameAs(&b.SigningKey) == false {
		return false
	}
	if e.Status != b.Status {
		return false
	}

	if len(e.AnchorKeys) != len(b.AnchorKeys) {
		return false
	}
	for i := range e.AnchorKeys {
		if e.AnchorKeys[i].IsSameAs(&b.AnchorKeys[i]) == false {
			return false
		}
	}
	if len(e.KeyHistory) != len(b.KeyHistory) {
		return false
	}
	for i := range e.KeyHistory {
		if e.KeyHistory[i].IsSameAs(&b.KeyHistory[i]) == false {
			return false
		}
	}

	return true
}

func (e *Authority) Init() {
	if e.AuthorityChainID == nil {
		e.AuthorityChainID = primitives.NewZeroHash()
	}
	if e.ManagementChainID == nil {
		e.ManagementChainID = primitives.NewZeroHash()
	}
	if e.MatryoshkaHash == nil {
		e.MatryoshkaHash = primitives.NewZeroHash()
	}
}

func (e *Authority) MarshalBinary() ([]byte, error) {
	e.Init()
	buf := primitives.NewBuffer(nil)

	err := buf.PushBinaryMarshallable(e.AuthorityChainID)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(e.ManagementChainID)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(e.MatryoshkaHash)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(&e.SigningKey)
	if err != nil {
		return nil, err
	}
	err = buf.PushByte(byte(e.Status))
	if err != nil {
		return nil, err
	}

	l := len(e.AnchorKeys)
	err = buf.PushVarInt(uint64(l))
	if err != nil {
		return nil, err
	}
	for _, v := range e.AnchorKeys {
		err = buf.PushBinaryMarshallable(&v)
		if err != nil {
			return nil, err
		}
	}

	l = len(e.KeyHistory)
	err = buf.PushVarInt(uint64(l))
	if err != nil {
		return nil, err
	}
	for _, v := range e.KeyHistory {
		err = buf.PushBinaryMarshallable(&v)
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), nil
}

func (e *Authority) UnmarshalBinaryData(p []byte) (newData []byte, err error) {
	e.Init()
	newData = p
	buf := primitives.NewBuffer(p)

	err = buf.PopBinaryMarshallable(e.AuthorityChainID)
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(e.ManagementChainID)
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(e.MatryoshkaHash)
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(&e.SigningKey)
	if err != nil {
		return
	}
	status, err := buf.PopByte()
	if err != nil {
		return
	}
	e.Status = uint8(status)

	l, err := buf.PopVarInt()
	if err != nil {
		return
	}
	for i := 0; i < int(l); i++ {
		var ask AnchorSigningKey
		err = buf.PopBinaryMarshallable(&ask)
		if err != nil {
			return
		}
		e.AnchorKeys = append(e.AnchorKeys, ask)
	}

	l, err = buf.PopVarInt()
	if err != nil {
		return
	}
	for i := 0; i < int(l); i++ {
		var hk HistoricKey
		err = buf.PopBinaryMarshallable(&hk)
		if err != nil {
			return
		}
		e.KeyHistory = append(e.KeyHistory, hk)
	}

	newData = buf.DeepCopyBytes()
	return
}

func (e *Authority) UnmarshalBinary(p []byte) error {
	_, err := e.UnmarshalBinaryData(p)
	return err
}

func (auth *Authority) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		AuthorityChainID  interfaces.IHash   `json:"chainid"`
		ManagementChainID interfaces.IHash   `json:"manageid"`
		MatryoshkaHash    interfaces.IHash   `json:"matroyshka"`
		SigningKey        string             `json:"signingkey"`
		Status            string             `json:"status"`
		AnchorKeys        []AnchorSigningKey `json:"anchorkeys"`
	}{
		AuthorityChainID:  auth.AuthorityChainID,
		ManagementChainID: auth.ManagementChainID,
		MatryoshkaHash:    auth.MatryoshkaHash,
		SigningKey:        auth.SigningKey.String(),
		Status:            statusToJSONString(auth.Status),
		AnchorKeys:        auth.AnchorKeys,
	})
}

// 1 if fed, 0 if audit, -1 if neither
func (auth *Authority) Type() int {
	if auth.Status == constants.IDENTITY_FEDERATED_SERVER {
		return 1
	} else if auth.Status == constants.IDENTITY_AUDIT_SERVER {
		return 0
	}
	return -1
}

func (auth *Authority) VerifySignature(msg []byte, sig *[constants.SIGNATURE_LENGTH]byte) (bool, error) {
	//return true, nil // Testing
	var pub [32]byte
	tmp, err := auth.SigningKey.MarshalBinary()
	if err != nil {
		return false, err
	} else {
		copy(pub[:], tmp)
		valid := ed.VerifyCanonical(&pub, msg, sig)
		if !valid {
			for _, histKey := range auth.KeyHistory {
				histTemp, err := histKey.SigningKey.MarshalBinary()
				if err != nil {
					continue
				}
				copy(pub[:], histTemp)
				if ed.VerifyCanonical(&pub, msg, sig) {
					return true, nil
				}
			}
		} else {
			return true, nil
		}
	}
	return false, nil
}

// Checks the signature of a message. Returns an int based on who signed it:
// 			1  -> Federated Signature
//			0  -> Audit Signature
//			-1 -> Neither Fed or Audit Signature
func (st *State) VerifyAuthoritySignature(msg []byte, sig *[constants.SIGNATURE_LENGTH]byte, dbheight uint32) (int, error) {
	feds := st.GetFedServers(dbheight)
	if feds == nil {
		return -1, fmt.Errorf("Federated Servers are unknown at directory block hieght %d", dbheight)
	}
	auds := st.GetAuditServers(dbheight)

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
		return 0, fmt.Errorf("Federated Servers are unknown at directory block hieght %d", dbheight)
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
		st.Authorities[AuthorityIndex].Status = constants.IDENTITY_FEDERATED_SERVER
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
		st.Authorities[AuthorityIndex].Status = constants.IDENTITY_AUDIT_SERVER
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
					st.Identities[IdentityIndex].Status = constants.IDENTITY_SKELETON
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
	newAuth.Status = constants.IDENTITY_PENDING_FULL

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
				s.Identities[idIndex].Status = s.Authorities[i].Status
			}
		}
	}

	// Fix any missing keys
	for _, id := range s.Identities {
		if !id.IsFull() {
			id.FixMissingKeys(s)
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
	oneASK.Key = signingKey

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
