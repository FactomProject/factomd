// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package simplecoin

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	//    "github.com/agl/ed25519"
)

/**************************
 * IRCD  Interface for Redeem Condition Datastructures (RCD)
 * 
 * https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#factoid-transaction
 **************************/

type IRCD interface {
	IBlock
	Validate(addr IAddress, transaction []byte) bool // Validate the address, transaction
}

/**************************
 * RCD_1 Simple Signature
 **************************/

// In this case, we are simply validating one address to ensure it signed
// this transaction.
type RCD_1 struct {
	IRCD
	signature []byte
}

var _ IRCD = (*RCD_1)(nil)

func (a1 RCD_1) IsEqual(addr IBlock) bool {
    a2, ok := addr.(*   RCD_1)
    if 
        !ok ||                                              // Not the right kind of IBlock
        bytes.Compare(a1.signature, a2.signature) != 0 {    // Not the right sigature
            return false
    }
        
    return true
}

func (t *RCD_1) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
    
    typ := int8(data[0])
    data = data[1:]
    
    if typ != 1 {
        PrtStk()
        return nil, fmt.Errorf("Bad type byte: %d",typ)
    }
    
    if(len(data)<SIGNATURE_LENGTH) {
        PrtStk()
        return nil, fmt.Errorf("Data source too short to unmarshal a Signature: %d",len(data))
    }
    
    t.signature = make([]byte,SIGNATURE_LENGTH,SIGNATURE_LENGTH)
    copy(t.signature,data[:SIGNATURE_LENGTH])
    
    return data[SIGNATURE_LENGTH:], nil
}

func (a RCD_1) MarshalBinary() (data []byte, err error) {
	var out bytes.Buffer

	out.WriteByte(byte(1)) // The First Authorization method
	out.Write(a.signature)

	return out.Bytes(), nil
}

// Check this signature against this transaction.
func (a RCD_1) Validate(addr IAddress, t []byte) bool {
	return true
}

func (a RCD_1) MarshalText() (text []byte, err error) {
	var out bytes.Buffer
	out.WriteString("RCD 1: ")
	WriteNumber8(&out, uint8(1)) // Type Zero Authorization
	out.WriteString(" ")
	out.WriteString(hex.EncodeToString(a.signature[:]))
	out.WriteString("\n")

	return out.Bytes(), nil
}

/*******************
 * Interface for Sign 
 * 
 * Data structure to support multisig
 *******************/

type ISign interface {
	IBlock
	GetIndex() int
	GetAuthorization() IRCD
}

// We need an index into m.  We could search, but that could make transaction
// processing time slow.
type Sign struct {
	ISign
	index         int            // Index into m for this signature
	authorization IRCD // The authorization to test
}

var _ ISign = (*Sign)(nil)

func (s1 Sign) IsEqual(sig IBlock) bool {
    s2, ok := sig.(*Sign)
    if 
        !ok ||                                          // Not the right kind of IBlock
        s1.index != s2.index ||                         // Not the same index
        !s1.authorization.IsEqual(s2.authorization) {   // Not the right authorization
            return false
    }
        
    return true
}



func (s *Sign) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
    
    if(len(data)<2) {
        return nil, fmt.Errorf("Data source too short to unmarshal a Signature: %d",len(data))
    }
    
    s.index, data = int(binary.BigEndian.Uint16(data[0:2])), data[2:]
    
    s.authorization, data, err = UnmarshalBinaryAuth(data)
    
    return data, nil
}



func (s Sign) GetIndex() int {
	return s.index
}

func (s Sign) GetAuthorization() IRCD {
	return s.authorization
}

func (s Sign) MarshalBinary() ([]byte, error) {
	var out bytes.Buffer

	binary.Write(&out, binary.BigEndian, uint16(s.index))
	data, err := s.authorization.MarshalBinary()
	out.Write(data)

	return out.Bytes(), err
}

func (s Sign) MarshalText() ([]byte, error) {
	var out bytes.Buffer

	out.WriteString("index: ")
	WriteNumber16(&out, uint16(s.index))
	out.WriteString(" ")
	text, _ := s.authorization.MarshalText()
	out.Write(text)

	return out.Bytes(), nil
}

/************************
 * RCD 2
 ************************/

// Type 2 authorization blocks implement multisig
// m of n
// Must have m addresses from which to choose, no fewer, no more
// Must have n authorization blocks, no fewer no more.
// NOTE: This does mean you can have a multisig nested in a
// multisig.  It just works.

type RCD_2 struct {
	IRCD
	m            int        // Number signatures required
	n            int        // Total sigatures possible
	n_addresses  []IAddress // m addresses
	m_signatures []ISign    // n sigatures.
}

var _ IRCD = (*RCD_2)(nil)

func (a1 RCD_2) IsEqual(addr IBlock) bool {
    a2, ok := addr.(*   RCD_2)
    if 
        !ok ||                                          // Not the right kind of IBlock
        a1.n != a2.n ||                                 // Size of sig has to match
        a1.m != a2.m ||                                 // Size of sig has to match
        len(a1.n_addresses) != len(a2.n_addresses) ||   // Size of arrays has to match
        len(a1.m_signatures) != len(a2.m_signatures)  { // Size of arrays has to match
            return false
    }
    
    for i,addr := range a1.n_addresses {
        if !addr.IsEqual(a2.n_addresses[i]){
            return false
        }
    }
    for i,sig := range a1.m_signatures {
        if !sig.IsEqual(a2.m_signatures[i]){
            return false
        }
    }
    
    return true
}


func (t *RCD_2) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
    
    typ := int8(data[0])
    data = data[1:]
    if typ!=2 {
        if err != nil { return nil,fmt.Errorf("Bad data fed to RCD_2 UnmarshalBinaryData()") }
    }
    
    t.n, data = int(binary.BigEndian.Uint16(data[0:2])), data[2:]
    t.m, data = int(binary.BigEndian.Uint16(data[0:2])), data[2:]
    
    t.n_addresses = make([]IAddress,t.m,t.m)
    t.m_signatures = make([]ISign,t.n,t.n)
    
    for i,_ := range t.n_addresses {
        t.n_addresses[i] = new(Address)
        data,err = t.n_addresses[i].UnmarshalBinaryData(data)
        if err != nil { return nil,err }
    }

    for i,_ := range t.m_signatures {
        t.m_signatures[i] = new(Sign)
        data,err = t.m_signatures[i].UnmarshalBinaryData(data)
        if err != nil { return nil,err }
    }
    return data, nil
}
    
    
    
func (a RCD_2) MarshalBinary() ([]byte, error) {
	var out bytes.Buffer

	binary.Write(&out, binary.BigEndian, uint8(2))
	binary.Write(&out, binary.BigEndian, uint16(a.n))
	binary.Write(&out, binary.BigEndian, uint16(a.m))
	for i := 0; i < a.m; i++ {
		data, err := a.n_addresses[i].MarshalBinary()
		if err != nil {
			return nil, err
		}
		out.Write(data)
	}
	for i := 0; i < a.n; i++ {
		data, err := a.m_signatures[i].MarshalBinary()
		if err != nil {
			return nil, err
		}
		out.Write(data)
	}

	return out.Bytes(), nil
}

func (a RCD_2) MarshalText() ([]byte, error) {
	var out bytes.Buffer

	WriteNumber8(&out, uint8(2)) // Type 2 Authorization
	out.WriteString("\n n: ")
	WriteNumber16(&out, uint16(a.n))
	out.WriteString(" m: ")
	WriteNumber16(&out, uint16(a.m))
	out.WriteString("\n")
	for i := 0; i < a.m; i++ {
		out.WriteString("  m: ")
		out.WriteString(hex.EncodeToString(a.n_addresses[i].Bytes()))
		out.WriteString("\n")
	}
	for i := 0; i < a.n; i++ {
		text, _ := a.m_signatures[i].MarshalText()
		out.Write(text)
	}

	return out.Bytes(), nil
}

// We are going to require all the sigatures to be valid, if they are provided.
// We will expect only n signatures if only n are required.
func (a RCD_2) Validate(addr IAddress, t []byte) bool {
	if len(a.m_signatures) < a.n { // Gotta have at least n signatures
		return false
	}

	// Marshal Binary, Hash, and make sure this signature matches
	// The address passed to us.  TODO

	for i := 0; i < a.n; i++ {
		addr2 := a.n_addresses[a.m_signatures[i].GetIndex()]
		if !a.m_signatures[i].GetAuthorization().Validate(addr2, t) {
			return false
		}
	}

	return true
}

/***********************
 * Helper Functions
 ***********************/

func UnmarshalBinaryAuth(data []byte) (a IRCD, newData []byte, err error) {
    
    t := data[0] 
    
    var auth IRCD
    switch int(t) {
        case 1:
            auth = new(RCD_1)
        case 2:
            auth = new(RCD_2)
        default:
            PrtStk()
            return nil, nil, fmt.Errorf("Invalid type byte for authorizations %x ", int(t))
    }
    data, err = auth.UnmarshalBinaryData(data)
    return auth, data, err
}

func NewSignature1(sign []byte) (IRCD, error) {
	a := new(RCD_1)
	a.signature = make([]byte, len(sign), len(sign))
	copy(a.signature[:], sign)
	return a, nil
}

func NewSignature2(n int, m int, addresses []IAddress, signs []ISign) (IRCD, error) {
	if len(addresses) != m {
		return nil, fmt.Errorf("Improper number of addresses.  m = %d n = %d #addresses = %d", m, n, len(addresses))
	}
	if len(signs) != n {
		return nil, fmt.Errorf("Improper number of authorizations.  m = %d n = %d #authorizations = %d", m, n, len(signs))
	}

	au := new(RCD_2)
	au.n = n
	au.m = m

	au.n_addresses = addresses
	au.m_signatures = signs

	return au, nil
}
