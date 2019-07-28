// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives_test

import (
	"fmt"
	"math/rand"
	"testing"

	"bytes"

	"encoding/hex"

	"github.com/FactomProject/factomd/common/factoid"
	. "github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/testHelper"
)

// TestCalculateCoinbasePayout given a set of input efficiencies and their known outputs
func TestCalculateCoinbasePayout(t *testing.T) {
	testIt := func(eff uint16, exp uint64) {
		if CalculateCoinbasePayout(eff) != exp {
			t.Errorf("Expected %d, got %d", exp, CalculateCoinbasePayout(eff))
		}
	}

	testIt(10000, 0)
	testIt(0, 6.4*1e8)
	testIt(4952, 3.23072*1e8)

}

// TestEfficiencyToString checks that input efficiencies are properly converted to strings
func TestEfficiencyToString(t *testing.T) {
	testIt := func(n uint16, e string) {
		if EfficiencyToString(n) != e {
			t.Errorf("Expected %s, got %s", e, EfficiencyToString(n))
		}
	}

	testIt(10000, "100.00")
	testIt(4952, "49.52")
	testIt(0x1358, "49.52")
	testIt(100, "1.00")
	testIt(2945, "29.45")
	testIt(10, "0.10")
}

// TestPrintHelp checks that input numbers will have commas inseted in the proper places
func TestPrintHelp(test *testing.T) {
	t := func(v int64, str string) {
		s := AddCommas(v)
		if s != str {
			fmt.Println("For", v, "Expected", str, "and got", s)
			test.Fail()
		}
	}
	t(0, "0")
	t(1, "1")
	t(99, "99")
	t(1000, "1,000")
	t(1001, "1,001")
	t(1100, "1,100")
	t(300002100, "300,002,100")
	t(4300002100, "4,300,002,100")
	t(1001, "1,001")
	t(1100, "1,100")
	t(-1, "-1")
	t(-99, "-99")
	t(-1000, "-1,000")
	t(-1001, "-1,001")
	t(-1100, "-1,100")
	t(-300002100, "-300,002,100")
	t(-4300002100, "-4,300,002,100")
}

// TestConversions tests a variety of Factoid to Factoshi inputs
func TestConversions(test *testing.T) {
	v, err := ConvertFixedPoint(".999")
	if err != nil || v != "99900000" {
		fmt.Println("1", v, err)
		test.Fail()
	}
	v, err = ConvertFixedPoint("0.999")
	if err != nil || v != "99900000" {
		fmt.Println("2", v, err)
		test.Fail()
	}
	v, err = ConvertFixedPoint("10.999")
	if err != nil || v != "1099900000" {
		fmt.Println("3", v, err)
		test.Fail()
	}
	v, err = ConvertFixedPoint(".99999999999999")
	if err != nil || v != "99999999" {
		fmt.Println("4", v, err)
		test.Fail()
	}
}

// TestWriteNumber checks that the WriteNumber* functions work properly
func TestWriteNumber(t *testing.T) {
	out := new(Buffer)

	WriteNumber8(out, 0x01)
	WriteNumber16(out, 0x0203)
	WriteNumber32(out, 0x04050607)
	WriteNumber64(out, 0x0809101112131415)

	answer := "010203040506070809101112131415"
	if out.String() != answer {
		t.Errorf("Failed WriteNumbers. Expected %v, got %v", out.String(), answer)
	}
}

// TestConversion checsk that the Factoshi to Factoid conversion works properly
func TestConversion(t *testing.T) {
	var num uint64 = 123456789
	if ConvertDecimalToString(num) != "1.23456789" {
		t.Error("Failed ConvertDecimalToString")
	}
	if ConvertDecimalToPaddedString(num) != "            1.23456789" {
		t.Errorf("Failed ConvertDecimalToPaddedString - '%v'", ConvertDecimalToPaddedString(num))
	}
}

// TestVariable_Integers does the following 1000 time:
// 1) create an array of 10 random rumbers of varying lengths and sizes
// 2) writes them to a buffer using the varint
// 3) decodes the array 1000 times (I don't know why it does this,
//    if you can decode it 1 time, is it expected to change? Or is this a mis-written test)
func TestVariable_Integers(test *testing.T) {
	for i := 0; i < 1000; i++ {
		var out Buffer

		v := make([]uint64, 10)

		for j := 0; j < len(v); j++ {
			var m uint64           // 64 bit mask
			sw := rand.Int63() % 4 // Pick a random choice
			switch sw {
			case 0:
				m = 0xFF // Random byte
			case 1:
				m = 0xFFFF // Random 16 bit integer
			case 2:
				m = 0xFFFFFFFF // Random 32 bit integer
			case 3:
				m = 0xFFFFFFFFFFFFFFFF // Random 64 bit integer
			}
			n := uint64(rand.Int63() + (rand.Int63() << 32))
			v[j] = n & m
		}

		for j := 0; j < len(v); j++ { // Encode our entire array of numbers
			err := EncodeVarInt(&out, v[j])
			if err != nil {
				fmt.Println(err)
				test.Fail()
				return
			}
			//              fmt.Printf("%x ",v[j])
		}
		//          fmt.Println( "Length: ",out.Len())

		data := out.Bytes()

		//          PrtData(data)
		//          fmt.Println()
		sdata := data // Decode our entire array of numbers, and
		var dv uint64 // check we got them back correctly.
		for k := 0; k < 1000; k++ {
			data = sdata
			for j := 0; j < len(v); j++ {
				dv, data = DecodeVarInt(data)
				if dv != v[j] {
					fmt.Printf("Values don't match: decode:%x expected:%x (%d)\n", dv, v[j], j)
					test.Fail()
					return
				}
			}
		}
	}
}

// TestValidateUserStr validates a set of fixed addresses, and creates 1000 FA addresses
// and verifies them
func TestValidateUserStr(t *testing.T) {
	fctAdd := "FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q"
	fctAddSecret := "Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK"
	ecAdd := "EC2DKSYyRcNWf7RS963VFYgMExoHRYLHVeCfQ9PGPmNzwrcmgm2r"

	ok := ValidateFUserStr(fctAdd)
	if ok == false {
		t.Errorf("Valid address not validating - %v", fctAdd)
	}

	ok = ValidateECUserStr(ecAdd)
	if ok == false {
		t.Errorf("Valid address not validating - %v", fctAdd)
	}

	ok = ValidateFPrivateUserStr(fctAddSecret)
	if ok == false {
		t.Errorf("Valid address not validating - %v", fctAdd)
	}

	factoidAddresses := []string{}
	//ecAddresses:=[]string{}

	max := 1000

	for i := 0; i < max; i++ {
		_, _, add := testHelper.NewFactoidAddressStrings(uint64(i))
		factoidAddresses = append(factoidAddresses, add)

		//ecAddresses = append(ecAddresses, add)
	}

	for _, v := range factoidAddresses {
		ok := ValidateFUserStr(v)
		if ok == false {
			t.Errorf("Valid address not validating - %v", v)
		}
	}
}

// TestAddressConversions creates one each of FA, EC, Fs, and Es from a key and verifies that the expected
// user facing address is returned
func TestAddressConversions(t *testing.T) {
	//https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#factoid-address
	pub := "0000000000000000000000000000000000000000000000000000000000000000"
	user := "FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu"

	h, err := NewShaHashFromStr(pub)
	if err != nil {
		t.Errorf("%v", err)
	}
	add := factoid.CreateAddress(h)

	converted := ConvertFctAddressToUserStr(add)
	if converted != user {
		t.Errorf("Wrong conversion - %v vs %v", converted, user)
	}

	//https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#entry-credit-address
	pub = "0000000000000000000000000000000000000000000000000000000000000000"
	user = "EC1m9mouvUQeEidmqpUYpYtXg8fvTYi6GNHaKg8KMLbdMBrFfmUa"

	h, err = NewShaHashFromStr(pub)
	if err != nil {
		t.Errorf("%v", err)
	}
	add = factoid.CreateAddress(h)

	converted = ConvertECAddressToUserStr(add)
	if converted != user {
		t.Errorf("Wrong conversion - %v vs %v", converted, user)
	}

	//https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#factoid-private-keys
	priv := "0000000000000000000000000000000000000000000000000000000000000000"
	user = "Fs1KWJrpLdfucvmYwN2nWrwepLn8ercpMbzXshd1g8zyhKXLVLWj"
	h, err = NewShaHashFromStr(priv)
	if err != nil {
		t.Errorf("%v", err)
	}
	add = factoid.CreateAddress(h)
	converted = ConvertFctPrivateToUserStr(add)
	if converted != user {
		t.Errorf("Wrong conversion - %v vs %v", converted, user)
	}

	//https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#entry-credit-private-keys
	priv = "0000000000000000000000000000000000000000000000000000000000000000"
	user = "Es2Rf7iM6PdsqfYCo3D1tnAR65SkLENyWJG1deUzpRMQmbh9F3eG"
	h, err = NewShaHashFromStr(priv)
	if err != nil {
		t.Errorf("%v", err)
	}
	add = factoid.CreateAddress(h)
	converted = ConvertECPrivateToUserStr(add)
	if converted != user {
		t.Errorf("Wrong conversion - %v vs %v", converted, user)
	}
}

// TestHumanReadableAddressConvert checks that the conversion from human readable to address works:
// 1) Address -> Human Readable -> Address
// 2) Human readable -> Address -> Human Readable
func TestHumanReadableAddressConvert(t *testing.T) {
	// Test going from Address -> Human Readable -> Address
	for i := 0; i < 1000; i++ {
		add := factoid.RandomAddress()
		hr := ConvertFctAddressToUserStr(add)
		if bytes.Compare(ConvertUserStrToAddress(hr), add.Bytes()) != 0 {
			t.Errorf("User string does not match")
		}
	}

	// Test Human readable -> Address -> Human Readable
	humanReadables := []string{
		"FA3oajkmHMfqkNMMShmqpwDThzMCuVrSsBwiXM2kYFVRz3MzxNAJ",
		"FA3Ga2XcaheS5NgQ3q22gBpLgE6tXmPu1GhjdU2FsdN2QPMzKJET",
		"FA3GH7VEFKqTdJcmwGgDrcY4Xh9njQ4EWiJxhJeim6BCA7QuB388",
		"FA2k8ULFMGVxgXaj2XwJeaQXHGRt5xd7DJ6wjGf4t2KtgwRkhfny",
		"FA2qG68LqMoyyv1iJaypaXWhKDJg76XU3i3vV4v5Ah8s5UryECR4",
		"FA3ipMRDLR9SCgdJApqKHiNzZRcB4tRQFZb3wCD8u5horBWcsJ3u",
		"FA21AsVeHMXSeDuHe9QrZxs3vDotoQV8Ri8WdKpb62wmKcYykKpb",
		"FA3ifuDtDgna6gCFfxX1kbDjzZ8EMK3fNFkPEZKALkNtPmWtXSoR",
	}
	for _, hr := range humanReadables {
		if hr != ConvertFctAddressToUserStr(factoid.NewAddress(ConvertUserStrToAddress(hr))) {
			t.Errorf("Human readable does not match")
		}
	}

	// Test Vectors
	// 	map[hr]add
	vectors := map[string]string{
		"FA3ifuDtDgna6gCFfxX1kbDjzZ8EMK3fNFkPEZKALkNtPmWtXSoR": "E6AD686AA2E1098DA3D6C44199AE7C51DD40625E47EDABFF3A03BA01CDAD6257",
		"FA23Lcqqm51BjvcVxY4w9XEvYQKxKfj3KXWPU97XuZNXgm8jAcHJ": "09AC108C627819760D78BACDFCF490D0F6C7C35DFF179EDB6F614B7F2D7C9E2D",
		"FA3JvZx1EKW6zM9rSZff9nqBjrKmjWK52FWEmieRm9K9XetPC7Bn": "B0C1B468EB5A8EA2A7740D98882757F2B80EBAE1F528D1877A8D14EA8EA6D185",
	}

	for k, v := range vectors {
		b, _ := hex.DecodeString(v)
		if k != ConvertFctAddressToUserStr(factoid.NewAddress(b)) {
			t.Errorf("Bytes to Human readable failed")
		}

		if bytes.Compare(ConvertUserStrToAddress(k), b) != 0 {
			t.Errorf("Human readable to bytes failed")
		}
	}

}
