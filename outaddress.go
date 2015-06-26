// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Output object for a factoid transaction.   contains an amount
// and the destination address.

package factoid

type IOutAddress interface {
	ITransAddress
}

type OutAddress struct {
	TransAddress
}

var _ IOutAddress = (*OutAddress)(nil)

func (b OutAddress) String() string {
	txt, err := b.MarshalText()
	if err != nil {
		return "<error>"
	}
	return string(txt)
}

func (OutAddress) GetDBHash() IHash {
	return Sha([]byte("OutAddress"))
}

func (w1 OutAddress) GetNewInstance() IBlock {
	return new(OutAddress)
}

func (oa OutAddress) GetName() string {
	return "out"
}

func (a OutAddress) MarshalText() (text []byte, err error) {
    return a.MarshalText2("output") 
}

/******************************
 * Helper functions
 ******************************/

func NewOutAddress(address IAddress, amount uint64) IOutAddress {
	oa := new(OutAddress)
	oa.amount = amount
	oa.address = address
	return oa
}
