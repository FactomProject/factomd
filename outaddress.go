// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// Output object for a Simplecoin transaction.   contains an amount
// and the destination address.

package simplecoin

import (
    "bytes"
)

type IOutAddress interface {
    IBlock
    GetAmount() uint64
    GetAddress() IAddress
}

type OutAddress struct {
    IOutAddress
    amount      uint64
    address     IAddress
}

func (oa OutAddress) GetAmount() uint64 {
    return oa.amount
}

func (oa OutAddress) GetAddress() IAddress {
    return oa.address
}

func (oa OutAddress) MarshalText() ([]byte, error) {
    var out bytes.Buffer

    out.WriteString("out   ")
    WriteNumber64(&out, oa.amount)
    out.WriteString(" ")
    
    text, _ := oa.address.MarshalText()
    out.Write(text)
    
    return out.Bytes(), nil
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