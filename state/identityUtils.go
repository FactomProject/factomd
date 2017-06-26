// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"encoding/binary"
	"errors"

	ed "github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
)

type AnchorSigningKey struct {
	BlockChain string `json:"blockchain"`
	KeyLevel   byte   `json:"level"`
	KeyType    byte   `json:"keytype"`
	Key        []byte `json:"key"` //if bytes, it is hex
}

var _ interfaces.BinaryMarshallable = (*AnchorSigningKey)(nil)

func RandomAnchorSigningKey() *AnchorSigningKey {
	ask := new(AnchorSigningKey)

	ask.BlockChain = random.RandomString()
	ask.KeyLevel = random.RandByte()
	ask.KeyType = random.RandByte()
	ask.Key = random.RandNonEmptyByteSlice()

	return ask
}

func (e *AnchorSigningKey) IsSameAs(b *AnchorSigningKey) bool {
	if e.BlockChain != b.BlockChain {
		return false
	}
	if e.KeyLevel != b.KeyLevel {
		return false
	}
	if e.KeyType != b.KeyType {
		return false
	}
	if primitives.AreBytesEqual(e.Key, b.Key) == false {
		return false
	}
	return true
}

func (e *AnchorSigningKey) MarshalBinary() ([]byte, error) {
	buf := primitives.NewBuffer(nil)

	err := buf.PushString(e.BlockChain)
	if err != nil {
		return nil, err
	}

	err = buf.PushByte(e.KeyLevel)
	if err != nil {
		return nil, err
	}
	err = buf.PushByte(e.KeyType)
	if err != nil {
		return nil, err
	}

	err = buf.PushBytes(e.Key)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (e *AnchorSigningKey) UnmarshalBinaryData(p []byte) (newData []byte, err error) {
	newData = p
	buf := primitives.NewBuffer(p)

	e.BlockChain, err = buf.PopString()
	if err != nil {
		return
	}
	e.KeyLevel, err = buf.PopByte()
	if err != nil {
		return
	}
	e.KeyType, err = buf.PopByte()
	if err != nil {
		return
	}
	e.Key, err = buf.PopBytes()
	if err != nil {
		return
	}

	newData = buf.DeepCopyBytes()
	return
}

func (e *AnchorSigningKey) UnmarshalBinary(p []byte) error {
	_, err := e.UnmarshalBinaryData(p)
	return err
}

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

func (id *Identity) FixMissingKeys(s *State) error {
	// This identity will always have blank keys
	if id.IdentityChainID.IsSameAs(s.GetNetworkBootStrapIdentity()) {
		return nil
	}
	if !statusIsFedOrAudit(id.Status) {
		//return
	}
	// Rebuilds identity
	err := s.AddIdentityFromChainID(id.IdentityChainID)
	if err != nil {
		return err
	}
	return nil
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

// Sig is signed message, msg is raw message
func CheckSig(idKey interfaces.IHash, pub []byte, msg []byte, sig []byte) bool {
	var pubFix [32]byte
	var sigFix [64]byte

	copy(pubFix[:], pub[:32])
	copy(sigFix[:], sig[:64])

	pre := make([]byte, 0)
	pre = append(pre, []byte{0x01}...)
	pre = append(pre, pubFix[:]...)
	id := primitives.Shad(pre)

	if id.IsSameAs(idKey) {
		return ed.VerifyCanonical(&pubFix, msg, &sigFix)
	} else {
		return false
	}
}

// Checking the external ids if they match the needed lengths
func CheckExternalIDsLength(extIDs [][]byte, lengths []int) bool {
	if len(extIDs) != len(lengths) {
		return false
	}
	for i := range extIDs {
		if !CheckLength(lengths[i], extIDs[i]) {
			return false
		}
	}
	return true
}

func CheckLength(length int, item []byte) bool {
	if len(item) != length {
		return false
	} else {
		return true
	}
}

func AppendExtIDs(extIDs [][]byte, start int, end int) ([]byte, error) {
	if len(extIDs) < (end + 1) {
		return nil, errors.New("Error: Index out of bound exception in AppendExtIDs()")
	}
	appended := make([]byte, 0)
	for i := start; i <= end; i++ {
		appended = append(appended, extIDs[i][:]...)
	}
	return appended, nil
}

// Makes sure the timestamp is within the designated window to be valid : 12 hours
// TimeEntered is in seconds
func CheckTimestamp(time []byte, timeEntered int64) bool {
	if len(time) < 8 {
		zero := []byte{00}
		add := make([]byte, 0)
		for i := len(time); i <= 8; i++ {
			add = append(add, zero...)
		}
		time = append(add, time...)
	}

	// In Seconds
	ts := binary.BigEndian.Uint64(time)
	var res uint64
	timeEnteredUint := uint64(timeEntered)
	if timeEnteredUint > ts {
		res = timeEnteredUint - ts
	} else {
		res = ts - timeEnteredUint
	}
	if res <= TWELVE_HOURS_S {
		return true
	} else {
		return false
	}
}

func statusIsFedOrAudit(status uint8) bool {
	if status == constants.IDENTITY_FEDERATED_SERVER ||
		status == constants.IDENTITY_AUDIT_SERVER ||
		status == constants.IDENTITY_PENDING_FEDERATED_SERVER ||
		status == constants.IDENTITY_PENDING_AUDIT_SERVER {
		return true
	}
	return false
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

// Only used for marshaling JSON
func statusToJSONString(status uint8) string {
	switch status {
	case constants.IDENTITY_UNASSIGNED:
		return "none"
	case constants.IDENTITY_FEDERATED_SERVER:
		return "federated"
	case constants.IDENTITY_AUDIT_SERVER:
		return "audit"
	case constants.IDENTITY_FULL:
		return "none"
	case constants.IDENTITY_PENDING_FEDERATED_SERVER:
		return "federated"
	case constants.IDENTITY_PENDING_AUDIT_SERVER:
		return "audit"
	case constants.IDENTITY_PENDING_FULL:
		return "none"
	case constants.IDENTITY_SKELETON:
		return "skeleton"
	}
	return "NA"
}
