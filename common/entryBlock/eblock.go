// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryBlock

import (
	"fmt"
	"os"
	"reflect"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// EBlock is the Entry Block. It holds the hashes of the Entries and its Merkle
// Root is written into the Directory Blocks. Each Entry Block represents all
// of the entries for a particular Chain during a 10 minute period.
type EBlock struct {
	Header interfaces.IEntryBlockHeader `json:"header"`
	Body   *EBlockBody                  `json:"body"`
}

var _ interfaces.Printable = (*EBlock)(nil)
var _ interfaces.BinaryMarshallableAndCopyable = (*EBlock)(nil)
var _ interfaces.BinaryMarshallable = (*EBlock)(nil)
var _ interfaces.DatabaseBatchable = (*EBlock)(nil)
var _ interfaces.IEntryBlock = (*EBlock)(nil)
var _ interfaces.DatabaseBlockWithEntries = (*EBlock)(nil)

func (c *EBlock) Init() {
	if c.Header == nil {
		h := new(EBlockHeader)
		h.Init()
		c.Header = h
	}
	if c.Body == nil {
		c.Body = new(EBlockBody)
	}
}

func (a *EBlock) IsSameAs(b interfaces.IEntryBlock) bool {
	if a == nil || b == nil {
		if a == nil && b == nil {
			return true
		}
		return false
	}

	if a.Header.IsSameAs(b.GetHeader()) == false {
		return false
	}
	if a.Body.IsSameAs(b.GetBody()) == false {
		return false
	}

	return true
}

func (c *EBlock) GetEntryHashes() []interfaces.IHash {
	return c.GetBody().GetEBEntries()
}

func (c *EBlock) GetEntrySigHashes() []interfaces.IHash {
	return nil
}

func (c *EBlock) New() interfaces.BinaryMarshallableAndCopyable {
	return NewEBlock()
}

func (c *EBlock) GetDatabaseHeight() uint32 {
	return c.GetHeader().GetDBHeight()
}

func (c *EBlock) GetChainID() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("EBlock.GetChainID() saw an interface that was nil")
		}
	}()

	return c.GetHeader().GetChainID()
}

func (c *EBlock) GetHashOfChainID() []byte {
	return primitives.DoubleSha(c.GetChainID().Bytes())
}

func (c *EBlock) GetHashOfChainIDHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("EBlock.GetHashOfChainIDHash() saw an interface that was nil")
		}
	}()

	hash := primitives.NewZeroHash()
	hash.SetBytes(c.GetHashOfChainID())
	return hash
}

func (c *EBlock) DatabasePrimaryIndex() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("EBlock.DatabasePrimaryIndex() saw an interface that was nil")
		}
	}()

	key, _ := c.KeyMR()
	return key
}

func (c *EBlock) DatabaseSecondaryIndex() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("EBlock.DatabaseSecondaryIndex() saw an interface that was nil")
		}
	}()

	h, _ := c.Hash()
	return h
}

func (c *EBlock) GetHeader() interfaces.IEntryBlockHeader {
	c.Init()
	return c.Header
}

func (c *EBlock) GetBody() interfaces.IEBlockBody {
	c.Init()
	return c.Body
}

// AddEBEntry creates a new Entry Block Entry from the provided Factom Entry
// and adds it to the Entry Block Body.
func (e *EBlock) AddEBEntry(entry interfaces.IEBEntry) error {
	e.Init()
	e.GetBody().AddEBEntry(entry.GetHash())
	if err := e.BuildHeader(); err != nil {
		return err
	}
	return nil
}

// AddEndOfMinuteMarker adds the End of Minute to the Entry Block. The End of
// Minut byte becomes the last byte in a 32 byte slice that is added to the
// Entry Block Body as an Entry Block Entry.
func (e *EBlock) AddEndOfMinuteMarker(m byte) error {
	e.Init()
	e.GetBody().AddEndOfMinuteMarker(m)
	if err := e.BuildHeader(); err != nil {
		return err
	}
	return nil
}

// BuildHeader updates the Entry Block Header to include information about the
// Entry Block Body. BuildHeader should be run after the Entry Block Body has
// included all of its EntryEntries.
func (e *EBlock) BuildHeader() error {
	e.GetHeader().SetBodyMR(e.GetBody().MR())
	e.GetHeader().SetEntryCount(uint32(len(e.GetEntryHashes())))
	return nil
}

func (e *EBlock) GetHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("EBlock.GetHash() saw an interface that was nil")
		}
	}()

	h, _ := e.Hash()
	return h
}

// Hash returns the simple Sha256 hash of the serialized Entry Block. Hash is
// used to provide the PrevFullHash to the next Entry Block in a Chain.
func (e *EBlock) Hash() (interfaces.IHash, error) {
	p, err := e.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return primitives.Sha(p), nil
}

func (e *EBlock) HeaderHash() (interfaces.IHash, error) {
	e.BuildHeader()
	header, err := e.GetHeader().MarshalBinary()
	if err != nil {
		return nil, err
	}
	h := primitives.Sha(header)
	return h, nil
}

func (e *EBlock) BodyKeyMR() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("EBlock.BodyKeyMR() saw an interface that was nil")
		}
	}()

	e.BuildHeader()
	return e.GetHeader().GetBodyMR()
}

// KeyMR returns the hash of the hash of the Entry Block Header concatenated
// with the Merkle Root of the Entry Block Body. The Body Merkle Root is
// calculated by the func (e *EBlockBody) MR() which is called by the func
// (e *EBlock) BuildHeader().
func (e *EBlock) KeyMR() (interfaces.IHash, error) {
	// Sha(Sha(header) + BodyMR)
	e.BuildHeader()
	h, err := e.HeaderHash()
	if err != nil {
		return nil, err
	}
	return primitives.Sha(append(h.Bytes(), e.GetHeader().GetBodyMR().Bytes()...)), nil
}

// MarshalBinary returns the serialized binary form of the Entry Block.
func (e *EBlock) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "EBlock.MarshalBinary err:%v", *pe)
		}
	}(&err)
	e.Init()
	buf := primitives.NewBuffer(nil)

	err = e.BuildHeader()
	if err != nil {
		return nil, err
	}

	err = buf.PushBinaryMarshallable(e.GetHeader())
	if err != nil {
		return nil, err
	}

	if p, err := e.marshalBodyBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(p)
	}

	return buf.DeepCopyBytes(), nil
}

func UnmarshalEBlock(data []byte) (interfaces.IEntryBlock, error) {
	block, _, err := UnmarshalEBlockData(data)
	return block, err
}

func UnmarshalEBlockData(data []byte) (interfaces.IEntryBlock, []byte, error) {
	block := NewEBlock()

	data, err := block.UnmarshalBinaryData(data)
	if err != nil {
		return nil, nil, err
	}

	return block, data, nil
}

// UnmarshalBinary populates the Entry Block object from the serialized binary
// data.
func (e *EBlock) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	newData = data

	if e.Header == nil {
		e.Header = new(EBlockHeader)
	}

	newData, err = e.GetHeader().UnmarshalBinaryData(newData)
	if err != nil {
		return
	}

	newData, err = e.unmarshalBodyBinaryData(newData)
	if err != nil {
		return
	}

	return
}

func (e *EBlock) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

// marshalBodyBinary returns a serialized binary Entry Block Body
func (e *EBlock) marshalBodyBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "EBlock.marshalBodyBinary err:%v", *pe)
		}
	}(&err)
	e.Init()
	buf := new(primitives.Buffer)

	for _, v := range e.GetEntryHashes() {
		buf.Write(v.Bytes())
	}

	return buf.DeepCopyBytes(), nil
}

// unmarshalBodyBinary builds the Entry Block Body from the serialized binary.
func (e *EBlock) unmarshalBodyBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	e.Init()

	buf := primitives.NewBuffer(data)
	hash := make([]byte, 32)

	for i := uint32(0); i < e.GetHeader().GetEntryCount(); i++ {
		if _, err := buf.Read(hash); err != nil {
			return nil, err
		}

		h := primitives.NewZeroHash()
		h.SetBytes(hash)
		e.GetBody().AddEBEntry(h)
	}

	newData = buf.DeepCopyBytes()
	return
}

func (e *EBlock) unmarshalBodyBinary(data []byte) (err error) {
	_, err = e.unmarshalBodyBinaryData(data)
	return
}

func (e *EBlock) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *EBlock) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *EBlock) String() string {
	e.Init()
	str := e.GetHeader().String()
	str = str + e.GetBody().String()
	return str
}

/*****************************************************
 * Support Routines
 *****************************************************/

// NewEBlock returns a blank initialized Entry Block with all fields zeroed.
func NewEBlock() *EBlock {
	e := new(EBlock)
	e.Header = NewEBlockHeader()
	e.Body = NewEBlockBody()
	return e
}
