// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package directoryBlock_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"

	. "github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

func TestMarshalUnmarshalDirectoryBlockHeader(t *testing.T) {
	fmt.Println("\n---\nTestMarshalUnmarshalDirectoryBlockHeader\n---\n")

	header := createTestDirectoryBlockHeader()

	bytes1, err := header.MarshalBinary()
	t.Logf("bytes1: %X\n", bytes1)

	header2 := new(DBlockHeader)
	header2.UnmarshalBinary(bytes1)

	bytes2, err := header2.MarshalBinary()
	if err != nil {
		t.Errorf("Error:%v", err)
	}
	t.Logf("bytes2: %X\n", bytes2)

	if bytes.Compare(bytes1, bytes2) != 0 {
		t.Errorf("Invalid output")
	}

}

func TestMarshalUnmarshalDirectoryBlock(t *testing.T) {
	fmt.Println("\n---\nTestMarshalUnmarshalDirectoryBlock\n---\n")

	dblock := createTestDirectoryBlock()

	bytes1, err := dblock.MarshalBinary()
	t.Logf("bytes1: %X\n", bytes1)

	dblock2 := new(DirectoryBlock)
	dblock2.UnmarshalBinary(bytes1)

	bytes2, err := dblock2.MarshalBinary()
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	t.Logf("bytes2: %X\n", bytes2)

	if bytes.Compare(bytes1, bytes2) != 0 {
		t.Errorf("Invalid output")
	}
}

var WeDidPanic bool

func CatchPanic() {
	if r := recover(); r != nil {
		WeDidPanic = true
	}
}

func TestInvalidUnmarshalDirectoryBlockHeader(t *testing.T) {
	fmt.Println("\n---\nTestInvalidUnmarshalDirectoryBlockHeader\n---\n")

	header := createTestDirectoryBlockHeader()

	bytes1, err := header.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	WeDidPanic = false
	defer CatchPanic()

	header2 := new(DBlockHeader)
	err = header2.UnmarshalBinary(nil)
	if err == nil {
		t.Error("We expected errors but we didn't get any")
	}
	if WeDidPanic == true {
		t.Error("We did panic and we shouldn't have")
		WeDidPanic = false
		defer CatchPanic()
	}

	header2 = new(DBlockHeader)
	err = header2.UnmarshalBinary([]byte{})
	if err == nil {
		t.Error("We expected errors but we didn't get any")
	}
	if WeDidPanic == true {
		t.Error("We did panic and we shouldn't have")
		WeDidPanic = false
		defer CatchPanic()
	}

	header2 = new(DBlockHeader)
	err = header2.UnmarshalBinary(bytes1[:len(bytes1)-1])
	if err == nil {
		t.Error("We expected errors but we didn't get any")
	}
	if WeDidPanic == true {
		t.Error("We did panic and we shouldn't have")
		WeDidPanic = false
		defer CatchPanic()
	}
}

func TestInvalidUnmarshalDirectoryBlock(t *testing.T) {
	fmt.Println("\n---\nTestInvalidUnmarshalDirectoryBlock\n---\n")
	dblock := createTestDirectoryBlock()

	bytes1, err := dblock.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	WeDidPanic = false
	defer CatchPanic()

	dblock2 := new(DirectoryBlock)
	err = dblock2.UnmarshalBinary(nil)
	if err == nil {
		t.Error("We expected errors but we didn't get any")
	}
	if WeDidPanic == true {
		t.Error("We did panic and we shouldn't have")
		WeDidPanic = false
		defer CatchPanic()
	}

	dblock2 = new(DirectoryBlock)
	err = dblock2.UnmarshalBinary([]byte{})
	if err == nil {
		t.Error("We expected errors but we didn't get any")
	}
	if WeDidPanic == true {
		t.Error("We did panic and we shouldn't have")
		WeDidPanic = false
		defer CatchPanic()
	}

	dblock2 = new(DirectoryBlock)
	err = dblock2.UnmarshalBinary(bytes1[:len(bytes1)-1])
	if err == nil {
		t.Error("We expected errors but we didn't get any")
	}
	if WeDidPanic == true {
		t.Error("We did panic and we shouldn't have")
		WeDidPanic = false
		defer CatchPanic()
	}
}

func TestMakeSureBlockCountIsNotDuplicates(t *testing.T) {
	fmt.Println("\n---\nTestMakeSureBlockCountIsNotDuplicates\n---\n")
	block := createTestDirectoryBlock()
	err := block.SetDBEntries([]interfaces.IDBEntry{})
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	min := 1000
	max := -1

	for i := 0; i < 100; i++ {
		//Update the BlockCount in header
		block.GetHeader().SetBlockCount(uint32(len(block.GetDBEntries())))
		//Marshal the block
		marshalled, err := block.MarshalBinary()
		if err != nil {
			t.Errorf("Error: %v", err)
		}
		//Get the byte representation of BlockCount
		var buf bytes.Buffer
		binary.Write(&buf, binary.BigEndian, block.GetHeader().GetBlockCount())
		hex := buf.Bytes()

		//How many times does BlockCount appear in the marshalled slice?
		count := bytes.Count(marshalled, hex)
		if count > max {
			max = count
		}
		if count < min {
			min = count
		}

		de := new(DBEntry)
		de.ChainID = primitives.NewZeroHash()
		de.KeyMR = primitives.NewZeroHash()

		err = block.SetDBEntries(append(block.GetDBEntries(), de))
		if err != nil {
			t.Errorf("Error: %v", err)
		}
	}
	t.Logf("Min count - %v, max count - %v", min, max)
	if min != 1 {
		t.Errorf("Invalid number of BlockCount occurances")
	}
}

func createTestDirectoryBlock() *DirectoryBlock {
	dblock := new(DirectoryBlock)

	dblock.SetHeader(createTestDirectoryBlockHeader())

	de := new(DBEntry)
	de.ChainID = primitives.NewZeroHash()
	de.KeyMR = primitives.NewZeroHash()

	err := dblock.SetDBEntries(append(make([]interfaces.IDBEntry, 0, 5), de))
	if err != nil {
		panic(err)
	}
	dblock.GetHeader().SetBlockCount(uint32(len(dblock.GetDBEntries())))

	return dblock
}

func createTestDirectoryBlockHeader() *DBlockHeader {
	header := new(DBlockHeader)

	header.SetDBHeight(1)
	header.SetBodyMR(primitives.Sha(primitives.NewZeroHash().Bytes()))
	header.SetBlockCount(0)
	header.SetNetworkID(0xffff)
	header.SetPrevFullHash(primitives.NewZeroHash())
	header.SetPrevKeyMR(primitives.NewZeroHash())
	header.SetTimestamp(1234)
	header.SetVersion(1)

	return header
}

func TestKeyMRs(t *testing.T) {
	entries := []string{"44c9f3a6d6f6b2ab5efb29e7d6159c4e3fca13fc5dd03b94ae3dea8bf30173cb",
		"41a36ab01a9b8e8d78d6b43b8e7e6671916a93b43b8fec48a627d0cb51f012f1",
		"905740850540f1d17fcb1fc7fd0c61a33150b2cdc0f88334f6a891ec34bd1cfc",
		"9c9610e09673c9136508112fe447c8b9c1e042a95bd140ec161ade4995cd0f73",
		"fbc3a4b40464049c999e99feff2bf36996f27869b045a0374bc47b7c2cda9e7c"}

	chainID, err := primitives.NewShaHashFromStr("0000000000000000000000000000000000000000000000000000000000000001")
	if err != nil {
		t.Error(err)
	}

	dbEntries := []interfaces.IDBEntry{}
	for _, e := range entries {
		h, err := primitives.NewShaHashFromStr(e)
		if err != nil {
			t.Error(err)
		}
		entry := new(DBEntry)
		entry.ChainID = chainID
		entry.KeyMR = h
		dbEntries = append(dbEntries, entry)
	}

	dBlock := NewDirectoryBlock(0, nil)
	err = dBlock.SetDBEntries(dbEntries)
	if err != nil {
		t.Error(err)
	}

	if dBlock.GetKeyMR().String() != "1710a017d0aaa29e03cdce767f2442a8519a512769777eb5c93d0167ad788104" {
		t.Error("Wrong DBlock KeyMR")
	}
}

/* func TestDBlockMisc(t *testing.T) {
	b, err := CreateDBlock(0, nil, 10)
	if err != nil {
		t.Error(err)
	}
	if b == nil {
		t.Error("CreateDBlock returned nil nil")
	}
	b2, err := CreateDBlock(1, b, 10)
	if err != nil {
		t.Error(err)
	}
	if b2 == nil {
		t.Error("CreateDBlock returned nil nil")
	}
	_, err = b2.BuildBodyMR()
	if err != nil {
		t.Error(err)
	}

}
*/
