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

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
	llog "github.com/FactomProject/factomd/log"
)

// An Entry is the element which carries user data to be stored in the blockchain
// https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#entry
type Entry struct {
	Version uint8                  `json:"version"` // The entry version, only currently supported number is 0
	ChainID interfaces.IHash       `json:"chainid"` // The chain id associated with this entry
	ExtIDs  []primitives.ByteSlice `json:"extids"`  // External ids used to create the chain id above ( see ExternalIDsToChainID() )
	Content primitives.ByteSlice   `json:"content"` // BytesSlice for holding generic data for this entry

	// cache
	hash interfaces.IHash
}

var _ interfaces.IEBEntry = (*Entry)(nil)
var _ interfaces.DatabaseBatchable = (*Entry)(nil)
var _ interfaces.BinaryMarshallable = (*Entry)(nil)

// RandomEntry produces an entry object with randomly initialized values
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

// DeterministicEntry creates a new entry deterministically based on the input integer
func DeterministicEntry(i int) interfaces.IEBEntry {
	e := NewEntry()
	e.Version = 0
	bs := fmt.Sprintf("%x", i)
	if len(bs)%2 == 1 {
		bs = "0" + bs
	}

	e.ExtIDs = []primitives.ByteSlice{*primitives.StringToByteSlice(bs)}
	//e.ExtIDs = append(e.ExtIDs, *primitives.StringToByteSlice(fmt.Sprintf("%d", i)))
	e.ChainID = ExternalIDsToChainID([][]byte{e.ExtIDs[0].Bytes})

	return e
}

// IsSameAs returns true iff the input entry is identical to this entry
func (c *Entry) IsSameAs(b interfaces.IEBEntry) bool {
	if b == nil {
		if c != nil {
			return false
		}
		return true
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

// KSize returns the size of the entry subject to payment in K.  So anything up
// to 1K returns 1, everything up to and including 2K returns 2, etc.
// An error returns 100 (an invalid size)
func (c *Entry) KSize() int {
	data, err := c.MarshalBinary()
	if err != nil {
		return 100
	}
	return (len(data) - 35 + 1023) / 1024
}

// New creates a new entry
func (c *Entry) New() interfaces.BinaryMarshallableAndCopyable {
	return NewEntry()
}

// GetDatabaseHeight always returns 0
func (c *Entry) GetDatabaseHeight() uint32 {
	return 0
}

// GetWeld returns the double sha of the entry's hash and chain id appended ('welded') together
func (c *Entry) GetWeld() []byte {
	return primitives.DoubleSha(append(c.GetHash().Bytes(), c.GetChainID().Bytes()...))
}

// GetWeldHash returns the doble sha of the entry's hash and chain id appended ('welded') together
func (c *Entry) GetWeldHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "Entry.GetWeldHash") }()

	hash := primitives.NewZeroHash()
	hash.SetBytes(c.GetWeld())
	return hash
}

// GetChainID returns the chain id of this entry
func (c *Entry) GetChainID() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "Entry.GetChainID") }()

	return c.ChainID
}

// DatabasePrimaryIndex returns the hash of the entry object
func (c *Entry) DatabasePrimaryIndex() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "Entry.DatabasePrimaryIndex") }()

	return c.GetHash()
}

// DatabaseSecondaryIndex always returns nil (ie, no secondary index)
func (c *Entry) DatabaseSecondaryIndex() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "Entry.DatabaseSecondaryIndex") }()

	return nil
}

// NewChainID generates a ChainID from an entry. ChainID = primitives.Sha(Sha(ExtIDs[0]) +
// Sha(ExtIDs[1] + ... + Sha(ExtIDs[n]))
func NewChainID(e interfaces.IEBEntry) interfaces.IHash {
	return ExternalIDsToChainID(e.ExternalIDs())
}

// ExternalIDsToChainID converts the input external ids into a chain id. ChainID = primitives.Sha(Sha(ExtIDs[0]) +
// Sha(ExtIDs[1] + ... + Sha(ExtIDs[n]))
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

// GetContent returns the content
func (c *Entry) GetContent() []byte {
	return c.Content.Bytes
}

// GetChainIDHash returns the chain id associated with this entry
func (c *Entry) GetChainIDHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "Entry.GetChainIDHash") }()

	return c.ChainID
}

// ExternalIDs returns an array of the external ids
func (c *Entry) ExternalIDs() [][]byte {
	answer := [][]byte{}
	for _, v := range c.ExtIDs {
		answer = append(answer, v.Bytes)
	}
	return answer
}

// IsValid checks whether an entry is considered valid. A valid entry must satisfy both conditions:
// 1) have version==0 AND
// 2) KSize() <= 10
func (c *Entry) IsValid() bool {
	//double check the version
	if c.Version != 0 {
		return false
	}

	if c.KSize() > 10 {
		return false
	}

	return true
}

// GetHash returns the hash of the entry: sha256(append(sha512(data),data))
func (c *Entry) GetHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "Entry.GetHash") }()

	if c.hash == nil || c.hash.PFixed() == nil {
		h := primitives.NewZeroHash()
		entry, err := c.MarshalBinary()
		if err != nil {
			fmt.Println("Failed to Marshal Entry", c.String())
			return nil
		}

		h1 := sha512.Sum512(entry)
		h2 := sha256.Sum256(append(h1[:], entry[:]...))
		h.SetBytes(h2[:])
		c.hash = h
	}
	return c.hash
}

// MarshalBinary marshals the object
func (c *Entry) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "Entry.MarshalBinary err:%v", *pe)
		}
	}(&err)
	buf := primitives.NewBuffer(nil)

	// 1 byte Version
	err = buf.PushByte(byte(c.Version))
	if err != nil {
		return nil, err
	}

	if c.ChainID == nil {
		c.ChainID = primitives.NewZeroHash()
	}
	// 32 byte ChainID
	err = buf.PushBinaryMarshallable(c.ChainID)
	if err != nil {
		return nil, err
	}

	// ExtIDs
	ext, err := c.MarshalExtIDsBinary()
	if err != nil {
		return nil, err
	}

	// 2 byte size of ExtIDs
	if err := binary.Write(buf, binary.BigEndian, int16(len(ext))); err != nil {
		return nil, err
	}

	// binary ExtIDs
	buf.Write(ext)

	// Content
	buf.Write(c.Content.Bytes)

	return buf.Bytes(), nil
}

// MarshalExtIDsBinary marshals the ExtIDs into a []byte containing a series of
// 2 byte size of each ExtID followed by the ExtID.
func (c *Entry) MarshalExtIDsBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "Entry.MarshalExtIDsBinary err:%v", *pe)
		}
	}(&err)
	buf := new(primitives.Buffer)

	for _, x := range c.ExtIDs {
		// 2 byte size of the ExtID
		if err := binary.Write(buf, binary.BigEndian, uint16(len(x.Bytes))); err != nil {
			return nil, err
		}

		// ExtID bytes
		buf.Write(x.Bytes)
	}

	return buf.DeepCopyBytes(), nil
}

// UnmarshalEntry unmarshals the input data into a new entry
func UnmarshalEntry(data []byte) (interfaces.IEBEntry, error) {
	entry := NewEntry()
	err := entry.UnmarshalBinary(data)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

// UnmarshalBinaryData unmarshals the input data into this object
func (c *Entry) UnmarshalBinaryData(data []byte) (_ []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
			llog.LogPrintf("recovery", "Error unmarshalling: %v", r)
		}
	}()

	buf := primitives.NewBuffer(data)

	// 1 byte Version
	c.Version, err = buf.PopByte()
	if err != nil {
		return nil, err
	}

	// 32 byte ChainID
	c.ChainID = primitives.NewZeroHash()
	err = buf.PopBinaryMarshallable(c.ChainID)
	if err != nil {
		return nil, err
	}

	// 2 byte size of ExtIDs
	var extSize uint16
	if err = binary.Read(buf, binary.BigEndian, &extSize); err != nil {
		return nil, err
	}

	// ExtIDs
	for i := int(extSize); i > 0; {
		var xsize int16
		binary.Read(buf, binary.BigEndian, &xsize)
		i -= 2
		if i < 0 {
			err = fmt.Errorf("Error parsing external IDs")
			return nil, err
		}
		// check that the payload size is not too big before we allocate the
		// buffer. Max payload size is 10KB
		if xsize > constants.MaxEntrySizeInBytes {
			return nil, fmt.Errorf(
				"Error: entry.UnmarshalBinary: ExtIDs size %d too high (uint "+
					" underflow?)",
				xsize,
			)

		}
		x := make([]byte, xsize)
		if n, err := buf.Read(x); err != nil {
			return nil, err
		} else {
			if cp := cap(x); n != cp {
				err = fmt.Errorf("Could not read ExtID: Read %d bytes of %d\n", n, cp)
				return nil, err
			}
			ex := primitives.ByteSlice{}
			err = ex.UnmarshalBinary(x)
			if err != nil {
				return nil, err
			}
			c.ExtIDs = append(c.ExtIDs, ex)
			i -= n
			if i < 0 {
				err = fmt.Errorf("Error parsing external IDs")
				return nil, err
			}
		}
	}

	// Content
	err = c.Content.UnmarshalBinary(buf.Bytes())
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// UnmarshalBinary unmarshals the input data into this object
func (c *Entry) UnmarshalBinary(data []byte) (err error) {
	_, err = c.UnmarshalBinaryData(data)
	return
}

// JSONByte returns the json encoded byte array
func (c *Entry) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(c)
}

// JSONString returns the json encoded string
func (c *Entry) JSONString() (string, error) {
	return primitives.EncodeJSONString(c)
}

// String returns this object as a string
func (c *Entry) String() string {
	str, _ := c.JSONString()
	return str
}

/***************************************************************
 * Helper Functions
 ***************************************************************/

// NewEntry returns a new entry initialized with empty interfaces and zero hashes
func NewEntry() *Entry {
	e := new(Entry)
	e.ChainID = primitives.NewZeroHash()
	e.ExtIDs = make([]primitives.ByteSlice, 0)
	e.Content = primitives.ByteSlice{}
	return e
}

// MarshalEntryList marshals the input list into a single byte array
func MarshalEntryList(list []interfaces.IEBEntry) ([]byte, error) {
	buf := primitives.NewBuffer(nil)
	l := len(list)
	buf.PushVarInt(uint64(l))
	for _, v := range list {
		bin, err := v.MarshalBinary()
		if err != nil {
			return nil, err
		}
		err = buf.PushBytes(bin)
		if err != nil {
			return nil, err
		}
	}
	return buf.DeepCopyBytes(), nil
}

// UnmarshalEntryList unmarshals the input byte array into a list of entries
func UnmarshalEntryList(data []byte) ([]interfaces.IEBEntry, []byte, error) {
	buf := primitives.NewBuffer(data)

	entryLimit := uint64(buf.Len())
	entryCount, err := buf.PopVarInt()
	if err != nil {
		return nil, nil, err
	}
	if entryCount > entryLimit {
		return nil, nil, fmt.Errorf(
			"Error: UnmarshalEntryList: entry count %d higher than space in "+
				"body %d (uint underflow?)",
			entryCount, entryLimit,
		)
	}

	list := make([]interfaces.IEBEntry, int(entryCount))

	for i := range list {
		e := NewEntry()
		x, err := buf.PopBytes()
		if err != nil {
			return nil, nil, err
		}
		err = e.UnmarshalBinary(x)
		if err != nil {
			return nil, nil, err
		}
		list[i] = e
	}

	return list, buf.DeepCopyBytes(), nil
}

func (e *Entry) GetVersion() uint8 {
	return e.Version
}

func (e *Entry) GetExtIDs() []primitives.ByteSlice {
	return e.ExtIDs
}
