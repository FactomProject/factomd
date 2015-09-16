// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Structure for managing Addresses.  Addresses can be literally the public
// key for holding some value, requiring a signature to release that value.
// Or they can be a Hash of an Authentication block.  In which case, if the
// the authentication block is valid, the value is released (and we can
// prove this is okay, because the hash of the authentication block must
// match this address.

package factoid

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

var _ = fmt.Println

type IAddress interface {
	IHash
}

type Address struct {
	Hash // Since Hash implements IHash, and IAddress is just a
} // alais for IHash, then I don't have to (nor can I) make
// Address implement IAddress... Weird, but that's the way it is.

var _ IAddress = (*Address)(nil)

func (b Address) String() string {
	txt, err := b.CustomMarshalText()
	if err != nil {
		return "<error>"
	}
	return string(txt)
}

func (Address) GetDBHash() IHash {
	return Sha([]byte("Address"))
}

func (a Address) CustomMarshalText() (text []byte, err error) {
	var out bytes.Buffer
	addr := hex.EncodeToString(a.Bytes())
	out.WriteString("addr  ")
	out.WriteString(addr)
	return out.Bytes(), nil
}

func NewAddress(b []byte) IAddress {
	a := new(Address)
	a.SetBytes(b)
	return a
}

func CreateAddress(hash IHash) IAddress {
	return NewAddress(hash.Bytes())
}
