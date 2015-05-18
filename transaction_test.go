// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package simplecoin


import (
    "testing"
    "fmt"
)

var adr1 = [ADDRESS_LENGTH]byte{ 
     0x61, 0xe3, 0x8c, 0x0a, 0xb6, 0xf1, 0xb3, 0x72,  0xc1, 0xa6, 0xa2, 0x46, 0xae, 0x63, 0xf7, 0x4f,
     0x93, 0x1e, 0x83, 0x65, 0xe1, 0x5a, 0x08, 0x9c,  0x68, 0xd6, 0x19, 0x00, 0x00, 0x00, 0x00, 0x00,
}

func TestTransaction(t *testing.T) {
    nb := transaction{}.NewBlock()
    cb := nb.(*transaction)
    
    addr := new(address)
    addr.SetBytes(adr1[:])
    for i:=0;i<10;i++ {
        cb.inputs = cb.AddInput(addr)
        addr2 := new(address)
        addr2.SetHash(Sha(addr.Bytes()))
        addr = addr2    
    }
    
    bytes,_ := nb.MarshalText()
    fmt.Printf("Transaction:\n%slen: %d\n", string(bytes), len(bytes))
}