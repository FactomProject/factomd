// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives_test

import (
	"math"
	"testing"

	. "github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
)

func TestPushPopInt64(t *testing.T) {
	b := NewBuffer(nil)

	var i int64
	err := b.PushInt64(i)
	if err != nil {
		t.Errorf("%v", err)
	}
	r, err := b.PopInt64()
	if err != nil {
		t.Errorf("%v", err)
	}
	if i != r {
		t.Errorf("Received wrong number - %v vs %v", i, r)
	}

	i = math.MaxInt64
	err = b.PushInt64(i)
	if err != nil {
		t.Errorf("%v", err)
	}
	r, err = b.PopInt64()
	if err != nil {
		t.Errorf("%v", err)
	}
	if i != r {
		t.Errorf("Received wrong number - %v vs %v", i, r)
	}

	i = math.MinInt64
	err = b.PushInt64(i)
	if err != nil {
		t.Errorf("%v", err)
	}
	r, err = b.PopInt64()
	if err != nil {
		t.Errorf("%v", err)
	}
	if i != r {
		t.Errorf("Received wrong number - %v vs %v", i, r)
	}

	for j := 0; j < 1000; j++ {
		i = random.RandInt64()
		err = b.PushInt64(i)
		if err != nil {
			t.Errorf("%v", err)
		}
		r, err = b.PopInt64()
		if err != nil {
			t.Errorf("%v", err)
		}
		if i != r {
			t.Errorf("Received wrong number - %v vs %v", i, r)
		}
	}
}

func TestPushPopUInt64(t *testing.T) {
	b := NewBuffer(nil)

	var i uint64
	err := b.PushUInt64(i)
	if err != nil {
		t.Errorf("%v", err)
	}
	r, err := b.PopUInt64()
	if err != nil {
		t.Errorf("%v", err)
	}
	if i != r {
		t.Errorf("Received wrong number - %v vs %v", i, r)
	}

	i = math.MaxInt64
	err = b.PushUInt64(i)
	if err != nil {
		t.Errorf("%v", err)
	}
	r, err = b.PopUInt64()
	if err != nil {
		t.Errorf("%v", err)
	}
	if i != r {
		t.Errorf("Received wrong number - %v vs %v", i, r)
	}

	for j := 0; j < 1000; j++ {
		i = random.RandUInt64()
		err = b.PushUInt64(i)
		if err != nil {
			t.Errorf("%v", err)
		}
		r, err = b.PopUInt64()
		if err != nil {
			t.Errorf("%v", err)
		}
		if i != r {
			t.Errorf("Received wrong number - %v vs %v", i, r)
		}
	}
}

func TestPushPopVarInt(t *testing.T) {
	b := NewBuffer(nil)

	var i uint64
	err := b.PushVarInt(i)
	if err != nil {
		t.Errorf("%v", err)
	}
	r, err := b.PopVarInt()
	if err != nil {
		t.Errorf("%v", err)
	}
	if i != r {
		t.Errorf("Received wrong number - %v vs %v", i, r)
	}

	i = math.MaxInt64
	err = b.PushVarInt(i)
	if err != nil {
		t.Errorf("%v", err)
	}
	r, err = b.PopVarInt()
	if err != nil {
		t.Errorf("%v", err)
	}
	if i != r {
		t.Errorf("Received wrong number - %v vs %v", i, r)
	}

	for j := 0; j < 1000; j++ {
		i = random.RandUInt64()
		err = b.PushVarInt(i)
		if err != nil {
			t.Errorf("%v", err)
		}
		r, err = b.PopVarInt()
		if err != nil {
			t.Errorf("%v", err)
		}
		if i != r {
			t.Errorf("Received wrong number - %v vs %v", i, r)
		}
	}
}

func TestPushPopString(t *testing.T) {
	b := NewBuffer(nil)
	for i := 0; i < 1000; i++ {
		str := random.RandomString()
		err := b.PushString(str)
		if err != nil {
			t.Errorf("%v", err)
		}
		r, err := b.PopString()
		if err != nil {
			t.Errorf("%v", err)
		}
		if str != r {
			t.Errorf("Received wrong string - %v vs %v", str, r)
		}
	}
}

func TestPushPopBytes(t *testing.T) {
	buf := NewBuffer(nil)
	for i := 0; i < 1000; i++ {
		b := random.RandByteSlice()
		err := buf.PushBytes(b)
		if err != nil {
			t.Errorf("%v", err)
		}
		r, err := buf.PopBytes()
		if err != nil {
			t.Errorf("%v", err)
		}
		if AreBytesEqual(b, r) == false {
			t.Errorf("Received wrong byte slice - %x vs %x", b, r)
		}
		err = buf.Push(b)
		if err != nil {
			t.Errorf("%v", err)
		}
		h := make([]byte, len(b))
		err = buf.Pop(h)
		if err != nil {
			t.Errorf("%v", err)
		}
		if AreBytesEqual(b, h) == false {
			t.Errorf("Received wrong byte slice for push/pop - %x vs %x", b, h)
		}
		_ = buf.Push(b) // No error, already checked
		h2, err := buf.PopLen(len(b))
		if AreBytesEqual(b, h2) == false {
			t.Errorf("Received wrong byte slice for PopLen - %x vs %x", b, h2)
		}

	}
}

func TestPushPopUInt32(t *testing.T) {
	b := NewBuffer(nil)

	var i uint32
	err := b.PushUInt32(i)
	if err != nil {
		t.Errorf("%v", err)
	}
	r, err := b.PopUInt32()
	if err != nil {
		t.Errorf("%v", err)
	}
	if i != r {
		t.Errorf("Received wrong number - %v vs %v", i, r)
	}

	i = math.MaxUint32
	err = b.PushUInt32(i)
	if err != nil {
		t.Errorf("%v", err)
	}
	r, err = b.PopUInt32()
	if err != nil {
		t.Errorf("%v", err)
	}
	if i != r {
		t.Errorf("Received wrong number - %v vs %v", i, r)
	}

	for j := 0; j < 1000; j++ {
		i = random.RandUInt32()
		err = b.PushUInt32(i)
		if err != nil {
			t.Errorf("%v", err)
		}
		r, err = b.PopUInt32()
		if err != nil {
			t.Errorf("%v", err)
		}
		if i != r {
			t.Errorf("Received wrong number - %v vs %v", i, r)
		}
	}
}

func TestPushPopBinaryMarshallable(t *testing.T) {
	b := NewBuffer(nil)
	for i := 0; i < 1000; i++ {
		h := RandomHash()
		err := b.PushBinaryMarshallable(h)
		if err != nil {
			t.Errorf("%v", err)
		}
		h2 := new(Hash)
		err = b.PopBinaryMarshallable(h2)
		if err != nil {
			t.Errorf("%v", err)
		}
		if h.IsSameAs(h2) == false {
			t.Errorf("Received wrong hash - %v vs %v", h, h2)
		}
		err = b.PushIHash(h)
		if err != nil {
			t.Errorf("%v", err)
		}
		h3, err := b.PopIHash()
		if err != nil {
			t.Errorf("%v", err)
		}
		if h.IsSameAs(h3) == false {
			t.Errorf("Received wrong hash - %v vcs %v", h, h3)
		}
	}
}

// TestRandomPushPop tests a host of push and pop functions
func TestRandomPushPop(t *testing.T) {
	b := NewBuffer(nil)

	r8 := random.RandUInt8()
	b.PushUInt8(r8)
	b.PushUInt16(32)
	rb := random.RandByte()
	b.PushByte(rb)
	rint := random.RandInt()
	b.PushInt(rint)
	b.PushBool(true)
	b.PushBool(false)

	nr8, err := b.PopUInt8()
	if err != nil || nr8 != r8 {
		t.Error("Unable to PopUInt8")
	}
	nr16, err := b.PopUInt16()
	if err != nil || nr16 != 32 {
		t.Error("Unable to PopUInt16")
	}
	mybyte, err := b.PeekByte() // Check the byte before popping for a test
	if err != nil || mybyte != rb {
		t.Error("Unable to PeekByte")
	}
	nrb, err := b.PopByte()
	if err != nil || nrb != rb {
		t.Error("Unable to PopByte")
	}
	nrint, err := b.PopInt()
	if err != nil || nrint != rint {
		t.Error("Unable to PopInt")
	}
	nboo, err := b.PopBool()
	if err != nil || nboo != true {
		t.Error("Unable to PopBool true")
	}
	nboo, err = b.PopBool()
	if err != nil || nboo != false {
		t.Error("Unable to PopBool false")
	}
}
