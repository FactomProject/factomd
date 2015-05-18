// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package simplecoin

import (
    "bytes"
    "encoding/hex"
)

type IAuthorization interface {
    IBlock
    GetAddress() address
    Validate(addr address, transaction []byte) bool     // Validate the address, transaction
}

// In this case, we are simply validating one address to ensure it signed
// this transaction.
type authorize_1 struct {
    signature           [64]byte
}

// Check this signature against this transaction.
func (a authorize_1) Validate(addr address, t []byte) bool {
    return true
}

func (a authorize_1) MarshalText() (text []byte, err error) {
    var out bytes.Buffer
    
    WriteNumber8(&out, uint8(0))            // Type Zero Authorization
    out.WriteString("\n")
    out.WriteString(hex.EncodeToString(a.signature[:]))
    out.WriteString("\n")
    
    return out.Bytes(), nil
}

// We need an index into m.  We could search, but that could make transaction
// processing time slow.
type sign struct {
    index               int                 // Index into m for this signature
    authorization       IAuthorization      // The authorization to test
}

type authorize_2 struct {
    m                   int                 // Total sigatures possible
    n                   int                 // Number signatures required
    m_addresses         []address           // m addresses
    n_signatures        []sign              // n sigatures. 
}

func (a authorize_2) MarshalText() (text []byte, err error) {
    var out bytes.Buffer
    
    WriteNumber8(&out, uint8(1))            // Type Zero Authorization
    out.WriteString("\n")
    WriteNumber16(&out, uint16(a.m))
    out.WriteString("\n")
    WriteNumber16(&out, uint16(a.n))
    out.WriteString("\n")
    for i:=0; i<a.m; i++ {
        out.WriteString(hex.EncodeToString(a.m_addresses[i].Bytes()))
        out.WriteString("\n")
    }
    for i:=0; i<a.n; i++ {
        WriteNumber16(&out,uint16(a.n_signatures[i].index))
        out.WriteString("\n")        
        a.n_signatures[i].authorization.MarshalText()
    }
    
    return out.Bytes(), nil
}



// We are going to require all the sigatures to be valid, if they are provided.
// We will only check n signatures, even if the user provides more.
func (a authorize_2) Validate(addr address, t []byte) bool {
    if len(a.n_signatures) < a.n {                    // Gotta have at least n signatures
        return false
    }
    
    // Marshal Binary, Hash, and make sure this signature matches 
    // The address passed to us.  TODO
    
    for i:=0 ; i < a.n; i++ {
        addr2 := a.m_addresses[a.n_signatures[i].index]
        if !a.n_signatures[i].authorization.Validate(addr2,t) {
            return false
        }
    }

    return true
}
            