// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// Input object for a Simplecoin transaction.   contains an amount
// and the destination address.

package simplecoin

type IInAddress interface {
	ITransAddress
}

type InAddress struct {
	TransAddress
}

func (oa InAddress) GetName() string {
	return "in"
}

/******************************
 * Helper functions
 ******************************/

func NewInAddress(amount uint64, address IAddress) IInAddress {
	oa := new(InAddress)
	oa.amount = amount
	oa.address = address
	return oa
}
