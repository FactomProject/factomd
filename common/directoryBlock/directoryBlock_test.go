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

func TestBuildBlock(t *testing.T) {
	db1 := new(DirectoryBlock)
	db1.Init()
	//fmt.Println(db1)

	k, _ := primitives.HexToHash("e118d53659a92c69ad37602827bfaf256428867c6827f10829d678d7f8ddab33")

	if !k.IsSameAs(db1.GetKeyMR()) { //expected an empty directoryblock
		fmt.Println(k)
		fmt.Println(db1.GetKeyMR())
		t.Fail()
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
		fmt.Println(k)
		fmt.Println(db.GetKeyMR())
		t.Fail()
	}

	es := db.GetEBlockDBEntries()

	//fmt.Println(es[1].GetChainID())

	if !c.IsSameAs(es[1].GetChainID()) {
		fmt.Println(c)
		fmt.Println(es[1].GetChainID())
		t.Fail()
	}

	es2 := db.GetEntryHashes()
	//fmt.Println(es2)
	if !h.IsSameAs(es2[4]) {
		fmt.Println(h)
		fmt.Println(es2[4])
		t.Fail()
	}

	es3 := db.GetEntryHashesForBranch()
	list := fmt.Sprintf("%v", es3)
	expectedList := "[000000000000000000000000000000000000000000000000000000000000000a 4fb409d5369fad6aa7768dc620f11cd219f9b885956b631ad050962ca934052e 000000000000000000000000000000000000000000000000000000000000000c a566023a9d7b824e4a12121ee38bc4d3c4987988f04eb8cfecc63570936d7c56 000000000000000000000000000000000000000000000000000000000000000f c9ab808e3d1d5eb2b7d3fa946dca27c2d250d782dab05a729fe99e9aaf656330 3e3eb61fb20e71d8211882075d404f5929618a189d23aba8c892b22228aa0d71 9daad42e5efedf3075fa2cf51908babdb568f431a3c13b9a496ffbfb7160ad2e df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604 b926da5ea5840b34189c37c55db9eb482f6e370bd097a16d6e890bc000c10898]"
	if list != expectedList {
		fmt.Println(h)
		fmt.Println(es2[4])
		t.Fail()
	}

	printout := db.String()
	expectedString1 := fmt.Sprintf(`              KeyMR: %s
             BodyMR: 01004ae2e96c0344a3c30a0704383c5c90ca2663921a9c1b8dc50658d52850a3
           FullHash: 857d121b40c0763cd310c68963d23ebf6fa4241ef6ba26861d9b80aa71c9f3a9
  Version:         0
  NetworkID:       0
  BodyMR:          01004ae2e96c0344a3c30a0704383c5c90ca2663921a9c1b8dc50658d52850a3
  PrevKeyMR:       0000000000000000000000000000000000000000000000000000000000000000
  PrevFullHash:    0000000000000000000000000000000000000000000000000000000000000000
  Timestamp:       0
  Timestamp Str:   `, k.String()) // Use KeyMR from above
	epoch := time.Unix(0, 0)
	expectedString2 := epoch.Format("2006-01-02 15:04:05")

	expectedString3 := `
  DBHeight:        0
  BlockCount:      5
Entries: 
    0 ChainID: 000000000000000000000000000000000000000000000000000000000000000a
      KeyMR:   4fb409d5369fad6aa7768dc620f11cd219f9b885956b631ad050962ca934052e
    1 ChainID: 000000000000000000000000000000000000000000000000000000000000000c
      KeyMR:   a566023a9d7b824e4a12121ee38bc4d3c4987988f04eb8cfecc63570936d7c56
    2 ChainID: 000000000000000000000000000000000000000000000000000000000000000f
      KeyMR:   c9ab808e3d1d5eb2b7d3fa946dca27c2d250d782dab05a729fe99e9aaf656330
    3 ChainID: 3e3eb61fb20e71d8211882075d404f5929618a189d23aba8c892b22228aa0d71
      KeyMR:   9daad42e5efedf3075fa2cf51908babdb568f431a3c13b9a496ffbfb7160ad2e
    4 ChainID: df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604
      KeyMR:   b926da5ea5840b34189c37c55db9eb482f6e370bd097a16d6e890bc000c10898
`
	expectedString := expectedString1 + expectedString2 + expectedString3
	if printout != expectedString {
		fmt.Println(printout)
		fmt.Println(expectedString)
		t.Fail()
	}

	m := db.GetDatabaseHeight()
	if m != 0 {
		t.Fail()
	}

	n := db.GetChainID()
	cid, _ := primitives.HexToHash("000000000000000000000000000000000000000000000000000000000000000d")
	if !n.IsSameAs(cid) {
		fmt.Println(n)
		fmt.Println(cid)
		t.Fail()
	}

	o := db.DatabasePrimaryIndex()
	r, _ := primitives.HexToHash("eadf05b85c7ad70390c72783a9a3a29ae253f4f7d45d36f176bbc56d56bab9cc")
	if !o.IsSameAs(r) {
		fmt.Println(o)
		fmt.Println(r)
		t.Fail()
	}

	p := db.DatabaseSecondaryIndex()
	s, _ := primitives.HexToHash("857d121b40c0763cd310c68963d23ebf6fa4241ef6ba26861d9b80aa71c9f3a9")
	if !p.IsSameAs(s) {
		fmt.Println(p)
		fmt.Println(s)
		t.Fail()
	}

	returnVal, _ := db.JSONString()
	//fmt.Println(returnVal)

	expectedString = `{"DBHash":"857d121b40c0763cd310c68963d23ebf6fa4241ef6ba26861d9b80aa71c9f3a9","KeyMR":"eadf05b85c7ad70390c72783a9a3a29ae253f4f7d45d36f176bbc56d56bab9cc","Header":{"Version":0,"NetworkID":0,"BodyMR":"01004ae2e96c0344a3c30a0704383c5c90ca2663921a9c1b8dc50658d52850a3","PrevKeyMR":"0000000000000000000000000000000000000000000000000000000000000000","PrevFullHash":"0000000000000000000000000000000000000000000000000000000000000000","Timestamp":0,"DBHeight":0,"BlockCount":5,"ChainID":"000000000000000000000000000000000000000000000000000000000000000d"},"DBEntries":[{"ChainID":"000000000000000000000000000000000000000000000000000000000000000a","KeyMR":"4fb409d5369fad6aa7768dc620f11cd219f9b885956b631ad050962ca934052e"},{"ChainID":"000000000000000000000000000000000000000000000000000000000000000c","KeyMR":"a566023a9d7b824e4a12121ee38bc4d3c4987988f04eb8cfecc63570936d7c56"},{"ChainID":"000000000000000000000000000000000000000000000000000000000000000f","KeyMR":"c9ab808e3d1d5eb2b7d3fa946dca27c2d250d782dab05a729fe99e9aaf656330"},{"ChainID":"3e3eb61fb20e71d8211882075d404f5929618a189d23aba8c892b22228aa0d71","KeyMR":"9daad42e5efedf3075fa2cf51908babdb568f431a3c13b9a496ffbfb7160ad2e"},{"ChainID":"df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604","KeyMR":"b926da5ea5840b34189c37c55db9eb482f6e370bd097a16d6e890bc000c10898"}]}`
	if returnVal != expectedString {
		fmt.Println("got", returnVal)
		fmt.Println("expected", expectedString)
		t.Fail()
	}

	//fmt.Println(q)

	returnByte, _ := db.JSONByte()
	by := string(returnByte)
	if by != expectedString {
		fmt.Println("got", by)
		fmt.Println("expected", expectedString)
		t.Fail()
	}

	if nil == CheckBlockPairIntegrity(nil, nil) {
		t.Fail()
	}

	if nil != CheckBlockPairIntegrity(db, nil) {
		t.Fail()
	}

	db2 := NewDirectoryBlock(db1)
	j, _ := primitives.HexToHash("df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604")
	i, _ := primitives.HexToHash("b926da5ea5840b34189c37c55db9eb482f6e370bd097a16d6e890bc000c10898")
	db2.SetEntryHash(j, i, 3)

	l, _ := primitives.HexToHash("3e3eb61fb20e71d8211882075d404f5929618a189d23aba8c892b22228aa0d71")
	q, _ := primitives.HexToHash("9daad42e5efedf3075fa2cf51908babdb568f431a3c13b9a496ffbfb7160ad2e")
	db2.SetEntryHash(l, q, 4)

	_, err := db2.MarshalBinary()

	if nil != err {
		t.Fail()
	}

	if nil != CheckBlockPairIntegrity(db2, db1) {
		t.Fail()
	}

	bm := db.New()
	bms := fmt.Sprintln(bm)

	expectedString1 = `              KeyMR: e118d53659a92c69ad37602827bfaf256428867c6827f10829d678d7f8ddab33
             BodyMR: e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
           FullHash: e9f7dd92d52ff4efa31d0fb6521e64a30e211a262a6d407588f3b9229343b06b
  Version:         0
  NetworkID:       0
  BodyMR:          e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
  PrevKeyMR:       0000000000000000000000000000000000000000000000000000000000000000
  PrevFullHash:    0000000000000000000000000000000000000000000000000000000000000000
  Timestamp:       0
  Timestamp Str:   `
	epoch = time.Unix(0, 0)
	expectedString2 = epoch.Format("2006-01-02 15:04:05")

	expectedString3 = `
  DBHeight:        0
  BlockCount:      0
Entries: 

`
	expectedString = expectedString1 + expectedString2 + expectedString3

	if bms != expectedString {
		fmt.Println(bms)
		fmt.Println(expectedString)
		t.Fail()
	}

}
