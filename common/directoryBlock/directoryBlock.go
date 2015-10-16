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
	Header    *DBlockHeader
	DBEntries []*DBEntry

	//Not Marshalized
	IsSealed    bool
	DBHash      interfaces.IHash
	KeyMR       interfaces.IHash
	IsSavedInDB bool
	IsValidated bool
}

var _ interfaces.Printable = (*DirectoryBlock)(nil)
var _ interfaces.BinaryMarshallableAndCopyable = (*DirectoryBlock)(nil)
var _ interfaces.IDirectoryBlock = (*DirectoryBlock)(nil)
var _ interfaces.DatabaseBatchable = (*DirectoryBlock)(nil)

func (c *DirectoryBlock) New() interfaces.BinaryMarshallableAndCopyable {
	return new(DirectoryBlock)
}

func (c *DirectoryBlock) GetDatabaseHeight() uint32 {
	return c.Header.DBHeight
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

	data, err = b.Header.MarshalBinary()
	if err != nil {
		return
	}
	buf.Write(data)

	count := uint32(len(b.DBEntries))
	for i := uint32(0); i < count; i = i + 1 {
		data, err = b.DBEntries[i].MarshalBinary()
		if err != nil {
			return
		}
		buf.Write(data)
	}

	return buf.Bytes(), err
}

func (b *DirectoryBlock) BuildBodyMR() (mr interfaces.IHash, err error) {
	hashes := make([]interfaces.IHash, len(b.DBEntries))
	for i, entry := range b.DBEntries {
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
	binaryEBHeader, _ := b.Header.MarshalBinary()
	hashes = append(hashes, primitives.Sha(binaryEBHeader))
	hashes = append(hashes, b.Header.BodyMR)
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

	fbh := new(DBlockHeader)
	newData, err = fbh.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}
	b.Header = fbh

	count := b.Header.BlockCount
	b.DBEntries = make([]*DBEntry, count)
	for i := uint32(0); i < count; i++ {
		b.DBEntries[i] = new(DBEntry)
		newData, err = b.DBEntries[i].UnmarshalBinaryData(newData)
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

func (b *DirectoryBlock) GetDBHeight() uint32 {
	return b.Header.DBHeight
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
	if b.KeyMR == nil {
		b.BuildKeyMerkleRoot()
	}
	return b.KeyMR
}

/**
func (b *DirectoryBlock) EncodableFields() map[string]reflect.Value {
	fields := map[string]reflect.Value{
		`Header`:    reflect.ValueOf(b.Header),
		`DBEntries`: reflect.ValueOf(b.DBEntries),
		`DBHash`:    reflect.ValueOf(b.DBHash),
	}
	return fields
}
**/

/************************************************
 * Support Functions
 ************************************************/

func NewDirectoryBlock() *DirectoryBlock {
	d := new(DirectoryBlock)
	d.Header = NewDBlockHeader()

	d.DBEntries = make([]*DBEntry, 0)
	d.DBHash = primitives.NewZeroHash()
	d.KeyMR = primitives.NewZeroHash()

	return d
}

func NewDBlock() *DirectoryBlock {
	return NewDirectoryBlock()
}

func CreateDBlock(nextDBHeight uint32, prev *DirectoryBlock, cap uint) (b *DirectoryBlock, err error) {
	if prev == nil && nextDBHeight != 0 {
		return nil, errors.New("Previous block cannot be nil")
	} else if prev != nil && nextDBHeight == 0 {
		return nil, errors.New("Origin block cannot have a parent block")
	}

	b = new(DirectoryBlock)

	b.Header = new(DBlockHeader)
	b.Header.Version = constants.VERSION_0

	if prev == nil {
		b.Header.PrevLedgerKeyMR = primitives.NewZeroHash()
		b.Header.PrevKeyMR = primitives.NewZeroHash()
	} else {
		b.Header.PrevLedgerKeyMR, err = primitives.CreateHash(prev)
		if prev.KeyMR == nil {
			prev.BuildKeyMerkleRoot()
		}
		b.Header.PrevKeyMR = prev.KeyMR
	}

	b.Header.DBHeight = nextDBHeight
	b.DBEntries = make([]*DBEntry, 0, cap)
	b.IsSealed = false

	return b, err
}
