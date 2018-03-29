// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identity

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sync"

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
	Authorities          map[string]*Authority
	Identities           map[string]*Identity
	AuthorityServerCount int

	OldEntries []*OldEntry
}

func NewIdentityManager() *IdentityManager {
	im := new(IdentityManager)
	im.Authorities = make(map[string]*Authority)
	im.Identities = make(map[string]*Identity)
	return im
}

func (im *IdentityManager) SetBootstrapIdentity(id interfaces.IHash, key interfaces.IHash) error {
	auth := NewAuthority()
	auth.AuthorityChainID = id

	var pub primitives.PublicKey
	pub = key.Fixed()
	auth.SigningKey = pub
	auth.Status = constants.IDENTITY_FEDERATED_SERVER

	im.SetAuthority(auth.AuthorityChainID, auth)
	return nil
}

func (im *IdentityManager) SetSkeletonKey(key string) error {
	auth := new(Authority)
	err := auth.SigningKey.UnmarshalText([]byte(key))
	if err != nil {
		return err
	}
	auth.Status = constants.IDENTITY_FEDERATED_SERVER

	im.SetAuthority(primitives.NewZeroHash(), auth)
	return nil
}

func (im *IdentityManager) SetSkeletonKeyMainNet() error {
	//Skeleton key:
	//"0000000000000000000000000000000000000000000000000000000000000000":"0426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a"

	return im.SetSkeletonKey("0426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a")
}

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

	auth := im.GetAuthority(chainID)
	if auth == nil {
		return false
	}
	auth.Status = constants.IDENTITY_UNASSIGNED
	im.SetAuthority(chainID, auth)

	return true
}

func (im *IdentityManager) GetAuthority(chainID interfaces.IHash) *Authority {
	im.Init()
	im.Mutex.RLock()
	defer im.Mutex.RUnlock()
	return im.Authorities[chainID.String()]
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

func (im *IdentityManager) ProcessOldEntries() error {
	for {
		allErrors := true
		for i := 0; i < len(im.OldEntries); i++ {
			oe := im.OldEntries[i]
			entry := new(entryBlock.Entry)
			err := entry.UnmarshalBinary(oe.EntryBinary)
			if err != nil {
				return err
			}
			t := primitives.NewTimestampFromMilliseconds(oe.DBlockTimestamp)
			err = im.ProcessIdentityEntry(entry, oe.DBlockHeight, t, false)
			if err == nil {
				//entry has been finally processed, now we can delete it
				allErrors = false
				im.OldEntries = append(im.OldEntries[:i], im.OldEntries[i+1:]...)
				i--
			}
		}
		//loop over and over until no entries have been removed in a whole loop
		if allErrors == true {
			return nil
		}
	}
	return nil
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
