// +build all

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid_test

import (
	crand "crypto/rand"
	"math/rand"
	"testing"

	"github.com/FactomProject/ed25519"
	. "github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
)

func TestUnmarshalNilRCD_1(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(RCD_1)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestJSONMarshal(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := newRCD_1()
	s, err := a.JSONString()
	if err != nil {
		t.Error(err)
	}

	if s[1:3] != "01" {
		t.Errorf("Not prepended by rcd type, found %s", s)
	}

	if len(s) != 68 {
		t.Error("Not the correct length")
	}
}

func TestMarshal(t *testing.T) {
	for i := 0; i < 10; i++ {
		a := newRandRCD_1()
		data, err := a.MarshalBinary()
		if err != nil {
			t.Error(err)
		}

		var b = interfaces.IRCD(new(RCD_1))

		n, err := b.UnmarshalBinaryData(data)
		if err != nil {
			t.Error(err)
		}

		if len(n) > 0 {
			t.Errorf("Should have 0 bytes left, found %d", len(n))
		}
	}
}

func TestBadPublic(t *testing.T) {
	if !testCreate() {
		t.Error("Should have paniced")
	}
}

func testCreate() (fail bool) {
	defer func() {
		if r := recover(); r != nil {
			fail = true
			return
		}
	}()

	CreateRCD([]byte{})
	return false
}

type zeroReader1 struct{}

var zero1 zeroReader1

func (zeroReader1) Read(buf []byte) (int, error) {
	//if r==nil { r = rand.New(rand.NewSource(time.Now().Unix())) }
	//if r == nil {
	r := rand.New(rand.NewSource(1))
	//}
	for i := range buf {
		buf[i] = byte(r.Int())
	}
	return len(buf), nil
}

func newRandRCD_1() *RCD_1 {
	public, _, _ := ed25519.GenerateKey(crand.Reader)
	rcd := NewRCD_1(public[:])

	return rcd.(*RCD_1)
}

func newRCD_1() *RCD_1 {
	public, _, _ := ed25519.GenerateKey(zero1)
	rcd := NewRCD_1(public[:])

	return rcd.(*RCD_1)
}
