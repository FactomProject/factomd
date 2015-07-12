// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Structure for managing Addresses.  Addresses can be literally the public
// key for holding some value, requiring a signature to release that value.
// Or they can be a Hash of an Authentication block.  In which case, if the
// the authentication block is valid, the value is released (and we can
// prove this is okay, because the hash of the authentication block must
// match this address.

package factoid

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"
)

type IAddress interface {
	IHash
}

type Address struct {
	Hash // Since Hash implements IHash, and IAddress is just a
} // alais for IHash, then I don't have to (nor can I) make
// Address implement IAddress... Weird, but that's the way it is.

var _ IAddress = (*Address)(nil)

func (b Address) String() string {
	txt, err := b.MarshalText()
	if err != nil {
		return "<error>"
	}
	return string(txt)
}

func (Address) GetDBHash() IHash {
	return Sha([]byte("Address"))
}

func (a Address) MarshalText() (text []byte, err error) {
	var out bytes.Buffer
	addr := hex.EncodeToString(a.Bytes())
	out.WriteString("addr  ")
	out.WriteString(addr)
	return out.Bytes(), nil
}

func NewAddress(b []byte) IAddress {
    a := new(Address)
    a.SetBytes(b)
    return a
}

func CreateAddress(hash IHash) IAddress {
    return NewAddress(hash.Bytes())
}

func Test_Factoid_Addresses(test *testing.T) {
   
    addr := NewAddress(Sha([]byte("A fake address")).Bytes())
    fmt.Println( addr)
    
    uaddr := ConvertFctAddressToUserStr(addr) 
    fmt.Println(uaddr)

    if !ValidateFUserStr(uaddr) { test.Fail() }
    
    addrBack := ConvertUserStrToAddress(uaddr)
    
    if bytes.Compare(addrBack,addr.Bytes()) != 0 { test.Fail() }
    
    buaddr := []byte(uaddr)
    
    for i,v := range buaddr {
        for j:= uint(0); j<8; j++ {
            if !ValidateFUserStr(string(buaddr)) { test.Fail() }
            buaddr[i] = v^(01<<j)
            if ValidateFUserStr(string(buaddr)) { test.Fail() }
            buaddr[i] = v
        }
    }
}

func Test_Entry_Credit_Addresses(test *testing.T) {
    
    addr := NewAddress(Sha([]byte("A fake address")).Bytes())
    fmt.Println( addr)
    
    uaddr := ConvertECAddressToUserStr(addr) 
    fmt.Println(uaddr)
    
    if !ValidateECUserStr(uaddr) {fmt.Printf("1"); test.Fail() }
    
    addrBack := ConvertUserStrToAddress(uaddr)
    
    if bytes.Compare(addrBack,addr.Bytes()) != 0 {fmt.Printf("2"); test.Fail() }
    
    buaddr := []byte(uaddr)
    
    for i,v := range buaddr {
        for j:= uint(0); j<8; j++ {
            if !ValidateECUserStr(string(buaddr)) { fmt.Printf("3"); test.Fail(); return}
            buaddr[i] = v^(01<<j)
            if ValidateECUserStr(string(buaddr)) { fmt.Printf("4"); test.Fail();return }
            buaddr[i] = v
        }
    }
}
