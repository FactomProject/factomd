// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identity

import (
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
	Key1                 interfaces.IHash
	Key2                 interfaces.IHash
	Key3                 interfaces.IHash
	Key4                 interfaces.IHash
	SigningKey           interfaces.IHash
	Status               uint8
	AnchorKeys           []AnchorSigningKey
}

var _ interfaces.Printable = (*Identity)(nil)
var _ interfaces.BinaryMarshallable = (*Identity)(nil)

func NewIdentity() *Identity {
	i := new(Identity)
	i.IdentityChainID = primitives.NewZeroHash()
	i.ManagementChainID = primitives.NewZeroHash()
	i.MatryoshkaHash = primitives.NewZeroHash()
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

/*
	IdentityRegistered   uint32
	IdentityCreated      uint32
	ManagementRegistered uint32
	ManagementCreated    uint32
	Status               uint8
	AnchorKeys           []AnchorSigningKey
*/

func (e *Identity) Clone() *Identity {
	b := NewIdentity()
	b.IdentityChainID.SetBytes(e.IdentityChainID.Bytes())
	b.ManagementChainID.SetBytes(e.ManagementChainID.Bytes())
	b.MatryoshkaHash.SetBytes(e.MatryoshkaHash.Bytes())
	b.Key1.SetBytes(e.Key1.Bytes())
	b.Key2.SetBytes(e.Key2.Bytes())
	b.Key3.SetBytes(e.Key3.Bytes())
	b.Key4.SetBytes(e.Key4.Bytes())

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
