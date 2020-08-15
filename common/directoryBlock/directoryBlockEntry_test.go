// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package directoryBlock_test

import (
	"fmt"
	"testing"

	. "github.com/PaulSnow/factom2d/common/directoryBlock"
	"github.com/PaulSnow/factom2d/common/primitives"
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

func TestHash(t *testing.T) {
	dbe := new(DBEntry)

	h, _ := primitives.HexToHash("3e3eb61fb20e71d8211882075d404f5929618a189d23aba8c892b22228aa0d71")
	dbe.SetChainID(h)
	h, _ = primitives.HexToHash("9daad42e5efedf3075fa2cf51908babdb568f431a3c13b9a496ffbfb7160ad2e")
	dbe.SetKeyMR(h)
	hash := dbe.ShaHash()

	keymr, _ := primitives.HexToHash("7509a84bcda2045a400b4650135613685449e05b6b1cb578f152ae4682d9d6ea")

	if !hash.IsSameAs(keymr) {
		fmt.Println(hash)
		fmt.Println(keymr)
		t.Fail()
	}
}

func TestPrintsE(t *testing.T) {
	dbe := new(DBEntry)
	h, _ := primitives.HexToHash("3e3eb61fb20e71d8211882075d404f5929618a189d23aba8c892b22228aa0d71")
	dbe.SetChainID(h)
	h, _ = primitives.HexToHash("9daad42e5efedf3075fa2cf51908babdb568f431a3c13b9a496ffbfb7160ad2e")
	dbe.SetKeyMR(h)
	returnVal := dbe.String()

	expectedString := `chainid: 3e3eb61fb20e71d8211882075d404f5929618a189d23aba8c892b22228aa0d71
      keymr:   9daad42e5efedf3075fa2cf51908babdb568f431a3c13b9a496ffbfb7160ad2e
`

	if returnVal != expectedString {
		fmt.Println(returnVal)
		fmt.Println(expectedString)
		t.Fail()
	}

	returnVal, _ = dbe.JSONString()
	//fmt.Println(returnVal)

	expectedString = `{"chainid":"3e3eb61fb20e71d8211882075d404f5929618a189d23aba8c892b22228aa0d71","keymr":"9daad42e5efedf3075fa2cf51908babdb568f431a3c13b9a496ffbfb7160ad2e"}`
	if returnVal != expectedString {
		fmt.Println("got", returnVal)
		fmt.Println("expected", expectedString)
		t.Fail()
	}

	returnBytes, _ := dbe.JSONByte()
	s := string(returnBytes)
	if s != expectedString {
		fmt.Println("got", s)
		fmt.Println("expected", expectedString)
		t.Fail()
	}
}

func TestCheckErrorsMarshal(t *testing.T) {
	dbe := new(DBEntry)

	h, _ := primitives.HexToHash("3e3eb61fb20e71d8211882075d404f5929618a189d23aba8c892b22228aa0d71")
	dbe.SetChainID(h)

	_, err := dbe.MarshalBinary()
	if err != nil {
		fmt.Println("expected better revocery from missing keymr", err)
		t.Fail()
	}

}
