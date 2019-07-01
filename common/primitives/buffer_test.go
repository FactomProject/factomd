// +build all 

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
	}
}
