// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package directoryBlock_test

import (
	"fmt"
	"testing"
	"time"

	. "github.com/PaulSnow/factom2d/common/directoryBlock"
	"github.com/PaulSnow/factom2d/common/primitives"
)

func TestUnmarshalNilDBlockHeader(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(DBlockHeader)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestVersion(t *testing.T) {
	dbh := new(DBlockHeader)

	currentVersion := byte(0)
	dbh.SetVersion(currentVersion)

	returnVal := dbh.GetVersion()
	if currentVersion != returnVal {
		t.Fail()
	}

	futureVersion := byte(9)
	dbh.SetVersion(futureVersion)

	returnVal = dbh.GetVersion()
	if futureVersion != returnVal {
		t.Fail()
	}
}

func TestNetID(t *testing.T) {
	dbh := new(DBlockHeader)

	mainnet := uint32(0xFA92E5A2)
	dbh.SetNetworkID(mainnet)

	returnVal := dbh.GetNetworkID()
	if mainnet != returnVal {
		t.Fail()
	}
}

func TestMRs(t *testing.T) {
	dbh := new(DBlockHeader)

	dbh.Init()
	returnVal := dbh.GetBodyMR()
	if !returnVal.IsZero() {
		t.Fail()
	}

	returnVal = dbh.GetPrevKeyMR()
	if !returnVal.IsZero() {
		t.Fail()
	}

	returnVal = dbh.GetPrevFullHash()
	if !returnVal.IsZero() {
		t.Fail()
	}

	//testhash := new(primitives.Hash)
	testhash, _ := primitives.HexToHash("1934687145014f234b3451c151345a14350e13462568c4146317181456256526")

	dbh.SetBodyMR(testhash)
	returnVal = dbh.GetBodyMR()
	if testhash != returnVal {
		t.Fail()
	}

	dbh.SetPrevKeyMR(testhash)
	returnVal = dbh.GetPrevKeyMR()
	if testhash != returnVal {
		t.Fail()
	}

	dbh.SetPrevFullHash(testhash)
	returnVal = dbh.GetPrevFullHash()
	if testhash != returnVal {
		t.Fail()
	}

}

func TestTimestamp(t *testing.T) {
	dbh := new(DBlockHeader)
	ts := primitives.NewTimestampFromMinutes(24018960) //genesis block time in minutes

	dbh.SetTimestamp(ts)
	returnVal := dbh.GetTimestamp()

	if returnVal.GetTimeSeconds() != 1441137600 { //genesis block time in seconds
		t.Fail()
	}
}

func TestHeight(t *testing.T) {
	dbh := new(DBlockHeader)

	dbh.SetDBHeight(1234)
	returnVal := dbh.GetDBHeight()

	if returnVal != 1234 {
		t.Fail()
	}

}

func TestPrints(t *testing.T) {
	dbh := new(DBlockHeader)
	dbh.Init()

	returnVal := dbh.String()

	expectedString1 := `  version:         0
  networkid:       0
  bodymr:          000000
  prevkeymr:       000000
  prevfullhash:    000000
  timestamp:       0
  timestamp str:   `

	//1969-12-31 18:00:00
	epoch := time.Unix(0, 0)
	expectedString2 := epoch.Format("2006-01-02 15:04:05")

	expectedString3 := `
  dbheight:        0
  blockcount:      0
`

	expectedString := expectedString1 + expectedString2 + expectedString3

	if returnVal != expectedString {
		fmt.Println(returnVal)
		fmt.Println(expectedString)
		t.Fail()
	}

	returnVal, _ = dbh.JSONString()
	//fmt.Println(returnVal)

	expectedString = `{"version":0,"networkid":0,"bodymr":"0000000000000000000000000000000000000000000000000000000000000000","prevkeymr":"0000000000000000000000000000000000000000000000000000000000000000","prevfullhash":"0000000000000000000000000000000000000000000000000000000000000000","timestamp":0,"dbheight":0,"blockcount":0,"chainid":"000000000000000000000000000000000000000000000000000000000000000d"}`
	if returnVal != expectedString {
		fmt.Println("got", returnVal)
		fmt.Println("expected", expectedString)
		t.Fail()
	}

	returnBytes, _ := dbh.JSONByte()
	s := string(returnBytes)
	if s != expectedString {
		fmt.Println("got", s)
		fmt.Println("expected", expectedString)
		t.Fail()
	}
}
