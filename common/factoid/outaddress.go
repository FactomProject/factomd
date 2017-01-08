// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Output object for a factoid transaction.   contains an amount
// and the destination address.

package factoid

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type OutAddress struct {
	TransAddress
}

var _ interfaces.IOutAddress = (*OutAddress)(nil)

func (b OutAddress) String() string {
	txt, err := b.CustomMarshalText()
	if err != nil {
		return "<error>"
	}
	return string(txt)
}

func (oa OutAddress) GetName() string {
	return "out"
}

func (a OutAddress) CustomMarshalText() (text []byte, err error) {
	return a.CustomMarshalText2("output")
}

/******************************
 * Helper functions
 ******************************/

func NewOutAddress(address interfaces.IAddress, amount uint64) interfaces.IOutAddress {
	ta := new(OutAddress)
	ta.Amount = amount
	ta.Address = address
	ta.UserAddress = primitives.ConvertFctAddressToUserStr(address)
	return ta
}
