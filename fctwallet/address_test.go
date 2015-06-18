// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
    "encoding/hex"
    "encoding/binary"
    "bytes"
    "fmt"
    fct "github.com/FactomProject/factoid"
    "github.com/agl/ed25519"
    "math/rand"
    "testing"
    
)

var _ = hex.EncodeToString
var _ = fmt.Printf
var _ = ed25519.Sign
var _ = rand.New
var _ = binary.Write
var _ = fct.Prtln   


func Test_test_Factoid_Addresses(test *testing.T) {

    addr := fct.NewAddress(fct.Sha([]byte("A fake address")).Bytes())
    fmt.Println( addr)
    
    uaddr := ConvertFAddressToUserStr(addr) 
    fmt.Println(uaddr)

    if !ValidateFUserStr(uaddr) { test.Fail() }
    
    addrBack := ConvertUserStrToAddress(uaddr)
    
    if bytes.Compare(addrBack,addr.Bytes()) != 0 { test.Fail() }
    
    buaddr := []byte(uaddr)
    
    for i,v := range buaddr {
        for j:= uint(0); j<8; j++ {
            if !ValidateFUserStr(string(buaddr)) { test.Fail() }
            buaddr[i] = v^(01<<j)
            if ValidateFUserStr(string(buaddr)) { test.Fail() }
            buaddr[i] = v
        }
    }
}

func Test_test_Entry_Credit_Addresses(test *testing.T) {
    
    addr := fct.NewAddress(fct.Sha([]byte("A fake address")).Bytes())
    fmt.Println( addr)
    
    uaddr := ConvertECAddressToUserStr(addr) 
    fmt.Println(uaddr)
    
    if !ValidateECUserStr(uaddr) { test.Fail() }
    
    addrBack := ConvertUserStrToAddress(uaddr)
    
    if bytes.Compare(addrBack,addr.Bytes()) != 0 { test.Fail() }
    
    buaddr := []byte(uaddr)
    
    for i,v := range buaddr {
        for j:= uint(0); j<8; j++ {
            if !ValidateECUserStr(string(buaddr)) { test.Fail() }
            buaddr[i] = v^(01<<j)
            if ValidateECUserStr(string(buaddr)) { test.Fail() }
            buaddr[i] = v
        }
    }
}
