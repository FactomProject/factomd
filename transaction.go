// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package simplecoin

import (
    // "fmt"
    "bytes"
)

type ITransaction interface {
    IBlock
}

type transaction struct {
    ITransaction
    inputs              [] IAddress
    outputs             [] IOutAddress
    outECs              [] IOutECAddress
    authorizations      [] IAuthorization
}
    
var _ IBlock = (*transaction)(nil)

func (cb transaction) NewBlock() (IBlock) {
    blk := new (transaction)
    return blk
}

func (cb *transaction) AddInput(input IAddress) {
      if(cb.inputs == nil) {
          cb.inputs = make([]IAddress,0,5)
      }
      cb.inputs = append(cb.inputs, input)
}

func (cb *transaction) AddOutput(amount uint64, input IAddress) {
    if(cb.outputs == nil) {
        cb.outputs = make([]IOutAddress,0,5)
    }
    out := NewOutAddress(amount, input)
    cb.outputs = append(cb.outputs, out)
    
}

func (cb *transaction) AddECOutput(amount uint64, input IAddress)  {
    if(cb.outECs == nil) {
        cb.outECs = make([]IOutECAddress,0,5)
    }
    out := NewOutECAddress(amount, input)
    cb.outECs = append(cb.outECs, out)
    
}

func (cb *transaction) AddAuthorization(auth IAuthorization) {
    if(cb.authorizations == nil) {
        cb.authorizations = make([]IAuthorization,0,5)
    }
    cb.authorizations = append(cb.authorizations, auth)
}    

func (cb transaction) MarshalText() (text []byte, err error) {
   var out bytes.Buffer
    
    out.WriteString("in  ")
    WriteNumber16(&out, uint16( len(cb.inputs) ))
    out.WriteString("\nout ")
    WriteNumber16(&out, uint16( len(cb.outputs)))
    out.WriteString("\nec  ")
    WriteNumber16(&out, uint16( len(cb.outECs)))
    out.WriteString("\n")
    
    for  _,address := range cb.inputs {
        text, _ := address.MarshalText()
        out.Write(text)
    }
    for _,address := range cb.outputs {
        text, _ := address.MarshalText()
        out.Write(text)
    }
    for _,ecaddress := range cb.outECs {
        text, _ := ecaddress.MarshalText()
        out.Write(text)
    }
    for _,auth := range cb.authorizations {
        text, _ := auth.MarshalText()
        out.Write(text)
    }
    
    return out.Bytes(), nil
}