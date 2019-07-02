// +build all

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package directoryBlock_test

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
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

func TestMarshalUnmarshalBadDirectoryBlock(t *testing.T) {
	dblock := createTestDirectoryBlock()
	p, err := dblock.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	// set a bad Block Count in the dblock header
	p[111] = 0xff

	dblock2 := new(DirectoryBlock)
	if err := dblock2.UnmarshalBinary(p); err == nil {
		t.Error("DirectoryBlock should have errored on unmarshal", dblock2)
	} else {
		t.Log(err)
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
		t.Errorf("Invalid number of BlockCount occurrences")
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

	dblock.AddEntry(primitives.NewHash(constants.ADMIN_CHAINID), primitives.NewZeroHash())
	dblock.AddEntry(primitives.NewHash(constants.EC_CHAINID), primitives.NewZeroHash())
	dblock.AddEntry(primitives.NewHash(constants.FACTOID_CHAINID), primitives.NewZeroHash())
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
	if !strings.Contains(j, `"chainid":"000000000000000000000000000000000000000000000000000000000000000d"`) {
		t.Error("Header does not contain ChainID")
	}
}

func TestBuildBlock(t *testing.T) {
	db1 := NewDirectoryBlock(nil)
	db1.(*DirectoryBlock).Init()
	//fmt.Println(db1)

	k, _ := primitives.HexToHash("7b2b988cd5308f76d2a44c564ade986213929b7fcfab6f2fc7694b595c71012e")

	t.Logf(db1.GetKeyMR().String())

	if !k.IsSameAs(db1.GetKeyMR()) { //expected an empty directoryblock
		t.Errorf("Invalid KeyMR - %v vs %v", k, db1.GetKeyMR())
	}

	db := NewDirectoryBlock(nil)

	if db.GetEntrySigHashes() != nil {
		t.Errorf("Invalid GetEntrySigHashes")
	}

	//h, _ := primitives.HexToHash("ce733587b898421bb3efab257ac8d6679b520df217ec949e41faf231121cb9b8")
	a := new(adminBlock.AdminBlock)
	a.Init()
	//fmt.Println(a.DatabasePrimaryIndex())
	db.SetABlockHash(a)

	//h, _ = primitives.HexToHash("f294cd012b3c088740aa90b1fa8feead006c5a35176f57dd0bc7aac19c88f409")
	e := new(entryCreditBlock.ECBlock)
	e.Init()
	db.SetECBlockHash(e)

	//h, _ = primitives.HexToHash("1ce2a6114650bc6695f6714526c5170e7f93def316a3ea21ab6e3fa75007b770")
	f := new(factoid.FBlock)
	//f.Init()
	db.SetFBlockHash(f)

	c, _ := primitives.HexToHash("3e3eb61fb20e71d8211882075d404f5929618a189d23aba8c892b22228aa0d71")
	h, _ := primitives.HexToHash("9daad42e5efedf3075fa2cf51908babdb568f431a3c13b9a496ffbfb7160ad2e")
	db.SetEntryHash(h, c, 3)

	c, _ = primitives.HexToHash("df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604")
	h, _ = primitives.HexToHash("b926da5ea5840b34189c37c55db9eb482f6e370bd097a16d6e890bc000c10898")
	db.SetEntryHash(h, c, 4)

	k, _ = primitives.HexToHash("eadf05b85c7ad70390c72783a9a3a29ae253f4f7d45d36f176bbc56d56bab9cc")

	if !k.IsSameAs(db.GetKeyMR()) {
		t.Errorf("Invalid KeyMR - %v vs %v", k, db.GetKeyMR())
	}

	es := db.GetEBlockDBEntries()

	//fmt.Println(es[1].GetChainID())

	if !c.IsSameAs(es[1].GetChainID()) {
		t.Errorf("Invalid ChainID - %v vs %v", c, es[1].GetChainID())
	}

	es2 := db.GetEntryHashes()
	//fmt.Println(es2)
	if !h.IsSameAs(es2[4]) {
		t.Errorf("Invalid Entry Hash - %v vs %v", h, es2[4])
	}

	es3 := db.GetEntryHashesForBranch()
	list := fmt.Sprintf("%v", es3)
	expectedList := "[000000000000000000000000000000000000000000000000000000000000000a 4fb409d5369fad6aa7768dc620f11cd219f9b885956b631ad050962ca934052e 000000000000000000000000000000000000000000000000000000000000000c a566023a9d7b824e4a12121ee38bc4d3c4987988f04eb8cfecc63570936d7c56 000000000000000000000000000000000000000000000000000000000000000f c9ab808e3d1d5eb2b7d3fa946dca27c2d250d782dab05a729fe99e9aaf656330 3e3eb61fb20e71d8211882075d404f5929618a189d23aba8c892b22228aa0d71 9daad42e5efedf3075fa2cf51908babdb568f431a3c13b9a496ffbfb7160ad2e df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604 b926da5ea5840b34189c37c55db9eb482f6e370bd097a16d6e890bc000c10898]"
	if list != expectedList {
		fmt.Printf("Invalid expectedList - %v vs %v", h, es2[4])
	}

	printout := db.String()
	/*
		              keymr: eadf05b85c7ad70390c72783a9a3a29ae253f4f7d45d36f176bbc56d56bab9cc
		             bodymr: 01004ae2e96c0344a3c30a0704383c5c90ca2663921a9c1b8dc50658d52850a3
		           fullhash: 857d121b40c0763cd310c68963d23ebf6fa4241ef6ba26861d9b80aa71c9f3a9
		  version:         0
		  networkid:       0
		  bodymr:          01004ae2e96c0344a3c30a0704383c5c90ca2663921a9c1b8dc50658d52850a3
		  prevkeymr:       0000000000000000000000000000000000000000000000000000000000000000
		  prevfullhash:    0000000000000000000000000000000000000000000000000000000000000000
		  timestamp:       0
		  timestamp str:   1969-12-31 18:00:00
		  dbheight:        0
		  blockcount:      5
		entries:
		    0 chainid: 000000000000000000000000000000000000000000000000000000000000000a
		      keymr:   4fb409d5369fad6aa7768dc620f11cd219f9b885956b631ad050962ca934052e
		    1 chainid: 000000000000000000000000000000000000000000000000000000000000000c
		      keymr:   a566023a9d7b824e4a12121ee38bc4d3c4987988f04eb8cfecc63570936d7c56
		    2 chainid: 000000000000000000000000000000000000000000000000000000000000000f
		      keymr:   c9ab808e3d1d5eb2b7d3fa946dca27c2d250d782dab05a729fe99e9aaf656330
		    3 chainid: 3e3eb61fb20e71d8211882075d404f5929618a189d23aba8c892b22228aa0d71
		      keymr:   9daad42e5efedf3075fa2cf51908babdb568f431a3c13b9a496ffbfb7160ad2e
		    4 chainid: df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604
		      keymr:   b926da5ea5840b34189c37c55db9eb482f6e370bd097a16d6e890bc000c10898

	*/
	expectedString1 := fmt.Sprintf(`              keymr: %s
             bodymr: 01004ae2e96c0344a3c30a0704383c5c90ca2663921a9c1b8dc50658d52850a3
           fullhash: 857d121b40c0763cd310c68963d23ebf6fa4241ef6ba26861d9b80aa71c9f3a9
  version:         0
  networkid:       0
  bodymr:          01004a
  prevkeymr:       000000
  prevfullhash:    000000
  timestamp:       0
  timestamp str:   `, k.String()) // Use KeyMR from above
	epoch := time.Unix(0, 0)
	expectedString2 := epoch.Format("2006-01-02 15:04:05")

	expectedString3 := `
  dbheight:        0
  blockcount:      5
entries:
    0 chainid: 000000000000000000000000000000000000000000000000000000000000000a
      keymr:   4fb409d5369fad6aa7768dc620f11cd219f9b885956b631ad050962ca934052e
    1 chainid: 000000000000000000000000000000000000000000000000000000000000000c
      keymr:   a566023a9d7b824e4a12121ee38bc4d3c4987988f04eb8cfecc63570936d7c56
    2 chainid: 000000000000000000000000000000000000000000000000000000000000000f
      keymr:   c9ab808e3d1d5eb2b7d3fa946dca27c2d250d782dab05a729fe99e9aaf656330
    3 chainid: 3e3eb61fb20e71d8211882075d404f5929618a189d23aba8c892b22228aa0d71
      keymr:   9daad42e5efedf3075fa2cf51908babdb568f431a3c13b9a496ffbfb7160ad2e
    4 chainid: df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604
      keymr:   b926da5ea5840b34189c37c55db9eb482f6e370bd097a16d6e890bc000c10898
`
	expectedString := expectedString1 + expectedString2 + expectedString3
	if printout != expectedString {
		t.Errorf("Invalid printout:\n%v\n%v", printout, expectedString)
		limit := len(expectedString)
		if limit < len(printout) {
			limit = len(printout)
		}
		printout = strings.Replace(printout, "\n", " ", -1)
		expectedString = strings.Replace(expectedString, "\n", " ", -1)

		for i := 0; i < limit; i++ {
			if i > len(printout) {
				t.Errorf("Ran out of output at %d", i)
				break
			}
			if i > len(expectedString) {
				t.Errorf("Ran out of expectedString at %d", i)
				break
			}
			if expectedString[i] != printout[i] {
				t.Errorf("Mismatch at %d ", i)
				t.Errorf(fmt.Sprintf("[%%s]\n[%%s]\n %%%ds^\n", i), expectedString, printout, "")
				break
			}
		}
	}

	m := db.GetDatabaseHeight()
	if m != 0 {
		t.Fail()
	}

	n := db.GetChainID()
	cid, _ := primitives.HexToHash("000000000000000000000000000000000000000000000000000000000000000d")
	if !n.IsSameAs(cid) {
		t.Errorf("Invalid cid - %v vs %v", n, cid)
	}

	o := db.DatabasePrimaryIndex()
	r, _ := primitives.HexToHash("eadf05b85c7ad70390c72783a9a3a29ae253f4f7d45d36f176bbc56d56bab9cc")
	if !o.IsSameAs(r) {
		t.Errorf("Invalid DatabasePrimaryIndex - %v vs %v", o, r)
	}

	p := db.DatabaseSecondaryIndex()
	s, _ := primitives.HexToHash("857d121b40c0763cd310c68963d23ebf6fa4241ef6ba26861d9b80aa71c9f3a9")
	if !p.IsSameAs(s) {
		t.Errorf("Invalid DatabaseSecondaryIndex - %v vs %v", p, s)
		fmt.Println(p)
		fmt.Println(s)
		t.Fail()
	}

	returnVal, _ := db.JSONString()
	//fmt.Println(returnVal)

	expectedString = `{"dbhash":"857d121b40c0763cd310c68963d23ebf6fa4241ef6ba26861d9b80aa71c9f3a9","keymr":"eadf05b85c7ad70390c72783a9a3a29ae253f4f7d45d36f176bbc56d56bab9cc","headerhash":null,"header":{"version":0,"networkid":0,"bodymr":"01004ae2e96c0344a3c30a0704383c5c90ca2663921a9c1b8dc50658d52850a3","prevkeymr":"0000000000000000000000000000000000000000000000000000000000000000","prevfullhash":"0000000000000000000000000000000000000000000000000000000000000000","timestamp":0,"dbheight":0,"blockcount":5,"chainid":"000000000000000000000000000000000000000000000000000000000000000d"},"dbentries":[{"chainid":"000000000000000000000000000000000000000000000000000000000000000a","keymr":"4fb409d5369fad6aa7768dc620f11cd219f9b885956b631ad050962ca934052e"},{"chainid":"000000000000000000000000000000000000000000000000000000000000000c","keymr":"a566023a9d7b824e4a12121ee38bc4d3c4987988f04eb8cfecc63570936d7c56"},{"chainid":"000000000000000000000000000000000000000000000000000000000000000f","keymr":"c9ab808e3d1d5eb2b7d3fa946dca27c2d250d782dab05a729fe99e9aaf656330"},{"chainid":"3e3eb61fb20e71d8211882075d404f5929618a189d23aba8c892b22228aa0d71","keymr":"9daad42e5efedf3075fa2cf51908babdb568f431a3c13b9a496ffbfb7160ad2e"},{"chainid":"df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604","keymr":"b926da5ea5840b34189c37c55db9eb482f6e370bd097a16d6e890bc000c10898"}]}`
	if returnVal != expectedString {
		t.Errorf("Invalid returnVal:\n%v\n%v", returnVal, expectedString)
	}

	//fmt.Println(q)

	returnByte, _ := db.JSONByte()
	by := string(returnByte)
	if by != expectedString {
		t.Errorf("Invalid returnByte:\n%v\n%v", by, expectedString)
	}

	if nil == CheckBlockPairIntegrity(nil, nil) {
		t.Errorf("CheckBlockPairIntegrity(nil, nil) failed")
	}

	if nil != CheckBlockPairIntegrity(db, nil) {
		t.Errorf("CheckBlockPairIntegrity(db, nil) failed")
	}

	db2 := NewDirectoryBlock(db1)
	j, _ := primitives.HexToHash("df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604")
	i, _ := primitives.HexToHash("b926da5ea5840b34189c37c55db9eb482f6e370bd097a16d6e890bc000c10898")
	db2.SetEntryHash(j, i, 3)

	l, _ := primitives.HexToHash("3e3eb61fb20e71d8211882075d404f5929618a189d23aba8c892b22228aa0d71")
	q, _ := primitives.HexToHash("9daad42e5efedf3075fa2cf51908babdb568f431a3c13b9a496ffbfb7160ad2e")
	db2.SetEntryHash(l, q, 4)

	_, err := db2.MarshalBinary()

	if err != nil {
		t.Errorf("%v", err)
	}

	if nil != CheckBlockPairIntegrity(db2, db1) {
		t.Errorf("CheckBlockPairIntegrity(db2, db1) failed")
	}
}

func TestSortFunc(t *testing.T) {
	db1 := NewDirectoryBlock(nil)
	db1.(*DirectoryBlock).Init()

	ecb_kmr, _ := primitives.HexToHash("0000000000000000000000000000000000000000000000000000000000000050")
	cid, _ := primitives.HexToHash("0000000000000000000000000000000000000000000000000000000000000150")
	db1.SetEntryHash(ecb_kmr, cid, 3)

	ecb_kmr, _ = primitives.HexToHash("0000000000000000000000000000000000000000000000000000000000000040")
	cid, _ = primitives.HexToHash("0000000000000000000000000000000000000000000000000000000000000140")
	db1.SetEntryHash(ecb_kmr, cid, 4)

	ecb_kmr, _ = primitives.HexToHash("0000000000000000000000000000000000000000000000000000000000000030")
	cid, _ = primitives.HexToHash("0000000000000000000000000000000000000000000000000000000000000130")
	db1.SetEntryHash(ecb_kmr, cid, 5)

	//should be out of order at this point
	//fmt.Println(db1.GetEntryHashes())

	//marshal will call the sort function, which does not have an interface
	db1.MarshalBinary()
	//fmt.Println(db1.GetEntryHashes())
	expected := "[0000000000000000000000000000000000000000000000000000000000000000 0000000000000000000000000000000000000000000000000000000000000000 0000000000000000000000000000000000000000000000000000000000000000 0000000000000000000000000000000000000000000000000000000000000030 0000000000000000000000000000000000000000000000000000000000000040 0000000000000000000000000000000000000000000000000000000000000050]"

	sorted := fmt.Sprint(db1.GetEntryHashes())
	if sorted != expected {
		t.Errorf("Sort function failed")
	}
}

func TestMerkleTree(t *testing.T) {
	db1 := NewDirectoryBlock(nil)
	db1.(*DirectoryBlock).Init()

	ecb_kmr, _ := primitives.HexToHash("0000000000000000000000000000000000000000000000000000000000000010")
	cid, _ := primitives.HexToHash("0000000000000000000000000000000000000000000000000000000000000100")
	db1.SetEntryHash(ecb_kmr, cid, 3)

	ecb_kmr, _ = primitives.HexToHash("0000000000000000000000000000000000000000000000000000000000000020")
	cid, _ = primitives.HexToHash("0000000000000000000000000000000000000000000000000000000000000200")
	db1.SetEntryHash(ecb_kmr, cid, 4)

	ecb_kmr, _ = primitives.HexToHash("0000000000000000000000000000000000000000000000000000000000000030")
	cid, _ = primitives.HexToHash("0000000000000000000000000000000000000000000000000000000000000300")
	db1.SetEntryHash(ecb_kmr, cid, 5)

	//should be out of order at this point
	//fmt.Println(db1.GetEntryHashes())
	mr, err := db1.BuildBodyMR()
	if err != nil {
		t.Errorf("tree building function failed %v", err)
	}

	//fmt.Println(mr)

	expected := "a8caad450c7f721526087bf33d251d5fd537e51885f718e89dd9780cfeee3e66"

	value := fmt.Sprint(mr)
	if value != expected {
		t.Errorf("Sort function failed")
	}
}

func TestSameAs(t *testing.T) {
	//block 1000
	db1kbytes, _ := hex.DecodeString("00fa92e5a2f2eaf170a2da9e4956a40231ed7255c6c6e5ada1ed746fc5ab3a0b79b8c700367a49467be900ba00daedd7d9cf2b1a07f839360e859e1f3d78c46701d3ad1507974595bf9b73dbec9ff5d5744cbf6410d66b837924208a0b8b84e54fc4aad660016ea716000003e800000004000000000000000000000000000000000000000000000000000000000000000a3d92dc70f4cfd4fe464e18962057d71924679cc866fe37f4b8023d292d9f34ce000000000000000000000000000000000000000000000000000000000000000c0526c1fdb9e0813e297a331891815ed893cb5a9cff15529197f13932ed9f9547000000000000000000000000000000000000000000000000000000000000000f526aca5f63bfb59188bae1fc367411a123bcc1d5a3c23c710b66b46703542855df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604f08c42bc44c09ac26c349bef8ee80d2ffb018cfa3e769107b2413792fa9bd642")
	db1k, err := UnmarshalDBlock(db1kbytes)
	if err != nil {
		t.Errorf("db unmarshall failed")
	}

	db1k1bytes, _ := hex.DecodeString("00fa92e5a21b06a03dcc94ab7d6d991bfc9937cf72c744095f46316ef2281f9f20d3738c03cd45e38f53c090a03513f0c67afb93c774a064a5614a772cd079f31b3db4d01106e8d2d429fe728c4a90a3b6fbd910eb97e543c460c762a72d1563302bb401b1016ea720000003e900000004000000000000000000000000000000000000000000000000000000000000000abfd0f8ab0ea25a16d68c0533401eaea53cb84b9b597f2ad85006120836ac1788000000000000000000000000000000000000000000000000000000000000000c9495cc2988fb7d598cec0c91ea00447e9f2e6b0f5cd3b808e50eb7010cae4202000000000000000000000000000000000000000000000000000000000000000f1f67dc478fc7a5f8d9d612e362040a114dd437228c64b11a10df3a3e31d01c45df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604cbf7179a054e6a40dbbebdb4ac29e5185052889907c8607f35a3aca84eeb72f6")
	db1k1, err := UnmarshalDBlock(db1k1bytes)
	if err != nil {
		t.Errorf("db unmarshall failed")
	}

	if db1k.IsSameAs(db1k1) {
		t.Errorf("directory block same as comparison failed")
	}

	db1kclone, err := UnmarshalDBlock(db1kbytes)
	if err != nil {
		t.Errorf("db unmarshall failed")
	}

	if !db1k.IsSameAs(db1kclone) {
		t.Errorf("directory block same as comparison failed")
	}
	//fmt.Println(db1k1)
}

type TestDBlock struct {
	Raw   string
	KeyMR string
	Hash  string
}

func TestMarshalUnmarshalFixedDirectoryBlocks(t *testing.T) {
	ts := []TestDBlock{}

	t1 := TestDBlock{}
	t1.Raw = "00fa92e5a24d0789d16890ec1f96f617d8c802a40ee876d761b076da330d784356ac80f9ab00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000016e80100000000000000003000000000000000000000000000000000000000000000000000000000000000a4fb409d5369fad6aa7768dc620f11cd219f9b885956b631ad050962ca934052e000000000000000000000000000000000000000000000000000000000000000cf87cfc073df0e82cdc2ed0bb992d7ea956fd32b435b099fc35f4b0696948507a000000000000000000000000000000000000000000000000000000000000000fa164ccbb77a21904edc4f2bb753aa60635fb2b60279c06ae01aa211f37541736"
	t1.KeyMR = "64d4352b134280305599363ea388c2a9c3c64dc3ee6e0100893262e372bf064b"
	t1.Hash = "cbd3d09db6defdc25dfc7d57f3479b339a077183cd67022e6d1ef6c041522b40"
	ts = append(ts, t1)

	t2 := TestDBlock{}
	t2.Raw = "00fa92e5a2c533096d61aad4d07e85546c813a14ff6aeccf09af5af888e22c66a4bd8d8e682079bf4c6ae92c398eb16cf3b10ba05d0ac0691c07e5e023d527dc09c562b9c48fb437fae06ef1b38d971eb0edf53e7b9e4d5737aef4aafa6d6e18b33723cd9d017c4ec300015e5e0000003d000000000000000000000000000000000000000000000000000000000000000ad93677fa2660ee262b0ac4f941e71c2925c3193de5bb6ef5480c790c3cc3f1fc000000000000000000000000000000000000000000000000000000000000000c76bd8907808e6d94e1c37c2338a02bcbc21dd87dac1c1d75bdbc524663e082eb000000000000000000000000000000000000000000000000000000000000000f3689c1bdc32b75c43091944d789149bcbf8fe75eff741a8abb5827e269a1d09f0464bf13a66ed62d8196c51292caedaecbd8dfe245acdbd1aafdac9ed9d77b1bad5aa08b6c601a2f116374818fb99ee82edd0b9caecb443cb97bebc6c970e07204e57c0c2fee5834c885c16b88301ffaa4596b40fca01e91a1f28e4492ce09e8148a8e45ad34a64225c4da7d7b5bf1b126cfc6a9c8f0c1114265c987e3ea4d0106a40590f536293bdecc3d7e69a5c21785c6ed454a59caf7b2e083a1a88ac85b497ac429de804506d7828d173898ed061c2d6153a0eb001cd226756abc3563310caff62ea5b5aa015c706add7b2463a5be07e1f0537617f553558090f23c7f567ef00d23f177b34bd04dd80fb8226581bb24d2e8d1ea938d320c89a1ac8dd0c90dc47da47563719e91de700d54f7f0423dd142d2f339a88c238509b30ab2d353be86b8fd7f451ca0aa6394151373624d4d3050e4f3ad1f441e6ced1c528c87d11892156d68608e521d78fe86fbe93a7dde64d54bbef2afe149300cec78b002098a344109522bca1778c560f8b2d2738a6a2ee0dae117a21f0bee034466b284e318d3a59606dfc89aa3bcc0b696b8d55eab24391f03bba66fcba51f71996708540ead932ed9f340a72ed6983245d8bffd301f0a553b6411d2adcb8fb4a21a29851c194b75dbbe826190882c38f2d24016176bec7e5b7b5813ebfca3a7372c8a120c9ad696508795775890a9184db8315f02b20177d44156ad9c6f18f1b34fe0791ebaeba6aecd32368c2cc49a906f9fd20bab73f3a1d768128dcfdcf9d773ca713e9a9b774bffb6c443316be20de7119a757350c31b7efc84b91479b264d31e5529f8148d9bcf88ec268fed34458af7eb0a2ec819e2375824d785b731bc51aafc3328f3a8324a108b090895c944d357c41fe9405e178401516a12f09bc3af84232ee789afd7bd2f82c57784f556b5cc07c371c0361057123c68e3b2d496b26778cc6751730b16ac4fb334ff434f74a5dd5826ca5583a27193ecb44426d12d011d37b8c9eda0d8d4e2c25a100c9443139be8e1c07312da2d5a852ce00d284bcc4c8aea4b91e0dd662d7e2c0d37beed49cda7455c096a67afc8bb3ba9e6426da4c537d36e0fa75e74ecb59180a0e34d2b5e044f1e011fe01f3ce3e863cacbce686d47231b1ad2906c1ac2c1f54d8b88bef37510aec086b89fc1053b418c258217f2383d772059bbea739e28a0a9e8a36333de663a3654fd8e42c6cffab37d60cfde0f64df97db109b08b21a216f7a7c1a43506bdb3c80bd1fa39e7cab059d772abf406af18cdaec4b6d22ec962e534140922c450fec2924497fc9e487eb25f90e0f5bcc8213be348ef03ccaca8cd52be9f867d53fe8e0909d46a547050ed7b64a494b4f0806446070f627f850925521b901edc613ba57e5ee734785e22c6a8054678581a68d01c526892034ae2261fd4313dd2e6969ea9dea908be42fef8c7a3e9c4bf71c177e71504032ab84023d8afc16e302de970e6be110dac20adbf9a197460cdcc0e3e7ea1560a29717d9c9148ec772677bdcd9ba76dd6e1b41f65315af915c7a44a37870ca729d03820339379955781a863bc461545a9df06bbc15110bdb4c328f61fddd0d5964ebdfb3fde5c410a7b0e3cb675a03eedbc614cc5d5cf92a5db403e0d07e9c00a1719443b1327a4adb936ec0f98d9f650a0d897fbc260a2fefbbc51ed7826a6f350ecd511e87cafbf8e731ae4327aea8f79e8b993a220c6b6127cf66f2ab347e7d4fd6f810e8e6d5ecf17189612feb9d00b097c86b99c4bb163e99402ddc3ff2e669293079a60662fbfc26e76df7adab6ff910c4dab0452e62e05a098cd11191cd4364aede266eba0157d0b193f35caa1a71ea1443a1417f947803eabf0aab269fca3dd07b6b338ba0733ea3aa93dc39e137f25e31310b3264917e7307c789d5150f8276f82ce85c569e7ba3b1987a06caebc1cb6fd75d6d125c739aed4845a8251c889b5353f1a3b1845cf03e324a82794f96f1bc61df156909765ff072c322c56a7c4bfa8911ee4fdefacca711d30a9ad2a8672a3cc9592f75a0c19f790502c6e7f9bd6fc2447d3c238410f2d27dd0f4bf1e815958dc5b6cc25373ad9850b7765accb29ab175aafb006b9b973cfafdada223c508d2bf36fb97f22e97e17a584273b5efbea43038df0427d014bae6356680847d092f1e1072c4442720c29ad2dd384bed3ca6a13a687f492f23503aaf31e5c130aab3b33d1e85f9b1cc192393fabe48c48201744c59fd211c51e5bf39788ad69569142c7b72e94d133e2e1fece7c15ba73db951742262ffb91e05e79b9f62e7a469a367c5948e8de658579f60914c522b312749b70594953cc185e18606b356a40d58563076c6df20cb25ba336d4117cc2a2a816b782f6e6f6d24d65391a27d0815f78316437e6eb98560b13e0123f001084b5e2523748c06f7e97a2369367ccf8f3369ad78f634366b6a0fbd54654d94161c83e5bc802f5b54a91684ca176ea4eafef9479d725dabc551ed597fd647859cdb4eaac2e135d96b8e633d1884985f2f6fdce07c7a6dc985ccffa1455e15cc6194ff1ec0f48bd388ba9db56c12013d3c7a3f95c946df30f1eeb8f2f25ee3ac6949130ffd40197e647b54bd9ba994bb10954a0485cc94c58b7e018d2a93513d88dbb50ec60a76086120d5a639cf38d7327e36490975be5237975dc2c00482d862a157b66e4fc41995b75f167c094e8491faa534861adf5be5c9dd73ac476517d1f36265e4db9e9c3b57a96c7e5343c87a6cabbb9167fefd5cd7b288f5f5f7eaa02433d1a802ddb6420ab2713fb2d88dbec9858b86da74d5ef07852de0ead976be62b61acee3f3e862322dafd8f18e2319769e3aa02704fa5ad7cfd3d82468ae59ae6defa9df256f26ff24400457b556352b4d7e8871ec49d09429680da37ec203aa4124460975a8c6fd1926c3688e0ff6d35a9ee3b4bdcd8a0c1037f3ccf7c96c299e1836a40e992f620d46158282909ab5ab278945fc6e36bb6f674aa8bb6087bd7d4aa901eb49394f2e4ee7a68d68d11435dc732a5350f831b7e9c13228251a777f42bd31c24c72baf79c5d27834f830f44a895a66903790c2cb44df6123f5ba519b872bee3a407514966037324564ea84fda07744e00f6ff4576c48069ba99d18887ebc84044561da5d43d6393bc1ba48ade96d8be48199641129b9704479ae28879d31a16b914324258c28e034af0ae2a0c7d4095b9a3924c28530c347b7a0a3d72e05ab16db15b47ec248e8e5459279546992ef716de08dab0857ff9d2a15d39fb68fc255f7cda8f5b685fe159d9edfc41a793a3caec4118a109e3145011dcde193c99cb9cf69a5a96955f59732e0fcee8af0e03d899527e8e1ee8768b7af6f45a2beea3982f2dd8e455377832ee1191bf065da226aa91f2965dd8a38f627fb252e91a3b2da3491db8228b8760500ea085b43ebeb6444686be26b2dbe30b11d233fcb67455a024f10ac099a28524ca592b10278e2d895801970e8f5f05958ae725d3a05a96f0b93f09737d18197202f2a3c3142d1579ccc63387ee6db7f9e28fd4046f147c8e92d1ac5891a420b41e896ed8da85522fb32aa4794dfc0143f4d47765f618a79af126601970a83deb185717c5edc2ce6a00e779022c2e743e6100a395caed3fe3ce8d0ffe86607f38d0360b5ff7438626ac3f1a6a0e5f01611e8dadbed30c5fecaac2ab612630a0e2fa7076cb8b848558fcab528c9cabc26bd415c44dcc89c1d4ac67ef6ce73cca44d34e31fb0d70cc0ae8320e3ef7f49069312904b216b607dad9faad9a1d85244110678fccb10c7d8b40cc2e3b4308cf882826cecab582286c9c0abcd15eaa6d2bfd0403fee4cd017cc94c04ea13283e145fe07bf25a7c86585f3c5fccf87cfff435af25d06f794925d778193b11c1af6666ab2792f2d948cc78fba41d2b70e458953b970023b589642316412a97563e7c6f7029d6f5f9b1054b9129b9b4f2dbf6e7d198d279dd7e92c067bffd51dc1d2dadc5ec04f2dbe105abefae97d1af09d78cc1b77b4e967c07e481d63127db71cd89e5def3bd05477aa70c80209aa0591786c8b1d3abab36f0abe172b08df64396e6e4b4129bcaf7b0b3e1b94653414c68249386a8403b4270acd0c3ad7cc9e582a25a67f3ed3b645db8591aefc8b1eca3570271d6183f63fd0143c68afc789472c439b71a2fec562c1b47adcf936ad5b0b07712f1145101f3ce546f5a7f5abe0395c761ac179283288a9c52ad5e57d1c0656129decdd5bcabded92f52967a0e11f173f84c48b3aa00b03f4a556f8594b4d08405b87d20e62ca1b3a95570376826c82b33c25fb823f9f373d0a31aa5449ca50077e13b1ab341b4aa3394082aff1bb0589e4f0ed6e4fefea882c57ff1fbe36d1d2bf9c2f72e0d24f47105cab653d7c54b48e4f756a0ca7d8a9449b8953742a07befe47ee66842d799f6cdb826e3b8d90fce73832186477d693d9851a2ac85e76072b106ed1a5fee5c6ca7182048da677b8d14b3f0e98737f5f6ff859b4e9bff995ee4d0ae91c17f5d5ece46dca6c497c8ba2d5ff0481bf39fa62b59d94b3a2f526337a8cd0af88cd632f630164c36a72047a0f6ca960c1def8f9d1a976c9c4a8f0ae4dd765ea2ec33b361d2a1426484a5ab98a85ef4c3f83aa5d7b9b5b914b62810775e29bcdeeaeca82e343f82996b3b1ebd8ebc2f86bb3644b8916c941ea7605de79a515024b3a5eea39976aa13b9d987c1c23cc9ae399fccaf6f96f64a27a08f73516ee8ef8432032b4f533b00f80777f65e1baed8e3c8f69b872fd96f555f8ee821eaeb87972beec8783beca6762559a8a73591b593172eebcb3a5698f6acae5c5e60f7290c61b837e9f9f93b8fc2c7ad0d6ef37fc2ad797662743e2b08cbc7f52b26aa85d2e086e0455e80742a3b7555a1ad19a6f2b766f97197ac500763f50cee70113b7a64c50835be6815a8839d1265e3831aa5e3b5b6031bcb24370975f6df2b40bf16be03deddc169a6429804a67a761ed5f2d5499f7e56a7202f854b472c8d8efbd044db54b07705df1cc9ea2c192ac8132291606d88689553cec7d0f9164cd66af9d5773b4523a510b5eefb9a5e626480feeb6671ef2d17510ca30085d2af7b419b358f288065b7c6b16c646d071125c0690b68ace2106d56e3d042fbf1bb7ffa4ec0bbb0f7dc18cbeb47514102c2eb38fd1f985be3254156b286770dcdcc7c3f73191a6b68a62bf5c74ec1c9b13ebc430e6b7de3f0400bd448694a"
	t2.KeyMR = "bdc4d0def175d1373c4932c056f930d43ac037057da1bcf13972da31bfc669ff"
	t2.Hash = "b26795a9b218fce9aec67ad453719e8b09fc850b53db398e1db208dd0494f566"
	ts = append(ts, t2)

	for _, tBlock := range ts {
		rawStr := tBlock.Raw
		raw, err := hex.DecodeString(rawStr)
		if err != nil {
			t.Errorf("%v", err)
		}

		f := new(DirectoryBlock)
		rest, err := f.UnmarshalBinaryData(raw)
		if err != nil {
			t.Errorf("%v", err)
		}
		if len(rest) > 0 {
			t.Errorf("Returned too much data - %x", rest)
		}

		b, err := f.MarshalBinary()
		if err != nil {
			t.Errorf("%v", err)
		}

		if primitives.AreBytesEqual(raw, b) == false {
			t.Errorf("Marshalled bytes are not equal - %x vs %x", raw, b)
		}

		//f.CalculateHashes()

		if f.DatabasePrimaryIndex().String() != tBlock.KeyMR {
			t.Errorf("Wrong PrimaryIndex - %v vs %v", f.DatabasePrimaryIndex().String(), tBlock.KeyMR)
		}
		if f.DatabaseSecondaryIndex().String() != tBlock.Hash {
			t.Errorf("Wrong SecondaryIndex - %v vs %v", f.DatabaseSecondaryIndex().String(), tBlock.Hash)
		}
	}
}
