// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
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
    IBlock
}

type outECAddress struct {
    OutAddress
}

func (oa outECAddress) MarshalText() ([]byte, error) {
    var out bytes.Buffer
    
    out.WriteString("ec_out ")
    WriteNumber64(&out, oa.amount)
    out.WriteString(" ")
    
    text, _ := oa.address.MarshalText()
    out.Write(text)
    
    return out.Bytes(), nil
} 

/******************************
 * Helper functions
 ******************************/

func NewOutECAddress(amount uint64, address IAddress) IOutAddress {
    oa := new(outECAddress)
    oa.amount = amount
    oa.address = address
    return oa
}