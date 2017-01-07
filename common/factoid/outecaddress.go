// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid

// Entry Credit Addresses are the same as Addresses in factoid
// They just get printed out differently when we output them in
// human readable form.
//
// Entry Credit Addresses are always outputs.

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type OutECAddress struct {
	TransAddress
}

var _ interfaces.IOutECAddress = (*OutECAddress)(nil)

func (b OutECAddress) String() string {
	txt, err := b.CustomMarshalText()
	if err != nil {
		return "<error>"
	}
	return string(txt)
}

func (oa OutECAddress) GetName() string {
	return "outEC"
}

func (a OutECAddress) CustomMarshalText() (text []byte, err error) {
	return a.CustomMarshalTextEC2("ecoutput")
}

/******************************
 * Helper functions
 ******************************/

func NewOutECAddress(address interfaces.IAddress, amount uint64) interfaces.IOutECAddress {
	ta := new(OutECAddress)
	ta.Amount = amount
	ta.Address = address
	ta.UserAddress = primitives.ConvertECAddressToUserStr(address)
	return ta
}
