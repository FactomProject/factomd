// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Input object for a factoid transaction.   contains an amount
// and the destination address.

package factoid

type IInAddress interface {
	ITransAddress
}

type InAddress struct {
	TransAddress
}

var _ IInAddress = (*InAddress)(nil)

func (b InAddress) String() string {
	txt, err := b.CustomMarshalText()
	if err != nil {
		return "<error>"
	}
	return string(txt)
}

func (InAddress) GetDBHash() IHash {
	return Sha([]byte("InAddress"))
}

func (i InAddress) GetNewInstance() IBlock {
	return new(InAddress)
}

func (a InAddress) CustomMarshalText() (text []byte, err error) {
	return a.CustomMarshalText2("input")
}

/******************************
 * Helper functions
 ******************************/

func NewInAddress(address IAddress, amount uint64) IInAddress {
	oa := new(InAddress)
	oa.amount = amount
	oa.address = address
	return oa
}
