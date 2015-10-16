// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package block

import (
	"encoding/binary"
	"fmt"
	"github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/factoid/block"
	"github.com/FactomProject/factomd/common/factoid/wallet"
	"github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/primitives"
	"math/rand"
	cv "strconv"
	"testing"
	"time"
)

var _ = Prt
var _ = fmt.Printf
var _ = ed25519.Sign
var _ = rand.New
var _ = binary.Write

var fakeAddr interfaces.IHash = new(Hash)

func newFakeAddr() interfaces.IAddress {
	fakeAddr = Sha(fakeAddr.Bytes())
	return fakeAddr
}

func Test_create_block(test *testing.T) {
	w := new(wallet.SCWallet) // make me a wallet
	w.Init()
	w.NewSeed([]byte("slfkjasdlfjasflajsfl"))
	scb := block.NewFBlock(1000, 0)
	cb := w.CreateTransaction(uint64(time.Now().UnixNano() / 1000000))
	scb.AddCoinbase(cb)

	for i := 0; i < 3; i++ {
		h0, err := w.GenerateFctAddress([]byte("test "+cv.Itoa(i)+"-0"), 1, 1)
		if err != nil {
			Prtln("Error 1")
			test.Fail()
		}
		h1, err := w.GenerateFctAddress([]byte("test "+cv.Itoa(i)+"-1"), 1, 1)
		if err != nil {
			Prtln("Error 2")
			test.Fail()
		}
		h2, err := w.GenerateFctAddress([]byte("test "+cv.Itoa(i)+"-2"), 1, 1)
		if err != nil {
			Prtln("Error 3")
			test.Fail()
		}
		h3, err := w.GenerateFctAddress([]byte("test "+cv.Itoa(i)+"-3"), 1, 1)
		if err != nil {
			Prtln("Error 4")
			test.Fail()
		}
		h4, err := w.GenerateFctAddress([]byte("test "+cv.Itoa(i)+"-4"), 1, 1)
		if err != nil {
			Prtln("Error 5")
			test.Fail()
		}
		h5, err := w.GenerateFctAddress([]byte("test "+cv.Itoa(i)+"-5"), 1, 1)
		if err != nil {
			Prtln("Error 6")
			test.Fail()
		}

		t := w.CreateTransaction(uint64(time.Now().UnixNano() / 1000000))

		w.AddInput(t, h1, 1000000)
		w.AddInput(t, h2, 1000000)
		w.AddOutput(t, h3, 1000000)
		w.AddOutput(t, h4, 500000)
		w.AddECOutput(t, h5, 500000)
		w.AddInput(t, h0, 0)
		fee, err := t.CalculateFee(1000)
		w.UpdateInput(t, 2, h0, fee)

		signed, err := w.SignInputs(t)
		if err != nil {
			Prtln("Error found: ", err)
			test.Fail()
			return
		}
		if !signed {
			Prtln("Not valid")
			test.Fail()
			return
		}

		err = scb.AddTransaction(t)
		if err != nil {
			Prtln("Error found: ", err)
			test.Fail()
			return
		}
	}
	data, err := scb.MarshalBinary()
	if err != nil {
		fmt.Println(err)
		test.Fail()
		return
	}
	scb2 := new(block.FBlock)
	_, err = scb2.UnmarshalBinaryData(data)

	fmt.Println("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\n", scb2)

	if err != nil {
		fmt.Println(err)
		test.Fail()
		return
	}
	//Prtln("FIRST\n",scb,"SECOND\n",scb2)
	if scb.IsEqual(scb2) != nil {
		fmt.Println(err)
		test.Fail()
		return
	}

}
