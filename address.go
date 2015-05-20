// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package simplecoin

import (
    // "fmt"
    "bytes"
    "encoding/hex"

)

type IAddress interface {
    IBlock
    Bytes() []byte
    SetBytes([]byte)
    SetHash(IHash)
}

type address struct {
    IAddress
    theBytes   IHash
}

func (a address) MarshalBinary() ( []byte,  error) {
    
    data,err := a.theBytes.MarshalBinary()
    
    return data,err
}

func (cb *address) NewBlock() (IBlock) {
    blk := new (address)
    return blk
}

func (a *address) SetBytes(b []byte) {
    if(a.theBytes == nil) {
        a.theBytes = Hash{}.NewBlock().(IHash)
    }
    a.theBytes.SetBytes(b)
}

func (a address) MarshalText() (text []byte, err error) {
    var out bytes.Buffer
    addr := hex.EncodeToString(a.theBytes.Bytes())
    out.WriteString("addr  ")
    out.WriteString(addr)
    out.WriteString("\n")
    return out.Bytes(),nil
}

func (a address) Bytes() ([]byte) {
    return a.theBytes.Bytes()
}

func (a *address) SetHash(h IHash) {
    a.theBytes = h
}