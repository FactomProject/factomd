// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryBlock

import (
	"bytes"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

const (
	EBHeaderSize = 140 // 32+32+32+32+4+4+4
)

// EBlock is the Entry Block. It holds the hashes of the Entries and its Merkel
// Root is written into the Directory Blocks. Each Entry Block represents all
// of the entries for a paticular Chain during a 10 minute period.
type EBlock struct {
	Header interfaces.IEntryBlockHeader
	Body   *EBlockBody
}

var _ interfaces.Printable = (*EBlock)(nil)
var _ interfaces.BinaryMarshallableAndCopyable = (*EBlock)(nil)
var _ interfaces.DatabaseBatchable = (*EBlock)(nil)
var _ interfaces.IEntryBlock = (*EBlock)(nil)

func (c *EBlock) New() interfaces.BinaryMarshallableAndCopyable {
	return NewEBlock()
}

func (c *EBlock) GetDatabaseHeight() uint32 {
	return c.Header.GetDBHeight()
}

func (c *EBlock) GetChainID() []byte {
	return c.Header.GetChainID().Bytes()
}

func (c *EBlock) DatabasePrimaryIndex() interfaces.IHash {
	key, _ := c.KeyMR()
	return key
}

func (c *EBlock) DatabaseSecondaryIndex() interfaces.IHash {
	h, _ := c.Hash()
	return h
}

func (c *EBlock) MarshalledSize() uint64 {
	return uint64(EBHeaderSize)
}

func (c *EBlock) GetHeader() interfaces.IEntryBlockHeader {
	return c.Header
}

// NewEBlock returns a blank initialized Entry Block with all of its fields
// zeroed.
func NewEBlock() *EBlock {
	e := new(EBlock)
	e.Header = NewEBlockHeader()
	e.Body = NewEBlockBody()
	return e
}

// AddEBEntry creates a new Entry Block Entry from the provided Factom Entry
// and adds it to the Entry Block Body.
func (e *EBlock) AddEBEntry(entry interfaces.IEBEntry) error {
	e.Body.EBEntries = append(e.Body.EBEntries, entry.GetHash())
	return nil
}

// AddEndOfMinuteMarker adds the End of Minute to the Entry Block. The End of
// Minut byte becomes the last byte in a 32 byte slice that is added to the
// Entry Block Body as an Entry Block Entry.
func (e *EBlock) AddEndOfMinuteMarker(m byte) {
	h := make([]byte, 32)
	h[len(h)-1] = m
	hash := primitives.NewZeroHash()
	hash.SetBytes(h)
	e.Body.EBEntries = append(e.Body.EBEntries, hash)
}

// BuildHeader updates the Entry Block Header to include information about the
// Entry Block Body. BuildHeader should be run after the Entry Block Body has
// included all of its EntryEntries.
func (e *EBlock) BuildHeader() error {
	e.Header.SetBodyMR(e.Body.MR())
	e.Header.SetEntryCount(uint32(len(e.Body.EBEntries)))
	return nil
}

// Hash returns the simple Sha256 hash of the serialized Entry Block. Hash is
// used to provide the PrevLedgerKeyMR to the next Entry Block in a Chain.
func (e *EBlock) Hash() (interfaces.IHash, error) {
	p, err := e.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return primitives.Sha(p), nil
}

// KeyMR returns the hash of the hash of the Entry Block Header concatinated
// with the Merkle Root of the Entry Block Body. The Body Merkle Root is
// calculated by the func (e *EBlockBody) MR() which is called by the func
// (e *EBlock) BuildHeader().
func (e *EBlock) KeyMR() (interfaces.IHash, error) {
	// Sha(Sha(header) + BodyMR)
	e.BuildHeader()
	header, err := e.Header.MarshalBinary()
	if err != nil {
		return nil, err
	}
	h := primitives.Sha(header)
	return primitives.Sha(append(h.Bytes(), e.Header.GetBodyMR().Bytes()...)), nil
}

// MarshalBinary returns the serialized binary form of the Entry Block.
func (e *EBlock) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := e.BuildHeader(); err != nil {
		return buf.Bytes(), err
	}
	if p, err := e.Header.MarshalBinary(); err != nil {
		return buf.Bytes(), err
	} else {
		buf.Write(p)
	}

	if p, err := e.marshalBodyBinary(); err != nil {
		return buf.Bytes(), err
	} else {
		buf.Write(p)
	}

	return buf.Bytes(), nil
}

// UnmarshalBinary populates the Entry Block object from the serialized binary
// data.
func (e *EBlock) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	newData = data

	newData, err = e.Header.UnmarshalBinaryData(newData)
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
func (e *EBlock) marshalBodyBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	for _, v := range e.Body.EBEntries {
		buf.Write(v.Bytes())
	}

	return buf.Bytes(), nil
}

// unmarshalBodyBinary builds the Entry Block Body from the serialized binary.
func (e *EBlock) unmarshalBodyBinaryData(data []byte) (newData []byte, err error) {
	buf := bytes.NewBuffer(data)
	hash := make([]byte, 32)

	for i := uint32(0); i < e.Header.GetEntryCount(); i++ {
		if _, err = buf.Read(hash); err != nil {
			return buf.Bytes(), err
		}

		h := primitives.NewZeroHash()
		h.SetBytes(hash)
		e.Body.EBEntries = append(e.Body.EBEntries, h)
	}

	newData = buf.Bytes()
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

func (e *EBlock) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *EBlock) String() string {
	str, _ := e.JSONString()
	return str
}

// EBlockBody is the series of Hashes that form the Entry Block Body.
type EBlockBody struct {
	EBEntries []interfaces.IHash
}

var _ interfaces.Printable = (*EBlockBody)(nil)

// NewEBlockBody initalizes an empty Entry Block Body.
func NewEBlockBody() *EBlockBody {
	e := new(EBlockBody)
	e.EBEntries = make([]interfaces.IHash, 0)
	return e
}

// MR calculates the Merkle Root of the Entry Block Body. See func
// primitives.BuildMerkleTreeStore(hashes []interfaces.IHash) (merkles []interfaces.IHash) in common/merkle.go.
func (e *EBlockBody) MR() interfaces.IHash {
	mrs := primitives.BuildMerkleTreeStore(e.EBEntries)
	r := mrs[len(mrs)-1]
	return r
}

func (e *EBlockBody) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *EBlockBody) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *EBlockBody) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *EBlockBody) String() string {
	str, _ := e.JSONString()
	return str
}
