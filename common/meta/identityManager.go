// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package meta

import (
	"bytes"
	"encoding/gob"
	"sync"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
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
