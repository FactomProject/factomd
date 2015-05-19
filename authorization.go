// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package simplecoin

import (
    "bytes"
    "encoding/hex"
    "fmt"
)

type IAuthorization interface {
    IBlock
    Validate(addr IAddress, transaction []byte) bool     // Validate the address, transaction
}

// In this case, we are simply validating one address to ensure it signed
// this transaction.
type authorize_1 struct {
    IAuthorization
    signature           [64]byte
}

// Check this signature against this transaction.
func (a authorize_1) Validate(addr IAddress, t []byte) bool {
    return true
}

func (a authorize_1) MarshalText() (text []byte, err error) {
    var out bytes.Buffer
    out.WriteString("Authorize 1: ")    
    WriteNumber8(&out, uint8(1))            // Type Zero Authorization
    out.WriteString(" ")
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

func (s sign) MarshalText() ([]byte, error) {
    var out bytes.Buffer

    out.WriteString("index: ")
    WriteNumber16(&out,uint16(s.index))
    out.WriteString(" ")        
    text, _ := s.authorization.MarshalText()
    out.Write(text)

    return out.Bytes(), nil
}

type authorize_2 struct {
    IAuthorization
    m                   int                 // Total sigatures possible
    n                   int                 // Number signatures required
    m_addresses         []IAddress          // m addresses
    n_signatures        []sign              // n sigatures. 
}

func (a authorize_2) MarshalText() ([]byte, error) {
    var out bytes.Buffer
    
    WriteNumber8(&out, uint8(2))            // Type 2 Authorization
    out.WriteString("\n n: ")
    WriteNumber16(&out, uint16(a.n))
    out.WriteString(" m: ")
    WriteNumber16(&out, uint16(a.m))
    out.WriteString("\n")
    for i:=0; i<a.m; i++ {
        out.WriteString("  m: ")
        out.WriteString(hex.EncodeToString(a.m_addresses[i].Bytes()))
        out.WriteString("\n")
    }
    for i:=0; i<a.n; i++ {
        text,_ := a.n_signatures[i].MarshalText()
        out.Write(text)
    }
    
    return out.Bytes(), nil
}



// We are going to require all the sigatures to be valid, if they are provided.
// We will only check n signatures, even if the user provides more.
func (a authorize_2) Validate(addr IAddress, t []byte) bool {
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
      
      
/***********************
 * Helper Functions
 ***********************/

func NewSignature1( sign []byte ) (IAuthorization, error) {
    a := new(authorize_1)
    copy(a.signature[:], sign)
    return a, nil
}

func NewSignature2(n int, m int, addresses []IAddress,signs []sign) (IAuthorization, error) {
    if len(addresses) != m {
        return nil, fmt.Errorf("Improper number of addresses.  m = %d n = %d #addresses = %d",m,n,len(addresses))
    }
    if len(signs) != n {
        return nil, fmt.Errorf("Improper number of authorizations.  m = %d n = %d #authorizations = %d",m,n,len(signs))
    }

    au := new(authorize_2)
    au.n = n
    au.m = m
    
    au.m_addresses = addresses    
    au.n_signatures = signs
    
    return au, nil
}