// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives_test

import (
	"math/rand"
	"testing"

	. "github.com/FactomProject/factomd/common/primitives"
)

func TestVarIntLength(t *testing.T) {
	if VarIntLength(0x00) != 1 {
		t.Errorf("Invalid VarInt length")
	}
	if VarIntLength(0x01) != 1 {
		t.Errorf("Invalid VarInt length")
	}

	if VarIntLength(0x7F) != 1 {
		t.Errorf("Invalid VarInt length")
	}
	if VarIntLength(0x80) != 2 {
		t.Errorf("Invalid VarInt length")
	}

	if VarIntLength(0x3FFF) != 2 {
		t.Errorf("Invalid VarInt length")
	}
	if VarIntLength(0x4000) != 3 {
		t.Errorf("Invalid VarInt length")
	}

	if VarIntLength(0x1FFFFF) != 3 {
		t.Errorf("Invalid VarInt length")
	}
	if VarIntLength(0x200000) != 4 {
		t.Errorf("Invalid VarInt length")
	}

	if VarIntLength(0x0FFFFFFF) != 4 {
		t.Errorf("Invalid VarInt length")
	}
	if VarIntLength(0x10000000) != 5 {
		t.Errorf("Invalid VarInt length")
	}

	if VarIntLength(0x7FFFFFFFF) != 5 {
		t.Errorf("Invalid VarInt length")
	}
	if VarIntLength(0x800000000) != 6 {
		t.Errorf("Invalid VarInt length")
	}

	if VarIntLength(0x3FFFFFFFFFF) != 6 {
		t.Errorf("Invalid VarInt length")
	}
	if VarIntLength(0x40000000000) != 7 {
		t.Errorf("Invalid VarInt length")
	}

	if VarIntLength(0x1FFFFFFFFFFFF) != 7 {
		t.Errorf("Invalid VarInt length")
	}
	if VarIntLength(0x2000000000000) != 8 {
		t.Errorf("Invalid VarInt length")
	}

	if VarIntLength(0x0FFFFFFFFFFFFFF) != 8 {
		t.Errorf("Invalid VarInt length")
	}
	if VarIntLength(0x100000000000000) != 9 {
		t.Errorf("Invalid VarInt length")
	}

	if VarIntLength(0x7FFFFFFFFFFFFFFF) != 9 {
		t.Errorf("Invalid VarInt length")
	}
	if VarIntLength(0x8000000000000000) != 10 {
		t.Errorf("Invalid VarInt length")
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
