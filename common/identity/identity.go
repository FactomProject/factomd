// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identity

import (
	"errors"
	"fmt"

	ed "github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
)

//https://github.com/FactomProject/FactomDocs/blob/master/Identity.md

type Identity struct {
	IdentityChainID      interfaces.IHash
	IdentityRegistered   uint32
	IdentityCreated      uint32
	ManagementChainID    interfaces.IHash
	ManagementRegistered uint32
	ManagementCreated    uint32
	MatryoshkaHash       interfaces.IHash

	// All 4 levels keys, 0 indexed.
	//		Keys[0] --> Key 1
	//		Keys[1] --> Key 2
	//		Keys[2] --> Key 3
	//		Keys[3] --> Key 4
	Keys       [4]interfaces.IHash
	Key1       interfaces.IHash
	Key2       interfaces.IHash
	Key3       interfaces.IHash
	Key4       interfaces.IHash
	SigningKey interfaces.IHash
	Status     uint8
	AnchorKeys []AnchorSigningKey
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

	i.Key1 = primitives.NewZeroHash()
	i.Key2 = primitives.NewZeroHash()
	i.Key3 = primitives.NewZeroHash()
	i.Key4 = primitives.NewZeroHash()
	i.SigningKey = primitives.NewZeroHash()

	return i
}

func RandomIdentity() *Identity {
	id := new(Identity)

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

	id.Key1 = primitives.RandomHash()
	id.Key2 = primitives.RandomHash()
	id.Key3 = primitives.RandomHash()
	id.Key4 = primitives.RandomHash()
	id.SigningKey = primitives.RandomHash()
	id.Status = random.RandUInt8()

	l := random.RandIntBetween(0, 10)
	for i := 0; i < l; i++ {
		id.AnchorKeys = append(id.AnchorKeys, *RandomAnchorSigningKey())
	}

	return id
}

// IsPromteable will return if the identity is able to be promoted.
//		Checks if the Identity is complete
//		Checks if the registration is valid
func (id *Identity) IsPromteable() (bool, error) {
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
		missing = append(missing, "block signing key")
	}

	if len(id.AnchorKeys) == 0 {
		missing = append(missing, "block signing key")

	}

	if isNil(id.MatryoshkaHash) {
		missing = append(missing, "block signing key")
	}

	// There are additional requirements we will enforce
	for c := range id.Keys {
		if isNil(id.Keys[c]) {
			missing = append(missing, fmt.Sprintf("id key %d", c+1))
		}
	}

	if isNil(id.IdentityChainID) {
		missing = append(missing, "identity chain")
	}

	if isNil(id.ManagementChainID) {
		missing = append(missing, "identity chain")
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
	b.Key1.SetBytes(e.Key1.Bytes())
	b.Key2.SetBytes(e.Key2.Bytes())
	b.Key3.SetBytes(e.Key3.Bytes())
	b.Key4.SetBytes(e.Key4.Bytes())

	b.SigningKey = e.SigningKey
	b.IdentityRegistered = e.IdentityRegistered
	b.IdentityCreated = e.IdentityCreated
	b.ManagementRegistered = e.ManagementRegistered
	b.ManagementCreated = e.ManagementCreated
	b.Status = e.Status

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
	if e.Key1.IsSameAs(b.Key1) == false {
		return false
	}
	if e.Key2.IsSameAs(b.Key2) == false {
		return false
	}
	if e.Key3.IsSameAs(b.Key3) == false {
		return false
	}
	if e.Key4.IsSameAs(b.Key4) == false {
		return false
	}
	if e.SigningKey.IsSameAs(b.SigningKey) == false {
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
	if e.Key1 == nil {
		e.Key1 = primitives.NewZeroHash()
	}
	if e.Key2 == nil {
		e.Key2 = primitives.NewZeroHash()
	}
	if e.Key3 == nil {
		e.Key3 = primitives.NewZeroHash()
	}
	if e.Key4 == nil {
		e.Key4 = primitives.NewZeroHash()
	}
	if e.SigningKey == nil {
		e.SigningKey = primitives.NewZeroHash()
	}
}

func (e *Identity) MarshalBinary() ([]byte, error) {
	e.Init()
	buf := primitives.NewBuffer(nil)

	err := buf.PushBinaryMarshallable(e.IdentityChainID)
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
	err = buf.PushBinaryMarshallable(e.MatryoshkaHash)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(e.Key1)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(e.Key2)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(e.Key3)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(e.Key4)
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
	err = buf.PopBinaryMarshallable(e.MatryoshkaHash)
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(e.Key1)
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(e.Key2)
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(e.Key3)
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(e.Key4)
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
	if id.Key1.IsSameAs(zero) {
		return false
	}
	if id.Key2.IsSameAs(zero) {
		return false
	}
	if id.Key3.IsSameAs(zero) {
		return false
	}
	if id.Key4.IsSameAs(zero) {
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
