// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identity

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sync"

	"sort"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type IdentityManager struct {
	Mutex sync.RWMutex
	IdentityManagerWithoutMutex
}

type IdentityManagerWithoutMutex struct {
	Authorities          map[[32]byte]*Authority
	Identities           map[[32]byte]*Identity
	AuthorityServerCount int

	OldEntries []*OldEntry
}

func NewIdentityManager() *IdentityManager {
	im := new(IdentityManager)
	im.Authorities = make(map[[32]byte]*Authority)
	im.Identities = make(map[[32]byte]*Identity)
	return im
}

func (im *IdentityManager) Clone() *IdentityManager {
	b := NewIdentityManager()
	for k, v := range im.Authorities {
		b.Authorities[k] = v.Clone()
	}
	for k, v := range im.Identities {
		b.Identities[k] = v.Clone()
	}

	b.AuthorityServerCount = im.AuthorityServerCount
	for k, v := range im.OldEntries {
		copy := *v
		b.OldEntries[k] = &copy
	}

	return b
}

func (im *IdentityManager) SetBootstrapIdentity(id interfaces.IHash, key interfaces.IHash) error {
	// Identity
	i := NewIdentity()
	i.IdentityChainID = id
	i.SigningKey = key
	i.Status = constants.IDENTITY_FEDERATED_SERVER
	im.SetIdentity(id, i)

	// Authority
	auth := NewAuthority()
	auth.AuthorityChainID = id

	var pub primitives.PublicKey
	pub = key.Fixed()
	auth.SigningKey = pub
	auth.Status = constants.IDENTITY_FEDERATED_SERVER
	auth.ManagementChainID, _ = primitives.HexToHash("88888800000000000000000000000000")

	im.SetAuthority(auth.AuthorityChainID, auth)
	return nil
}

func (im *IdentityManager) SetSkeletonIdentity(chain interfaces.IHash) error {
	// Skeleton is in the identity list
	//	The key comes from the blockchain, and must be parsed
	id := NewIdentity()
	id.IdentityChainID = chain
	id.Status = constants.IDENTITY_SKELETON

	im.SetIdentity(chain, id)
	return nil

	// Skeleton is not an authority
	/*
		auth := new(Authority)
		err := auth.SigningKey.UnmarshalText([]byte(key))
		if err != nil {
			return err
		}
		auth.Status = constants.IDENTITY_FEDERATED_SERVER

		im.SetAuthority(primitives.NewZeroHash(), auth)
		return nil*/
}

//func (im *IdentityManager) SetSkeletonKeyMainNet() error {
//	//Skeleton key:
//	//"0000000000000000000000000000000000000000000000000000000000000000":"0426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a"
//
//	return im.SetSkeletonKey("0426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a")
//}

func (im *IdentityManager) FedServerCount() int {
	im.Mutex.RLock()
	defer im.Mutex.RUnlock()
	answer := 0
	for _, v := range im.Authorities {
		if v.Type() == int(constants.IDENTITY_FEDERATED_SERVER) {
			answer++
		}
	}
	return answer
}

func (im *IdentityManager) AuditServerCount() int {
	im.Mutex.RLock()
	defer im.Mutex.RUnlock()
	answer := 0
	for _, v := range im.Authorities {
		if v.Type() == int(constants.IDENTITY_AUDIT_SERVER) {
			answer++
		}
	}
	return answer
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
	// Do nothing, it used to init the maps if they were empty, but we init the Identity control with non-empty maps
}

func (im *IdentityManager) SetIdentity(chainID interfaces.IHash, id *Identity) {
	im.Init()
	im.Mutex.Lock()
	defer im.Mutex.Unlock()
	c := chainID.String()
	var _ = c
	im.Identities[chainID.Fixed()] = id
}

func (im *IdentityManager) RemoveIdentity(chainID interfaces.IHash) bool {
	im.Init()
	im.Mutex.Lock()
	defer im.Mutex.Unlock()
	_, ok := im.Identities[chainID.Fixed()]
	if ok == false {
		return false
	}
	delete(im.Identities, chainID.Fixed())
	return true
}

func (im *IdentityManager) GetIdentity(chainID interfaces.IHash) *Identity {
	im.Init()
	im.Mutex.RLock()
	defer im.Mutex.RUnlock()
	// First check identity chain ids
	id := im.Identities[chainID.Fixed()]
	if id != nil {
		return id
	}

	// Then check management chains
	for _, id := range im.Identities {
		if id.ManagementChainID.IsSameAs(chainID) {
			return id
		}
	}

	return nil
}

func (im *IdentityManager) GetSortedIdentities() []*Identity {
	list := im.GetIdentities()
	sort.Sort(IdentitySort(list))
	return list

}

func (im *IdentityManager) GetSortedAuthorities() []interfaces.IAuthority {
	list := im.GetAuthorities()
	sort.Sort(AuthoritySort(list))
	return list

}

func (im *IdentityManager) GetIdentities() []*Identity {
	im.Mutex.RLock()
	defer im.Mutex.RUnlock()

	ids := make([]*Identity, 0)
	for _, id := range im.Identities {
		ids = append(ids, id)
	}

	return ids
}

func (im *IdentityManager) SetAuthority(chainID interfaces.IHash, auth *Authority) {
	im.Init()
	im.Mutex.Lock()
	defer im.Mutex.Unlock()
	im.Authorities[chainID.Fixed()] = auth
}

func (im *IdentityManager) RemoveAuthority(chainID interfaces.IHash) bool {
	im.Init()
	_, ok := im.Authorities[chainID.Fixed()]
	if !ok {
		return false
	}

	delete(im.Authorities, chainID.Fixed())
	return true
}

func (im *IdentityManager) GetAuthorities() []interfaces.IAuthority {
	im.Mutex.RLock()
	defer im.Mutex.RUnlock()

	auths := make([]interfaces.IAuthority, 0)
	for _, auth := range im.Authorities {
		auths = append(auths, auth)
	}

	return auths
}

func (im *IdentityManager) GetAuthority(chainID interfaces.IHash) *Authority {
	im.Init()
	im.Mutex.RLock()
	defer im.Mutex.RUnlock()
	return im.Authorities[chainID.Fixed()]
}

type OldEntry struct {
	EntryBinary     []byte
	DBlockHeight    uint32
	DBlockTimestamp uint64
}

func (im *IdentityManager) PushEntryForLater(entry interfaces.IEBEntry, dBlockHeight uint32, dBlockTimestamp interfaces.Timestamp) error {
	oe := new(OldEntry)
	b, err := entry.MarshalBinary()
	if err != nil {
		return err
	}
	oe.EntryBinary = b
	oe.DBlockHeight = dBlockHeight
	oe.DBlockTimestamp = dBlockTimestamp.GetTimeMilliUInt64()

	im.OldEntries = append(im.OldEntries, oe)
	return nil
}

func (im *IdentityManager) ProcessOldEntries() (bool, error) {
	var change bool
	for {
		allErrors := true
		for i := 0; i < len(im.OldEntries); i++ {
			oe := im.OldEntries[i]
			entry := new(entryBlock.Entry)
			err := entry.UnmarshalBinary(oe.EntryBinary)
			if err != nil {
				return false, err
			}
			t := primitives.NewTimestampFromMilliseconds(oe.DBlockTimestamp)
			change, err = im.ProcessIdentityEntry(entry, oe.DBlockHeight, t, false)
			if err == nil {
				//entry has been finally processed, now we can delete it
				allErrors = false
				im.OldEntries = append(im.OldEntries[:i], im.OldEntries[i+1:]...)
				i--
			}
		}
		//loop over and over until no entries have been removed in a whole loop
		if allErrors == true {
			return change, nil
		}
	}
	return change, nil
}

func (im *IdentityManager) CheckDBSignatureEntries(aBlock interfaces.IAdminBlock, dBlock interfaces.IDirectoryBlock, prevHeader []byte) error {
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
			pub := dbs.PrevDBSig.Pub
			signingKey := ""

			auth := im.GetAuthority(dbs.IdentityAdminChainID)
			signingKey = auth.SigningKey.String()

			if signingKey != pub.String() {
				return fmt.Errorf("Invalid Public Key in DBSignatureEntry %v - expected %v, got %v", v.Hash().String(), signingKey, pub.String())
			}

			if dbs.PrevDBSig.Verify(prevHeader) == false {
				//return fmt.Errorf("Invalid signature in DBSignatureEntry %v", v.Hash().String())
			} else {
				foundSigs[dbs.IdentityAdminChainID.String()] = "ok"
			}
		}
	}
	fedServerCount := im.FedServerCount()
	if len(foundSigs) < fedServerCount/2 {
		return fmt.Errorf("Invalid number of DBSignatureEntries found in aBlock %v - %v vs %v", aBlock.DatabasePrimaryIndex().String(), len(foundSigs), fedServerCount)
	}
	return nil
}
