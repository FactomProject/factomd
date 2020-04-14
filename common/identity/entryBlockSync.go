package identity

import (
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
)

// EntryBlockSync has the current eblock synced to, and the target Eblock
//	It also has the blocks in between in order. This makes it so traversing only needs to happen once
type EntryBlockSync struct {
	Current          EntryBlockMarker
	Target           EntryBlockMarker
	BlocksToBeParsed []EntryBlockMarker
}

// RandomEntryBlockSync returns a new random EntryBlockSync with 0-9 EntryBlockMarkers
func RandomEntryBlockSync() *EntryBlockSync {
	s := NewEntryBlockSync()

	for i := 0; i < random.RandIntBetween(0, 10); i++ {
		m := RandomEntryBlockMarker()
		m.Sequence = uint32(i)
		m.DBHeight = uint32(i)
		s.AddNewHeadMarker(*m)
	}

	return s
}

// NewEntryBlockSync returns a new object
func NewEntryBlockSync() *EntryBlockSync {
	e := new(EntryBlockSync)
	e.Current = *NewEntryBlockMarker()
	e.Target = *NewEntryBlockMarker()

	return e
}

// Synced returns true if fully synced (current == target)
func (e *EntryBlockSync) Synced() bool {
	return e.Current.IsSameAs(&e.Target)
}

// NextEBlock returns the next eblock that is needed to be parsed
func (e *EntryBlockSync) NextEBlock() *EntryBlockMarker {
	if len(e.BlocksToBeParsed) == 0 {
		return nil
	}
	return &e.BlocksToBeParsed[0]
}

// BlockParsed indicates a block has been parsed. We update our current block to the input
func (e *EntryBlockSync) BlockParsed(block EntryBlockMarker) {
	if !e.BlocksToBeParsed[0].IsSameAs(&block) {
		panic("This block should be next in the list")
	}
	e.Current = block
	e.BlocksToBeParsed = e.BlocksToBeParsed[1:]
}

// AddNewHead creates and adds a new EntryBlockMarker from the input parameters to the head (tail) of 'to be parsed' list
func (e *EntryBlockSync) AddNewHead(keymr interfaces.IHash, seq uint32, ht uint32, dblockTimestamp interfaces.Timestamp) {
	e.AddNewHeadMarker(EntryBlockMarker{keymr, seq, ht, dblockTimestamp})
}

// AddNewHeadMarker will add a new eblock to be parsed to the head (tail of list)
//	Since the block needs to be parsed, it is the new target and added to the blocks to be parsed
func (e *EntryBlockSync) AddNewHeadMarker(marker EntryBlockMarker) {
	if marker.DBHeight < e.Target.DBHeight {
		return // Already added this target
	}
	e.BlocksToBeParsed = append(e.BlocksToBeParsed, marker)
	e.Target = marker
}

// IsSameAs return true iff the input object is identical to this object
func (e *EntryBlockSync) IsSameAs(b *EntryBlockSync) bool {
	if !e.Current.IsSameAs(&b.Current) {
		return false
	}
	if !e.Target.IsSameAs(&b.Target) {
		return false
	}

	if len(e.BlocksToBeParsed) != len(b.BlocksToBeParsed) {
		return false
	}

	for i := range e.BlocksToBeParsed {
		if !e.BlocksToBeParsed[i].IsSameAs(&b.BlocksToBeParsed[i]) {
			return false
		}
	}

	return true
}

// Clone makes an identical copy of this object
func (e *EntryBlockSync) Clone() *EntryBlockSync {
	b := new(EntryBlockSync)
	b.Current = *e.Current.Clone()
	b.Target = *e.Target.Clone()

	b.BlocksToBeParsed = make([]EntryBlockMarker, len(e.BlocksToBeParsed))
	for i, eb := range e.BlocksToBeParsed {
		b.BlocksToBeParsed[i] = *eb.Clone()
	}

	return b
}

// MarshalBinary marshals this object
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
		err = buf.PushBinaryMarshallable(&v)
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinary unmarshals the input data into this object
func (e *EntryBlockSync) UnmarshalBinary(p []byte) error {
	_, err := e.UnmarshalBinaryData(p)
	return err
}

// UnmarshalBinaryData unmarshals the input data into this object
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

	// blockLimit is the maximum number of Entry Blocks that could fit in the
	// buffer.
	tmp := NewEntryBlockMarker()
	blockLimit := buf.Len() / tmp.Size()
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

// EntryBlockMarkerList slice of EntryBlockMarkers
type EntryBlockMarkerList []EntryBlockMarker

// Len returns the length of this object
func (p EntryBlockMarkerList) Len() int {
	return len(p)
}

// Swap swaps the values stored at indices 'i' and 'j'
func (p EntryBlockMarkerList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// Less returns true if the object at 'i' is les than the object at 'j'
func (p EntryBlockMarkerList) Less(i, j int) bool {
	return p[i].Sequence < p[j].Sequence
}

type EntryBlockMarker struct {
	KeyMr           interfaces.IHash
	Sequence        uint32
	DBHeight        uint32
	DblockTimestamp interfaces.Timestamp
}

// NewEntryBlockMarker returns a new entry block marker
func NewEntryBlockMarker() *EntryBlockMarker {
	e := new(EntryBlockMarker)
	e.KeyMr = primitives.NewZeroHash()
	e.DblockTimestamp = new(primitives.Timestamp)
	return e
}

// RandomEntryBlockMarker returns a new entry block marker with random initialized values
func RandomEntryBlockMarker() *EntryBlockMarker {
	m := NewEntryBlockMarker()
	m.KeyMr = primitives.RandomHash()
	m.DblockTimestamp = primitives.NewTimestampNow()
	return m
}

// IsSameAs returns true iff the input object is identical
func (e *EntryBlockMarker) IsSameAs(b *EntryBlockMarker) bool {
	if !e.KeyMr.IsSameAs(b.KeyMr) {
		return false
	}
	if e.Sequence != b.Sequence {
		return false
	}
	if e.DBHeight != b.DBHeight {
		return false
	}
	if !e.DblockTimestamp.IsSameAs(b.DblockTimestamp) {
		return false
	}
	return true
}

// Clone returns an identical copy of the this object
func (e *EntryBlockMarker) Clone() *EntryBlockMarker {
	b := new(EntryBlockMarker)
	b.KeyMr = e.KeyMr.Copy()
	b.Sequence = e.Sequence
	b.DBHeight = e.DBHeight
	b.DblockTimestamp = e.DblockTimestamp
	return b
}

// MarshalBinary marshals this object
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

// Size returns the byte size when marshaled
func (e *EntryBlockMarker) Size() int {
	// If you count it, it's 46. However, PushIHash is actually 33 bytes. and PushTimestamp is actually 8, rather than 6.
	return 49
}

// UnmarshalBinary unmarshals the input data into this object
func (e *EntryBlockMarker) UnmarshalBinary(p []byte) error {
	_, err := e.UnmarshalBinaryData(p)
	return err
}

// UnmarshalBinaryData unmarshals the input data into this object
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
