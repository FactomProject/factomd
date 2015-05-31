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
    gb := GetGenesisBlock()
    txt,err := gb.MarshalText()
    if err != nil { test.Fail() }
    sc.Prtln(string(txt))
}