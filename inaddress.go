// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Input object for a Simplecoin transaction.   contains an amount
// and the destination address.

package simplecoin

import (
    "bytes"
)

type IInAddress interface {
	ITransAddress
}

type InAddress struct {
	TransAddress
}

var _ IInAddress = (*InAddress)(nil)

func (i InAddress)GetDBHash() IHash {
    return Sha([]byte("InAddress"))
}

func (i InAddress)GetNewInstance() IBlock {
    return new(InAddress)
}


func (a InAddress) MarshalText() (text []byte, err error) {
    var out bytes.Buffer
    out.WriteString("input: ")
    a.MarshalText2(&out)
    return out.Bytes(), nil
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
