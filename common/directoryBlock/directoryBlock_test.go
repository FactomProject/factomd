// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package directoryBlock_test

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

func TestUnmarshalNilDirectoryBlock(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(DirectoryBlock)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestMarshalUnmarshalDirectoryBlockHeader(t *testing.T) {
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
		var buf primitives.Buffer
		binary.Write(&buf, binary.BigEndian, block.GetHeader().GetBlockCount())
		hex := buf.DeepCopyBytes()

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
	header.SetNetworkID(constants.MAIN_NETWORK_ID)
	header.SetPrevFullHash(primitives.NewZeroHash())
	header.SetPrevKeyMR(primitives.NewZeroHash())
	header.SetTimestamp(primitives.NewTimestampFromSeconds(1234))
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

	dBlock := NewDirectoryBlock(nil)
	err = dBlock.SetDBEntries(dbEntries)
	if err != nil {
		t.Error(err)
	}

	if dBlock.GetKeyMR().String() != "1710a017d0aaa29e03cdce767f2442a8519a512769777eb5c93d0167ad788104" {
		t.Error("Wrong DBlock KeyMR")
	}
}

func TestDBlockTimestamp(t *testing.T) {
	dbStr := "010000ffff45acc1e2847302b80d0558aac1504c54253c28293a92bab6c7f8bb984a1e696fcd63b26d12e9d397a545fd50e26b53ab8b1fb555f824edb1f71937a6288d59014d1b7854253ec712124c9862f3aece068fe8b56b33e540dd6e8f7bb30efdb4f7000004da0000000800000005000000000000000000000000000000000000000000000000000000000000000a44a3b5f89f8f861815930b8442ed143d61163a8d5ad4cc3f792847c6c26e3543000000000000000000000000000000000000000000000000000000000000000c9d149c5213f91502ad50d9136792974987ad086309bf4d1462c68fe982284245000000000000000000000000000000000000000000000000000000000000000f1a708e863af21b5492563f6440cabfd2932653864f77cf4519cf361b107e4ce86e7e64ac45ff57edbf8537a0c99fba2e9ee351ef3d3f4abd93af9f01107e592c25c9e5963917c97ed988c571e703104b34d11f2f6241c0c69d9cfd6ad94491dbdf3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604027710061c785d0ffbf15f2fe4a42a744f78ef6a0ca39bcf38ed4ead6ab0cded"
	dbHex, err := hex.DecodeString(dbStr)
	if err != nil {
		t.Errorf("%v", err)
	}
	dBlock, err := UnmarshalDBlock(dbHex)
	if err != nil {
		t.Errorf("%v", err)
	}

	timestamp := dBlock.GetTimestamp()

	seconds := timestamp.GetTimeMilli()
	if seconds != 1242*60*1000 {
		t.Errorf("Invalid timestamp - %v vs %v", seconds, 1242*60*1000)
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

func TestExpandedDBlockHeader(t *testing.T) {
	block := createTestDirectoryBlock()
	j, err := block.JSONString()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !strings.Contains(j, `"ChainID":"000000000000000000000000000000000000000000000000000000000000000d"`) {
		t.Error("Header does not contain ChainID")
	}
}
