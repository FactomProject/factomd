// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package simplecoin

// Entry Credit Addresses are the same as Addresses in Simplecoin
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

func (oa OutECAddress) GetName() string {
	return "outEC"
}

/******************************
 * Helper functions
 ******************************/

func NewOutECAddress(amount uint64, address IAddress) IOutAddress {
	oa := new(OutECAddress)
	oa.amount = amount
	oa.address = address
	return oa
}
