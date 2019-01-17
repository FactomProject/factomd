package entryBlock_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
)

/*
func TestUnmarshal(t *testing.T) {
	e := new(Entry)

	data, err := hex.DecodeString("00954d5a49fd70d9b8bcdb35d252267829957f7ef7fa6c74f88419bdc5e82209f4000600110004746573745061796c6f616448657265")
	if err != nil {
		t.Error(err)
	}

	if err := e.UnmarshalBinary(data); err != nil {
		t.Error(err)
	}

	t.Log(e)
}*/

func TestUnmarshalNilEntry(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(Entry)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestFirstEntry(t *testing.T) {
	entry := new(Entry)

	entry.ExtIDs = make([]primitives.ByteSlice, 0, 5)
	entry.ExtIDs = append(entry.ExtIDs, primitives.ByteSlice{Bytes: []byte("1asdfadfasdf")})
	entry.ExtIDs = append(entry.ExtIDs, primitives.ByteSlice{Bytes: []byte("")})
	entry.ExtIDs = append(entry.ExtIDs, primitives.ByteSlice{Bytes: []byte("3")})
	entry.ChainID = new(primitives.Hash)
	err := entry.ChainID.SetBytes(constants.EC_CHAINID)
	if err != nil {
		t.Errorf("Error:%v", err)
	}

	entry.Content = primitives.ByteSlice{Bytes: []byte("1asdf asfas dfsg\"08908098(*)*^*&%&%&$^#%##%$$@$@#$!$#!$#@!~@!#@!%#@^$#^&$*%())_+_*^*&^&\"\"?>?<<>/./,")}

	bytes1, err := entry.MarshalBinary()
	t.Logf("bytes1:%v\n", bytes1)

	entry2 := new(Entry)
	entry2.UnmarshalBinary(bytes1)

	bytes2, _ := entry2.MarshalBinary()
	t.Logf("bytes2:%v\n", bytes2)

	if bytes.Compare(bytes1, bytes2) != 0 {
		t.Errorf("Invalid output")
	}

	if err != nil {
		t.Errorf("Error: %v", err)
	}
}

func TestEntry(t *testing.T) {
	entry := new(Entry)

	entry.ExtIDs = make([]primitives.ByteSlice, 0, 5)
	entry.ExtIDs = append(entry.ExtIDs, primitives.ByteSlice{Bytes: []byte("1asdfadfasdf")})
	entry.ExtIDs = append(entry.ExtIDs, primitives.ByteSlice{Bytes: []byte("2asdfas asfasfasfafas ")})
	entry.ExtIDs = append(entry.ExtIDs, primitives.ByteSlice{Bytes: []byte("3sd fasfas fsaf asf asfasfsafsfa")})
	entry.ChainID = new(primitives.Hash)
	err := entry.ChainID.SetBytes(constants.EC_CHAINID)
	if err != nil {
		t.Errorf("Error:%v", err)
	}

	entry.Content = primitives.ByteSlice{Bytes: []byte("1asdf asfas fasfadfasdfasfdfff12345")}

	bytes1, err := entry.MarshalBinary()
	t.Logf("bytes1:%v\n", bytes1)

	entry2 := new(Entry)
	entry2.UnmarshalBinary(bytes1)

	bytes2, _ := entry2.MarshalBinary()
	t.Logf("bytes2:%v\n", bytes2)

	if bytes.Compare(bytes1, bytes2) != 0 {
		t.Errorf("Invalid output")
	}

	if err != nil {
		t.Errorf("Error:%v", err)
	}

	entry3, err := UnmarshalEntry(bytes1)
	if err != nil {
		t.Errorf("Error:%v", err)
	}
	bytes3, _ := entry3.MarshalBinary()
	t.Logf("bytes3:%v\n", bytes3)

	if bytes.Compare(bytes1, bytes3) != 0 {
		t.Errorf("Invalid output")
	}
}

func TestEntryMisc(t *testing.T) {
	e := newEntry()
	if e.IsValid() == false {
		t.Fail()
	}
	if e.GetHash().String() != "24674e6bc3094eb773297de955ee095a05830e431da13a37382dcdc89d73c7d7" {
		t.Fail()
	}
	//	if e.GetChainID().String() != "df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604" {
	//		t.Fail()
	//	}
	ids := e.ExternalIDs()
	if len(ids) != 1 {
		t.Fail()
	}
	if NewChainID(e).String() != "df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604" {
		t.Fail()
	}

	if e.GetDatabaseHeight() != 0 {
		t.Errorf("Returned wrong height")
	}

	if e.GetWeldHash().String() != "c255e5da4dd6202448db0ed8e938d0c6a2a0f370c527c27f96efb602935e9c9f" {
		t.Errorf("Returned wrong WeldHash - %v", e.GetWeldHash().String())
	}

	if e.GetChainID().String() != "df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604" {
		t.Errorf("Returned wrong ChainID - %v", e.GetChainID().String())
	}

	if e.DatabasePrimaryIndex().String() != "24674e6bc3094eb773297de955ee095a05830e431da13a37382dcdc89d73c7d7" {
		t.Errorf("Returned wrong DatabasePrimaryIndex - %v", e.DatabasePrimaryIndex().String())
	}

	if e.DatabaseSecondaryIndex() != nil {
		t.Errorf("Returned wrong DatabaseSecondaryIndex")
	}

	if e.GetChainID().String() != "df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604" {
		t.Errorf("Returned wrong GetChainID - %v", e.GetChainID().String())
	}

	if fmt.Sprintf("%x", e.GetContent()) != "546869732069732074686520466163746f6d20616e63686f7220636861696e2c207768696368207265636f7264732074686520616e63686f727320466163746f6d2070757473206f6e20426974636f696e20616e64206f74686572206e6574776f726b732e0a" {
		t.Errorf("Returned wrong GetChainID - %x", e.GetContent())
	}
}

func TestKSize(t *testing.T) {
	e := NewEntry()
	content := []byte{}
	for i := 0; i < 256; i++ {
		content = append(content, []byte{0x11, 0x22, 0x33, 0x44}...)
	}
	e.Content = primitives.ByteSlice{Bytes: content}
	if e.KSize() != 1 {
		t.Fail()
	}
}

func newEntry() *Entry {
	e := NewEntry()
	entryStr := "00df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e60400130011466163746f6d416e63686f72436861696e546869732069732074686520466163746f6d20616e63686f7220636861696e2c207768696368207265636f7264732074686520616e63686f727320466163746f6d2070757473206f6e20426974636f696e20616e64206f74686572206e6574776f726b732e0a"
	h, err := hex.DecodeString(entryStr)
	if err != nil {
		panic(err)
	}
	err = e.UnmarshalBinary(h)
	if err != nil {
		panic(err)
	}
	return e
}

func TestMarshalUnmarshalEntryList(t *testing.T) {
	for i := 0; i < 1000; i++ {
		l := random.RandIntBetween(0, 30)
		es := []interfaces.IEBEntry{}
		for j := 0; j < l; j++ {
			es = append(es, RandomEntry())
		}
		b, err := MarshalEntryList(es)
		if err != nil {
			t.Errorf("%v", err)
		}
		list, rest, err := UnmarshalEntryList(b)
		if err != nil {
			t.Errorf("%v", err)
		}
		if len(rest) > 0 {
			t.Errorf("Too much data returned - %x", rest)
		}
		if len(list) != len(es) {
			t.Errorf("Wrong amount of entries returned - %v vs %v", len(list), len(es))
			continue
		}
		for i := range list {
			if list[i].IsSameAs(es[i]) == false {
				t.Errorf("Entries are not the same - %v vs %v", list[i], es[i])
			}
		}
	}
}

func TestUnmarshalBadEntryList(t *testing.T) {
	el := make([]interfaces.IEBEntry, 0)

	for i := 0; i < 100; i++ {
		l := random.RandIntBetween(0, 30)
		for j := 0; j < l; j++ {
			el = append(el, RandomEntry())
		}
	}

	p, err := MarshalEntryList(el)
	if err != nil {
		t.Errorf("%v", err)
	}
	// write bad value into list length
	p[1] = 0xff

	el2, p, err := UnmarshalEntryList(p)
	if err == nil {
		t.Error("EntryList should have errored on unmarshal", el2)
	} else {
		t.Logf("%x\n%s", p, err)
	}
}
