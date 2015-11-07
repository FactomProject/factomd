// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package directoryBlock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type DirectoryBlock struct {
	//Marshalized
	Header    interfaces.IDirectoryBlockHeader
	DBEntries []interfaces.IDBEntry

	//Not Marshalized
	IsSealed bool
	DBHash   interfaces.IHash
	KeyMR    interfaces.IHash
}

var _ interfaces.Printable = (*DirectoryBlock)(nil)
var _ interfaces.BinaryMarshallableAndCopyable = (*DirectoryBlock)(nil)
var _ interfaces.IDirectoryBlock = (*DirectoryBlock)(nil)
var _ interfaces.DatabaseBatchable = (*DirectoryBlock)(nil)

func (c *DirectoryBlock) GetDBEntries() []interfaces.IDBEntry {
	return c.DBEntries
}

func (c *DirectoryBlock) GetKeyMR() interfaces.IHash {
	keyMR, err := c.BuildKeyMerkleRoot()
	if err != nil {
		panic("Failed to build the key MR")
	}
	c.KeyMR = keyMR
	return c.KeyMR
}

func (c *DirectoryBlock) GetHeader() interfaces.IDirectoryBlockHeader {
	return c.Header
}

func (c *DirectoryBlock) SetHeader(header interfaces.IDirectoryBlockHeader) {
	c.Header = header
}

func (c *DirectoryBlock) SetDBEntries(dbEntries []interfaces.IDBEntry) {
	c.DBEntries = dbEntries
}

func (c *DirectoryBlock) New() interfaces.BinaryMarshallableAndCopyable {
	return new(DirectoryBlock)
}

func (c *DirectoryBlock) GetDatabaseHeight() uint32 {
	return c.GetHeader().GetDBHeight()
}

func (c *DirectoryBlock) GetChainID() []byte {
	return constants.D_CHAINID
}

func (c *DirectoryBlock) DatabasePrimaryIndex() interfaces.IHash {
	return c.GetKeyMR()
}

func (c *DirectoryBlock) DatabaseSecondaryIndex() interfaces.IHash {
	return c.GetHash()
}

func (e *DirectoryBlock) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *DirectoryBlock) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *DirectoryBlock) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *DirectoryBlock) String() string {
	pretty, _ := json.MarshalIndent(e, "", "  ")
	return string(pretty)
}

func (b *DirectoryBlock) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	count := uint32(len(b.GetDBEntries()))
	b.GetHeader().SetBlockCount(count)

	data, err = b.GetHeader().MarshalBinary()
	if err != nil {
		return
	}
	buf.Write(data)

	for i := uint32(0); i < count; i = i + 1 {
		data, err = b.GetDBEntries()[i].MarshalBinary()
		if err != nil {
			return
		}
		buf.Write(data)
	}

	return buf.Bytes(), err
}

func (b *DirectoryBlock) BuildBodyMR() (mr interfaces.IHash, err error) {
	hashes := make([]interfaces.IHash, len(b.GetDBEntries()))
	for i, entry := range b.GetDBEntries() {
		data, _ := entry.MarshalBinary()
		hashes[i] = primitives.Sha(data)
	}

	if len(hashes) == 0 {
		hashes = append(hashes, primitives.Sha(nil))
	}

	merkle := primitives.BuildMerkleTreeStore(hashes)
	return merkle[len(merkle)-1], nil
}

func (b *DirectoryBlock) BuildKeyMerkleRoot() (keyMR interfaces.IHash, err error) {
	// Create the Entry Block Key Merkle Root from the hash of Header and the Body Merkle Root
	hashes := make([]interfaces.IHash, 0, 2)
	binaryEBHeader, _ := b.GetHeader().MarshalBinary()
	hashes = append(hashes, primitives.Sha(binaryEBHeader))
	hashes = append(hashes, b.GetHeader().GetBodyMR())
	merkle := primitives.BuildMerkleTreeStore(hashes)
	keyMR = merkle[len(merkle)-1] // MerkleRoot is not marshalized in Dir Block

	return keyMR, nil
}

func (b *DirectoryBlock) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()

	newData = data

	var fbh interfaces.IDirectoryBlockHeader = new(DBlockHeader)

	newData, err = fbh.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}
	b.SetHeader(fbh)

	count := b.GetHeader().GetBlockCount()
	b.SetDBEntries(make([]interfaces.IDBEntry, count))
	for i := uint32(0); i < count; i++ {
		b.GetDBEntries()[i] = new(DBEntry)
		newData, err = b.GetDBEntries()[i].UnmarshalBinaryData(newData)
		if err != nil {
			return
		}
	}

	return
}

func (b *DirectoryBlock) UnmarshalBinary(data []byte) (err error) {
	_, err = b.UnmarshalBinaryData(data)
	return
}

func (b *DirectoryBlock) GetHash() interfaces.IHash {
	if b.DBHash == nil {
		binaryDblock, err := b.MarshalBinary()
		if err != nil {
			return nil
		}
		b.DBHash = primitives.Sha(binaryDblock)
	}
	return b.DBHash
}

func (b *DirectoryBlock) AddEntry(chainID interfaces.IHash, keyMR interfaces.IHash) {
	var dbentry interfaces.IDBEntry
	dbentry = new(DBEntry)
	dbentry.SetChainID(chainID)
	dbentry.SetKeyMR(keyMR)
	b.SetDBEntries(append(b.DBEntries, dbentry))
}
