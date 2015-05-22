// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Structure for managing Addresses.  Addresses can be literally the public
// key for holding some value, requiring a signature to release that value.
// Or they can be a Hash of an Authentication block.  In which case, if the
// the authentication block is valid, the value is released (and we can
// prove this is okay, because the hash of the authentication block must
// match this address.

package simplecoin

import (
	// "fmt"
	"bytes"
	"encoding/hex"
)

type IAddress interface {
	IBlock
	Bytes() []byte
	SetBytes([]byte)
	SetHash(IHash)  // Really this could be a Hash or an Address
	GetHash() IHash // Same here.
}

type Address struct {
	IAddress
	address IHash
}

var _ IAddress = (*Address)(nil)

func (t Address) IsEqual(addr IBlock) bool {
	a, ok := addr.(IAddress)
	if !ok || !a.GetHash().IsEqual(t.GetHash()) {
		return false
	}
	return true
}

func (t Address) GetHash() IHash {
	return t.address
}

func (t *Address) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	t.address = new(Hash)
	data, err = t.address.UnmarshalBinaryData(data)
	return data, err
}

func (a Address) MarshalBinary() ([]byte, error) {

	data, err := a.address.MarshalBinary()

	return data, err
}

func (cb *Address) NewBlock() IBlock {
	blk := new(Address)
	return blk
}

func (a *Address) SetBytes(b []byte) {
	if a.address == nil {
		a.address = Hash{}.NewBlock().(IHash)
	}
	a.address.SetBytes(b)
}

func (a Address) MarshalText() (text []byte, err error) {
	var out bytes.Buffer
	addr := hex.EncodeToString(a.address.Bytes())
	out.WriteString("addr  ")
	out.WriteString(addr)
	out.WriteString("\n")
	return out.Bytes(), nil
}

func (a Address) Bytes() []byte {
	return a.address.Bytes()
}

func (a *Address) SetHash(h IHash) {
	a.address = h
}
