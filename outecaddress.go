// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package simplecoin

import (
    "bytes"
)

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

var _ IOutECAddress = (*OutECAddress)(nil)

func (OutECAddress)GetDBHash() IHash {
    return Sha([]byte("OutECAddress"))
}

func (w1 OutECAddress)GetNewInstance() IBlock {
    return new(OutECAddress)
}

func (oa OutECAddress) GetName() string {
	return "outEC"
}

func (a OutECAddress) MarshalText() (text []byte, err error) {
    var out bytes.Buffer
    out.WriteString("ec: ")
    a.MarshalText2(&out)
    out.WriteString("\n")
    return out.Bytes(), nil
}
/******************************
 * Helper functions
 ******************************/

func NewOutECAddress( address IAddress, amount uint64 ) IOutAddress {
	oa := new(OutECAddress)
	oa.amount = amount
	oa.address = address
	return oa
}
