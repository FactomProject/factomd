// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/FactomProject/factomd/anchor"

	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/common/primitives"
)

// captureOutput redirects screen prints to a string for ease of testing
func captureOutput(f func()) string {
	// Change stdout to the pipe
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f() // Call your function that prints to stdout

	// Reset stdout
	w.Close()
	os.Stdout = old

	// Return stdout from function call as string
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// TestLog captures the output from the Log() function and compares to a expected, fixed string
func TestLog(t *testing.T) {
	// We use an AnchorRecord as an arbitrary data structure for testing
	record := `{"AnchorRecordVer":1,"DBHeight":5,"KeyMR":"980ab6d50d9fad574ad4df6dba06a8c02b1c67288ee5beab3fbfde2723f73ef6","RecordHeight":6,"Bitcoin":{"Address":"1K2SXgApmo9uZoyahvsbSanpVWbzZWVVMF","TXID":"e2ac71c9c0fd8edc0be8c0ba7098b77fb7d90dcca755d5b9348116f3f9d9f951","BlockHeight":372576,"BlockHash":"000000000000000003059382ed4dd82b2086e99ec78d1b6e811ebb9d53d8656d","Offset":1144}}`
	ar, _ := anchor.UnmarshalAnchorRecord([]byte(record))

	// Test the Log() function with arbitrary inputs
	screenOutput := captureOutput(func() { Log("My TestLog Print Message:\nNumber: %v\nString: %s\nAnchor: %v", 123, "abc", ar) })

	// Strip off the filepath to only get the Go file (since full paths differ based on individual repos)
	_, file := filepath.Split(screenOutput)
	// Expected string should exactly follow string format input to Log() above. Note '\n' added to
	// end of string as performed inside Log(). Also prepend with file and line number. How to not use const line#?
	expected := "log_test.go:47 - My TestLog Print Message:\nNumber: 123\nString: abc\nAnchor: " + record + "\n"
	if file != expected {
		t.Error(fmt.Printf("Screen Ouptut string does not equal expected:\nScreen:\n%s\nExpected:\n%s\n", file, expected))
	}
}

// TestLog captures the output from the LogJSONs() function and compares to a expected, fixed string
func TestLogJSONs(t *testing.T) {
	// We use an AnchorRecord as an arbitrary data structure for testing
	record := `{"AnchorRecordVer":1,"DBHeight":5,"KeyMR":"980ab6d50d9fad574ad4df6dba06a8c02b1c67288ee5beab3fbfde2723f73ef6","RecordHeight":6,"Bitcoin":{"Address":"1K2SXgApmo9uZoyahvsbSanpVWbzZWVVMF","TXID":"e2ac71c9c0fd8edc0be8c0ba7098b77fb7d90dcca755d5b9348116f3f9d9f951","BlockHeight":372576,"BlockHash":"000000000000000003059382ed4dd82b2086e99ec78d1b6e811ebb9d53d8656d","Offset":1144}}`
	ar, _ := anchor.UnmarshalAnchorRecord([]byte(record))

	// Test the LogJSONs() function with arbitrary inputs
	screenOutput := captureOutput(func() { LogJSONs("My TestLog Print Message:\nNumber: %v\nString: %s\nAnchor: %v", 123, "abc", ar) })

	// Strip off the filepath to only get the Go file (since full paths differ based on individual repos)
	_, file := filepath.Split(screenOutput)
	// Expected string should exactly follow string format input to Log() above. Note '\n' added to
	// end of string as performed inside LogJSONs(). Also prepend with file and line number. How to not use const line#?
	// NOTE: This expected string is different from the TestLog() above because "abc" has explicit ""s around it
	expected := "log_test.go:66 - My TestLog Print Message:\nNumber: 123\nString: \"abc\"\nAnchor: " + record + "\n"
	if file != expected {
		t.Error(fmt.Printf("Screen Ouptut string does not equal expected:\nScreen:\n%s\nExpected:\n%s\n", file, expected))
	}
}

// TestEncodeJSON test the json formating by encoding and decoding a struct and making sure its the same after
// both operations
func TestEncodeJSON(t *testing.T) {
	ar := new(anchor.AnchorRecord)
	ar.AnchorRecordVer = 1
	ar.DBHeight = 5
	ar.KeyMR = "980ab6d50d9fad574ad4df6dba06a8c02b1c67288ee5beab3fbfde2723f73ef6"
	ar.RecordHeight = 6
	ar.Bitcoin = new(anchor.BitcoinStruct)
	ar.Bitcoin.Address = "1K2SXgApmo9uZoyahvsbSanpVWbzZWVVMF"
	ar.Bitcoin.TXID = "e2ac71c9c0fd8edc0be8c0ba7098b77fb7d90dcca755d5b9348116f3f9d9f951"
	ar.Bitcoin.BlockHeight = 372576
	ar.Bitcoin.BlockHash = "000000000000000003059382ed4dd82b2086e99ec78d1b6e811ebb9d53d8656d"
	ar.Bitcoin.Offset = 1144

	data, err := primitives.EncodeJSON(ar)
	if err != nil {
		t.Error("TestEncodeJSON: Failure to encode")
	}
	ar2 := new(anchor.AnchorRecord)
	err = primitives.DecodeJSON(data, ar2)
	if err != nil || ar.IsSame(ar2) == false {
		t.Error("TestEncodeJSON: Failure to decode")
	}

	arstring, err := primitives.EncodeJSONString(ar)
	if err != nil {
		t.Error("TestEncodeJSON: Failure to encode string")
	}
	ar3 := new(anchor.AnchorRecord)
	err = primitives.DecodeJSONString(arstring, ar3)
	if err != nil || ar.IsSame(ar3) == false {
		t.Error("TestEncodeJSON: Failure to decode string")
	}

	buf := new(bytes.Buffer)
	err = primitives.EncodeJSONToBuffer(ar, buf)
	if err != nil {
		t.Error("TestEncodeJSON: Failure to encode buffer")
	}
	if arstring != buf.String() {
		t.Error("TestEncodeJSON: Failure to decode buffer")
	}
}
