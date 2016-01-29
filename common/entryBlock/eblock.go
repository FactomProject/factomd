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
var _ interfaces.DatabaseBlockWithEntries = (*EBlock)(nil)

func (c *EBlock) GetEntryHashes() []interfaces.IHash {
	return c.Body.EBEntries[:]
}

func (c *EBlock) New() interfaces.BinaryMarshallableAndCopyable {
	return NewEBlock()
}

func (e *EBlock) GetWelds() [][]byte {
	var answer [][]byte
	for _, entry := range e.Body.EBEntries {
		answer = append(answer, primitives.DoubleSha(append(entry.Bytes(), e.GetChainID().Bytes()...)))
	}
	return answer
}

func (e *EBlock) GetWeldHashes() []interfaces.IHash {
	var answer []interfaces.IHash
	for _, h := range e.GetWelds() {
		hash := primitives.NewZeroHash()
		hash.SetBytes(h)
		answer = append(answer, hash)
	}
	return answer
}

func (c *EBlock) GetDatabaseHeight() uint32 {
	return c.Header.GetDBHeight()
}

func (c *EBlock) GetChainID() interfaces.IHash {
	return c.Header.GetChainID()
}

func (c *EBlock) GetHashOfChainID() []byte {
	return primitives.DoubleSha(c.GetChainID().Bytes())
}

func (c *EBlock) GetHashOfChainIDHash() interfaces.IHash {
	hash := primitives.NewZeroHash()
	hash.SetBytes(c.GetHashOfChainID())
	return hash
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

func (c *EBlock) GetBody() interfaces.IEBlockBody {
	return c.Body
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

func (e *EBlock) HeaderHash() (interfaces.IHash, error) {
	e.BuildHeader()
	header, err := e.Header.MarshalBinary()
	if err != nil {
		return nil, err
	}
	h := primitives.Sha(header)
	return h, nil
}

func (e *EBlock) BodyKeyMR() interfaces.IHash {
	e.BuildHeader()
	return e.Header.GetBodyMR()
}

// KeyMR returns the hash of the hash of the Entry Block Header concatinated
// with the Merkle Root of the Entry Block Body. The Body Merkle Root is
// calculated by the func (e *EBlockBody) MR() which is called by the func
// (e *EBlock) BuildHeader().
func (e *EBlock) KeyMR() (interfaces.IHash, error) {
	// Sha(Sha(header) + BodyMR)
	h, err := e.HeaderHash()
	if err != nil {
		return nil, err
	}
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

/*****************************************************
 * Support Routines
 *****************************************************/

// NewEBlock returns a blank initialized Entry Block with all of its fields
// zeroed.
func NewEBlock() *EBlock {
	e := new(EBlock)
	e.Header = NewEBlockHeader()
	e.Body = NewEBlockBody()
	return e
}
