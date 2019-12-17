// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid_test

import (
	"strings"
	"testing"

	"math"

	. "github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
	"github.com/FactomProject/factomd/testHelper"
)

// TestUnmarshalNilTransAddress checks that unmarshalling nil or the empty interface returns proper errors
func TestUnmarshalNilTransAddress(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(TransAddress)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

// TestTAddressEquals creates 1000 random trans addresses does various manipulations to check that they are equal when they should be
// and not equal when different. Also checks for integrity of marshal and unmarshal
func TestTAddressEquals(t *testing.T) {
	for i := 0; i < 1000; i++ {
		a1 := RandomTransAddress()
		a2 := new(TransAddress)

		if a1.IsSameAs(a2) == true {
			t.Errorf("Addresses are equal while they shouldn't be")
		}

		a2.SetAddress(a1.GetAddress())
		a2.SetAmount(a1.GetAmount())
		if a1.IsSameAs(a2) == false {
			t.Errorf("Addresses are not equal while they should be")
		}

		a2.Amount = a1.GetAmount() - 1
		if a1.IsSameAs(a2) == true {
			t.Errorf("Addresses are equal while they shouldn't be")
		}

		a2 = new(TransAddress)
		if a1.IsSameAs(a2) == true {
			t.Errorf("Addresses are equal while they shouldn't be")
		}

		b, err := a1.MarshalBinary()
		if err != nil {
			t.Errorf("%v", err)
		}
		err = a2.UnmarshalBinary(b)
		if err != nil {
			t.Errorf("%v", err)
		}

		if a1.IsSameAs(a2) == false {
			t.Errorf("Addresses are not equal while they should be")
		}
		a2.Address = RandomAddress()
		if a1.IsSameAs(a2) == true {
			t.Errorf("Addresses are equal while they shouldn't be")
		}

	}
}

// TestTransMarshalUnmarshal checks that trans address can be marshalled and unmarshalled correctly
func TestTransAddressMarshalUnmarshal(t *testing.T) {
	ta := new(TransAddress)
	ta.SetAmount(12345678)
	h, err := primitives.HexToHash("ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973")
	if err != nil {
		t.Error(err)
	}
	add := h.(interfaces.IAddress)
	ta.SetAddress(add)

	hex, err := ta.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	ta2 := new(TransAddress)
	err = ta2.UnmarshalBinary(hex)
	if err != nil {
		t.Error(err)
	}
	json1, err := ta.JSONString()
	if err != nil {
		t.Error(err)
	}
	json2, err := ta2.JSONString()
	if err != nil {
		t.Error(err)
	}
	if json1 != json2 {
		t.Error("JSONs are not identical")
	}
}

// TestOutECAddress checks that NewOutECAddress returns the proper object
func TestOutECAddress(t *testing.T) {
	h, err := primitives.HexToHash("ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973")
	if err != nil {
		t.Error(err)
	}
	add := h.(interfaces.IAddress)
	outECAdd := NewOutECAddress(add, 12345678)
	str := outECAdd.StringECOutput()

	t.Logf("outECAdd str - %v", str)

	if strings.Contains(str, "ecoutput") == false {
		t.Error("'ecoutput' not found")
	}
	if strings.Contains(str, "0.12345678") == false {
		t.Error("'0.12345678' not found")
	}
	if strings.Contains(str, "EC3ZMxDt8xUBKBmrmzLwSpnMHkdptLS8gTSf8NQhVf7vpAWqNE2p") == false {
		t.Error("'EC3ZMxDt8xUBKBmrmzLwSpnMHkdptLS8gTSf8NQhVf7vpAWqNE2p' not found")
	}
	if strings.Contains(str, "0000000000bc614e") == false {
		t.Error("'0000000000bc614e' not found")
	}
	if strings.Contains(str, "ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973") == false {
		t.Error("'ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973' not found")
	}
}

// TestOutAddress checks that NewOutAddress returns the proper object
func TestOutAddress(t *testing.T) {
	h, err := primitives.HexToHash("ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973")
	if err != nil {
		t.Error(err)
	}
	add := h.(interfaces.IAddress)
	outAdd := NewOutAddress(add, 12345678)
	str := outAdd.StringOutput()

	t.Logf("outAdd str - %v", str)

	if strings.Contains(str, "out") == false {
		t.Error("'out' not found")
	}
	if strings.Contains(str, "0.12345678") == false {
		t.Error("'0.12345678' not found")
	}
	if strings.Contains(str, "FA3mHjgsVvQJjVbvJpy67deDKzEsqc8FsLU122i8Tj76rmakpqRL") == false {
		t.Error("'FA3mHjgsVvQJjVbvJpy67deDKzEsqc8FsLU122i8Tj76rmakpqRL' not found")
	}
	if strings.Contains(str, "0000000000bc614e") == false {
		t.Error("'0000000000bc614e' not found")
	}
	if strings.Contains(str, "ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973") == false {
		t.Error("'ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973' not found")
	}
}

// TestInAddress checks that NewInAddress returns the proper object
func TestInAddress(t *testing.T) {
	h, err := primitives.HexToHash("ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973")
	if err != nil {
		t.Error(err)
	}
	add := h.(interfaces.IAddress)
	inAdd := NewInAddress(add, 12345678)
	str := inAdd.StringInput()

	t.Logf("InAdd str - %v", str)

	if strings.Contains(str, "input") == false {
		t.Error("'input' not found")
	}
	if strings.Contains(str, "0.12345678") == false {
		t.Error("'0.12345678' not found")
	}
	if strings.Contains(str, "FA3mHjgsVvQJjVbvJpy67deDKzEsqc8FsLU122i8Tj76rmakpqRL") == false {
		t.Error("'FA3mHjgsVvQJjVbvJpy67deDKzEsqc8FsLU122i8Tj76rmakpqRL' not found")
	}
	if strings.Contains(str, "0000000000bc614e") == false {
		t.Error("'0000000000bc614e' not found")
	}
	if strings.Contains(str, "ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973") == false {
		t.Error("'ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973' not found")
	}
}

// TestTransAddressBlob tests multiple trans addresses in the same binary blob
func TestTransAddressBlob(t *testing.T) {
	var err error
	// TODO: Add unit test to cover unmarshaling multiple transaddresses from same binary blob
	for i := 0; i < 100; i++ {
		n := random.RandIntBetween(2, 10)
		array := make([]interfaces.ITransAddress, n)
		var data []byte
		for i := range array {
			array[i] = RandomTransAddress()
			d, err := array[i].MarshalBinary()
			if err != nil {
				t.Error(err)
			}
			data = append(data, d...)
		}

		new_array := make([]interfaces.ITransAddress, n)
		for i := range new_array {
			new_array[i] = new(TransAddress)
			data, err = new_array[i].UnmarshalBinaryData(data)
			if err != nil {
				t.Error(err)
			}
		}

		if len(data) != 0 {
			t.Errorf("%d bytes remain, expected 0", len(data))
		}

		for i := range array {
			if !array[i].IsSameAs(new_array[i]) {
				t.Error("Transaddress differs")
			}
		}
	}
}

// TestVectorTransAddressMarshalling checks trans addresses can be marshalled and unmarshalled properly
func TestVectorTransAddressMarshalling(t *testing.T) {
	amounts := []uint64{
		0, 1, 2, 3, 4, 5, 100, 1000, 1e10, math.MaxUint32, math.MaxUint64 - 1, math.MaxUint64,
	}
	for _, amt := range amounts {
		a := RandomTransAddress()
		a.SetAmount(amt)
		b := new(TransAddress)
		testHelper.TestMarshaling(a, b, random.RandIntBetween(0, 100), t)

		if !a.IsSameAs(b) {
			t.Error("Unmarshaled value does not match marshalled")
		}
	}

}

// TestRandomTransAddressMarshalling checks that 100 random trans addresses can be marshalled and unmarshalled properly
func TestRandomTransAddressMarshalling(t *testing.T) {
	// Test Random addresses with random amounts
	for i := 0; i < 100; i++ {
		a := RandomTransAddress()
		b := new(TransAddress)
		testHelper.TestMarshaling(a, b, random.RandIntBetween(0, 100), t)

		if !a.IsSameAs(b) {
			t.Error("Unmarshaled value does not match marshalled")
		}
	}
}
