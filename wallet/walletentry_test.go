// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wallet

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
var _ = binary.Write

func Test_create_walletentry(test *testing.T) {
    w := new(SCWallet)          // make me a wallet
    we := new(WalletEntry)
    rcd := new(simplecoin.RCD_1)
    name := "John Smith"
    pub, pri, err := w.generateKey()
    
    if err != nil {
        simplecoin.Prtln("Generate Failed")
        test.Fail()
    }
    
    we.SetRCD(rcd)
    we.AddKey(pub,pri)
    we.SetName(name)

}