// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

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
	SetHash(IHash)
}

type Address struct {
	IAddress
	theBytes IHash
}

func (a Address) MarshalBinary() ([]byte, error) {

	data, err := a.theBytes.MarshalBinary()

	return data, err
}

func (cb *Address) NewBlock() IBlock {
	blk := new(Address)
	return blk
}

func (a *Address) SetBytes(b []byte) {
	if a.theBytes == nil {
		a.theBytes = Hash{}.NewBlock().(IHash)
	}
	a.theBytes.SetBytes(b)
}

func (a Address) MarshalText() (text []byte, err error) {
	var out bytes.Buffer
	addr := hex.EncodeToString(a.theBytes.Bytes())
	out.WriteString("addr  ")
	out.WriteString(addr)
	out.WriteString("\n")
	return out.Bytes(), nil
}

func (a Address) Bytes() []byte {
	return a.theBytes.Bytes()
}

func (a *Address) SetHash(h IHash) {
	a.theBytes = h
}
