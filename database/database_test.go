// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
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
        
        var sig1 [64]byte                               // Make a "random" 64 byte thing
        copy(sig1[:32],Sha([]byte("one")).Bytes())
        copy(sig1[32:],Sha([]byte("two")).Bytes())
        var sig2 [64]byte                               // Make another 64 byte thing
        copy(sig2[:32],Sha([]byte("three")).Bytes())
        copy(sig2[32:],Sha([]byte("four")).Bytes())