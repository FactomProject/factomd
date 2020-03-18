// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identity

import (
	"bytes"
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
)

//https://github.com/FactomProject/FactomDocs/blob/master/Identity.md

// AnchorSigningKeySort is a slice of AnchorSigningKeys
// sort.Sort interface implementation
type AnchorSigningKeySort []AnchorSigningKey

// Len returns the length of the slice
func (p AnchorSigningKeySort) Len() int {
	return len(p)
}

// Swap swaps the data at the input indices 'i' and 'j' in the slice
func (p AnchorSigningKeySort) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// Less returns true if the data at the ith index is less than the data at the jth index
func (p AnchorSigningKeySort) Less(i, j int) bool {
	return bytes.Compare(p[i].SigningKey[:], p[j].SigningKey[:]) < 0
}

// AnchorSigningKey
type AnchorSigningKey struct {
	BlockChain string                 `json:"blockchain"`
	KeyLevel   byte                   `json:"level"`
	KeyType    byte                   `json:"keytype"`
	SigningKey primitives.ByteSlice20 `json:"key"` //if bytes, it is hex
}

var _ interfaces.BinaryMarshallable = (*AnchorSigningKey)(nil)

// RandomAnchorSigningKey returns a new AnchorSigningKey with random starting values
func RandomAnchorSigningKey() *AnchorSigningKey {
	ask := new(AnchorSigningKey)

	ask.BlockChain = random.RandomString()
	ask.KeyLevel = random.RandByte()
	ask.KeyType = random.RandByte()
	copy(ask.SigningKey[:], random.RandByteSliceOfLen(20))

	return ask
}

// IsSameAs returns true iff the input object is identical to this object
func (p *AnchorSigningKey) IsSameAs(b *AnchorSigningKey) bool {
	if p.BlockChain != b.BlockChain {
		return false
	}
	if p.KeyLevel != b.KeyLevel {
		return false
	}
	if p.KeyType != b.KeyType {
		return false
	}
	if primitives.AreBytesEqual(p.SigningKey[:], b.SigningKey[:]) == false {
		return false
	}
	return true
}

// MarshalBinary marshals this object
func (p *AnchorSigningKey) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "AnchorSigningKey.MarshalBinary err:%v", *pe)
		}
	}(&err)
	buf := primitives.NewBuffer(nil)

	err = buf.PushString(p.BlockChain)
	if err != nil {
		return nil, err
	}

	err = buf.PushByte(p.KeyLevel)
	if err != nil {
		return nil, err
	}
	err = buf.PushByte(p.KeyType)
	if err != nil {
		return nil, err
	}

	err = buf.Push(p.SigningKey[:])
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinaryData unmarshals the input data into this object
func (p *AnchorSigningKey) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	newData = data
	buf := primitives.NewBuffer(data)

	p.BlockChain, err = buf.PopString()
	if err != nil {
		return
	}
	p.KeyLevel, err = buf.PopByte()
	if err != nil {
		return
	}
	p.KeyType, err = buf.PopByte()
	if err != nil {
		return
	}
	h, err := buf.PopLen(20)
	if err != nil {
		return
	}
	copy(p.SigningKey[:], h)

	newData = buf.DeepCopyBytes()
	return
}

// UnmarshalBinary unmarshals the input data into this object
func (p *AnchorSigningKey) UnmarshalBinary(data []byte) error {
	_, err := p.UnmarshalBinaryData(data)
	return err
}
