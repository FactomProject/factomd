// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package simplecoin

import (
    "bytes"
    "encoding/hex"
    "fmt"
    "encoding/binary"
//    "github.com/agl/ed25519"
)

/**************************
 * IAuthorization
 **************************/


type IAuthorization interface {
    IBlock
    Validate(addr IAddress, transaction []byte) bool     // Validate the address, transaction
}

/**************************
 * Authorize_1
 **************************/


// In this case, we are simply validating one address to ensure it signed
// this transaction.
type authorize_1 struct {
    IAuthorization
    signature           []byte
}

func (a authorize_1) MarshalBinary() (data []byte, err error) {
    var out bytes.Buffer
    
    out.WriteByte(byte(1))              // The First Authorization method
    out.Write(a.signature)
    
    return out.Bytes(), nil
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


/*******************
 * sign
 *******************/

// We need an index into m.  We could search, but that could make transaction
// processing time slow.
type sign struct {
    index               int                 // Index into m for this signature
    authorization       IAuthorization      // The authorization to test
}

func (s sign) MarshalBinary() ([]byte, error) {
    var out bytes.Buffer
    
    binary.Write(&out, binary.BigEndian, uint16(s.index))    
    data,err := s.authorization.MarshalBinary()
    out.Write(data)
    
    return out.Bytes(),err
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

/************************
 * Authorize 2
 ************************/

// Type 2 authorization blocks implement multisig
// m of n
// Must have m addresses from which to choose, no fewer, no more
// Must have n authorization blocks, no fewer no more.
// NOTE: This does mean you can have a multisig nested in a 
// multisig.  It just works.

type authorize_2 struct {
    IAuthorization
    n                   int                 // Number signatures required
    m                   int                 // Total sigatures possible
    m_addresses         []IAddress          // m addresses
    n_signatures        []sign              // n sigatures. 
}

func (a authorize_2) MarshalBinary() ([]byte, error) {
    var out bytes.Buffer
    
    binary.Write(&out, binary.BigEndian, uint8(2))    
    binary.Write(&out, binary.BigEndian, uint16(a.n))    
    binary.Write(&out, binary.BigEndian, uint16(a.m))    
    for i:=0;i<a.m;i++ {
        data,err := a.m_addresses[i].MarshalBinary()
        if err != nil {
            return nil, err
        }
        out.Write(data)
    }
    for i:=0;i<a.n;i++ {
        data,err := a.n_signatures[i].MarshalBinary()
        if err != nil {
            return nil, err
        }
        out.Write(data)
    }
    
    return out.Bytes(),nil
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
// We will expect only n signatures if only n are required.
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
    a.signature = make([]byte,len(sign),len(sign))
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