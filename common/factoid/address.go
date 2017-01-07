// Copyright 2017 Factom Foundation
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
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

var _ = fmt.Println

type Address struct {
	primitives.Hash // Since Hash implements interfaces.IHash, and interfaces.IAddress is just a
} // alais for interfaces.IHash, then I don't have to (nor can I) make
// Address implement interfaces.IAddress... Weird, but that's the way it is.

var _ interfaces.IAddress = (*Address)(nil)

/*
func (b *Address) String() string {
	txt, err := b.CustomMarshalText()
	if err != nil {
		return "<error>"
	}
	return string(txt)
}*/

func (a *Address) CustomMarshalText() (text []byte, err error) {
	var out primitives.Buffer
	addr := hex.EncodeToString(a.Bytes())
	out.WriteString("addr  ")
	out.WriteString(addr)
	return out.DeepCopyBytes(), nil
}

func NewAddress(b []byte) interfaces.IAddress {
	a := new(Address)
	a.SetBytes(b)
	return a
}

func CreateAddress(hash interfaces.IHash) interfaces.IAddress {
	return NewAddress(hash.Bytes())
}
