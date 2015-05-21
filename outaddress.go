// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// Output object for a Simplecoin transaction.   contains an amount
// and the destination address.

package simplecoin

type IOutAddress interface {
	ITransAddress
}

type OutAddress struct {
	TransAddress
}

var _ IOutAddress = (*OutAddress)(nil)

func (oa OutAddress) GetName() string {
	return "out"
}

/******************************
 * Helper functions
 ******************************/

func NewOutAddress(amount uint64, address IAddress) IOutAddress {
	oa := new(OutAddress)
	oa.amount = amount
	oa.address = address
	return oa
}
