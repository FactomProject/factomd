// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"testing"

	//"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
)

func TestCheckSig(t *testing.T) {
	priv := testHelper.NewPrivKey(1)
	msg := []byte("Hello!")
	pub := testHelper.PrivateKeyToEDPub(priv)

	pre := []byte{0x01}
	pre = append(pre, pub...)
	id := primitives.Shad(pre)

	sig := primitives.Sign(priv, msg)

	if CheckSig(id, pub, msg, sig) == false {
		t.Error("Valid signature not valid")
	}

	sig[0] += 1

	if CheckSig(id, pub, msg, sig) == true {
		t.Error("Invalid signature valid")
	}
}

func TestAppendExtIDs(t *testing.T) {
	ids := [][]byte{
		[]byte{0x01, 0x02},
		[]byte{0x03, 0x04},
		[]byte{0x05, 0x06},
		[]byte{0x07, 0x08},
		[]byte{0x09, 0x0a},
	}

	appended := []byte{
		0x03, 0x04,
		0x05, 0x06,
		0x07, 0x08,
		0x09, 0x0a,
	}

	resp, err := AppendExtIDs(ids, 1, 4)
	if err != nil {
		t.Errorf("%v", err)
	}
	if primitives.AreBytesEqual(resp, appended) == false {
		t.Errorf("AppendExtIDs are not equal - %x vs %x", resp, appended)
	}

	resp, err = AppendExtIDs(ids, 1, 5)
	if err == nil {
		t.Error("Err is not when it should not be")
	}
	if resp != nil {
		t.Error("Resp is not nil when it should be")
	}
}

func TestCheckLength(t *testing.T) {
	// Invalid
	x := []byte{0x00, 0x00, 0x00, 0x00, 0x00}
	if CheckLength(6, x) {
		t.Error("CheckLength check failed")
	}

	// Invalid
	if CheckLength(4, x) {
		t.Error("CheckLength check failed")
	}

	// Invalid
	if CheckLength(0, x) {
		t.Error("CheckLength check failed")
	}

	// Valid
	x = []byte{}
	for i := 0; i < 15; i++ {
		if !CheckLength(len(x), x) {
			t.Error("CheckLength check failed")
		}
		x = append(x, []byte{0x00}...)
	}
}
