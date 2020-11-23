package state_test

import (
	"bytes"
	"testing"

	"github.com/FactomProject/factomd/common/primitives/random"
	. "github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
)

func TestUInt32Bytes(t *testing.T) {
	var i uint32 = 0
	for ; i < 1000; i++ {
		a := uint32(random.RandInt())
		data, err := Uint32ToBytes(a)
		if err != nil {
			t.Error(err)
		}

		b, err := BytesToUint32(data)
		if err != nil {
			t.Error(err)
		}

		if b != a {
			t.Error("Failed, should be same")
		}
	}
}

func TestWholeBlocks(t *testing.T) {
	blocks := makeDBStateList(100)
	all := make([]byte, 0)

	var err error

	for _, a := range blocks {
		b := NewWholeBlock()
		data, err := a.MarshalBinary()
		if err != nil {
			t.Error(err)
		}
		all = append(all, data...)
		r := bytes.NewReader(data)

		newData, err := b.UnmarshalBinaryData(data)
		if err != nil {
			t.Error(err)
		}

		if len(newData) > 0 {
			t.Error("Should have no bytes left")
		}

		if !a.IsSameAs(b) {
			t.Error("[Slice] Should be same")
		}

		c := NewWholeBlock()
		_, err = c.UnmarshalBinaryDataBuffer(r, 0)
		if err != nil {
			t.Error(err)
		}

		if !a.IsSameAs(c) {
			t.Error("[Buffer] Should be same")
		}
		var _ = r

	}

	var off int64 = 0
	ra := bytes.NewReader(all)
	// Test block of binary unmarshal into individual blocks
	for i, a := range blocks {
		b := NewWholeBlock()
		all, err = b.UnmarshalBinaryData(all)
		if err != nil {
			t.Error(err)
		}

		if !a.IsSameAs(b) {
			t.Error("[Slice] Should be same")
		}

		c := NewWholeBlock()
		n, err := c.UnmarshalBinaryDataBuffer(ra, int(off))
		if err != nil {
			t.Errorf("Block: %d, Offset: %d :: %s", i, off, err.Error())
		}
		off += n
		if !a.IsSameAs(c) {
			t.Error("[Buffer] Should be same")
		}

	}
	if len(all) > 0 {
		t.Error("Bytes left over")
	}

	var _ = off
	var _ = ra
}

func makeDBStateList(l int) []*WholeBlock {
	blocks := make([]*testHelper.BlockSet, l)
	one := testHelper.CreateTestBlockSet(nil)
	msgList := make([]*WholeBlock, l)

	blocks[0] = one
	msg := blockToWholeBlock(blocks[0])
	msgList[0] = msg
	for i := 1; i < l; i++ {
		blocks[i] = testHelper.CreateTestBlockSet(blocks[i-1])
		msg := blockToWholeBlock(blocks[i])
		msgList[i] = msg

		if (i+1)%1000 == 0 {

		}
	}

	return msgList
}

func blockToWholeBlock(set *testHelper.BlockSet) *WholeBlock {
	block := NewWholeBlock()
	block.DBlock = set.DBlock
	block.ABlock = set.ABlock
	block.FBlock = set.FBlock
	block.ECBlock = set.ECBlock
	block.AddEblock(set.EBlock)
	block.AddEblock(set.AnchorEBlock)
	for _, e := range set.Entries {
		block.AddEntry(e)
	}

	return block
}
