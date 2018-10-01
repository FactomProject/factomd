// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"encoding/binary"
	"fmt"
	"os"
	"testing"

	. "github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
)

func TestUnmarshalNilBounce(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(Bounce)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestBounceAttributes(t *testing.T) {
	b := new(Bounce)

	v := b.Validate(nil)
	if v != 1 {
		t.Errorf("Should be 1, found %d", v)
	}

	b.SetValid(0)
	v = b.Validate(nil)
	if v != 0 {
		t.Errorf("Should be 0, found %d", v)
	}

	b.SetValid(-1)
	v = b.Validate(nil)
	if v != -1 {
		t.Errorf("Should be -1, found %d", v)
	}

	if b.Processed() {
		t.Error("Processed should be false")
	}

	b.FollowerExecute(nil)
	if !b.Processed() {
		t.Error("Processed should be true")
	}
}

func TestBounceMisc(t *testing.T) {
	b := new(Bounce)

	b.Process(1, nil)
	if !b.Processed() {
		t.Error("Processed should be true")
	}

	err := b.Sign(nil)
	if err != nil {
		t.Error(err)
	}

	nilSig := b.GetSignature()
	if nilSig != nil {
		t.Error("Bounce signature should be nil")
	}

	_, err = b.JSONByte()
	if err != nil {
		t.Error(err)
	}

	_, err = b.JSONString()
	if err != nil {
		t.Error(err)
	}

	sigVer, err := b.VerifySignature()
	if err != nil {
		t.Error(err)
	}
	if !sigVer {
		t.Error("VerifySignature should be true")
	}

	b2 := new(Bounce)
	if !b.IsSameAs(b2) {
		t.Error("Bounces should always be considered the same")
	}

	b2.LeaderExecute(nil)
	if !b2.Processed() {
		t.Error("Processed should be true")
	}
}

func TestMarshalUnmarshalBounce(t *testing.T) {
	b1 := new(Bounce)
	b1.Timestamp = primitives.NewTimestampNow()
	b1.Data = append(b1.Data, 1)

	p, err := b1.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	b2 := new(Bounce)
	if err := b2.UnmarshalBinary(p); err != nil {
		t.Error(err)
	}

	if !b1.IsSameAs(b2) {
		t.Error("unmarshaled bounce did not match original", b1, b2)
	}
}

func TestMarshalUnmarshalMaliciousBounce(t *testing.T) {
	b1 := new(Bounce)
	b1.Timestamp = primitives.NewTimestampNow()
	b1.Data = append(b1.Data, 1)

	good, err := b1.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	bad, err := BadMarshal(b1)
	if err != nil {
		t.Error(err)
	}

	t.Logf("good: %x\n", good)
	t.Logf("bad:  %x\n", bad)

	b2 := new(Bounce)
	b2.UnmarshalBinary(bad) // this should fail badly
}

func BadMarshal(m *Bounce) (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "Bounce.MarshalForSignature err:%v", *pe)
		}
	}(&err)
	var buf primitives.Buffer

	binary.Write(&buf, binary.BigEndian, m.Type())

	var buff [32]byte

	copy(buff[:32], []byte(fmt.Sprintf("%32s", m.Name)))
	buf.Write(buff[:])

	binary.Write(&buf, binary.BigEndian, m.Number)

	t := m.GetTimestamp()
	data, err := t.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	// Bad value -1 used instad of correct value
	binary.Write(&buf, binary.BigEndian, int32(-1))
	// binary.Write(&buf, binary.BigEndian, int32(len(m.Stamps)))

	for _, ts := range m.Stamps {
		data, err := ts.MarshalBinary()
		if err != nil {
			return nil, err
		}
		buf.Write(data)
	}

	// Bad write wrong length for Data
	binary.Write(&buf, binary.BigEndian, int32(-1))
	// binary.Write(&buf, binary.BigEndian, int32(len(m.Data)))
	buf.Write(m.Data)

	return buf.DeepCopyBytes(), nil
}
