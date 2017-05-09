// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"

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

func TestCheckTimestamp(t *testing.T) {
	var out bytes.Buffer
	now := time.Now()
	binary.Write(&out, binary.BigEndian, uint64(now.Unix()))
	hex := out.Bytes()

	if CheckTimestamp(hex, now.Unix()) == false {
		t.Error("Timestamp check failed")
	}

	var delta uint64 = (11*60 + 59) * 60
	out.Reset()
	binary.Write(&out, binary.BigEndian, uint64(now.Unix())-delta)
	hex = out.Bytes()

	if CheckTimestamp(hex, now.Unix()) == false {
		t.Error("Timestamp check failed")
	}
	out.Reset()
	binary.Write(&out, binary.BigEndian, uint64(now.Unix())+delta)
	hex = out.Bytes()

	if CheckTimestamp(hex, now.Unix()) == false {
		t.Error("Timestamp check failed")
	}

	delta = (11*60 + 61) * 60
	out.Reset()
	binary.Write(&out, binary.BigEndian, uint64(now.Unix())-delta)
	hex = out.Bytes()

	if CheckTimestamp(hex, now.Unix()) == true {
		t.Error("Timestamp check failed")
	}
	out.Reset()
	binary.Write(&out, binary.BigEndian, uint64(now.Unix())+delta)
	hex = out.Bytes()

	if CheckTimestamp(hex, now.Unix()) == true {
		t.Error("Timestamp check failed")
	}

	delta = (12 * 60 * 60) + 10
	out.Reset()
	binary.Write(&out, binary.BigEndian, uint64(now.Unix())-delta)
	hex = out.Bytes()
	if CheckTimestamp(hex, now.Unix()) == true {
		t.Error("Timestamp check failed")
	}
}

func TestCheckExternalIDsLength(t *testing.T) {
	extIDs := [][]byte{
		{0x00, 0x00, 0x00, 0x00, 0x00},                               // 5
		{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, // 10
		{0x00, 0x00, 0x00},                                           // 3
		{0x00},                                                       // 1
		{},                                                           // 0
	}
	lengths := []int{5, 10, 3, 1, 0}
	lengthsBad := []int{5, 10, 3, 1, 1}
	if CheckExternalIDsLength(extIDs, lengthsBad) {
		t.Error("1: CheckExternalIDsLength check failed")
	}

	lengthsBad = []int{}
	if CheckExternalIDsLength(extIDs, lengthsBad) {
		t.Error("2: CheckExternalIDsLength check failed")
	}

	lengthsBad = []int{5, 10, 3, 1}
	if CheckExternalIDsLength(extIDs, lengthsBad) {
		t.Error("3: CheckExternalIDsLength check failed")
	}

	if !CheckExternalIDsLength(extIDs, lengths) {
		t.Error("4: CheckExternalIDsLength check failed")
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

func TestAnchorSigningKeyMarshalUnmarshal(t *testing.T) {
	for i := 0; i < 1000; i++ {
		ask := RandomAnchorSigningKey()
		h, err := ask.MarshalBinary()
		if err != nil {
			t.Errorf("%v", err)
		}
		ask2 := new(AnchorSigningKey)
		err = ask2.UnmarshalBinary(h)
		if err != nil {
			t.Errorf("%v", err)
		}
		if ask.IsSameAs(ask2) == false {
			t.Errorf("AnchorSigningKeys are not the same")
		}
	}
}

func TestIdentityMarshalUnmarshal(t *testing.T) {
	for i := 0; i < 1000; i++ {
		id := RandomIdentity()
		h, err := id.MarshalBinary()
		if err != nil {
			t.Errorf("%v", err)
		}
		id2 := new(Identity)
		err = id2.UnmarshalBinary(h)
		if err != nil {
			t.Errorf("%v", err)
		}
		if id.IsSameAs(id2) == false {
			t.Errorf("Identities are not the same")
		}
	}
}
