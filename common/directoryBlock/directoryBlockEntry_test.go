// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package directoryBlock_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/primitives"
)

func TestUnmarshalNilDBEntry(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(DBEntry)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestDBSEMisc(t *testing.T) {
	dbe := new(DBEntry)
	hash, err := primitives.HexToHash("000000000000000000000000000000000000000000000000000000000000000a")
	if err != nil {
		t.Error(err)
	}
	dbe.ChainID = hash
	hash, err = primitives.HexToHash("000000000000000000000000000000000000000000000000000000000000000b")
	if err != nil {
		t.Error(err)
	}
	dbe.KeyMR = hash

	hash = dbe.GetChainID()
	if hash.String() != "000000000000000000000000000000000000000000000000000000000000000a" {
		t.Fail()
	}
	hash = dbe.GetKeyMR()
	if hash.String() != "000000000000000000000000000000000000000000000000000000000000000b" {
		t.Fail()
	}
	/*
		dbe2, err := NewDBEntry(dbe)
		if err != nil {
			t.Error(err)
		}
		if dbe2 == nil {
			t.Fail()
		}

		hash = dbe2.GetChainID()
		if hash.String() != "000000000000000000000000000000000000000000000000000000000000000a" {
			t.Fail()
		}
		hash, err = dbe2.GetKeyMR()
		if err != nil {
			t.Error(err)
		}
		if hash.String() != "000000000000000000000000000000000000000000000000000000000000000b" {
			t.Fail()
		}
	*/
}

func TestDBSEMarshalUnmarshal(t *testing.T) {
	dbe := new(DBEntry)

	hash, err := primitives.HexToHash("000000000000000000000000000000000000000000000000000000000000000a")
	if err != nil {
		t.Error(err)
	}
	dbe.ChainID = hash
	hash, err = primitives.HexToHash("000000000000000000000000000000000000000000000000000000000000000b")
	if err != nil {
		t.Error(err)
	}
	dbe.KeyMR = hash

	hex, err := dbe.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	dbe2 := new(DBEntry)
	err = dbe2.UnmarshalBinary(hex)
	if err != nil {
		t.Error(err)
	}

	hex2, err := dbe2.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	if len(hex) != len(hex2) {
		t.Fail()
	}

	for i := range hex {
		if hex[i] != hex2[i] {
			t.Fail()
		}
	}
}
