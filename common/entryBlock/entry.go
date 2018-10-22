// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryBlock

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
)

// An Entry is the element which carries user data
// https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#entry
type Entry struct {
	Version uint8                  `json:"version"`
	ChainID interfaces.IHash       `json:"chainid"`
	ExtIDs  []primitives.ByteSlice `json:"extids"`
	Content primitives.ByteSlice   `json:"content"`

	// cache
	hash interfaces.IHash
}

var _ interfaces.IEBEntry = (*Entry)(nil)
var _ interfaces.DatabaseBatchable = (*Entry)(nil)
var _ interfaces.BinaryMarshallable = (*Entry)(nil)

func RandomEntry() interfaces.IEBEntry {
	e := NewEntry()
	e.Version = random.RandUInt8()
	e.ChainID = primitives.RandomHash()
	l := random.RandIntBetween(0, 20)
	for i := 0; i < l; i++ {
		e.ExtIDs = append(e.ExtIDs, *primitives.RandomByteSlice())
	}
	e.Content = *primitives.RandomByteSlice()
	return e
}

func (c *Entry) IsSameAs(b interfaces.IEBEntry) bool {
	if b == nil {
		if c != nil {
			return false
		} else {
			return true
		}
	}
	a := b.(*Entry)
	if c.Version != a.Version {
		return false
	}
	if len(c.ExtIDs) != len(a.ExtIDs) {
		return false
	}
	for i := range c.ExtIDs {
		if c.ExtIDs[i].IsSameAs(&a.ExtIDs[i]) == false {
			return false
		}
	}
	if c.ChainID.IsSameAs(a.ChainID) == false {
		return false
	}
	if c.Content.IsSameAs(&a.Content) == false {
		return false
	}
	return true
}

// Returns the size of the entry subject to payment in K.  So anything up
// to 1K returns 1, everything up to and including 2K returns 2, etc.
// An error returns 100 (an invalid size)
func (c *Entry) KSize() int {
	data, err := c.MarshalBinary()
	if err != nil {
		return 100
	}
	return (len(data) - 35 + 1023) / 1024
}

func (c *Entry) New() interfaces.BinaryMarshallableAndCopyable {
	return NewEntry()
}

func (c *Entry) GetDatabaseHeight() uint32 {
	return 0
}

func (e *Entry) GetWeld() []byte {
	return primitives.DoubleSha(append(e.GetHash().Bytes(), e.GetChainID().Bytes()...))
}

func (e *Entry) GetWeldHash() interfaces.IHash {
	hash := primitives.NewZeroHash()
	hash.SetBytes(e.GetWeld())
	return hash
}

func (c *Entry) GetChainID() interfaces.IHash {
	return c.ChainID
}

func (c *Entry) DatabasePrimaryIndex() interfaces.IHash {
	return c.GetHash()
}

func (c *Entry) DatabaseSecondaryIndex() interfaces.IHash {
	return nil
}

// NewChainID generates a ChainID from an entry. ChainID = primitives.Sha(Sha(ExtIDs[0]) +
// Sha(ExtIDs[1] + ... + Sha(ExtIDs[n]))
func NewChainID(e interfaces.IEBEntry) interfaces.IHash {
	return ExternalIDsToChainID(e.ExternalIDs())
}

func ExternalIDsToChainID(extIDs [][]byte) interfaces.IHash {
	id := new(primitives.Hash)
	sum := sha256.New()
	for _, v := range extIDs {
		x := sha256.Sum256(v)
		sum.Write(x[:])
	}
	id.SetBytes(sum.Sum(nil))

	return id
}

func (e *Entry) GetContent() []byte {
	return e.Content.Bytes
}

func (e *Entry) GetChainIDHash() interfaces.IHash {
	return e.ChainID
}

func (e *Entry) ExternalIDs() [][]byte {
	answer := [][]byte{}
	for _, v := range e.ExtIDs {
		answer = append(answer, v.Bytes)
	}
	return answer
}

func (e *Entry) IsValid() bool {
	//double check the version
	if e.Version != 0 {
		return false
	}

	if e.KSize() > 10 {
		return false
	}

	return true
}

func (e *Entry) GetHash() interfaces.IHash {
	if e.hash == nil || e.hash.PFixed() == nil {
		h := primitives.NewZeroHash()
		entry, err := e.MarshalBinary()
		if err != nil {
			fmt.Println("Failed to Marshal Entry", e.String())
			return nil
		}

		h1 := sha512.Sum512(entry)
		h2 := sha256.Sum256(append(h1[:], entry[:]...))
		h.SetBytes(h2[:])
		e.hash = h
	}
	return e.hash
}

func (e *Entry) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "Entry.MarshalBinary err:%v", *pe)
		}
	}(&err)
	buf := primitives.NewBuffer(nil)

	// 1 byte Version
	err = buf.PushByte(byte(e.Version))
	if err != nil {
		return nil, err
	}

	if e.ChainID == nil {
		e.ChainID = primitives.NewZeroHash()
	}
	// 32 byte ChainID
	err = buf.PushBinaryMarshallable(e.ChainID)
	if err != nil {
		return nil, err
	}

	// ExtIDs
	if ext, err := e.MarshalExtIDsBinary(); err != nil {
		return nil, err
	} else {
		// 2 byte size of ExtIDs
		if err := binary.Write(buf, binary.BigEndian, int16(len(ext))); err != nil {
			return nil, err
		}

		// binary ExtIDs
		buf.Write(ext)
	}

	// Content
	buf.Write(e.Content.Bytes)

	return buf.Bytes(), nil
}

// MarshalExtIDsBinary marshals the ExtIDs into a []byte containing a series of
// 2 byte size of each ExtID followed by the ExtID.
func (e *Entry) MarshalExtIDsBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "Entry.MarshalExtIDsBinary err:%v", *pe)
		}
	}(&err)
	buf := new(primitives.Buffer)

	for _, x := range e.ExtIDs {
		// 2 byte size of the ExtID
		if err := binary.Write(buf, binary.BigEndian, uint16(len(x.Bytes))); err != nil {
			return nil, err
		}

		// ExtID bytes
		buf.Write(x.Bytes)
	}

	return buf.DeepCopyBytes(), nil
}

func UnmarshalEntry(data []byte) (interfaces.IEBEntry, error) {
	entry := NewEntry()
	err := entry.UnmarshalBinary(data)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func (e *Entry) UnmarshalBinaryData(data []byte) ([]byte, error) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()

	buf := primitives.NewBuffer(data)

	// 1 byte Version
	e.Version, err = buf.PopByte()
	if err != nil {
		return nil, err
	}

	// 32 byte ChainID
	e.ChainID = primitives.NewZeroHash()
	err = buf.PopBinaryMarshallable(e.ChainID)
	if err != nil {
		return nil, err
	}

	// 2 byte size of ExtIDs
	var extSize uint16
	if err = binary.Read(buf, binary.BigEndian, &extSize); err != nil {
		return nil, err
	}

	// ExtIDs
	for i := int16(extSize); i > 0; {
		var xsize int16
		binary.Read(buf, binary.BigEndian, &xsize)
		i -= 2
		if i < 0 {
			err = fmt.Errorf("Error parsing external IDs")
			return nil, err
		}

		// check that the payload size is not too big before we allocate the
		// buffer.
		if xsize > 10240 {
			// TODO: replace this message with a proper error
			return nil, fmt.Errorf("Error: entry.UnmarshalBinary: ExtIDs size too high (uint underflow?)")
		}
		x := make([]byte, xsize)
		if n, err := buf.Read(x); err != nil {
			return nil, err
		} else {
			if c := cap(x); n != c {
				err = fmt.Errorf("Could not read ExtID: Read %d bytes of %d\n", n, c)
				return nil, err
			}
			ex := primitives.ByteSlice{}
			err = ex.UnmarshalBinary(x)
			if err != nil {
				return nil, err
			}
			e.ExtIDs = append(e.ExtIDs, ex)
			i -= int16(n)
			if i < 0 {
				err = fmt.Errorf("Error parsing external IDs")
				return nil, err
			}
		}
	}

	// Content
	err = e.Content.UnmarshalBinary(buf.Bytes())
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (e *Entry) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

func (e *Entry) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *Entry) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *Entry) String() string {
	str, _ := e.JSONString()
	return str
}

/***************************************************************
 * Helper Functions
 ***************************************************************/

func NewEntry() *Entry {
	e := new(Entry)
	e.ChainID = primitives.NewZeroHash()
	e.ExtIDs = make([]primitives.ByteSlice, 0)
	e.Content = primitives.ByteSlice{}
	return e
}

func MarshalEntryList(list []interfaces.IEBEntry) ([]byte, error) {
	b := primitives.NewBuffer(nil)
	l := len(list)
	b.PushVarInt(uint64(l))
	for _, v := range list {
		bin, err := v.MarshalBinary()
		if err != nil {
			return nil, err
		}
		err = b.PushBytes(bin)
		if err != nil {
			return nil, err
		}
	}
	return b.DeepCopyBytes(), nil
}

func UnmarshalEntryList(bin []byte) ([]interfaces.IEBEntry, []byte, error) {
	b := primitives.NewBuffer(bin)

	l, err := b.PopVarInt()
	if err != nil {
		return nil, nil, err
	}
	e := int(l)
	// TODO: remove printing unmarshal count numbers once we have good data on
	// what they should be.
	//log.Print("UnmarshalEntryList unmarshaled entry count: ", e)
	if e > 1000 {
		// TODO: replace this message with a proper error
		return nil, nil, fmt.Errorf("Error: UnmarshalEntryList: entry count too high (uint underflow?)")
	}

	list := make([]interfaces.IEBEntry, e)
	for i := range list {
		e := NewEntry()
		x, err := b.PopBytes()
		if err != nil {
			return nil, nil, err
		}
		err = e.UnmarshalBinary(x)
		if err != nil {
			return nil, nil, err
		}
		list[i] = e
	}

	return list, b.DeepCopyBytes(), nil
}
