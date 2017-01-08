// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Input object for a factoid transaction.   contains an amount
// and the destination address.

package factoid

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type InAddress struct {
	TransAddress
}

var _ interfaces.IInAddress = (*InAddress)(nil)

func (b InAddress) String() string {
	txt, err := b.CustomMarshalText()
	if err != nil {
		return "<error>"
	}
	return string(txt)
}

func (a InAddress) CustomMarshalText() (text []byte, err error) {
	return a.CustomMarshalText2("input")
}

/******************************
 * Helper functions
 ******************************/

func NewInAddress(address interfaces.IAddress, amount uint64) interfaces.IInAddress {
	ta := new(InAddress)
	ta.Amount = amount
	ta.Address = address
	//  at this point we know this address is an EC address.
	//  so fill useraddress with a factoid formatted human readable address
	ta.UserAddress = primitives.ConvertFctAddressToUserStr(address)
	return ta
}
