// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
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

var _ IInAddress = (*InAddress)(nil)

func (w1 InAddress)GetDBHash() IHash {
    return Sha([]byte("InAddress"))
}

func (w1 InAddress)GetNewInstance() IBlock {
    return new(InAddress)
}

func (oa InAddress) GetName() string {
	return "in"
}

/******************************
 * Helper functions
 ******************************/

func NewInAddress( address IAddress, amount uint64) IInAddress {
	oa := new(InAddress)
	oa.amount = amount
	oa.address = address
	return oa
}
