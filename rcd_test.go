// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package simplecoin

import (
    "fmt"
    "github.com/agl/ed25519"
    "math/rand"
    "testing"
)

var _ = fmt.Printf
var _ = ed25519.Sign
var _ = rand.New

func Test_Auth1_Equals(test *testing.T) {
    
    var sig1 [64]byte
    copy(sig1[:32],Sha([]byte("one")).Bytes())
    copy(sig1[32:],Sha([]byte("two")).Bytes())
    var sig2 [64]byte
    copy(sig2[:32],Sha([]byte("three")).Bytes())
    copy(sig2[32:],Sha([]byte("four")).Bytes())
    
    a1 := new(RCD_1)
    a1.signature = sig1[:]
    a2 := new(RCD_1)
    a2.signature = sig1[:]
    
    if !a1.IsEqual(a2) {        
        PrtStk()
        test.Fail()
    }
    
    a1.signature = sig2[:]
    
    if a1.IsEqual(a2) {         
        PrtStk()
        test.Fail()
    }
}

func Test_Auth2_Equals(test *testing.T) {
    
    a1 := nextAuth2()
    a2 := a1
    
    if !a1.IsEqual(a2) {        
        PrtStk()
        test.Fail()
    }
    
    a1 = nextAuth2()
    
    if a1.IsEqual(a2) {         
        PrtStk()
        test.Fail()
    }
}