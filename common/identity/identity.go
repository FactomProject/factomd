// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identity

import (
	"errors"
	"fmt"
	"os"

	"bytes"

	"math/rand"

	ed "github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
)

// IdentitySort is sort.Sort interface implementation
type IdentitySort []*Identity

// Len returns the length of this object
func (p IdentitySort) Len() int {
	return len(p)
}

// Swap swaps the objects at indices 'i' and 'i'
func (p IdentitySort) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// Less returns true if the object at index 'i' is less than the object at index 'j'
func (p IdentitySort) Less(i, j int) bool {
	return bytes.Compare(p[i].IdentityChainID.Bytes(), p[j].IdentityChainID.Bytes()) < 0

}

// Identity is a very important data structure for managing server identities. See a more detailed description here:
// https://github.com/FactomProject/FactomDocs/blob/master/Identity.md
type Identity struct {
	IdentityChainID    interfaces.IHash `json:"identity_chainid"` // The chain id associated with this id
	IdentityChainSync  EntryBlockSync   `json:"-"`
	IdentityRegistered uint32           `json:"identity_registered` // Directory block height of when the identity was registered
	IdentityCreated    uint32           `json:"identity_created`    // Directory block height of when the identity was created

	ManagementChainID    interfaces.IHash `json:"management_chaind` // Messages related to being a federated, audit, or candidate server live in this subchain
	ManagementChainSync  EntryBlockSync   `json:"-"`
	ManagementRegistered uint32           `json:"management_registered` // Directory block height of when the managment chain id was registered
	ManagementCreated    uint32           `json:"management_created`    // Directory block height of when the managment chain id was created
	MatryoshkaHash       interfaces.IHash `json:"matryoshka_hash`       // (no longer used since M2 and M3) The MatryoshkaHash of the id

	// All 4 levels keys, 0 indexed.
	//		Keys[0] --> Key 1
	//		Keys[1] --> Key 2
	//		Keys[2] --> Key 3
	//		Keys[3] --> Key 4
	Keys            [4]interfaces.IHash `json:"identity_keys"`    // The four secret keys associated with this identity
	SigningKey      interfaces.IHash    `json:"signing_key"`      // The key this id uses to sign blocks in Factom
	Status          uint8               `json:"status"`           // Identity status type, see constants.go (is this a federated / audit / etc )
	AnchorKeys      []AnchorSigningKey  `json:"anchor_keys"`      // (no longer used since M2 and M3) Bitcoin anchor keys
	Efficiency      uint16              `json:"efficiency"`       // The efficiency of the server in hundredths of a percent: ie 10000 is 100%
	CoinbaseAddress interfaces.IHash    `json:"coinbase_address"` // The Factoid address associated with this identity for receiving Factoids from the protocol
}

var _ interfaces.Printable = (*Identity)(nil)
var _ interfaces.BinaryMarshallable = (*Identity)(nil)

// NewIdentity returns a new identity with zero hashes
func NewIdentity() *Identity {
	i := new(Identity)
	i.IdentityChainID = primitives.NewZeroHash()
	i.ManagementChainID = primitives.NewZeroHash()
	i.MatryoshkaHash = primitives.NewZeroHash()

	for c := range i.Keys {
		i.Keys[c] = primitives.NewZeroHash()
	}

	i.SigningKey = primitives.NewZeroHash()
	i.Efficiency = 10000
	i.IdentityChainSync = *NewEntryBlockSync()
	i.ManagementChainSync = *NewEntryBlockSync()
	i.CoinbaseAddress = primitives.NewZeroHash()

	return i
}

// ToAuthority should ONLY be used in TESTING. Returns a new authority with the input identity
// Helpful for unit tests, useless for anything else
func (id *Identity) ToAuthority() *Authority {
	a := NewAuthority()
	a.AuthorityChainID = id.IdentityChainID
	a.ManagementChainID = id.ManagementChainID
	a.Efficiency = id.Efficiency
	//a.SigningKey = id.SigningKey
	a.CoinbaseAddress = id.CoinbaseAddress
	a.AnchorKeys = id.AnchorKeys
	a.Status = id.Status
	a.MatryoshkaHash = id.MatryoshkaHash
	return a
}

// RandomIdentity creates new identity with random initialized values
func RandomIdentity() *Identity {
	id := NewIdentity()

	id.IdentityChainID = primitives.RandomHash()
	id.IdentityRegistered = random.RandUInt32()
	id.IdentityCreated = random.RandUInt32()
	id.ManagementChainID = primitives.RandomHash()
	id.ManagementRegistered = random.RandUInt32()
	id.ManagementCreated = random.RandUInt32()
	id.MatryoshkaHash = primitives.RandomHash()

	for c := range id.Keys {
		id.Keys[c] = primitives.RandomHash()
	}

	id.SigningKey = primitives.RandomHash()
	id.Status = random.RandUInt8()

	l := random.RandIntBetween(1, 10)
	for i := 0; i < l; i++ {
		id.AnchorKeys = append(id.AnchorKeys, *RandomAnchorSigningKey())
	}
	id.CoinbaseAddress = primitives.RandomHash()
	id.Efficiency = uint16(rand.Intn(10000))

	id.IdentityChainSync = *RandomEntryBlockSync()
	id.ManagementChainSync = *RandomEntryBlockSync()

	return id
}

// GetCoinbaseHumanReadable returns the coinbasae address in human readable format if it exists, else returns 'No Address"
func (id *Identity) GetCoinbaseHumanReadable() string {
	if id.CoinbaseAddress.IsZero() {
		return "No Address"
	}
	add := factoid.NewAddress(id.CoinbaseAddress.Bytes())
	//primitives.ConvertFctAddressToUserStr(add)
	return primitives.ConvertFctAddressToUserStr(add)
}

// IsPromteable will return if the identity is able to be promoted.
//		Checks if the Identity is complete
//		Checks if the registration is valid
func (id *Identity) IsPromteable() (bool, error) {
	if id == nil {
		return false, fmt.Errorf("Identity does not exist")
	}

	if ok, err := id.IsComplete(); !ok {
		return ok, err
	}

	if ok, err := id.IsRegistrationValid(); !ok {
		return ok, err
	}

	return true, nil
}

// IsRegistrationValid will return if the registration of the identity is
// valid. It is determined by the block heights of registration to creation
// and is all time based (where time is measured in blocks)
func (id *Identity) IsRegistrationValid() (bool, error) {
	// Check the time window on registration
	dif := id.IdentityCreated - id.IdentityRegistered
	if id.IdentityRegistered > id.IdentityCreated { // Uint underflow
		dif = id.IdentityRegistered - id.IdentityCreated
	}

	if dif > constants.IDENTITY_REGISTRATION_BLOCK_WINDOW {
		return false, errors.New("Time window of identity create and register invalid")
	}

	// Also check Management registration
	dif = id.ManagementCreated - id.ManagementRegistered
	if id.ManagementRegistered > id.ManagementCreated { // Uint underflow
		dif = id.ManagementRegistered - id.ManagementCreated
	}

	if dif > constants.IDENTITY_REGISTRATION_BLOCK_WINDOW {
		return false, errors.New("Time window of identity managment create and register invalid")
	}

	return true, nil
}

// IsComplete returns if the identity is complete, meaning it has
// all of the required information for an authority server.
// If the identity is not valid, a list of missing things will be
// returned in the error
func (id *Identity) IsComplete() (bool, error) {
	isNil := func(hash interfaces.IHash) bool {
		if hash == nil || hash.IsZero() {
			return true
		}
		return false
	}

	// A list of all missing things for a helpful error
	missing := []string{}

	// Required for Admin Block
	if isNil(id.SigningKey) {
		missing = append(missing, "id.SigningKey block signing key, ")
	}

	if len(id.AnchorKeys) == 0 {
		missing = append(missing, "id.AnchorKeys block signing key, ")

	}

	if isNil(id.MatryoshkaHash) {
		missing = append(missing, "id.MatryoshkaHash block signing key, ")
	}

	// There are additional requirements we will enforce
	for c := range id.Keys {
		if isNil(id.Keys[c]) {
			missing = append(missing, fmt.Sprintf("id key %d, ", c+1))
		}
	}

	if isNil(id.IdentityChainID) {
		missing = append(missing, "id.IdentityChainID identity chain, ")
	}

	if isNil(id.ManagementChainID) {
		missing = append(missing, "id.ManagementChainID identity chain, ")
	}

	if len(missing) > 0 {
		return false, fmt.Errorf("missing: %v", missing)
	}

	return true, nil
}

// Clone returns an identical copy of this object
func (id *Identity) Clone() *Identity {
	b := NewIdentity()
	b.IdentityChainID.SetBytes(id.IdentityChainID.Bytes())
	b.ManagementChainID.SetBytes(id.ManagementChainID.Bytes())
	b.MatryoshkaHash.SetBytes(id.MatryoshkaHash.Bytes())
	for i := range b.Keys {
		b.Keys[i].SetBytes(id.Keys[i].Bytes())
	}

	b.SigningKey = id.SigningKey
	b.IdentityRegistered = id.IdentityRegistered
	b.IdentityCreated = id.IdentityCreated
	b.ManagementRegistered = id.ManagementRegistered
	b.ManagementCreated = id.ManagementCreated
	b.Status = id.Status
	b.Efficiency = id.Efficiency
	b.IdentityChainSync = *id.IdentityChainSync.Clone()
	b.ManagementChainSync = *id.ManagementChainSync.Clone()
	b.CoinbaseAddress = id.CoinbaseAddress.Copy()

	b.AnchorKeys = make([]AnchorSigningKey, len(id.AnchorKeys))
	for i := range id.AnchorKeys {
		b.AnchorKeys[i] = id.AnchorKeys[i]
	}

	return b
}

// IsSameAs returns true iff the input object is identical to this object
func (id *Identity) IsSameAs(b *Identity) bool {
	if id.IdentityChainID.IsSameAs(b.IdentityChainID) == false {
		return false
	}
	if id.IdentityRegistered != b.IdentityRegistered {
		return false
	}
	if id.IdentityCreated != b.IdentityCreated {
		return false
	}
	if id.ManagementChainID.IsSameAs(b.ManagementChainID) == false {
		return false
	}
	if id.ManagementRegistered != b.ManagementRegistered {
		return false
	}
	if id.ManagementCreated != b.ManagementCreated {
		return false
	}
	if id.MatryoshkaHash.IsSameAs(b.MatryoshkaHash) == false {
		return false
	}
	if id.Keys[0].IsSameAs(b.Keys[0]) == false {
		return false
	}
	if id.Keys[1].IsSameAs(b.Keys[1]) == false {
		return false
	}
	if id.Keys[2].IsSameAs(b.Keys[2]) == false {
		return false
	}
	if id.Keys[3].IsSameAs(b.Keys[3]) == false {
		return false
	}
	if id.SigningKey.IsSameAs(b.SigningKey) == false {
		return false
	}
	if id.Status != b.Status {
		return false
	}
	if id.Efficiency != b.Efficiency {
		return false
	}
	if len(id.AnchorKeys) != len(b.AnchorKeys) {
		return false
	}
	for i := range id.AnchorKeys {
		if id.AnchorKeys[i].IsSameAs(&b.AnchorKeys[i]) == false {
			return false
		}
	}
	if !id.IdentityChainSync.IsSameAs(&b.IdentityChainSync) {
		return false
	}
	if !id.ManagementChainSync.IsSameAs(&b.ManagementChainSync) {
		return false
	}
	return true
}

// Init initializes this object with zero hashes
func (id *Identity) Init() {
	if id.IdentityChainID == nil {
		id.IdentityChainID = primitives.NewZeroHash()
	}
	if id.ManagementChainID == nil {
		id.ManagementChainID = primitives.NewZeroHash()
	}
	if id.MatryoshkaHash == nil {
		id.MatryoshkaHash = primitives.NewZeroHash()
	}
	if id.Keys[0] == nil {
		id.Keys[0] = primitives.NewZeroHash()
	}
	if id.Keys[1] == nil {
		id.Keys[1] = primitives.NewZeroHash()
	}
	if id.Keys[2] == nil {
		id.Keys[2] = primitives.NewZeroHash()
	}
	if id.Keys[3] == nil {
		id.Keys[3] = primitives.NewZeroHash()
	}
	if id.SigningKey == nil {
		id.SigningKey = primitives.NewZeroHash()
	}
}

// MarshalBinary marshals this object
func (id *Identity) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "Identity.MarshalBinary err:%v", *pe)
		}
	}(&err)
	id.Init()
	buf := primitives.NewBuffer(nil)

	err = buf.PushBinaryMarshallable(id.IdentityChainID)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(id.IdentityRegistered)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(id.IdentityCreated)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(&id.IdentityChainSync)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(id.ManagementChainID)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(id.ManagementRegistered)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(id.ManagementCreated)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(&id.ManagementChainSync)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(id.MatryoshkaHash)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(id.Keys[0])
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(id.Keys[1])
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(id.Keys[2])
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(id.Keys[3])
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(id.SigningKey)
	if err != nil {
		return nil, err
	}
	err = buf.PushByte(byte(id.Status))
	if err != nil {
		return nil, err
	}

	l := len(id.AnchorKeys)
	err = buf.PushVarInt(uint64(l))
	for _, v := range id.AnchorKeys {
		err = buf.PushBinaryMarshallable(&v)
		if err != nil {
			return nil, err
		}
	}

	err = buf.PushUInt16(id.Efficiency)
	if err != nil {
		return nil, err
	}

	err = buf.PushIHash(id.CoinbaseAddress)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinaryData unmarshals the input data into this object
func (id *Identity) UnmarshalBinaryData(p []byte) (newData []byte, err error) {
	id.Init()
	buf := primitives.NewBuffer(p)
	newData = p

	err = buf.PopBinaryMarshallable(id.IdentityChainID)
	if err != nil {
		return
	}
	id.IdentityRegistered, err = buf.PopUInt32()
	if err != nil {
		return
	}
	id.IdentityCreated, err = buf.PopUInt32()
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(&id.IdentityChainSync)
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(id.ManagementChainID)
	if err != nil {
		return
	}
	id.ManagementRegistered, err = buf.PopUInt32()
	if err != nil {
		return
	}
	id.ManagementCreated, err = buf.PopUInt32()
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(&id.ManagementChainSync)
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(id.MatryoshkaHash)
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(id.Keys[0])
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(id.Keys[1])
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(id.Keys[2])
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(id.Keys[3])
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(id.SigningKey)
	if err != nil {
		return
	}
	b, err := buf.PopByte()
	if err != nil {
		return
	}
	id.Status = uint8(b)

	l, err := buf.PopVarInt()
	if err != nil {
		return
	}

	for i := 0; i < int(l); i++ {
		var ak AnchorSigningKey
		err = buf.PopBinaryMarshallable(&ak)
		if err != nil {
			return
		}
		id.AnchorKeys = append(id.AnchorKeys, ak)
	}

	id.Efficiency, err = buf.PopUInt16()
	if err != nil {
		return
	}

	id.CoinbaseAddress, err = buf.PopIHash()
	if err != nil {
		return
	}

	newData = buf.DeepCopyBytes()
	return
}

// UnmarshalBinary unmarshals the input data into this object
func (id *Identity) UnmarshalBinary(p []byte) error {
	_, err := id.UnmarshalBinaryData(p)
	return err
}

// VerifySignature verifies the input message and signature using the public key from this id
func (id *Identity) VerifySignature(msg []byte, sig *[constants.SIGNATURE_LENGTH]byte) (bool, error) {
	//return true, nil // Testing
	var pub [32]byte
	tmp, err := id.SigningKey.MarshalBinary()
	if err != nil {
		return false, err
	}
	copy(pub[:], tmp)
	valid := ed.VerifyCanonical(&pub, msg, sig)

	return valid, nil
}

// JSONByte returns the json encoded byte array
func (id *Identity) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(id)
}

// JSONString returns the json encoded string
func (id *Identity) JSONString() (string, error) {
	return primitives.EncodeJSONString(id)
}

// String returns the json encoded string
func (id *Identity) String() string {
	str, _ := id.JSONString()
	return str
}

// IsFull returns false if any hash is the zero hash
func (id *Identity) IsFull() bool {
	zero := primitives.NewZeroHash()
	if id.IdentityChainID.IsSameAs(zero) {
		return false
	}
	if id.ManagementChainID.IsSameAs(zero) {
		return false
	}
	if id.MatryoshkaHash.IsSameAs(zero) {
		return false
	}
	if id.Keys[0].IsSameAs(zero) {
		return false
	}
	if id.Keys[1].IsSameAs(zero) {
		return false
	}
	if id.Keys[2].IsSameAs(zero) {
		return false
	}
	if id.Keys[3].IsSameAs(zero) {
		return false
	}
	if id.SigningKey.IsSameAs(zero) {
		return false
	}
	if len(id.AnchorKeys) == 0 {
		return false
	}
	return true
}
