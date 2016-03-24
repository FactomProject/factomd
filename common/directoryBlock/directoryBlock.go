// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package directoryBlock

import (
	"bytes"
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

var _ = fmt.Print

type DirectoryBlock struct {
	//Marshalized
	Header    interfaces.IDirectoryBlockHeader
	DBEntries []interfaces.IDBEntry

	//Not Marshalized
	DBHash interfaces.IHash
	KeyMR  interfaces.IHash
}

var _ interfaces.Printable = (*DirectoryBlock)(nil)
var _ interfaces.BinaryMarshallableAndCopyable = (*DirectoryBlock)(nil)
var _ interfaces.IDirectoryBlock = (*DirectoryBlock)(nil)
var _ interfaces.DatabaseBatchable = (*DirectoryBlock)(nil)
var _ interfaces.DatabaseBlockWithEntries = (*DirectoryBlock)(nil)

func (c *DirectoryBlock) GetEntryHashes() []interfaces.IHash {
	entries := c.DBEntries[:]
	answer := make([]interfaces.IHash, len(entries))
	for i, entry := range entries {
		answer[i] = entry.GetKeyMR()
	}
	return answer
}

func (c *DirectoryBlock) GetEntryHashesForBranch() []interfaces.IHash {
	entries := c.DBEntries[:]
	answer := make([]interfaces.IHash, 2*len(entries))
	for i, entry := range entries {
		answer[2*i] = entry.GetChainID()
		answer[2*i+1] = entry.GetKeyMR()
	}
	return answer
}

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

func (c *DirectoryBlock) SetDBEntries(dbEntries []interfaces.IDBEntry) error {
	c.DBEntries = dbEntries
	c.GetHeader().SetBlockCount(uint32(len(dbEntries)))
	_, err := c.BuildBodyMR()
	if err != nil {
		return err
	}
	return nil
}

func (c *DirectoryBlock) New() interfaces.BinaryMarshallableAndCopyable {
	dBlock := new(DirectoryBlock)
	dBlock.Header = NewDBlockHeader()
	dBlock.DBHash = primitives.NewZeroHash()
	dBlock.KeyMR = primitives.NewZeroHash()
	return dBlock
}

func (c *DirectoryBlock) GetDatabaseHeight() uint32 {
	return c.GetHeader().GetDBHeight()
}

func (c *DirectoryBlock) GetChainID() interfaces.IHash {
	return primitives.NewHash(constants.D_CHAINID)
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
	var out bytes.Buffer
	kmr, err := e.BuildKeyMerkleRoot()

	if err != nil {
		out.WriteString(fmt.Sprintf("%20s %v\n", "KeyMR:", err))
	} else {
		out.WriteString(fmt.Sprintf("%20s %v\n", "KeyMR:", kmr.String()))
	}

	kmr, err = e.BuildBodyMR()
	if err != nil {
		out.WriteString(fmt.Sprintf("%20s %v\n", "BodyMR:", err))
	} else {
		out.WriteString(fmt.Sprintf("%20s %v\n", "BodyMR:", kmr.String()))
	}

	out.WriteString(e.Header.String())
	out.WriteString("Entries: \n")
	for _, entry := range e.DBEntries {
		out.WriteString(entry.String())
	}

	return (string)(out.Bytes())

}

func (b *DirectoryBlock) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	b.BuildBodyMR()

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

func (b *DirectoryBlock) BuildBodyMR() (interfaces.IHash, error) {
	hashes := make([]interfaces.IHash, len(b.GetDBEntries()))
	for i, entry := range b.GetDBEntries() {
		data, err := entry.MarshalBinary()
		if err != nil {
			return nil, err
		}
		hashes[i] = primitives.Sha(data)
	}

	if len(hashes) == 0 {
		hashes = append(hashes, primitives.Sha(nil))
	}

	merkleTree := primitives.BuildMerkleTreeStore(hashes)
	merkleRoot := merkleTree[len(merkleTree)-1]

	b.GetHeader().SetBodyMR(merkleRoot)

	return merkleRoot, nil
}

func (b *DirectoryBlock) HeaderHash() (interfaces.IHash, error) {
	binaryEBHeader, err := b.GetHeader().MarshalBinary()
	if err != nil {
		return nil, err
	}
	return primitives.Sha(binaryEBHeader), nil
}

func (b *DirectoryBlock) BodyKeyMR() interfaces.IHash {
	return b.GetHeader().GetBodyMR()
}

func (b *DirectoryBlock) BuildKeyMerkleRoot() (keyMR interfaces.IHash, err error) {
	// Create the Entry Block Key Merkle Root from the hash of Header and the Body Merkle Root

	hashes := make([]interfaces.IHash, 0, 2)
	headerHash, err := b.HeaderHash()
	if err != nil {
		return nil, err
	}
	hashes = append(hashes, headerHash)
	hashes = append(hashes, b.BodyKeyMR())
	merkle := primitives.BuildMerkleTreeStore(hashes)
	keyMR = merkle[len(merkle)-1] // MerkleRoot is not marshalized in Dir Block

	b.KeyMR = keyMR

	b.GetFullHash() // Create the Full Hash when we create the keyMR

	return keyMR, nil
}

func (b *DirectoryBlock) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	/*defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()*/

	newData = data

	var fbh interfaces.IDirectoryBlockHeader = new(DBlockHeader)

	newData, err = fbh.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}
	b.SetHeader(fbh)

	count := b.GetHeader().GetBlockCount()
	entries := make([]interfaces.IDBEntry, count)
	for i := uint32(0); i < count; i++ {
		entries[i] = new(DBEntry)
		newData, err = entries[i].UnmarshalBinaryData(newData)
		if err != nil {
			return
		}
	}

	err = b.SetDBEntries(entries)
	if err != nil {
		return
	}

	return
}

func (b *DirectoryBlock) UnmarshalBinary(data []byte) (err error) {
	_, err = b.UnmarshalBinaryData(data)
	return
}

func (b *DirectoryBlock) GetHash() interfaces.IHash {
	binaryDblock, err := b.MarshalBinary()
	if err != nil {
		return nil
	}
	b.DBHash = primitives.Sha(binaryDblock)
	return b.DBHash
}

func (b *DirectoryBlock) GetFullHash() interfaces.IHash {
	b.Header.SetFullHash(b.GetHash())
	return b.Header.GetFullHash()
}

func (b *DirectoryBlock) AddEntry(chainID interfaces.IHash, keyMR interfaces.IHash) error {
	var dbentry interfaces.IDBEntry
	dbentry = new(DBEntry)
	dbentry.SetChainID(chainID)
	dbentry.SetKeyMR(keyMR)

	if b.DBEntries == nil {
		b.DBEntries = []interfaces.IDBEntry{}
	}

	return b.SetDBEntries(append(b.DBEntries, dbentry))
}

/*********************************************************************
 * Support
 *********************************************************************/

func NewDirectoryBlock(dbheight uint32, prev *DirectoryBlock) interfaces.IDirectoryBlock {
	newdb := new(DirectoryBlock)

	newdb.Header = new(DBlockHeader)
	newdb.Header.SetVersion(constants.VERSION_0)
	newdb.Header.SetPrevFullHash(primitives.NewZeroHash())
	newdb.Header.SetPrevKeyMR(primitives.NewZeroHash())

	if prev != nil {
		newdb.GetHeader().SetPrevFullHash(prev.GetFullHash())
		newdb.GetHeader().SetPrevKeyMR(prev.GetKeyMR())
	}

	newdb.GetHeader().SetDBHeight(dbheight)
	newdb.SetDBEntries(make([]interfaces.IDBEntry, 0))

	newdb.AddEntry(primitives.NewHash(constants.ADMIN_CHAINID), primitives.NewZeroHash())
	newdb.AddEntry(primitives.NewHash(constants.EC_CHAINID), primitives.NewZeroHash())
	newdb.AddEntry(primitives.NewHash(constants.FACTOID_CHAINID), primitives.NewZeroHash())

	return newdb
}
