package identity

import (
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// EntryBlockSync has the current eblock synced to, and the target Eblock
//	It also has the blocks in between in order. This makes it so traversing only needs to happen once
type EntryBlockSync struct {
	Current          EntryBlockMarker
	Target           EntryBlockMarker
	BlocksToBeParsed []EntryBlockMarker
}

func NewEntryBlockSync() *EntryBlockSync {
	e := new(EntryBlockSync)
	e.Current = *NewEntryBlockMarker()
	e.Target = *NewEntryBlockMarker()

	return e
}

// Synced returns if fully synced (current == target)
func (a *EntryBlockSync) Synced() bool {
	return a.Current.IsSameAs(&a.Target)
}

// NextEBlock returns the next eblock that is needed to be parsed
func (a *EntryBlockSync) NextEBlock() *EntryBlockMarker {
	if len(a.BlocksToBeParsed) == 0 {
		return nil
	}
	return &a.BlocksToBeParsed[0]
}

// BlockParsed indicates a block has been parsed. We update our current
func (a *EntryBlockSync) BlockParsed(block EntryBlockMarker) {
	if !a.BlocksToBeParsed[0].IsSameAs(&block) {
		panic("This block should be next in the list")
	}
	a.Current = block
	a.BlocksToBeParsed = a.BlocksToBeParsed[1:]
}

func (a *EntryBlockSync) AddNewHead(keymr interfaces.IHash, seq uint32, ht uint32, dblockTimestamp interfaces.Timestamp) {
	a.AddNewHeadMarker(EntryBlockMarker{keymr, seq, ht, dblockTimestamp})
}

// AddNewHead will add a new eblock to be parsed to the head (tail of list)
//	Since the block needs to be parsed, it is the new target and added to the blocks to be parsed
func (a *EntryBlockSync) AddNewHeadMarker(marker EntryBlockMarker) {
	if marker.DBHeight < a.Target.DBHeight {
		return // Already added this target
	}
	a.BlocksToBeParsed = append(a.BlocksToBeParsed, marker)
	a.Target = marker
}

func (a *EntryBlockSync) IsSameAs(b *EntryBlockSync) bool {
	if !a.Current.IsSameAs(&b.Current) {
		return false
	}
	if !a.Target.IsSameAs(&b.Target) {
		return false
	}

	if len(a.BlocksToBeParsed) != len(b.BlocksToBeParsed) {
		return false
	}

	for i := range a.BlocksToBeParsed {
		if !a.BlocksToBeParsed[1].IsSameAs(&b.BlocksToBeParsed[i]) {
			return false
		}
	}

	return true
}

func (e *EntryBlockSync) Clone() *EntryBlockSync {
	b := new(EntryBlockSync)
	b.Current = *e.Current.Clone()
	b.Target = *e.Target.Clone()
	return b
}

func (e *EntryBlockSync) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "EntryBlockSync.MarshalBinary err:%v", *pe)
		}
	}(&err)
	buf := primitives.NewBuffer(nil)
	err = buf.PushBinaryMarshallable(&e.Current)
	if err != nil {
		return nil, err
	}

	err = buf.PushBinaryMarshallable(&e.Target)
	if err != nil {
		return nil, err
	}

	err = buf.PushInt(len(e.BlocksToBeParsed))
	if err != nil {
		return nil, err
	}

	for _, v := range e.BlocksToBeParsed {
		buf.PushBinaryMarshallable(&v)
	}

	return buf.DeepCopyBytes(), nil
}

func (e *EntryBlockSync) UnmarshalBinary(p []byte) error {
	_, err := e.UnmarshalBinaryData(p)
	return err
}

func (e *EntryBlockSync) UnmarshalBinaryData(p []byte) (newData []byte, err error) {
	buf := primitives.NewBuffer(p)
	newData = p

	err = buf.PopBinaryMarshallable(&e.Current)
	if err != nil {
		return
	}

	err = buf.PopBinaryMarshallable(&e.Target)
	if err != nil {
		return
	}

	// blockLimit is the mazimum number of Entry Blocks that could fit in the
	// buffer. Smallest possible Entry Block is 140 bytes.
	blockLimit := buf.Len() / 140
	blockCount, err := buf.PopInt()
	if err != nil {
		return
	}
	if blockCount > blockLimit {
		return nil, fmt.Errorf(
			"Error: EntryBlockSync.UnmarshalBinary: block count %d is greater "+
				"than remaining space in buffer %d (uint underflow?)",
			blockCount, blockLimit,
		)
	}

	e.BlocksToBeParsed = make([]EntryBlockMarker, blockCount)

	for i := 0; i < blockCount; i++ {
		var b EntryBlockMarker
		err = buf.PopBinaryMarshallable(&b)
		if err != nil {
			return
		}
		e.BlocksToBeParsed[i] = b
	}

	newData = buf.DeepCopyBytes()
	return
}

type EntryBlockMarkerList []EntryBlockMarker

func (p EntryBlockMarkerList) Len() int {
	return len(p)
}

func (p EntryBlockMarkerList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p EntryBlockMarkerList) Less(i, j int) bool {
	return p[i].Sequence < p[j].Sequence
}

type EntryBlockMarker struct {
	KeyMr           interfaces.IHash
	Sequence        uint32
	DBHeight        uint32
	DblockTimestamp interfaces.Timestamp
}

func NewEntryBlockMarker() *EntryBlockMarker {
	e := new(EntryBlockMarker)
	e.KeyMr = primitives.NewZeroHash()
	e.DblockTimestamp = new(primitives.Timestamp)
	return e
}

func (a *EntryBlockMarker) IsSameAs(b *EntryBlockMarker) bool {
	if !a.KeyMr.IsSameAs(b.KeyMr) {
		return false
	}
	if a.Sequence != b.Sequence {
		return false
	}
	if a.DBHeight != b.DBHeight {
		return false
	}
	if !a.DblockTimestamp.IsSameAs(b.DblockTimestamp) {
		return false
	}
	return true
}

func (e *EntryBlockMarker) Clone() *EntryBlockMarker {
	b := new(EntryBlockMarker)
	b.KeyMr = e.KeyMr.Copy()
	b.Sequence = e.Sequence
	b.DBHeight = e.DBHeight
	b.DblockTimestamp = e.DblockTimestamp
	return b
}

func (e *EntryBlockMarker) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "EntryBlockMarker.MarshalBinary err:%v", *pe)
		}
	}(&err)
	buf := primitives.NewBuffer(nil)
	err = buf.PushIHash(e.KeyMr)
	if err != nil {
		return nil, err
	}

	err = buf.PushUInt32(e.Sequence)
	if err != nil {
		return nil, err
	}

	err = buf.PushUInt32(e.DBHeight)
	if err != nil {
		return nil, err
	}

	err = buf.PushTimestamp(e.DblockTimestamp)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (e *EntryBlockMarker) UnmarshalBinary(p []byte) error {
	_, err := e.UnmarshalBinaryData(p)
	return err
}

func (e *EntryBlockMarker) UnmarshalBinaryData(p []byte) (newData []byte, err error) {
	buf := primitives.NewBuffer(p)
	newData = p

	e.KeyMr, err = buf.PopIHash()
	if err != nil {
		return
	}

	e.Sequence, err = buf.PopUInt32()
	if err != nil {
		return
	}

	e.DBHeight, err = buf.PopUInt32()
	if err != nil {
		return
	}

	e.DblockTimestamp, err = buf.PopTimestamp()
	if err != nil {
		return
	}

	newData = buf.DeepCopyBytes()
	return
}
