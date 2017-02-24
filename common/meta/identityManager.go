// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package meta

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sync"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type IdentityManager struct {
	Mutex sync.RWMutex
	IdentityManagerWithoutMutex
}

type IdentityManagerWithoutMutex struct {
	Authorities          map[string]*Authority
	Identities           map[string]*Identity
	AuthorityServerCount int
}

func (im *IdentityManager) GobDecode(data []byte) error {
	//Circumventing Gob's "gob: type sync.RWMutex has no exported fields"
	b := bytes.NewBuffer(data)
	dec := gob.NewDecoder(b)
	return dec.Decode(&im.IdentityManagerWithoutMutex)
}

func (im *IdentityManager) GobEncode() ([]byte, error) {
	//Circumventing Gob's "gob: type sync.RWMutex has no exported fields"
	b := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(b)
	err := enc.Encode(im.IdentityManagerWithoutMutex)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (im *IdentityManager) Init() {
	im.Mutex.Lock()
	defer im.Mutex.Unlock()
	if im.Authorities == nil {
		im.Authorities = map[string]*Authority{}
	}
	if im.Identities == nil {
		im.Identities = map[string]*Identity{}
	}
}

func (im *IdentityManager) SetIdentity(chainID interfaces.IHash, id *Identity) {
	im.Init()
	im.Mutex.Lock()
	defer im.Mutex.Unlock()
	im.Identities[chainID.String()] = id
}

func (im *IdentityManager) RemoveIdentity(chainID interfaces.IHash) bool {
	im.Init()
	im.Mutex.Lock()
	defer im.Mutex.Unlock()
	_, ok := im.Identities[chainID.String()]
	if ok == false {
		return false
	}
	delete(im.Identities, chainID.String())
	return true
}

func (im *IdentityManager) GetIdentity(chainID interfaces.IHash) *Identity {
	im.Init()
	im.Mutex.RLock()
	defer im.Mutex.RUnlock()
	return im.Identities[chainID.String()]
}

func (im *IdentityManager) SetAuthority(chainID interfaces.IHash, auth *Authority) {
	im.Init()
	im.Mutex.Lock()
	defer im.Mutex.Unlock()
	im.Authorities[chainID.String()] = auth
}

func (im *IdentityManager) RemoveAuthority(chainID interfaces.IHash) bool {
	im.Init()
	im.Mutex.Lock()
	defer im.Mutex.Unlock()
	_, ok := im.Authorities[chainID.String()]
	if ok == false {
		return false
	}
	delete(im.Authorities, chainID.String())
	return true
}

func (im *IdentityManager) GetAuthority(chainID interfaces.IHash) *Authority {
	im.Init()
	im.Mutex.RLock()
	defer im.Mutex.RUnlock()
	return im.Authorities[chainID.String()]
}

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

	im.SetAuthority(chainID, newAuth)
}

func (im *IdentityManager) ApplyIdentityChainStructure(ic *IdentityChainStructure, chainID interfaces.IHash, dBlockHeight uint32) error {
	id := im.GetIdentity(chainID)
	if id != nil {
		return fmt.Errorf("ChainID already exists! %v", chainID.String())
	}

	id = new(Identity)

	id.Key1 = ic.Key1
	id.Key2 = ic.Key2
	id.Key3 = ic.Key3
	id.Key4 = ic.Key4

	id.IdentityCreated = dBlockHeight

	id.IdentityChainID = chainID

	im.SetIdentity(chainID, id)
	return nil
}

func (im *IdentityManager) ApplyNewBitcoinKeyStructure(bnk *NewBitcoinKeyStructure, subChainID interfaces.IHash, BlockChain string, dBlockTimestamp interfaces.Timestamp) error {
	chainID := bnk.RootIdentityChainID

	id := im.GetIdentity(chainID)
	if id == nil {
		return fmt.Errorf("ChainID doesn't exists! %v", chainID.String())
	}
	err := bnk.VerifySignature(id.Key1)
	if err != nil {
		return err
	}

	if id.ManagementChainID.IsSameAs(subChainID) == false {
		return fmt.Errorf("Identity Error: Entry was not placed in the correct management chain - %v vs %v", subChainID.String(), id.ManagementChainID.String())
	}

	var ask []AnchorSigningKey
	var newAsk []AnchorSigningKey
	var oneAsk AnchorSigningKey

	ask = id.AnchorKeys
	newAsk = make([]AnchorSigningKey, len(ask)+1)

	oneAsk.BlockChain = BlockChain
	oneAsk.KeyLevel = bnk.BitcoinKeyLevel
	oneAsk.KeyType = bnk.KeyType
	oneAsk.SigningKey = bnk.NewKey

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

	if contains {
		id.AnchorKeys = ask
	} else {
		id.AnchorKeys = newAsk
	}

	/*
		dbase := st.GetAndLockDB()
		dblk, err := dbase.FetchDBlockByHeight(height)
		st.UnlockDB()
		if err == nil && dblk != nil && dblk.GetHeader().GetTimestamp().GetTimeSeconds() != 0 {
			if !CheckTimestamp(extIDs[6], dblk.GetHeader().GetTimestamp().GetTimeSeconds()) {
				return errors.New("New Anchor key for identity [" + chainID.String()[:10] + "] timestamp is too old")
			}
		} else {
			if !CheckTimestamp(extIDs[6], st.GetTimestamp().GetTimeSeconds()) {
				return errors.New("New Anchor key for identity [" + chainID.String()[:10] + "] timestamp is too old")
			}
		}
	*/

	/*
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
	*/

	im.SetIdentity(chainID, id)
	return nil
}

func (im *IdentityManager) ApplyNewBlockSigningKeyStruct(nbsk *NewBlockSigningKeyStruct, subchainID interfaces.IHash, dBlockTimestamp interfaces.Timestamp) error {
	chainID := nbsk.RootIdentityChainID
	id := im.GetIdentity(chainID)
	if id == nil {
		return fmt.Errorf("ChainID doesn't exists! %v", nbsk.RootIdentityChainID.String())
	}
	err := nbsk.VerifySignature(id.Key1)
	if err != nil {
		return err
	}

	if id.ManagementChainID.IsSameAs(subchainID) == false {
		return fmt.Errorf("Identity Error: Entry was not placed in the correct management chain - %v vs %v", id.ManagementChainID.String(), subchainID.String())
	}

	//Check Timestamp??

	key := primitives.NewZeroHash()
	err = key.UnmarshalBinary(nbsk.NewPublicKey)
	if err != nil {
		return err
	}
	id.SigningKey = key

	im.SetIdentity(nbsk.RootIdentityChainID, id)
	return nil

	/*

		func RegisterBlockSigningKey(entry interfaces.IEBEntry, initial bool, height uint32, st *State) error {





		dbase := st.GetAndLockDB()
		dblk, err := dbase.FetchDBlockByHeight(height)
		st.UnlockDB()

		if err == nil && dblk != nil && dblk.GetHeader().GetTimestamp().GetTimeSeconds() != 0 {
			if !CheckTimestamp(extIDs[4], dblk.GetHeader().GetTimestamp().GetTimeSeconds()) {
				return errors.New("New Block Signing key for identity  [" + chainID.String()[:10] + "] timestamp is too old")
			}
		} else {
			if !CheckTimestamp(extIDs[4], st.GetTimestamp().GetTimeSeconds()) {
				return errors.New("New Block Signing key for identity  [" + chainID.String()[:10] + "] timestamp is too old")
			}
		}
	*/

	/*
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

	*/
}

func (im *IdentityManager) ApplyNewMatryoshkaHashStructure(nmh *NewMatryoshkaHashStructure, dBlockTimestamp interfaces.Timestamp) error {
	id := im.GetIdentity(nmh.RootIdentityChainID)
	if id == nil {
		return fmt.Errorf("ChainID doesn't exists! %v", nmh.RootIdentityChainID.String())
	}
	err := nmh.VerifySignature(id.Key1)
	if err != nil {
		return err
	}
	//Check Timestamp??

	id.MatryoshkaHash = nmh.OutermostMHash

	im.SetIdentity(nmh.RootIdentityChainID, id)
	return nil
}

func (im *IdentityManager) ApplyRegisterFactomIdentityStructure(rfi *RegisterFactomIdentityStructure, dBlockHeight uint32) error {
	id := im.GetIdentity(rfi.IdentityChainID)
	if id == nil {
		return fmt.Errorf("ChainID doesn't exists! %v", rfi.IdentityChainID.String())
	}

	err := rfi.VerifySignature(id.Key1)
	if err != nil {
		return err
	}

	id.ManagementRegistered = dBlockHeight

	im.SetIdentity(id.IdentityChainID, id)
	return nil
}

func (im *IdentityManager) ApplyRegisterServerManagementStructure(rsm *RegisterServerManagementStructure, chainID interfaces.IHash, dBlockHeight uint32) error {
	id := im.GetIdentity(chainID)
	if id == nil {
		return fmt.Errorf("ChainID doesn't exists! %v", chainID.String())
	}

	err := rfi.VerifySignature(id.Key1)
	if err != nil {
		return err
	}

	if id.ManagementRegistered == 0 {
		id.ManagementRegistered = dBlockHeight
	}
	id.ManagementChainID = rsm.SubchainChainID

	im.SetIdentity(id.IdentityChainID, id)
	return nil
}

func (im *IdentityManager) ApplyServerManagementStructure(sm *ServerManagementStructure, chainID interfaces.IHash, dBlockHeight uint32) error {
	id := im.GetIdentity(chainID)
	if id == nil {
		return fmt.Errorf("ChainID doesn't exists! %v", chainID.String())
	}

	id.ManagementCreated = dBlockHeight

	im.SetIdentity(chainID, id)
	return nil
}
