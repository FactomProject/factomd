// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package block

import (
    "encoding/binary"
    "fmt"
    "github.com/agl/ed25519"
    sc "github.com/FactomProject/simplecoin"
    "math/rand"
    "testing"
    
)

var _ = sc.Prt
var _ = fmt.Printf
var _ = ed25519.Sign
var _ = rand.New
var _ = binary.Write

func Test_create_genesis_block(test *testing.T) {
    gb := GetGenesisBlock(1000000,10,200000000000)
    var _ = gb
    sc.Prtln(gb)
    sc.Prtln("Hash: ",gb.GetHash())
}