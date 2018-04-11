// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identity

import (
	"bytes"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
)

//https://github.com/FactomProject/FactomDocs/blob/master/Identity.md

// sort.Sort interface implementation
type AnchorSigningKeySort []AnchorSigningKey

func (p AnchorSigningKeySort) Len() int {
	return len(p)
}
func (p AnchorSigningKeySort) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
func (p AnchorSigningKeySort) Less(i, j int) bool {
	return bytes.Compare(p[i].SigningKey[:], p[j].SigningKey[:]) < 0
}

type AnchorSigningKey struct {
	BlockChain string                 `json:"blockchain"`
	KeyLevel   byte                   `json:"level"`
	KeyType    byte                   `json:"keytype"`
	SigningKey primitives.ByteSlice20 `json:"key"` //if bytes, it is hex
}

var _ interfaces.BinaryMarshallable = (*AnchorSigningKey)(nil)

func RandomAnchorSigningKey() *AnchorSigningKey {
	ask := new(AnchorSigningKey)

	ask.BlockChain = random.RandomString()
	ask.KeyLevel = random.RandByte()
	ask.KeyType = random.RandByte()
	copy(ask.SigningKey[:], random.RandByteSliceOfLen(20))

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
	if primitives.AreBytesEqual(e.SigningKey[:], b.SigningKey[:]) == false {
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

	err = buf.Push(e.SigningKey[:])
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
	h, err := buf.PopLen(20)
	if err != nil {
		return
	}
	copy(e.SigningKey[:], h)

	newData = buf.DeepCopyBytes()
	return
}

func (e *AnchorSigningKey) UnmarshalBinary(p []byte) error {
	_, err := e.UnmarshalBinaryData(p)
	return err
}
