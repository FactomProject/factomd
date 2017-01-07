// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid_test

import (
	"github.com/FactomProject/ed25519"
	. "github.com/FactomProject/factomd/common/factoid"
	"math/rand"
	//"testing"
)

type zeroReader1 struct{}

var zero1 zeroReader1

func (zeroReader1) Read(buf []byte) (int, error) {
	//if r==nil { r = rand.New(rand.NewSource(time.Now().Unix())) }
	if r == nil {
		r = rand.New(rand.NewSource(1))
	}
	for i := range buf {
		buf[i] = byte(r.Int())
	}
	return len(buf), nil
}

func newRCD_1() *RCD_1 {
	public, _, _ := ed25519.GenerateKey(zero1)
	rcd := NewRCD_1(public[:])

	return rcd.(*RCD_1)
}
