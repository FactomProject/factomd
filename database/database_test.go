// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package database

/*
import (
	"encoding/binary"
	"fmt"
	"github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/primitives"
	"math/rand"
	"testing"
)

var _ = fmt.Printf
var _ = ed25519.Sign
var _ = rand.New
var _ = binary.Read

type t_balance struct {
	interfaces.IBlock
	balance uint64
}

func Test_Auth1_Equals(test *testing.T) {

	scd := new(MapDB) // Get me a database
	scd.Init()

	ecAdr := Sha([]byte("ec one")) // Get me an address
	b := new(t_balance)            // Get a balance interfaces.IBlock
	b.balance = 1000               // Set the balance

	scd.Put("ec", ecAdr, b)    // Write balance to db
	b2 := scd.Get("ec", ecAdr) // Get it back.

	if b.balance != b2.(*t_balance).balance { // Make sure we got it back.
		test.Fail()
	}

}
*/
