// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package directoryblock

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type DirectoryBlock struct {
	//Marshalized
	header    interfaces.IDirectoryBlockHeader
	dbEntries []interfaces.IDBEntry

	//Not Marshalized
	IsSealed    bool
	DBHash      interfaces.IHash
	keyMR       interfaces.IHash
}

var _ interfaces.Printable = (*DirectoryBlock)(nil)
var _ interfaces.BinaryMarshallableAndCopyable = (*DirectoryBlock)(nil)
var _ interfaces.IDirectoryBlock = (*DirectoryBlock)(nil)
var _ interfaces.DatabaseBatchable = (*DirectoryBlock)(nil)

func (c *DirectoryBlock) DBEntries() []interfaces.IDBEntry {
	return c.dbEntries
}

func (c *DirectoryBlock) KeyMR() interfaces.IHash {
	return c.keyMR
}

func (c *DirectoryBlock) Header() interfaces.IDirectoryBlockHeader {
	return c.header
}

func (c *DirectoryBlock) SetHeader(header interfaces.IDirectoryBlockHeader) {
	c.header = header
}


func (c *DirectoryBlock) SetDBEntries(dbEntries []interfaces.IDBEntry) {
	c.dbEntries = dbEntries
}

func (c *DirectoryBlock) New() interfaces.BinaryMarshallableAndCopyable {
	return new(DirectoryBlock)
}

func (c *DirectoryBlock) GetDatabaseHeight() uint32 {
	return c.Header().DBHeight()
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
	str, _ := e.JSONString()
	return str
}

func (b *DirectoryBlock) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	data, err = b.Header().MarshalBinary()
	if err != nil {
		return
	}
	buf.Write(data)

	count := uint32(len(b.DBEntries()))
	for i := uint32(0); i < count; i = i + 1 {
		data, err = b.DBEntries()[i].MarshalBinary()
		if err != nil {
			return
		}
		buf.Write(data)
	}

	return buf.Bytes(), err
}

func (b *DirectoryBlock) BuildBodyMR() (mr interfaces.IHash, err error) {
	hashes := make([]interfaces.IHash, len(b.DBEntries()))
	for i, entry := range b.DBEntries() {
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
	binaryEBHeader, _ := b.Header().MarshalBinary()
	hashes = append(hashes, primitives.Sha(binaryEBHeader))
	hashes = append(hashes, b.Header().BodyMR())
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

	count := b.Header().BlockCount()
	b.SetDBEntries(make([]interfaces.IDBEntry, count))
	for i := uint32(0); i < count; i++ {
		b.DBEntries()[i] = new(DBEntry)
		newData, err = b.DBEntries()[i].UnmarshalBinaryData(newData)
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

func (b *DirectoryBlock) DBHeight() uint32 {
	return b.Header().DBHeight()
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

func (b *DirectoryBlock) GetKeyMR() interfaces.IHash {
	hash, err := b.BuildKeyMerkleRoot()
	if err != nil {
		return nil
	}
	return hash
}

/************************************************
 * Support Functions
 ************************************************/

func NewDirectoryBlock() *DirectoryBlock {
	d := new(DirectoryBlock)
	d.SetHeader(NewDBlockHeader())

	d.SetDBEntries( make([]interfaces.IDBEntry, 0))

	return d
}

func NewDBlock() *DirectoryBlock {
	return NewDirectoryBlock()
}

func CreateDBlock(nextDBHeight uint32, prev interfaces.IDirectoryBlock, cap uint) (b interfaces.IDirectoryBlock, err error) {
	if prev == nil && nextDBHeight != 0 {
		return nil, errors.New("Previous block cannot be nil")
	} else if prev != nil && nextDBHeight == 0 {
		return nil, errors.New("Origin block cannot have a parent block")
	}

	b = new(DirectoryBlock)

	b.SetHeader(new(DBlockHeader))
	b.Header().SetVersion(constants.VERSION_0)

	if prev == nil {
		b.Header().SetPrevLedgerKeyMR(primitives.NewZeroHash())
		b.Header().SetPrevKeyMR(primitives.NewZeroHash())
	} else {
		prevLedgerKeyMR, err := primitives.CreateHash(prev)
		if err != nil {
			return nil, err
		}
		b.Header().SetPrevLedgerKeyMR(prevLedgerKeyMR)
		keyMR, err := prev.BuildKeyMerkleRoot()
		if err != nil {
			return nil, err
		}
		b.Header().SetPrevKeyMR(keyMR)
	}

	b.Header().SetDBHeight(nextDBHeight)
	b.SetDBEntries(make([]interfaces.IDBEntry, 0, cap))

	return b, err
}
