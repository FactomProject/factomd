// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package database

import (
	"encoding/binary"
	"fmt"
	"github.com/FactomProject/simplecoin"
	"github.com/agl/ed25519"
	"math/rand"
	"testing"
)

var _ = fmt.Printf
var _ = ed25519.Sign
var _ = rand.New

func Test_Auth1_Equals(test *testing.T) {

	scd := new(SCDatabase) // Get me a database
	scd.Init()             
	
	ecAdr := simplecoin.Sha([]byte("ec one")) // Get me an address
	var balance uint64 = 1000               

	bal := new([8]byte)                // Someplace to put a balance
	binary.PutUvarint(bal[:], balance) // Write a balance

	scd.Put("ec", ecAdr, bal[:]) // Write balance to db
	rBal := scd.Get("ec", ecAdr) // Get it back.

	newBalance, _ := binary.Uvarint(rBal) //
	if newBalance != balance {            // Make sure we got it back.
		test.Fail() 
	}

}
