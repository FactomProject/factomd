// +build all 

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package databaseOverlay_test

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
	. "github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/mapdb"
	"github.com/FactomProject/factomd/testHelper"
)

/*
func TestListAllBuckets(t *testing.T) {
	dbo := testHelper.CreateAndPopulateTestDatabaseOverlay()
	buckets, err := dbo.ListAllBuckets()
	if err != nil {
		t.Error(err.Error())
	}
	for _, b := range buckets {
		t.Errorf("%x", b)
	}
}
*/

func TestGetEntryType(t *testing.T) {
	blocks := testHelper.CreateFullTestBlockSet()
	dbo := testHelper.CreateAndPopulateTestDatabaseOverlay()

	for _, block := range blocks {
		eType, err := dbo.GetEntryType(block.DBlock.DatabasePrimaryIndex())
		if err != nil {
			t.Error(err)
		}
		if eType == nil {
			t.Error("eType==nil")
		}
		if eType.IsSameAs(block.DBlock.GetChainID()) == false {
			t.Error("Block type mismatch")
		}

		eType, err = dbo.GetEntryType(block.ABlock.DatabasePrimaryIndex())
		if err != nil {
			t.Error(err)
		}
		if eType == nil {
			t.Error("eType==nil")
		}
		if eType.IsSameAs(block.ABlock.GetChainID()) == false {
			t.Error("Block type mismatch")
		}

		eType, err = dbo.GetEntryType(block.ECBlock.DatabasePrimaryIndex())
		if err != nil {
			t.Error(err)
		}
		if eType == nil {
			t.Error("eType==nil")
		}
		if eType.IsSameAs(block.ECBlock.GetChainID()) == false {
			t.Error("Block type mismatch")
		}

		eType, err = dbo.GetEntryType(block.FBlock.DatabasePrimaryIndex())
		if err != nil {
			t.Error(err)
		}
		if eType == nil {
			t.Error("eType==nil")
		}
		if eType.IsSameAs(block.FBlock.GetChainID()) == false {
			t.Error("Block type mismatch")
		}

		eType, err = dbo.GetEntryType(block.AnchorEBlock.DatabasePrimaryIndex())
		if err != nil {
			t.Error(err)
		}
		if eType == nil {
			t.Error("eType==nil")
		}
		if eType.IsSameAs(block.AnchorEBlock.GetChainID()) == false {
			t.Error("Block type mismatch")
		}

		eType, err = dbo.GetEntryType(block.EBlock.DatabasePrimaryIndex())
		if err != nil {
			t.Error(err)
		}
		if eType == nil {
			t.Error("eType==nil")
		}
		if eType.IsSameAs(block.EBlock.GetChainID()) == false {
			t.Error("Block type mismatch")
		}

		for _, entry := range block.Entries {
			eType, err = dbo.GetEntryType(entry.GetHash())
			if err != nil {
				t.Error(err)
			}
			if eType == nil {
				t.Error("eType==nil")
			}
			if eType.IsSameAs(entry.GetChainID()) == false {
				t.Error("Entry type mismatch")
			}
		}
	}
}

func TestMultiBatch(t *testing.T) {
	dbo := NewOverlay(new(mapdb.MapDB))

	var prev *testHelper.BlockSet = nil

	var err error

	for i := 0; i < 10; i++ {
		dbo.StartMultiBatch()
		prev = testHelper.CreateTestBlockSet(prev)

		err = dbo.ProcessABlockMultiBatch(prev.ABlock)
		if err != nil {
			t.Error(err)
		}

		err = dbo.ProcessEBlockMultiBatch(prev.EBlock, true)
		if err != nil {
			t.Error(err)
		}

		err = dbo.ProcessEBlockMultiBatch(prev.AnchorEBlock, true)
		if err != nil {
			t.Error(err)
		}

		err = dbo.ProcessECBlockMultiBatch(prev.ECBlock, false)
		if err != nil {
			t.Error(err)
		}

		err = dbo.ProcessFBlockMultiBatch(prev.FBlock)
		if err != nil {
			t.Error(err)
		}

		err = dbo.ProcessDBlockMultiBatch(prev.DBlock)
		if err != nil {
			t.Error(err)
		}

		for _, entry := range prev.Entries {
			err = dbo.InsertEntryMultiBatch(entry)
			if err != nil {
				t.Error(err)
			}
		}

		if err := dbo.ExecuteMultiBatch(); err != nil {
			t.Error(err)
		}
	}

	ahead, err := dbo.FetchABlockHead()
	if err != nil {
		t.Error(err)
	}
	if ahead == nil {
		t.Error("DBlock head is nil")
	}

	m1, err := prev.ABlock.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	m2, err := ahead.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	if primitives.AreBytesEqual(m1, m2) == false {
		t.Error("ABlocks are not equal")
	}

	fhead, err := dbo.FetchFBlockHead()
	if err != nil {
		t.Error(err)
	}
	if fhead == nil {
		t.Error("DBlock head is nil")
	}

	m1, err = prev.FBlock.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	m2, err = fhead.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	if primitives.AreBytesEqual(m1, m2) == false {
		t.Error("FBlocks are not equal")
	}

	echead, err := dbo.FetchECBlockHead()
	if err != nil {
		t.Error(err)
	}
	if echead == nil {
		t.Error("DBlock head is nil")
	}

	m1, err = prev.ECBlock.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	m2, err = echead.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	if primitives.AreBytesEqual(m1, m2) == false {
		t.Error("ECBlocks are not equal")
	}

	dhead, err := dbo.FetchDBlockHead()
	if err != nil {
		t.Error(err)
	}
	if dhead == nil {
		t.Error("DBlock head is nil")
	}

	m1, err = prev.DBlock.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	m2, err = dhead.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	if primitives.AreBytesEqual(m1, m2) == false {
		t.Error("DBlocks are not equal")
	}
}

func TestInsertFetch(t *testing.T) {
	dbo := createOverlay()
	defer dbo.Close()
	b := NewDBTestObject()
	b.Data = []byte{0x00, 0x01, 0x02, 0x03}

	err := dbo.Insert(TestBucket, b)
	if err != nil {
		t.Error(err)
	}

	index := b.DatabasePrimaryIndex()

	b2 := NewDBTestObject()
	resp, err := dbo.FetchBlock(TestBucket, index, b2)
	if err != nil {
		t.Error(err)
	}

	if resp == nil {
		t.Error("Response is nil while it shouldn't be.")
	}

	bResp := resp.(*DBTestObject)

	bytes1 := b.Data
	bytes2 := bResp.Data

	if primitives.AreBytesEqual(bytes1, bytes2) == false {
		t.Errorf("Bytes are not equal - %x vs %x", bytes1, bytes2)
	}
}

func TestFetchBy(t *testing.T) {
	dbo := createOverlay()
	defer dbo.Close()

	blocks := []*DBTestObject{}

	max := 10
	for i := 0; i < max; i++ {
		b := NewDBTestObject()
		b.ChainID = primitives.NewHash(CopyZeroHash())
		b.Data = []byte{byte(i), byte(i + 1), byte(i + 2), byte(i + 3)}

		primaryIndex := CopyZeroHash()
		primaryIndex[len(primaryIndex)-1] = byte(i)

		secondaryIndex := CopyZeroHash()
		secondaryIndex[0] = byte(i)

		b.PrimaryIndex = primitives.NewHash(primaryIndex)
		b.SecondaryIndex = primitives.NewHash(secondaryIndex)
		b.DatabaseHeight = uint32(i)
		blocks = append(blocks, b)

		err := dbo.ProcessBlockBatch(TestBucket, TestNumberBucket, TestSecondaryIndexBucket, b)
		if err != nil {
			t.Error(err)
		}
	}
	headIndex, err := dbo.FetchHeadIndexByChainID(primitives.NewHash(CopyZeroHash()))
	if err != nil {
		t.Error(err)
	}
	if headIndex.IsSameAs(blocks[max-1].PrimaryIndex) == false {
		t.Error("Wrong chain head")
	}

	head, err := dbo.FetchChainHeadByChainID(TestBucket, primitives.NewHash(CopyZeroHash()), new(DBTestObject))
	if err != nil {
		t.Error(err)
	}
	if blocks[max-1].IsEqual(head.(*DBTestObject)) == false {
		t.Error("Heads are not equal")
	}

	for i := 0; i < max; i++ {
		primaryIndex := CopyZeroHash()
		primaryIndex[len(primaryIndex)-1] = byte(i)

		secondaryIndex := CopyZeroHash()
		secondaryIndex[0] = byte(i)

		dbHeight := uint32(i)

		block, err := dbo.FetchBlockByHeight(TestNumberBucket, TestBucket, dbHeight, new(DBTestObject))
		if err != nil {
			t.Error(err)
		}
		if blocks[i].IsEqual(block.(*DBTestObject)) == false {
			t.Error("Blocks are not equal")
		}

		index, err := dbo.FetchBlockIndexByHeight(TestNumberBucket, dbHeight)
		if err != nil {
			t.Error(err)
		}
		if primitives.AreBytesEqual(index.Bytes(), primaryIndex) == false {
			t.Error("Wrong primary index returned")
		}

		index, err = dbo.FetchPrimaryIndexBySecondaryIndex(TestSecondaryIndexBucket, primitives.NewHash(secondaryIndex))
		if err != nil {
			t.Error(err)
		}
		if primitives.AreBytesEqual(index.Bytes(), primaryIndex) == false {
			t.Error("Wrong primary index returned")
		}

		block, err = dbo.FetchBlockBySecondaryIndex(TestSecondaryIndexBucket, TestBucket, primitives.NewHash(secondaryIndex), new(DBTestObject))
		if err != nil {
			t.Error(err)
		}
		if blocks[i].IsEqual(block.(*DBTestObject)) == false {
			t.Error("Blocks are not equal")
		}
	}

	fetchedBlocks, err := dbo.FetchAllBlocksFromBucket(TestBucket, new(DBTestObject))
	if err != nil {
		t.Error(err)
	}
	if len(fetchedBlocks) != len(blocks) {
		t.Error("Invalid amount of blocks returned")
	}
	for i := 0; i < max; i++ {
		if blocks[i].IsEqual(fetchedBlocks[i].(*DBTestObject)) == false {
			t.Error("Block from batch is not equal")
		}
	}

	startIndex := 3
	indexCount := 4
	fetchedIndexes, err := dbo.FetchBlockIndexesInHeightRange(TestNumberBucket, int64(startIndex), int64(startIndex+indexCount))
	if len(fetchedIndexes) != indexCount {
		t.Error("Invalid amount of indexes returned")
	}
	for i := 0; i < indexCount; i++ {
		primaryIndex := CopyZeroHash()
		primaryIndex[len(primaryIndex)-1] = byte(i + startIndex)

		if primitives.AreBytesEqual(primaryIndex, fetchedIndexes[i].Bytes()) == false {
			t.Error("Index from batch is not equal")
		}
	}
}

func createOverlay() *Overlay {
	return testHelper.CreateEmptyTestDatabaseOverlay()
}

var TestBucket []byte = []byte{0x01}
var TestNumberBucket []byte = []byte{0x02}
var TestSecondaryIndexBucket []byte = []byte{0x03}

type bareDBTestObject struct {
	Data           []byte
	DatabaseHeight uint32
	PrimaryIndex   interfaces.IHash
	SecondaryIndex interfaces.IHash
	ChainID        interfaces.IHash
}

func NewBareDBTestObject() *bareDBTestObject {
	d := new(bareDBTestObject)
	d.Data = []byte{}
	d.DatabaseHeight = 0
	d.PrimaryIndex = new(primitives.Hash)
	d.SecondaryIndex = new(primitives.Hash)
	d.ChainID = new(primitives.Hash)
	return d
}

type DBTestObject struct {
	Data           []byte
	DatabaseHeight uint32
	PrimaryIndex   interfaces.IHash
	SecondaryIndex interfaces.IHash
	ChainID        interfaces.IHash
}

func NewDBTestObject() *DBTestObject {
	d := new(DBTestObject)
	d.Data = []byte{}
	d.DatabaseHeight = 0
	d.PrimaryIndex = new(primitives.Hash)
	d.SecondaryIndex = new(primitives.Hash)
	d.ChainID = new(primitives.Hash)
	return d
}

var _ interfaces.DatabaseBatchable = (*DBTestObject)(nil)

func (d *DBTestObject) GetDatabaseHeight() uint32 {
	return d.DatabaseHeight
}

func (d *DBTestObject) DatabasePrimaryIndex() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("DBTestObject.DatabasePrimaryIndex() saw an interface that was nil")
		}
	}()

	return d.PrimaryIndex
}

func (d *DBTestObject) DatabaseSecondaryIndex() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("DBTestObject.DatabaseSecondaryIndex() saw an interface that was nil")
		}
	}()

	return d.SecondaryIndex
}

func (d *DBTestObject) GetChainID() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("DBTestObject.GetChainID() saw an interface that was nil")
		}
	}()

	return d.ChainID
}

func (d *DBTestObject) New() interfaces.BinaryMarshallableAndCopyable {
	return NewDBTestObject()
}

func (d *DBTestObject) UnmarshalBinaryData(data []byte) ([]byte, error) {
	dec := gob.NewDecoder(bytes.NewBuffer(data))

	tmp := NewBareDBTestObject()

	err := dec.Decode(tmp)
	if err != nil {
		return nil, err
	}

	d.Data = tmp.Data
	d.DatabaseHeight = tmp.DatabaseHeight
	d.PrimaryIndex = tmp.PrimaryIndex
	d.SecondaryIndex = tmp.SecondaryIndex
	d.ChainID = tmp.ChainID

	return nil, nil
}

func (d *DBTestObject) UnmarshalBinary(data []byte) error {
	_, err := d.UnmarshalBinaryData(data)
	return err
}

func (d *DBTestObject) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "DBTestObject.MarshalBinary err:%v", *pe)
		}
	}(&err)
	var data bytes.Buffer

	enc := gob.NewEncoder(&data)

	tmp := new(bareDBTestObject)

	tmp.Data = d.Data
	tmp.DatabaseHeight = d.DatabaseHeight
	tmp.PrimaryIndex = d.PrimaryIndex
	tmp.SecondaryIndex = d.SecondaryIndex
	tmp.ChainID = d.ChainID

	err = enc.Encode(tmp)
	if err != nil {
		return nil, err
	}

	return data.Bytes(), nil
}

func (d1 *DBTestObject) IsEqual(d2 *DBTestObject) bool {
	if d1.DatabaseHeight != d2.DatabaseHeight {
		return false
	}

	if primitives.AreBytesEqual(d1.Data, d2.Data) == false {
		return false
	}

	if primitives.AreBytesEqual(d1.PrimaryIndex.Bytes(), d2.PrimaryIndex.Bytes()) == false {
		return false
	}

	if primitives.AreBytesEqual(d1.SecondaryIndex.Bytes(), d2.SecondaryIndex.Bytes()) == false {
		return false
	}

	if primitives.AreBytesEqual(d1.ChainID.Bytes(), d2.ChainID.Bytes()) == false {
		return false
	}

	return true
}

func CopyZeroHash() []byte {
	answer := make([]byte, len(constants.ZERO_HASH))
	return answer
}

func TestDoesKeyExist(t *testing.T) {
	m := createOverlay()
	defer m.Close()

	for i := 0; i < 1000; i++ {
		key := random.RandNonEmptyByteSlice()
		bucket := random.RandNonEmptyByteSlice()

		test := NewDBTestObject()

		err := m.Put(bucket, key, test)
		if err != nil {
			t.Errorf("%v", err)
		}

		exists, err := m.DoesKeyExist(bucket, key)
		if err != nil {
			t.Errorf("%v", err)
		}

		if exists == false {
			t.Errorf("Key does not exist")
		}

		key = random.RandNonEmptyByteSlice()
		bucket = random.RandNonEmptyByteSlice()

		exists, err = m.DoesKeyExist(bucket, key)
		if err != nil {
			t.Errorf("%v", err)
		}

		if exists == true {
			t.Errorf("Key does exist while it shouldn't")
		}
	}
}

func TestFetchAllBlockKeysFromBucket(t *testing.T) {
	dbo := testHelper.CreateAndPopulateTestDatabaseOverlay()

	_, keys, err := dbo.GetAll(INCLUDED_IN, primitives.NewZeroHash())
	if err != nil {
		t.Errorf("%v", err)
	}
	if len(keys) != 150 {
		t.Errorf("Invalid amount of keys returned - expected 150, got %v", len(keys))
	}
	for i := range keys {
		for j := i + 1; j < len(keys); j++ {
			if primitives.AreBytesEqual(keys[i], keys[j]) {
				t.Errorf("Key %v is equal to key %v - %x", i, j, keys[i])
			}
		}
	}
}
