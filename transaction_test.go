// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package simplecoin

import (
	"fmt"
	"github.com/agl/ed25519"
	"math/rand"
	"testing"
)

// Random first "address".  It isn't a real one, but one we are using for now.
var adr1 = [ADDRESS_LENGTH]byte{
	0x61, 0xe3, 0x8c, 0x0a, 0xb6, 0xf1, 0xb3, 0x72, 0xc1, 0xa6, 0xa2, 0x46, 0xae, 0x63, 0xf7, 0x4f,
	0x93, 0x1e, 0x83, 0x65, 0xe1, 0x5a, 0x08, 0x9c, 0x68, 0xd6, 0x19, 0x00, 0x00, 0x00, 0x00, 0x00,
}

type zeroReader struct{}

var r *rand.Rand

func (zeroReader) Read(buf []byte) (int, error) {
    //if r==nil { r = rand.New(rand.NewSource(time.Now().Unix())) }
    if r==nil { r = rand.New(rand.NewSource(1)) }
    for i := range buf {
		buf[i] = byte(r.Int())
	}
	return len(buf), nil
}

var zero zeroReader

func nextAddress() IAddress {
    
    public, _, _ := ed25519.GenerateKey(zero)
        
        addr := new(Address)
        addr.SetBytes(public[:])
        return addr
}


func nextSig() []byte {
    // Get me a private key.
    public, private, _ := ed25519.GenerateKey(zero)
    
    
    bts := Sha(public[:]).Bytes()
    
    sig := ed25519.Sign(private,bts)
    return (*sig)[:]
}


func nextAuth2() IAuthorization {
    if r==nil { r = rand.New(rand.NewSource(1)) }
    n := r.Int()%4 + 1
    m := r.Int()%4 + n
    addresses := make([]IAddress, m, m)
    for j := 0; j < m; j++ {
        addresses[j] = nextAddress()
    }
    signs := make([]ISign, n, n)
    for j := 0; j < n; j++ {
        auth, _ := NewSignature1(nextSig())
        s := new(Sign)
        signs[j] = s
        s.index = j                         // Index into m for this signature
        s.authorization = auth
    }
    sig, _ := NewSignature2(n, m, addresses, signs)
    return sig
}


var nb IBlock

func getSignedTrans() IBlock {
    
    if nb != nil { return nb }
    
    nb = SignedTransaction{}.NewBlock()
    t := nb.(*SignedTransaction)
    
    for i := 0; i < 5; i++ {
        t.AddInput(uint64(rand.Int63n(10000000000)), nextAddress())
    }
    
    for i := 0; i < 3; i++ {
        t.AddOutput(uint64(rand.Int63n(10000000000)), nextAddress())
    }
    
    for i := 0; i < 3; i++ {
        t.AddECOutput(uint64(rand.Int63n(10000000)), nextAddress())
    }
    
    for i := 0; i < 3; i++ {
        sig, _ := NewSignature1(nextSig())
        t.AddAuthorization(sig)
    }
    
    for i := 0; i < 2; i++ {
        
        t.AddAuthorization(nextAuth2())
    }
    
    return nb
}


// This test prints bunches of stuff that must be visually checked.  
// Mostly we keep it commented out.
func xTestTransaction(test *testing.T) {
	nb = getSignedTrans()
	bytes, _ := nb.MarshalText()
	fmt.Printf("Transaction:\n%slen: %d\n", string(bytes), len(bytes))
    fmt.Println("\n---------------------------------------------------------------------")
}

func Test_Address_MarshalUnMarshal(test *testing.T){
    a := nextAddress()
    adr, err := a.MarshalBinary()
    if err != nil {
        Prtln(err)
        test.Fail()
    }
    _,err = a.UnmarshalBinaryData(adr)
    if err != nil {
        Prtln(err)
        test.Fail()
    }
}



func Test_Signatures_MarshalUnMarshal(test *testing.T){
    auth, _ := NewSignature1(nextSig())
    s := new(Sign)
    s.index = 1 // Index into m for this signature
    s.authorization = auth 
    sig, err := s.MarshalBinary()
    if err != nil {
        Prtln(err)
        test.Fail()
    }
    _,err=s.UnmarshalBinaryData(sig)
    if err != nil {
        Prtln(err)
        test.Fail()
    }
}

func Test_Multisig_MarshalUnMarshal(test *testing.T){
    auth := nextAuth2()
    auth2, err := auth.MarshalBinary()
    if err != nil {
        Prtln(err)
        test.Fail()
    }
  
    _,err = auth.UnmarshalBinaryData(auth2)
    
    if err != nil {
        Prtln(err)
        test.Fail()
    }
}

func Test_Transaction_MarshalUnMarshal(test *testing.T) {

    getSignedTrans()                    // Make sure we have a signed transaction
    data, err := nb.MarshalBinary()     // Marshal our signed transaction    
    if err != nil {                     // If we have an error, print our stack
        Prtln(err)                      //   and fail our test
        test.Fail()
    }
    
    xb := new(SignedTransaction)
    
    err = xb.UnmarshalBinary(data)      // Now Unmarshal
    if err != nil {
        Prtln(err)
        test.Fail()
    }
    
//     txt1,_ := xb.MarshalText()
//     txt2,_ := nb.MarshalText()
//     Prtln(string(txt1))
//     Prtln(string(txt2))
    
     if !xb.IsEqual(nb) {
         Prtln(err)
         test.Fail()
     }
    
}
