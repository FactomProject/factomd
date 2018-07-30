// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives_test

import (
	"math"
	"math/rand"
	"testing"

	. "github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
)

func TestUnmarshalNilVarInt(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a, rest := DecodeVarInt(nil)
	if rest != nil {
		t.Errorf("Returned extra data")
	}
	if a != 0 {
		t.Errorf("Wrong value returned")
	}

	a, rest = DecodeVarInt([]byte{})
	if len(rest) != 0 {
		t.Errorf("Returned extra data")
	}
	if a != 0 {
		t.Errorf("Wrong value returned")
	}
}

func TestVarIntLength(t *testing.T) {
	var vector map[uint64]int = map[uint64]int{
		0x00:               1,
		0x01:               1,
		0x7F:               1,
		0x80:               2,
		0x3FFF:             2,
		0x4000:             3,
		0x1FFFFF:           3,
		0x200000:           4,
		0x0FFFFFFF:         4,
		0x10000000:         5,
		0x7FFFFFFFF:        5,
		0x800000000:        6,
		0x3FFFFFFFFFF:      6,
		0x40000000000:      7,
		0x1FFFFFFFFFFFF:    7,
		0x2000000000000:    8,
		0x0FFFFFFFFFFFFFF:  8,
		0x100000000000000:  9,
		0x7FFFFFFFFFFFFFFF: 9,
		0x8000000000000000: 10,
		math.MaxUint64:     10,
	}
	for k, v := range vector {
		if VarIntLength(k) != uint64(v) {
			t.Errorf("Invalid VarInt length - %x is %v, not %v", k, VarIntLength(k), v)
		}
	}
}

func TestVarInt(t *testing.T) {
	for i := 0; i < 1000; i++ {
		var out Buffer
		v := make([]uint64, 10)

		for j := 0; j < len(v); j++ {
			var m uint64           // 64 bit mask
			sw := rand.Int63() % 4 // Pick a random choice
			switch sw {
			case 0:
				m = 0xFF // Random byte
			case 1:
				m = 0xFFFF // Random 16 bit integer
			case 2:
				m = 0xFFFFFFFF // Random 32 bit integer
			case 3:
				m = 0xFFFFFFFFFFFFFFFF // Random 64 bit integer
			}
			n := uint64(rand.Int63() + (rand.Int63() << 32))
			v[j] = n & m
		}

		for j := 0; j < len(v); j++ { // Encode our entire array of numbers
			err := EncodeVarInt(&out, v[j])
			if err != nil {
				t.Errorf("%v", err)
				t.FailNow()
			}
		}

		data := out.Bytes()

		sdata := data // Decode our entire array of numbers, and
		var dv uint64 // check we got them back correctly.
		for k := 0; k < 1000; k++ {
			data = sdata
			for j := 0; j < len(v); j++ {
				dv, data = DecodeVarInt(data)
				if dv != v[j] {
					t.Errorf("Values don't match: decode:%x expected:%x (%d)\n", dv, v[j], j)
					t.FailNow()
				}
			}
		}
	}
}

func TestRandomVarInt(t *testing.T) {
	for i := 0; i < 1000; i++ {
		vi := RandomVarInt()

		out := new(Buffer)
		EncodeVarInt(out, vi)

		dv, rest := DecodeVarInt(out.Bytes())
		if len(rest) > 0 {
			t.Errorf("Returned more bytes than expected for %v", vi)
		}
		if dv != vi {
			t.Errorf("VarInts are not equal - %v vs %v", dv, vi)
		}
	}
}

func TestRandomVarIntWithExtraData(t *testing.T) {
	for i := 0; i < 1000; i++ {
		vi := RandomVarInt()

		out := new(Buffer)
		EncodeVarInt(out, vi)
		extra := random.RandByteSlice()
		_, err := out.Write(extra)
		if err != nil {
			t.Errorf("Error writing extra bytes")
		}

		dv, rest := DecodeVarInt(out.Bytes())
		if len(rest) != len(extra) {
			t.Errorf("Returned wrong number of extra bytes for %v + %x", vi, extra)
		}
		if AreBytesEqual(extra, rest) == false {
			t.Errorf("Returned extra bytes are not equal - %x vs %x", extra, rest)
		}
		if dv != vi {
			t.Errorf("VarInts are not equal - %v vs %v", dv, vi)
		}
	}
}
