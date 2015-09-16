// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wallet

import (
	"encoding/binary"
	"fmt"
	"github.com/FactomProject/ed25519"
	"github.com/FactomProject/factoid"
	"math/rand"
	"testing"
)

var _ = fmt.Printf
var _ = ed25519.Sign
var _ = rand.New
var _ = binary.Write

func Test_create_walletentry(test *testing.T) {
	w := new(SCWallet) // make me a wallet
	w.Init()
	w.NewSeed([]byte("lkdfsgjlagkjlasd"))
	we := new(WalletEntry)
	rcd := new(factoid.RCD_1)
	name := "John Smith"
	adrtype := "fct"
	pub, pri, err := w.generateKey()

	if err != nil {
		factoid.Prtln("Generate Failed")
		test.Fail()
	}

	we.SetRCD(rcd)
	we.AddKey(pub, pri)
	we.SetName([]byte(name))
	we.SetType(adrtype)

	data, err := we.MarshalBinary()
	if err != nil {
		test.Fail()
	}

	w2 := new(WalletEntry)

	data, err = w2.UnmarshalBinaryData(data)
	if err != nil {
		test.Fail()
	}

	if we.IsEqual(w2) != nil {
		test.Fail()
	}
}
