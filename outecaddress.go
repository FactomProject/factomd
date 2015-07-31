// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid

// Entry Credit Addresses are the same as Addresses in factoid
// They just get printed out differently when we output them in
// human readable form.
//
// Entry Credit Addresses are always outputs.

type IOutECAddress interface {
	ITransAddress
}

type OutECAddress struct {
	TransAddress
}

var _ IOutECAddress = (*OutECAddress)(nil)

func (b OutECAddress) String() string {
	txt, err := b.CustomMarshalText()
	if err != nil {
		return "<error>"
	}
	return string(txt)
}

func (OutECAddress) GetDBHash() IHash {
	return Sha([]byte("OutECAddress"))
}

func (w1 OutECAddress) GetNewInstance() IBlock {
	return new(OutECAddress)
}

func (oa OutECAddress) GetName() string {
	return "outEC"
}

func (a OutECAddress) CustomMarshalText() (text []byte, err error) {
	return a.CustomMarshalText2("ecoutput")
}

/******************************
 * Helper functions
 ******************************/

func NewOutECAddress(address IAddress, amount uint64) IOutAddress {
	oa := new(OutECAddress)
	oa.amount = amount
	oa.address = address
	return oa
}
