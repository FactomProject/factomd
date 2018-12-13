// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identity

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math/rand"
	"sync"

	"sort"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/identityEntries"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/util/atomic"
)

type IdentityManager struct {
	Mutex sync.RWMutex
	IdentityManagerWithoutMutex
}

type IdentityManagerWithoutMutex struct {
	Authorities map[[32]byte]*Authority
	Identities  map[[32]byte]*Identity
	// All Identity Registrations.
	IdentityRegistrations   map[[32]byte]*identityEntries.RegisterFactomIdentityStructure
	MaxAuthorityServerCount int

	// Not Marshalled
	// Tracks cancellation of coinbases
	CancelManager *CoinbaseCancelManager

	// Map of all coinbase outputs that are cancelled.
	//	The map key is the block height of the DESCRIPTOR
	//	The list of ints are the indices of the outputs to be
	//	removed. The keys from the map should be deleted when the
	//	descriptor+declaration height is hit.
	//		[descriptorheight]List of cancelled outputs
	CanceledCoinbaseOutputs map[uint32][]uint32
	OldEntries              []*OldEntry
}

func NewIdentityManager() *IdentityManager {
	im := new(IdentityManager)
	im.Authorities = make(map[[32]byte]*Authority)
	im.Identities = make(map[[32]byte]*Identity)
	im.IdentityRegistrations = make(map[[32]byte]*identityEntries.RegisterFactomIdentityStructure)
	im.CancelManager = NewCoinbaseCancelManager(im)
	im.CanceledCoinbaseOutputs = make(map[uint32][]uint32)
	if im == nil {
		atomic.WhereAmIMsg("no identity manager")
	}

	return im
}

func RandomIdentityManagerWithCounts(fedCount, audCount int) *IdentityManager {
	im := NewIdentityManager()
	for i := 0; i < fedCount; i++ {
		id := RandomIdentity()
		id.Status = constants.IDENTITY_FEDERATED_SERVER
		im.Authorities[id.IdentityChainID.Fixed()] = id.ToAuthority()
		im.Identities[id.IdentityChainID.Fixed()] = id
	}

	for i := 0; i < audCount; i++ {
		id := RandomIdentity()
		id.Status = constants.IDENTITY_AUDIT_SERVER
		im.Authorities[id.IdentityChainID.Fixed()] = id.ToAuthority()
		im.Identities[id.IdentityChainID.Fixed()] = id
	}
	return im
}

func RandomIdentityManager() *IdentityManager {
	im := NewIdentityManager()
	for i := 0; i < rand.Intn(10); i++ {
		id := RandomIdentity()
		im.Identities[id.IdentityChainID.Fixed()] = id
	}

	for i := 0; i < rand.Intn(10); i++ {
		id := RandomAuthority()
		im.Authorities[id.AuthorityChainID.Fixed()] = id
	}
	for i := 0; i < rand.Intn(10); i++ {
		r := identityEntries.RandomRegisterFactomIdentityStructure()
		im.IdentityRegistrations[r.IdentityChainID.Fixed()] = r
	}
	return im
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

func (im *IdentityManager) SetIdentityRegistration(chain interfaces.IHash) error {
	id := NewIdentity()
	id.IdentityChainID = chain
	id.Status = constants.IDENTITY_REGISTRATION_CHAIN

	im.SetIdentity(chain, id)
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
}

//func (im *IdentityManager) SetSkeletonKeyMainNet() error {
//	//Skeleton key:
//	//"0000000000000000000000000000000000000000000000000000000000000000":"0426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a"
//
//	return im.SetSkeletonKey("0426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a")
//}

// AuthorityServerCount returns the total count of Federated + Audit Servers
func (im *IdentityManager) AuthorityServerCount() int {
	im.Mutex.RLock()
	defer im.Mutex.RUnlock()
	answer := 0
	for _, v := range im.Authorities {
		if v.Status == constants.IDENTITY_FEDERATED_SERVER ||
			v.Status == constants.IDENTITY_AUDIT_SERVER {
			answer++
		}
	}
	return answer
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
	// Do nothing, it used to init the maps if they were empty, but we init the Identity control with non-empty maps
	if im.Identities == nil {
		im.Identities = make(map[[32]byte]*Identity)
	}
	if im.Authorities == nil {
		im.Authorities = make(map[[32]byte]*Authority)
	}
}

func (im *IdentityManager) SetIdentity(chainID interfaces.IHash, id *Identity) {
	im.Init()
	im.Mutex.Lock()
	defer im.Mutex.Unlock()
	im.Identities[chainID.Fixed()] = id
}

func (im *IdentityManager) RemoveIdentity(chainID interfaces.IHash) bool {
	im.Init()
	im.Mutex.Lock()
	defer im.Mutex.Unlock()
	id, ok := im.Identities[chainID.Fixed()]
	if ok == false {
		return false
	}

	if id.Status == constants.IDENTITY_SKELETON {
		return true // Do not remove skeleton identity
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

func (im *IdentityManager) GetSortedRegistrations() []*identityEntries.RegisterFactomIdentityStructure {
	list := im.GetRegistrations()
	sort.Sort(identityEntries.RegisterFactomIdentityStructureSort(list))
	return list
}
func (im *IdentityManager) GetRegistrations() []*identityEntries.RegisterFactomIdentityStructure {
	im.Mutex.RLock()
	defer im.Mutex.RUnlock()

	rs := make([]*identityEntries.RegisterFactomIdentityStructure, 0)
	for _, r := range im.IdentityRegistrations {
		rs = append(rs, r)
	}

	return rs
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
	rval, ok := im.Authorities[chainID.Fixed()]
	if !ok {
		return nil
	}
	return rval
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

func (a *IdentityManager) IsSameAs(b *IdentityManager) bool {
	if len(a.Authorities) != len(b.Authorities) {
		return false
	}

	for k := range a.Authorities {
		if _, ok := b.Authorities[k]; !ok {
			return false
		}
		if !a.Authorities[k].IsSameAs(b.Authorities[k]) {
			return false
		}
	}

	if len(a.Identities) != len(b.Identities) {
		return false
	}

	for k := range a.Identities {
		if _, ok := b.Identities[k]; !ok {
			return false
		}
		if !a.Identities[k].IsSameAs(b.Identities[k]) {
			return false
		}
	}

	if len(a.IdentityRegistrations) != len(b.IdentityRegistrations) {
		return false
	}

	for k := range a.IdentityRegistrations {
		if _, ok := b.IdentityRegistrations[k]; !ok {
			return false
		}
		if !a.IdentityRegistrations[k].IsSameAs(b.IdentityRegistrations[k]) {
			return false
		}
	}
	return true
}

func (e *IdentityManager) UnmarshalBinary(p []byte) error {
	if e == nil {
		atomic.WhereAmIMsg("no identity manager")
	}

	_, err := e.UnmarshalBinaryData(p)
	return err
}

func (im *IdentityManager) UnmarshalBinaryData(p []byte) (newData []byte, err error) {
	if im == nil {
		atomic.WhereAmIMsg("no identity manager")
	}

	buf := primitives.NewBuffer(p)
	newData = p

	al, err := buf.PopInt()
	if err != nil {
		return
	}

	newData = buf.Bytes()
	for i := 0; i < al; i++ {
		a := NewAuthority()
		newData, err = a.UnmarshalBinaryData(newData)
		if err != nil {
			return
		}
		im.Authorities[a.AuthorityChainID.Fixed()] = a
	}
	buf = primitives.NewBuffer(newData)

	il, err := buf.PopInt()
	if err != nil {
		return
	}

	newData = buf.Bytes()
	for i := 0; i < il; i++ {
		a := NewIdentity()
		newData, err = a.UnmarshalBinaryData(newData)
		if err != nil {
			return
		}
		im.Identities[a.IdentityChainID.Fixed()] = a
	}
	buf = primitives.NewBuffer(newData)

	rl, err := buf.PopInt()
	if err != nil {
		return
	}

	newData = buf.Bytes()
	for i := 0; i < rl; i++ {
		r := new(identityEntries.RegisterFactomIdentityStructure)
		newData, err = r.UnmarshalBinaryData(newData)
		if err != nil {
			return
		}
		im.IdentityRegistrations[r.IdentityChainID.Fixed()] = r
	}
	buf = primitives.NewBuffer(newData)

	newData = buf.DeepCopyBytes()
	return
}

func (im *IdentityManager) MarshalBinary() ([]byte, error) {
	if im == nil {
		atomic.WhereAmIMsg("no identity manager")
	}

	buf := primitives.NewBuffer(nil)
	im.Init()

	err := buf.PushInt(len(im.Authorities))
	if err != nil {
		return nil, err
	}

	for _, a := range im.GetSortedAuthorities() {
		err = buf.PushBinaryMarshallable(a)
		if err != nil {
			return nil, err
		}
	}

	err = buf.PushInt(len(im.Identities))
	if err != nil {
		return nil, err
	}

	for _, i := range im.GetSortedIdentities() {
		err = buf.PushBinaryMarshallable(i)
		if err != nil {
			return nil, err
		}
	}

	err = buf.PushInt(len(im.IdentityRegistrations))
	if err != nil {
		return nil, err
	}

	for _, i := range im.GetSortedRegistrations() {
		err = buf.PushBinaryMarshallable(i)
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), nil
}

// Used when cloning state into sim nodes
func (im *IdentityManager) Clone() *IdentityManager {
	if im == nil {
		atomic.WhereAmIMsg("no identity manager")
	}
	b := NewIdentityManager()
	for k, v := range im.Authorities {
		b.Authorities[k] = v.Clone()
	}
	for k, v := range im.Identities {
		b.Identities[k] = v.Clone()
	}

	b.MaxAuthorityServerCount = im.MaxAuthorityServerCount
	b.OldEntries = make([]*OldEntry, len(im.OldEntries))
	for k, v := range im.OldEntries {
		copy := *v
		b.OldEntries[k] = &copy
	}

	b.IdentityRegistrations = make(map[[32]byte]*identityEntries.RegisterFactomIdentityStructure, len(im.IdentityRegistrations))
	for k, v := range im.IdentityRegistrations {
		b.IdentityRegistrations[k] = v
	}

	if b == nil {
		atomic.WhereAmIMsg("no identity manager")
	}

	return b
}
