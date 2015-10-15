// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package directoryblock

import (
	"bytes"
	"errors"
	"fmt"

	. "github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/primitives"
)

type DirectoryBlock struct {
	//Marshalized
	Header    *DBlockHeader
	DBEntries []*DBEntry

	//Not Marshalized
	IsSealed    bool
	DBHash      IHash
	KeyMR       IHash
	IsSavedInDB bool
	IsValidated bool
}

var _ Printable = (*DirectoryBlock)(nil)
var _ BinaryMarshallableAndCopyable = (*DirectoryBlock)(nil)
var _ IDirectoryBlock = (*DirectoryBlock)(nil)
var _ DatabaseBatchable = (*DirectoryBlock)(nil)

func (c *DirectoryBlock) New() BinaryMarshallableAndCopyable {
	return new(DirectoryBlock)
}

func (c *DirectoryBlock) GetDatabaseHeight() uint32 {
	return c.Header.DBHeight
}

func (c *DirectoryBlock) GetChainID() []byte {
	return D_CHAINID
}

func (c *DirectoryBlock) DatabasePrimaryIndex() IHash {
	return c.GetKeyMR()
}

func (c *DirectoryBlock) DatabaseSecondaryIndex() IHash {
	return c.GetHash()
}

func (c *DirectoryBlock) MarshalledSize() uint64 {
	panic("Function not implemented")
	return 0
}

func (e *DirectoryBlock) JSONByte() ([]byte, error) {
	return EncodeJSON(e)
}

func (e *DirectoryBlock) JSONString() (string, error) {
	return EncodeJSONString(e)
}

func (e *DirectoryBlock) JSONBuffer(b *bytes.Buffer) error {
	return EncodeJSONToBuffer(e, b)
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

func (b *DirectoryBlock) BuildBodyMR() (mr IHash, err error) {
	hashes := make([]IHash, len(b.DBEntries))
	for i, entry := range b.DBEntries {
		data, _ := entry.MarshalBinary()
		hashes[i] = Sha(data)
	}

	if len(hashes) == 0 {
		hashes = append(hashes, Sha(nil))
	}

	merkle := BuildMerkleTreeStore(hashes)
	return merkle[len(merkle)-1], nil
}

func (b *DirectoryBlock) BuildKeyMerkleRoot() (keyMR IHash, err error) {
	// Create the Entry Block Key Merkle Root from the hash of Header and the Body Merkle Root
	hashes := make([]IHash, 0, 2)
	binaryEBHeader, _ := b.Header.MarshalBinary()
	hashes = append(hashes, Sha(binaryEBHeader))
	hashes = append(hashes, b.Header.BodyMR)
	merkle := BuildMerkleTreeStore(hashes)
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

func (b *DirectoryBlock) GetHash() IHash {
	if b.DBHash == nil {
		binaryDblock, err := b.MarshalBinary()
		if err != nil {
			return nil
		}
		b.DBHash = Sha(binaryDblock)
	}
	return b.DBHash
}

func (b *DirectoryBlock) GetKeyMR() IHash {
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
	d.DBHash = NewZeroHash()
	d.KeyMR = NewZeroHash()

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
	b.Header.Version = VERSION_0

	if prev == nil {
		b.Header.PrevLedgerKeyMR = NewZeroHash()
		b.Header.PrevKeyMR = NewZeroHash()
	} else {
		b.Header.PrevLedgerKeyMR, err = CreateHash(prev)
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
