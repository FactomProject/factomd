// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package simplecoin

import (
    // "fmt"
    "encoding"
    "bytes"
)

type IBlock interface {
    encoding.BinaryMarshaler
    encoding.BinaryUnmarshaler
    encoding.TextMarshaler
    encoding.TextUnmarshaler

    newBlock(...interface{}) (IBlock)
    
}

type ITransaction interface {
    IBlock
}

type transaction struct {
    ITransaction
    inputs              [] IAddress
    outputs             [] IAddress
    eCAddrs             [] IECAddress
    authorizations      [] IAuthorization
}

var _ IBlock = (*transaction)(nil)

func (cb transaction) NewBlock() (IBlock) {
    blk := new (transaction)
    return blk
}

func (cb *transaction) AddInput(input IAddress) ([] IAddress) {
      if(cb.inputs == nil) {
          cb.inputs = make([]IAddress,0,5)
      }
      cb.inputs = append(cb.inputs, input)
      return cb.inputs
}

func (cb transaction) MarshalText() (text []byte, err error) {
   var out bytes.Buffer
    
    WriteNumber16(&out, uint16( len(cb.inputs) ))
    out.WriteString("\n")
    WriteNumber16(&out, uint16( len(cb.outputs)))
    out.WriteString("\n")
    WriteNumber16(&out, uint16( len(cb.eCAddrs)))
    out.WriteString("\n")
    
    for  _,address := range cb.inputs {
        text, _ := address.MarshalText()
        out.Write(text)
    }
    for _,address := range cb.outputs {
        text, _ := address.MarshalText()
        out.Write(text)
    }
    for _,ecaddress := range cb.eCAddrs {
        text, _ := ecaddress.MarshalText()
        out.Write(text)
    }
    for _,auth := range cb.authorizations {
        text, _ := auth.MarshalText()
        out.Write(text)
    }
    
    return out.Bytes(), nil
}