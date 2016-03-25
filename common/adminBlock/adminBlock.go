// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package adminBlock

import (
	"bytes"
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// Administrative Block
// This is a special block which accompanies this Directory Block.
// It contains the signatures and organizational data needed to validate previous and future Directory Blocks.
// This block is included in the DB body. It appears there with a pair of the Admin AdminChainID:SHA256 of the block.
// For more details, please go to:
// https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#administrative-block
type AdminBlock struct {
	//Marshalized
	Header    interfaces.IABlockHeader
	ABEntries []interfaces.IABEntry //Interface

	//Not Marshalized
	Full_Hash    interfaces.IHash //SHA512Half
	partialHash interfaces.IHash //SHA256
}

var _ interfaces.IAdminBlock = (*AdminBlock)(nil)
var _ interfaces.Printable = (*AdminBlock)(nil)
var _ interfaces.BinaryMarshallableAndCopyable = (*AdminBlock)(nil)
var _ interfaces.DatabaseBatchable = (*AdminBlock)(nil)

func (c *AdminBlock) GetHeader() interfaces.IABlockHeader {
	return c.Header
}

func (c *AdminBlock) SetHeader(header interfaces.IABlockHeader) {
	c.Header = header
}

func (c *AdminBlock) GetABEntries() []interfaces.IABEntry {
	return c.ABEntries
}

func (c *AdminBlock) GetDBHeight() uint32 {
	return c.Header.GetDBHeight()
}

func (c *AdminBlock) SetABEntries(abentries []interfaces.IABEntry) {
	c.ABEntries = abentries
}

func NewAdminBlock() interfaces.IAdminBlock {
	block := new(AdminBlock)
	block.Header = new(ABlockHeader)
	return block
}

func (c *AdminBlock) New() interfaces.BinaryMarshallableAndCopyable {
	return NewAdminBlock()
}

func (c *AdminBlock) GetDatabaseHeight() uint32 {
	return c.Header.GetDBHeight()
}

func (c *AdminBlock) GetChainID() interfaces.IHash {
	return c.Header.GetAdminChainID()
}

func (c *AdminBlock) DatabasePrimaryIndex() interfaces.IHash {
	key, _ := c.FullHash()
	return key
}

func (c *AdminBlock) DatabaseSecondaryIndex() interfaces.IHash {
	key, _ := c.PartialHash()
	return key
}

func (c *AdminBlock) GetHash() interfaces.IHash {
	h, _ := c.GetKeyMR()
	return h
}

func (c *AdminBlock) GetKeyMR() (interfaces.IHash, error) {
	return c.FullHash()
}

func (ab *AdminBlock) FullHash() (interfaces.IHash, error) {
	err := ab.BuildFullBHash()
	if err != nil {
		return nil, err
	}
	return ab.Full_Hash, nil
}

func (ab *AdminBlock) PartialHash() (interfaces.IHash, error) {
	if ab.partialHash == nil {
		err := ab.BuildPartialHash()
		if err != nil {
			return nil, err
		}
	}
	return ab.partialHash, nil
}

// Build the SHA512Half hash for the admin block
func (b *AdminBlock) BuildFullBHash() (err error) {
	var binaryAB []byte
	binaryAB, err = b.MarshalBinary()
	if err != nil {
		return
	}
	b.Full_Hash = primitives.Sha512Half(binaryAB)
	return
}

// Build the SHA256 hash for the admin block
func (b *AdminBlock) BuildPartialHash() (err error) {
	var binaryAB []byte
	binaryAB, err = b.MarshalBinary()
	if err != nil {
		return
	}
	b.partialHash = primitives.Sha(binaryAB)
	return
}

// Add an Admin Block entry to the block
func (b *AdminBlock) AddABEntry(e interfaces.IABEntry) (err error) {
	b.ABEntries = append(b.ABEntries, e)
	return
}

// Add the end-of-minute marker into the admin block
func (b *AdminBlock) AddEndOfMinuteMarker(eomType byte) (err error) {
	eOMEntry := &EndOfMinuteEntry{
		EntryType: constants.TYPE_MINUTE_NUM,
		EOM_Type:  eomType}

	b.AddABEntry(eOMEntry)

	return
}

// Write out the AdminBlock to binary.
func (b *AdminBlock) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	data, _ = b.Header.MarshalBinary()
	buf.Write(data)

	for _, v := range b.ABEntries {
		data, _ := v.MarshalBinary()
		buf.Write(data)
	}
	return buf.Bytes(), err
}

func (b *AdminBlock) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	newData = data
	h := new(ABlockHeader)
	newData, err = h.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}
	b.Header = h

	b.ABEntries = make([]interfaces.IABEntry, 0)
	for i := uint32(0); i < b.Header.GetMessageCount(); i++ {
		if newData[0] == constants.TYPE_DB_SIGNATURE {
			b.ABEntries[i] = new(DBSignatureEntry)
		} else if newData[0] == constants.TYPE_MINUTE_NUM {
			b.ABEntries[i] = new(EndOfMinuteEntry)
		}
		newData, err = b.ABEntries[i].UnmarshalBinaryData(newData)
		if err != nil {
			return
		}
	}
	return
}

// Read in the binary into the Admin block.
func (b *AdminBlock) UnmarshalBinary(data []byte) (err error) {
	_, err = b.UnmarshalBinaryData(data)
	return
}

// Read in the binary into the Admin block.
func (b *AdminBlock) GetDBSignature() interfaces.IABEntry {

	for i := uint32(0); i < b.Header.GetMessageCount(); i++ {
		if b.ABEntries[i].Type() == constants.TYPE_DB_SIGNATURE {
			return b.ABEntries[i]
		}
	}

	return nil
}

func (e *AdminBlock) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *AdminBlock) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *AdminBlock) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *AdminBlock) String() string {
	e.FullHash()
	str, _ := e.JSONString()
	return str
}
