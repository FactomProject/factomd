// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid_test

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"testing"

	"github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// Random first "address".  It isn't a real one, but one we are using for now.
var adr1 = [constants.ADDRESS_LENGTH]byte{
	0x61, 0xe3, 0x8c, 0x0a, 0xb6, 0xf1, 0xb3, 0x72, 0xc1, 0xa6, 0xa2, 0x46, 0xae, 0x63, 0xf7, 0x4f,
	0x93, 0x1e, 0x83, 0x65, 0xe1, 0x5a, 0x08, 0x9c, 0x68, 0xd6, 0x19, 0x00, 0x00, 0x00, 0x00, 0x00,
}

type zeroReader struct{}

var zero zeroReader

var r *rand.Rand

func (zeroReader) Read(buf []byte) (int, error) {
	//if r==nil { r = rand.New(rand.NewSource(time.Now().Unix())) }
	if r == nil {
		r = rand.New(rand.NewSource(1))
	}
	for i := range buf {
		buf[i] = byte(r.Int())
	}
	return len(buf), nil
}

func nextAddress() interfaces.IAddress {

	public, _, _ := ed25519.GenerateKey(zero)

	addr := new(Address)
	addr.SetBytes(public[:])
	return addr
}

func nextSig() []byte {
	// Get me a private key.
	public, _, _ := ed25519.GenerateKey(zero)

	return public[:]
}

func nextAuth2() interfaces.IRCD {
	if r == nil {
		r = rand.New(rand.NewSource(1))
	}
	n := r.Int()%4 + 1
	m := r.Int()%4 + n
	addresses := make([]interfaces.IAddress, m, m)
	for j := 0; j < m; j++ {
		addresses[j] = nextAddress()
	}

	rcd, _ := NewRCD_2(n, m, addresses)
	return rcd
}

var nb interfaces.IBlock

func getSignedTrans() interfaces.IBlock {

	if nb != nil {
		return nb
	}

	nb = new(Transaction)
	t := nb.(*Transaction)

	for i := 0; i < 5; i++ {
		t.AddInput(nextAddress(), uint64(rand.Int63n(10000000000)))
	}

	for i := 0; i < 3; i++ {
		t.AddOutput(nextAddress(), uint64(rand.Int63n(10000000000)))
	}

	for i := 0; i < 3; i++ {
		t.AddECOutput(nextAddress(), uint64(rand.Int63n(10000000)))
	}

	for i := 0; i < 3; i++ {
		sig := NewRCD_1(nextSig())
		t.AddAuthorization(sig)
	}

	for i := 0; i < 2; i++ {

		t.AddAuthorization(nextAuth2())
	}

	return nb
}

// This test prints bunches of stuff that must be visually checked.
// Mostly we keep it commented out.
func TestTransaction(test *testing.T) {
	nb = getSignedTrans()
	bytes, _ := nb.CustomMarshalText()
	fmt.Printf("Transaction:\n%slen: %d\n", string(bytes), len(bytes))
	fmt.Println("\n---------------------------------------------------------------------")
}

func Test_Address_MarshalUnMarshal(test *testing.T) {
	a := nextAddress()
	adr, err := a.MarshalBinary()
	if err != nil {
		primitives.Prtln(err)
		test.Fail()
	}
	_, err = a.UnmarshalBinaryData(adr)
	if err != nil {
		primitives.Prtln(err)
		test.Fail()
	}
}

func Test_Multisig_MarshalUnMarshal(test *testing.T) {
	rcd := nextAuth2()
	auth2, err := rcd.MarshalBinary()
	if err != nil {
		primitives.Prtln(err)
		test.Fail()
	}

	_, err = rcd.UnmarshalBinaryData(auth2)

	if err != nil {
		primitives.Prtln(err)
		test.Fail()
	}
}

func Test_Transaction_MarshalUnMarshal(test *testing.T) {

	getSignedTrans()                // Make sure we have a signed transaction
	data, err := nb.MarshalBinary() // Marshal our signed transaction
	if err != nil {                 // If we have an error, print our stack
		primitives.Prtln(err) //   and fail our test
		test.Fail()
	}

	xb := new(Transaction)

	err = xb.UnmarshalBinary(data) // Now Unmarshal

	if err != nil {
		primitives.Prtln(err)
		test.Fail()
		return
	}

	//     txt1,_ := xb.CustomMarshalText()
	//     txt2,_ := nb.CustomMarshalText()
	//     primitives.Prtln(string(txt1))
	//     primitives.Prtln(string(txt2))

	if xb.IsEqual(nb) != nil {
		primitives.Prtln("Trans\n", nb, "Unmarshal Trans\n", xb)
		test.Fail()
	}

}

func Test_ValidateAmounts(test *testing.T) {
	var zero uint64
	_, err := ValidateAmounts(zero - 1)
	if err != nil {
		test.Failed()
	}
	_, err = ValidateAmounts(1, 2, 3, 4, 5, zero-1)
	if err != nil {
		test.Failed()
	}
	_, err = ValidateAmounts(0x6FFFFFFFFFFFFFFF, 1)
	if err != nil {
		test.Failed()
	}
	_, err = ValidateAmounts(1, 0x6FFFFFFFFFFFFFFF, 1)
	if err != nil {
		test.Failed()
	}
}

func TestUnmarshalTransaction(t *testing.T) {
	str := "02014f8a7fcd1b000000"
	h, err := hex.DecodeString(str)
	if err != nil {
		t.Errorf("%v", err)
	}
	tr := new(Transaction)
	rest, err := tr.UnmarshalBinaryData(h)
	if err != nil {
		t.Errorf("%v", err)
	}
	if len(rest) > 0 {
		t.Errorf("Returned too much data - %x", rest)
	}

	if tr.GetSigHash().String() != "e321605afa458333cdded91644b0d9a21b4325bb3340b85a943974bf70aa1e99" {
		t.Errorf("Invalid SigHash - %v vs %v", tr.GetSigHash().String(), "")
	}
	if tr.GetFullHash().String() != "e321605afa458333cdded91644b0d9a21b4325bb3340b85a943974bf70aa1e99" {
		t.Errorf("Invalid FullHash - %v vs %v", tr.GetFullHash().String(), "")
	}

	str = "02014f8a851657010001e397a1607d4f56c528ab09da5bbf7b37b0b453f43db303730e28e9ebe02657dff431d4f7dfaf840017ef7a21d1a616d65e6b73f3c6a7ad5c49340a6c2592872020ec60767ff00d7d01a5be79b6ada79c0af4d6b7f91234ff321f3b647ed01e02ccbbc0fe9dcc63293482f22455b9756ee4b4db411a5d00e31b689c1bd1abe1d1e887cf4c52e67fc51fe4d9594c24643a91009c6ea91701b5b6df240248c2f39453162b61d71b982701"

	h, err = hex.DecodeString(str)
	if err != nil {
		t.Errorf("%v", err)
	}
	tr = new(Transaction)
	rest, err = tr.UnmarshalBinaryData(h)
	if err != nil {
		t.Errorf("%v", err)
	}
	if len(rest) > 0 {
		t.Errorf("Returned too much data - %x", rest)
	}

	if tr.GetSigHash().String() != "f1d9919829fa71ce18caf1bd8659cce8a06c0026d3f3fffc61054ebb25ebeaa0" {
		t.Errorf("Invalid SigHash - %v vs %v", tr.GetSigHash().String(), "")
	}
	if tr.GetFullHash().String() != "c3d09d10693eb867e2bd0a503746df370403c9451ae91a363046f2a68529c2fd" {
		t.Errorf("Invalid FullHash - %v vs %v", tr.GetFullHash().String(), "")
	}
}
