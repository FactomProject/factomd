// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package simplecoin

import (
	"fmt"
	"github.com/agl/ed25519"
	"math/rand"
	"testing"
	"time"
)

// Random first "address".  It isn't a real one, but one we are using for now.
var adr1 = [ADDRESS_LENGTH]byte{
	0x61, 0xe3, 0x8c, 0x0a, 0xb6, 0xf1, 0xb3, 0x72, 0xc1, 0xa6, 0xa2, 0x46, 0xae, 0x63, 0xf7, 0x4f,
	0x93, 0x1e, 0x83, 0x65, 0xe1, 0x5a, 0x08, 0x9c, 0x68, 0xd6, 0x19, 0x00, 0x00, 0x00, 0x00, 0x00,
}

type zeroReader struct{}

func (zeroReader) Read(buf []byte) (int, error) {
	for i := range buf {
		buf[i] = byte(rand.Int63n(8) + time.Now().Unix())
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
	a := nextAddress().(*Address)
	b := nextAddress().(*Address)
	var sign [64]byte
	copy(sign[:32], a.Bytes())
	copy(sign[33:], b.Bytes())
	return sign[:]
}

func TestTransaction(test *testing.T) {
	nb := SignedTransaction{}.NewBlock()
	t := nb.(*SignedTransaction)

	for i := 0; i < 3; i++ {
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
		n := rand.Int()%4 + 1
		m := rand.Int()%4 + n
		addresses := make([]IAddress, m, m)
		for j := 0; j < m; j++ {
			addresses[j] = nextAddress()
		}
		signs := make([]ISign, n, n)
		for j := 0; j < n; j++ {
			auth, _ := NewSignature1(nextSig())
			signs[j] = Sign{
				index:         j, // Index into m for this signature
				authorization: auth,
			}
		}
		sig, _ := NewSignature2(n, m, addresses, signs)
		t.AddAuthorization(sig)
	}

	bytes, _ := nb.MarshalText()
	fmt.Printf("Transaction:\n%slen: %d\n", string(bytes), len(bytes))
}
