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
	"github.com/PaulSnow/factom2d/common/constants"
	"github.com/PaulSnow/factom2d/common/factoid"
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/primitives"
	"github.com/PaulSnow/factom2d/common/primitives/random"
)

// sort.Sort interface implementation
type IdentitySort []*Identity

func (p IdentitySort) Len() int {
	return len(p)
}
func (p IdentitySort) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
func (p IdentitySort) Less(i, j int) bool {
	return bytes.Compare(p[i].IdentityChainID.Bytes(), p[j].IdentityChainID.Bytes()) < 0

}

//https://github.com/FactomProject/FactomDocs/blob/master/Identity.md
type Identity struct {
	IdentityChainID    interfaces.IHash `json:"identity_chainid"`
	IdentityChainSync  EntryBlockSync   `json:"-"`
	IdentityRegistered uint32           `json:"identity_registered`
	IdentityCreated    uint32           `json:"identity_created`

	ManagementChainID    interfaces.IHash `json:"management_chaind`
	ManagementChainSync  EntryBlockSync   `json:"-"`
	ManagementRegistered uint32           `json:"management_registered`
	ManagementCreated    uint32           `json:"management_created`
	MatryoshkaHash       interfaces.IHash `json:"matryoshka_hash`

	// All 4 levels keys, 0 indexed.
	//		Keys[0] --> Key 1
	//		Keys[1] --> Key 2
	//		Keys[2] --> Key 3
	//		Keys[3] --> Key 4
	Keys            [4]interfaces.IHash `json:"identity_keys"`
	SigningKey      interfaces.IHash    `json:"signing_key"`
	Status          uint8               `json:"status"`
	AnchorKeys      []AnchorSigningKey  `json:"anchor_keys"`
	Efficiency      uint16              `json:"efficiency"`
	CoinbaseAddress interfaces.IHash    `json:"coinbase_address"`
}

var _ interfaces.Printable = (*Identity)(nil)
var _ interfaces.BinaryMarshallable = (*Identity)(nil)

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

// ToAuthority should ONLY be used in TESTING
// 	Helpful for unit tests, useless for anything else
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

func (e *Identity) Clone() *Identity {
	b := NewIdentity()
	b.IdentityChainID.SetBytes(e.IdentityChainID.Bytes())
	b.ManagementChainID.SetBytes(e.ManagementChainID.Bytes())
	b.MatryoshkaHash.SetBytes(e.MatryoshkaHash.Bytes())
	for i := range b.Keys {
		b.Keys[i].SetBytes(e.Keys[i].Bytes())
	}

	b.SigningKey = e.SigningKey
	b.IdentityRegistered = e.IdentityRegistered
	b.IdentityCreated = e.IdentityCreated
	b.ManagementRegistered = e.ManagementRegistered
	b.ManagementCreated = e.ManagementCreated
	b.Status = e.Status
	b.Efficiency = e.Efficiency
	b.IdentityChainSync = *e.IdentityChainSync.Clone()
	b.ManagementChainSync = *e.ManagementChainSync.Clone()
	b.CoinbaseAddress = e.CoinbaseAddress.Copy()

	b.AnchorKeys = make([]AnchorSigningKey, len(e.AnchorKeys))
	for i := range e.AnchorKeys {
		b.AnchorKeys[i] = e.AnchorKeys[i]
	}

	return b
}

func (e *Identity) IsSameAs(b *Identity) bool {
	if e.IdentityChainID.IsSameAs(b.IdentityChainID) == false {
		return false
	}
	if e.IdentityRegistered != b.IdentityRegistered {
		return false
	}
	if e.IdentityCreated != b.IdentityCreated {
		return false
	}
	if e.ManagementChainID.IsSameAs(b.ManagementChainID) == false {
		return false
	}
	if e.ManagementRegistered != b.ManagementRegistered {
		return false
	}
	if e.ManagementCreated != b.ManagementCreated {
		return false
	}
	if e.MatryoshkaHash.IsSameAs(b.MatryoshkaHash) == false {
		return false
	}
	if e.Keys[0].IsSameAs(b.Keys[0]) == false {
		return false
	}
	if e.Keys[1].IsSameAs(b.Keys[1]) == false {
		return false
	}
	if e.Keys[2].IsSameAs(b.Keys[2]) == false {
		return false
	}
	if e.Keys[3].IsSameAs(b.Keys[3]) == false {
		return false
	}
	if e.SigningKey.IsSameAs(b.SigningKey) == false {
		return false
	}
	if e.Status != b.Status {
		return false
	}
	if e.Efficiency != b.Efficiency {
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
	if !e.IdentityChainSync.IsSameAs(&b.IdentityChainSync) {
		return false
	}
	if !e.ManagementChainSync.IsSameAs(&b.ManagementChainSync) {
		return false
	}
	return true
}

func (e *Identity) Init() {
	if e.IdentityChainID == nil {
		e.IdentityChainID = primitives.NewZeroHash()
	}
	if e.ManagementChainID == nil {
		e.ManagementChainID = primitives.NewZeroHash()
	}
	if e.MatryoshkaHash == nil {
		e.MatryoshkaHash = primitives.NewZeroHash()
	}
	if e.Keys[0] == nil {
		e.Keys[0] = primitives.NewZeroHash()
	}
	if e.Keys[1] == nil {
		e.Keys[1] = primitives.NewZeroHash()
	}
	if e.Keys[2] == nil {
		e.Keys[2] = primitives.NewZeroHash()
	}
	if e.Keys[3] == nil {
		e.Keys[3] = primitives.NewZeroHash()
	}
	if e.SigningKey == nil {
		e.SigningKey = primitives.NewZeroHash()
	}
}

func (e *Identity) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "Identity.MarshalBinary err:%v", *pe)
		}
	}(&err)
	e.Init()
	buf := primitives.NewBuffer(nil)

	err = buf.PushBinaryMarshallable(e.IdentityChainID)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(e.IdentityRegistered)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(e.IdentityCreated)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(&e.IdentityChainSync)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(e.ManagementChainID)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(e.ManagementRegistered)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(e.ManagementCreated)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(&e.ManagementChainSync)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(e.MatryoshkaHash)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(e.Keys[0])
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(e.Keys[1])
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(e.Keys[2])
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(e.Keys[3])
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(e.SigningKey)
	if err != nil {
		return nil, err
	}
	err = buf.PushByte(byte(e.Status))
	if err != nil {
		return nil, err
	}

	l := len(e.AnchorKeys)
	err = buf.PushVarInt(uint64(l))
	for _, v := range e.AnchorKeys {
		err = buf.PushBinaryMarshallable(&v)
		if err != nil {
			return nil, err
		}
	}

	err = buf.PushUInt16(e.Efficiency)
	if err != nil {
		return nil, err
	}

	err = buf.PushIHash(e.CoinbaseAddress)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (e *Identity) UnmarshalBinaryData(p []byte) (newData []byte, err error) {
	e.Init()
	buf := primitives.NewBuffer(p)
	newData = p

	err = buf.PopBinaryMarshallable(e.IdentityChainID)
	if err != nil {
		return
	}
	e.IdentityRegistered, err = buf.PopUInt32()
	if err != nil {
		return
	}
	e.IdentityCreated, err = buf.PopUInt32()
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(&e.IdentityChainSync)
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(e.ManagementChainID)
	if err != nil {
		return
	}
	e.ManagementRegistered, err = buf.PopUInt32()
	if err != nil {
		return
	}
	e.ManagementCreated, err = buf.PopUInt32()
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(&e.ManagementChainSync)
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(e.MatryoshkaHash)
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(e.Keys[0])
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(e.Keys[1])
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(e.Keys[2])
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(e.Keys[3])
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(e.SigningKey)
	if err != nil {
		return
	}
	b, err := buf.PopByte()
	if err != nil {
		return
	}
	e.Status = uint8(b)

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
		e.AnchorKeys = append(e.AnchorKeys, ak)
	}

	e.Efficiency, err = buf.PopUInt16()
	if err != nil {
		return
	}

	e.CoinbaseAddress, err = buf.PopIHash()
	if err != nil {
		return
	}

	newData = buf.DeepCopyBytes()
	return
}

func (e *Identity) UnmarshalBinary(p []byte) error {
	_, err := e.UnmarshalBinaryData(p)
	return err
}

func (id *Identity) VerifySignature(msg []byte, sig *[constants.SIGNATURE_LENGTH]byte) (bool, error) {
	//return true, nil // Testing
	var pub [32]byte
	tmp, err := id.SigningKey.MarshalBinary()
	if err != nil {
		return false, err
	} else {
		copy(pub[:], tmp)
		valid := ed.VerifyCanonical(&pub, msg, sig)
		if !valid {
		} else {
			return true, nil
		}
	}
	return false, nil
}

func (e *Identity) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *Identity) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *Identity) String() string {
	str, _ := e.JSONString()
	return str
}

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
