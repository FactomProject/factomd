// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid_test

import (
	"strings"
	"testing"

	. "github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

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
