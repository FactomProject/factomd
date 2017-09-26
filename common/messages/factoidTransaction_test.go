// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"math/rand"
	"testing"

	"github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/messages/msgsupport"
)

func TestUnmarshalNilFactoidTransaction(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(FactoidTransaction)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestMarshalUnmarshalFactoidTransaction(t *testing.T) {
	msg := newFactoidTransaction()

	hex, err := msg.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled - %x", hex)

	msg2, err := msgsupport.UnmarshalMessage(hex)
	if err != nil {
		t.Fatal(err)
	}
	//str := msg2.String()
	//t.Logf("str - %v", str)

	if msg2.Type() != constants.FACTOID_TRANSACTION_MSG {
		t.Error("Invalid message type unmarshalled")
	}

	hex2, err := msg2.(*FactoidTransaction).MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	if len(hex) != len(hex2) {
		t.Error("Hexes aren't of identical length")
	}
	for i := range hex {
		if hex[i] != hex2[i] {
			t.Error("Hexes do not match")
		}
	}

	if msg.IsSameAs(msg2.(*FactoidTransaction)) != true {
		t.Errorf("FactoidTransaction messages are not identical")
	}
}

func newFactoidTransaction() *FactoidTransaction {
	msg := new(FactoidTransaction)
	//msg.Timestamp = primitives.NewTimestampNow()

	t := new(factoid.Transaction)

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
		sig := factoid.NewRCD_1(nextSig())
		t.AddAuthorization(sig)
	}

	for i := 0; i < 2; i++ {
		t.AddAuthorization(nextAuth2())
	}

	msg.Transaction = t

	return msg
}

func nextAddress() interfaces.IAddress {
	public, _, _ := ed25519.GenerateKey(zero)

	addr := new(factoid.Address)
	addr.SetBytes(public[:])
	return addr
}

type zeroReader struct{}

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

var zero zeroReader

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

	rcd, _ := factoid.NewRCD_2(n, m, addresses)
	return rcd
}

func nextSig() []byte {
	// Get me a private key.
	public, _, _ := ed25519.GenerateKey(zero)

	return public[:]
}
