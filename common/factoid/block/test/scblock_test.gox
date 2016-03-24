// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package block

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	cv "strconv"
	"testing"
	"time"

	"github.com/FactomProject/ed25519"
	sc "github.com/FactomProject/factoid"
	"github.com/FactomProject/factoid/block"
	"github.com/FactomProject/factoid/wallet"
)

var _ = sc.Prt
var _ = fmt.Printf
var _ = ed25519.Sign
var _ = rand.New
var _ = binary.Write

var fakeAddr sc.IHash = new(sc.Hash)

func newFakeAddr() sc.IAddress {
	fakeAddr = sc.Sha(fakeAddr.Bytes())
	return fakeAddr
}

func Test_create_block(test *testing.T) {
	w := new(wallet.SCWallet) // make me a wallet
	w.Init()
	w.NewSeed([]byte("slfkjasdlfjasflajsfl"))
	scb := block.NewFBlock(1000, 0)
	cb := w.CreateTransaction(uint64(time.Now().UnixNano() / 1000000))
	scb.AddCoinbase(cb)

	for i := 1; i < 13; i++ {
		h0, err := w.GenerateFctAddress([]byte("test "+cv.Itoa(i)+"-0"), 1, 1)
		if err != nil {
			sc.Prtln("Error 1")
			test.Fail()
		}
		h1, err := w.GenerateFctAddress([]byte("test "+cv.Itoa(i)+"-1"), 1, 1)
		if err != nil {
			sc.Prtln("Error 2")
			test.Fail()
		}
		h2, err := w.GenerateFctAddress([]byte("test "+cv.Itoa(i)+"-2"), 1, 1)
		if err != nil {
			sc.Prtln("Error 3")
			test.Fail()
		}
		h3, err := w.GenerateFctAddress([]byte("test "+cv.Itoa(i)+"-3"), 1, 1)
		if err != nil {
			sc.Prtln("Error 4")
			test.Fail()
		}
		h4, err := w.GenerateFctAddress([]byte("test "+cv.Itoa(i)+"-4"), 1, 1)
		if err != nil {
			sc.Prtln("Error 5")
			test.Fail()
		}
		h5, err := w.GenerateFctAddress([]byte("test "+cv.Itoa(i)+"-5"), 1, 1)
		if err != nil {
			sc.Prtln("Error 6")
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
			sc.Prtln("Error found: ", err)
			test.Fail()
			return
		}
		if !signed {
			sc.Prtln("Not valid")
			test.Fail()
			return
		}

		err = scb.AddTransaction(t)
		if err != nil {
			sc.Prtln("Error found: ", err)
			test.Fail()
			return
		}

		if i == 2 {
			scb.EndOfPeriod(1)
		} else if i == 5 {
			scb.EndOfPeriod(2)
		} else if i == 6 {
			scb.EndOfPeriod(3)
			scb.EndOfPeriod(4)
		} else if i == 10 {
			scb.EndOfPeriod(5)
		} else if i == 12 {
			scb.EndOfPeriod(6)
		}
	}
	scb.EndOfPeriod(7)
	scb.EndOfPeriod(8)
	scb.EndOfPeriod(9)
	scb.EndOfPeriod(10)

	fmt.Println("SCB Period marks: ", scb.GetEndOfPeriod())

	data, err := scb.MarshalBinary()
	if err != nil {
		fmt.Println(err)
		test.Fail()
		return
	}
	scb2 := new(block.FBlock)
	_, err = scb2.UnmarshalBinaryData(data)

	data2, err2 := scb.MarshalBinary()
	if err2 != nil {
		fmt.Println(err2)
		test.Fail()
		return
	}

	result := bytes.Compare(data, data2)
	fmt.Println("data2==data: ", result)
	fmt.Println("SCB2 Period marks: ", scb2.GetEndOfPeriod())
	//fmt.Println("fblock hex: ", hex.EncodeToString(data))

	scb3 := new(block.FBlock)
	scb3.UnmarshalBinaryData(data2)

	data3, err3 := scb3.MarshalBinary()
	if err3 != nil {
		fmt.Println(err3)
		test.Fail()
		return
	}

	result = bytes.Compare(data2, data3)
	fmt.Println("data2==data3: ", result)
	fmt.Println("SCB3 Period marks: ", scb3.GetEndOfPeriod())

	//fmt.Println("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\n", scb2)

	if err != nil {
		fmt.Println(err)
		test.Fail()
		return
	}
	//sc.Prtln("FIRST\n", scb, "SECOND\n", scb2)
	if scb.IsEqual(scb2) != nil {
		fmt.Println("NOT EQUAL: scb.IsEqual(scb2)")
		//fmt.Println("scb: ", scb)
		//fmt.Println()
		//fmt.Println()
		//fmt.Println("scb2: ", scb)
		test.Fail()
		return
	}

	if scb3.IsEqual(scb2) != nil {
		fmt.Println("NOT EQUAL: scb3.IsEqual(scb2)")
		test.Fail()
		return
	}

}

func Test_gen_block(test *testing.T) {
	scb := block.GetGenesisFBlock()

	fmt.Println("SCB Period marks: ", scb.GetEndOfPeriod())

	data, err := scb.MarshalBinary()
	if err != nil {
		fmt.Println(err)
		test.Fail()
		return
	}
	scb2 := new(block.FBlock)
	_, err = scb2.UnmarshalBinaryData(data)

	data2, err2 := scb.MarshalBinary()
	if err2 != nil {
		fmt.Println(err2)
		test.Fail()
		return
	}

	result := bytes.Compare(data, data2)
	fmt.Println("data2==data: ", result)
	fmt.Println("SCB2 Period marks: ", scb2.GetEndOfPeriod())
	//fmt.Println("fblock hex: ", hex.EncodeToString(data))

	scb3 := new(block.FBlock)
	scb3.UnmarshalBinaryData(data2)

	data3, err3 := scb3.MarshalBinary()
	if err3 != nil {
		fmt.Println(err3)
		test.Fail()
		return
	}

	result = bytes.Compare(data2, data3)
	fmt.Println("data2==data3: ", result)
	fmt.Println("SCB3 Period marks: ", scb3.GetEndOfPeriod())

	//fmt.Println("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\n", scb2)

	if err != nil {
		fmt.Println(err)
		test.Fail()
		return
	}
	//sc.Prtln("FIRST\n", scb, "SECOND\n", scb2)
	if scb.IsEqual(scb2) != nil {
		fmt.Println("NOT EQUAL: scb.IsEqual(scb2)")
		//fmt.Println("scb: ", scb)
		//fmt.Println()
		//fmt.Println()
		//fmt.Println("scb2: ", scb)
		test.Fail()
		return
	}

	if scb3.IsEqual(scb2) != nil {
		fmt.Println("NOT EQUAL: scb3.IsEqual(scb2)")
		test.Fail()
		return
	}

}

func Test_alot_block(test *testing.T) {
	w := new(wallet.SCWallet) // make me a wallet
	w.Init()
	w.NewSeed([]byte("slfkjasdlfjasflajsfl"))

	h0, err := w.GenerateFctAddress([]byte("test "+cv.Itoa(1)+"-1"), 1, 1)
	if err != nil {
		sc.Prtln("Error 1")
		test.Fail()
	}
	h1, err := w.GenerateFctAddress([]byte("test "+cv.Itoa(2)+"-1"), 1, 1)
	if err != nil {
		sc.Prtln("Error 2")
		test.Fail()
	}
	h2, err := w.GenerateFctAddress([]byte("test "+cv.Itoa(3)+"-2"), 1, 1)
	if err != nil {
		sc.Prtln("Error 3")
		test.Fail()
	}
	h3, err := w.GenerateFctAddress([]byte("test "+cv.Itoa(4)+"-3"), 1, 1)
	if err != nil {
		sc.Prtln("Error 4")
		test.Fail()
	}
	h4, err := w.GenerateFctAddress([]byte("test "+cv.Itoa(5)+"-4"), 1, 1)
	if err != nil {
		sc.Prtln("Error 5")
		test.Fail()
	}
	h5, err := w.GenerateFctAddress([]byte("test "+cv.Itoa(6)+"-5"), 1, 1)
	if err != nil {
		sc.Prtln("Error 6")
		test.Fail()
	}

	for bcnt := 0; bcnt < 500; bcnt++ {
		scb := block.NewFBlock(1000, 0)
		cb := w.CreateTransaction(uint64(time.Now().UnixNano() / 1000000))
		scb.AddCoinbase(cb)

		for min := 1; min <= 10; min++ {
			max := 0
			switch rand.Int() % 4 {
				case 0:
					max = 0
				case 1:
					max = rand.Int()%1
				case 2:
					max = rand.Int()%2
				case 3:
					max = rand.Int()%3
			}
			
			for j := 0; j < max ; j++ {
				
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
					sc.Prtln("Error found: ", err)
					test.Fail()
					return
				}
				if !signed {
					sc.Prtln("Not valid")
					test.Fail()
					return
				}

				err = scb.AddTransaction(t)
				if err != nil {
					sc.Prtln("Error found: ", err)
					test.Fail()
					return
				}
				if err != nil {
					sc.Prtln("Error 1")
					test.Fail()
					return
				}
			}
			
			// Note:  All incomplete blocks have at least one period with 
			// a zero value.
			if scb.GetEndOfPeriod()[9]!= 0 {
				test.Fail()
				return
			}
			
			scb.EndOfPeriod(min)
		}

		fmt.Println("SCB Period marks for blk ",bcnt,": ", scb.GetEndOfPeriod(),"             ")

		data, err := scb.MarshalBinary()
		if err != nil {
			fmt.Println(err)
			test.Fail()
			return
		}
		
		scb2 := new(block.FBlock)
		_, err = scb2.UnmarshalBinaryData(data)

		data2, err2 := scb.MarshalBinary()
		if err2 != nil {
			fmt.Println(err2)
			test.Fail()
			return
		}

		result := bytes.Compare(data, data2)
		if result != 0 {
			fmt.Println("Failure at ",bcnt,".  Data does not compare")
			test.Fail()
			return
		}
		
		scb3 := new(block.FBlock)
		scb3.UnmarshalBinaryData(data2)

		data3, err3 := scb3.MarshalBinary()
		if err3 != nil {
			fmt.Println(err3)
			test.Fail()
			return
		}
		
		result = bytes.Compare(data2, data3)
		if result != 0 {
			fmt.Println("Failure at ",bcnt,".  Data does not compare")
			test.Fail()
			return
		}
			

		if err != nil {
			fmt.Println(err)
			test.Fail()
			return
		}
		//sc.Prtln("FIRST\n", scb, "SECOND\n", scb2)
		if scb.IsEqual(scb2) != nil {
			fmt.Println("NOT EQUAL: scb.IsEqual(scb2)")
			test.Fail()
			return
		}

		if scb3.IsEqual(scb2) != nil {
			fmt.Println("NOT EQUAL: scb3.IsEqual(scb2)")
			test.Fail()
			return
		}
		
		scbStr := scb.String()
		scb2Str := scb2.String()
		scb3Str := scb3.String()
		
		if scbStr != scb2Str || scbStr != scb3Str {
			fmt.Println("String prints don't match")
			test.Fail()
			return
		}
		
	}
}
