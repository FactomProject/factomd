// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package directoryBlock

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

var _ = fmt.Print

type DirectoryBlock struct {
	//Not Marshalized
	DBHash     interfaces.IHash `json:"dbhash"`
	KeyMR      interfaces.IHash `json:"keymr"`
	HeaderHash interfaces.IHash `json:"headerhash"`
	keyMRset   bool             `json:"keymrset"`

	//Marshalized
	Header    interfaces.IDirectoryBlockHeader `json:"header"`
	DBEntries []interfaces.IDBEntry            `json:"dbentries"`
}

var _ interfaces.Printable = (*DirectoryBlock)(nil)
var _ interfaces.BinaryMarshallableAndCopyable = (*DirectoryBlock)(nil)
var _ interfaces.IDirectoryBlock = (*DirectoryBlock)(nil)
var _ interfaces.DatabaseBatchable = (*DirectoryBlock)(nil)
var _ interfaces.DatabaseBlockWithEntries = (*DirectoryBlock)(nil)

func (c *DirectoryBlock) Init() {
	if c.Header == nil {
		h := new(DBlockHeader)
		h.Init()
		c.Header = h
	}
}

func (a *DirectoryBlock) IsSameAs(b interfaces.IDirectoryBlock) bool {
	if a == nil || b == nil {
		if a == nil && b == nil {
			return true
		}
		return false
	}

	if a.Header.IsSameAs(b.GetHeader()) == false {
		return false
	}
	bDBEntries := b.GetDBEntries()
	if len(a.DBEntries) != len(bDBEntries) {
		return false
	}
	for i := range a.DBEntries {
		if a.DBEntries[i].IsSameAs(bDBEntries[i]) == false {
			return false
		}
	}
	return true
}

func (c *DirectoryBlock) SetEntryHash(hash, chainID interfaces.IHash, index int) {
	if len(c.DBEntries) <= index {
		ent := make([]interfaces.IDBEntry, index+1)
		copy(ent, c.DBEntries)
		c.DBEntries = ent
	}
	dbe := new(DBEntry)
	dbe.ChainID = chainID
	dbe.KeyMR = hash
	c.DBEntries[index] = dbe
}

func (c *DirectoryBlock) SetABlockHash(aBlock interfaces.IAdminBlock) error {
	hash := aBlock.DatabasePrimaryIndex()
	c.SetEntryHash(hash, aBlock.GetChainID(), 0)
	return nil
}

func (c *DirectoryBlock) SetECBlockHash(ecBlock interfaces.IEntryCreditBlock) error {
	hash := ecBlock.DatabasePrimaryIndex()
	c.SetEntryHash(hash, ecBlock.GetChainID(), 1)
	return nil
}

func (c *DirectoryBlock) SetFBlockHash(fBlock interfaces.IFBlock) error {
	hash := fBlock.DatabasePrimaryIndex()
	c.SetEntryHash(hash, fBlock.GetChainID(), 2)
	return nil
}

func (c *DirectoryBlock) GetEntryHashes() []interfaces.IHash {
	entries := c.DBEntries[:]
	answer := make([]interfaces.IHash, len(entries))
	for i, entry := range entries {
		answer[i] = entry.GetKeyMR()
	}
	return answer
}

func (c *DirectoryBlock) GetEntrySigHashes() []interfaces.IHash {
	return nil
}

//bubble sort
func (c *DirectoryBlock) Sort() {
	done := false
	for i := 3; !done && i < len(c.DBEntries)-1; i++ {
		done = true
		for j := 3; j < len(c.DBEntries)-1-i+3; j++ {
			comp := bytes.Compare(c.DBEntries[j].GetChainID().Bytes(),
				c.DBEntries[j+1].GetChainID().Bytes())
			if comp > 0 {
				h := c.DBEntries[j]
				c.DBEntries[j] = c.DBEntries[j+1]
				c.DBEntries[j+1] = h
				done = false
			}
		}
	}
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

func (c *DirectoryBlock) GetEBlockDBEntries() []interfaces.IDBEntry {
	answer := []interfaces.IDBEntry{}
	for _, v := range c.DBEntries {
		if v.GetChainID().String() == "000000000000000000000000000000000000000000000000000000000000000a" {
			continue
		}
		if v.GetChainID().String() == "000000000000000000000000000000000000000000000000000000000000000f" {
			continue
		}
		if v.GetChainID().String() == "000000000000000000000000000000000000000000000000000000000000000c" {
			continue
		}
		answer = append(answer, v)
	}
	return answer
}

func (c *DirectoryBlock) CheckDBEntries() error {
	if len(c.DBEntries) < 3 {
		return fmt.Errorf("Not enough entries - %v", len(c.DBEntries))
	}
	if c.DBEntries[0].GetChainID().String() != "000000000000000000000000000000000000000000000000000000000000000a" {
		return fmt.Errorf("Invalid ChainID at position 0 - %v", c.DBEntries[0].GetChainID().String())
	}
	if c.DBEntries[1].GetChainID().String() != "000000000000000000000000000000000000000000000000000000000000000c" {
		return fmt.Errorf("Invalid ChainID at position 1 - %v", c.DBEntries[1].GetChainID().String())
	}
	if c.DBEntries[2].GetChainID().String() != "000000000000000000000000000000000000000000000000000000000000000f" {
		return fmt.Errorf("Invalid ChainID at position 2 - %v", c.DBEntries[2].GetChainID().String())
	}
	return nil
}

func (c *DirectoryBlock) GetKeyMR() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "DirectoryBlock.GetKeyMR") }()
	keyMR, err := c.BuildKeyMerkleRoot()
	if err != nil {
		panic("Failed to build the key MR")
	}

	//if c.keyMRset && c.KeyMR.Fixed() != keyMR.Fixed() {
	//	panic("keyMR changed!")
	//}

	c.KeyMR = keyMR
	c.keyMRset = true

	return c.KeyMR
}

func (c *DirectoryBlock) GetHeader() interfaces.IDirectoryBlockHeader {
	c.Init()
	return c.Header
}

func (c *DirectoryBlock) SetHeader(header interfaces.IDirectoryBlockHeader) {
	c.Header = header
}

func (c *DirectoryBlock) SetDBEntries(dbEntries []interfaces.IDBEntry) error {
	if dbEntries == nil {
		return errors.New("dbEntries cannot be nil")
	}

	c.DBEntries = dbEntries
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
	c.Init()
	return c.GetHeader().GetDBHeight()
}

func (c *DirectoryBlock) GetChainID() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "DirectoryBlock.GetChainID") }()
	return primitives.NewHash(constants.D_CHAINID)
}

func (c *DirectoryBlock) DatabasePrimaryIndex() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "DirectoryBlock.DatabasePrimaryIndex") }()
	return c.GetKeyMR()
}

func (c *DirectoryBlock) DatabaseSecondaryIndex() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "DirectoryBlock.DatabaseSecondaryIndex") }()
	return c.GetHash()
}

func (e *DirectoryBlock) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *DirectoryBlock) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *DirectoryBlock) String() string {
	e.Init()
	var out primitives.Buffer

	kmr := e.GetKeyMR()
	out.WriteString(fmt.Sprintf("%20s %v\n", "keymr:", kmr.String()))

	kmr = e.BodyKeyMR()
	out.WriteString(fmt.Sprintf("%20s %v\n", "bodymr:", kmr.String()))

	fh := e.GetFullHash()
	out.WriteString(fmt.Sprintf("%20s %v\n", "fullhash:", fh.String()))

	out.WriteString(e.GetHeader().String())
	out.WriteString("entries:\n")
	for i, entry := range e.DBEntries {
		out.WriteString(fmt.Sprintf("%5d %s", i, entry.String()))
	}

	return (string)(out.DeepCopyBytes())
}

func (b *DirectoryBlock) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "DirectoryBlock.MarshalBinary err:%v", *pe)
		}
	}(&err)
	b.Init()
	b.Sort()
	_, err = b.BuildBodyMR()
	if err != nil {
		return nil, err
	}

	buf := primitives.NewBuffer(nil)

	err = buf.PushBinaryMarshallable(b.GetHeader())
	if err != nil {
		return nil, err
	}

	for i := uint32(0); i < b.Header.GetBlockCount(); i++ {
		err = buf.PushBinaryMarshallable(b.GetDBEntries()[i])
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), err
}

func (b *DirectoryBlock) BuildBodyMR() (interfaces.IHash, error) {
	count := uint32(len(b.GetDBEntries()))
	b.GetHeader().SetBlockCount(count)
	if count == 0 {
		panic("Zero block size!")
	}

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

func (b *DirectoryBlock) GetHeaderHash() (interfaces.IHash, error) {
	b.Header.SetBlockCount(uint32(len(b.GetDBEntries())))
	return b.Header.GetHeaderHash()
}

func (b *DirectoryBlock) BodyKeyMR() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "DirectoryBlock.BodyKeyMR") }()
	key, _ := b.BuildBodyMR()
	return key
}

func (b *DirectoryBlock) BuildKeyMerkleRoot() (keyMR interfaces.IHash, err error) {
	// Create the Entry Block Key Merkle Root from the hash of Header and the Body Merkle Root

	hashes := make([]interfaces.IHash, 0, 2)
	bodyKeyMR := b.BodyKeyMR() //This needs to be called first to build the header properly!!
	headerHash, err := b.GetHeaderHash()
	if err != nil {
		return nil, err
	}
	b.HeaderHash = headerHash
	hashes = append(hashes, headerHash)
	hashes = append(hashes, bodyKeyMR)
	merkle := primitives.BuildMerkleTreeStore(hashes)
	keyMR = merkle[len(merkle)-1] // MerkleRoot is not marshalized in Dir Block

	b.KeyMR = keyMR

	b.GetFullHash() // Create the Full Hash when we create the keyMR

	return primitives.NewHash(keyMR.Bytes()), nil
}

func UnmarshalDBlock(data []byte) (interfaces.IDirectoryBlock, error) {
	dBlock := new(DirectoryBlock)
	dBlock.Header = NewDBlockHeader()
	dBlock.DBHash = primitives.NewZeroHash()
	dBlock.KeyMR = primitives.NewZeroHash()
	err := dBlock.UnmarshalBinary(data)
	if err != nil {
		return nil, err
	}
	return dBlock, nil
}

func (b *DirectoryBlock) UnmarshalBinaryData(data []byte) ([]byte, error) {
	newData := data
	var err error
	var fbh interfaces.IDirectoryBlockHeader = new(DBlockHeader)

	newData, err = fbh.UnmarshalBinaryData(data)
	if err != nil {
		return nil, err
	}
	b.SetHeader(fbh)

	// entryLimit is the maximum number of 32 byte entries that could fit in the body of the binary dblock
	entryLimit := uint32(len(newData) / 32)
	entryCount := b.GetHeader().GetBlockCount()
	if entryCount > entryLimit {
		return nil, fmt.Errorf(
			"Error: DirectoryBlock.UnmarshalBinary: Entry count %d is larger "+
				"than body size %d. (uint underflow?)",
			entryCount, entryLimit,
		)
	}

	entries := make([]interfaces.IDBEntry, entryCount)
	for i := uint32(0); i < entryCount; i++ {
		entries[i] = new(DBEntry)
		newData, err = entries[i].UnmarshalBinaryData(newData)
		if err != nil {
			return nil, err
		}
	}

	err = b.SetDBEntries(entries)
	if err != nil {
		return nil, err
	}

	err = b.CheckDBEntries()
	if err != nil {
		return nil, err
	}

	return newData, nil
}

func (h *DirectoryBlock) GetTimestamp() interfaces.Timestamp {
	return h.GetHeader().GetTimestamp().Clone()
}

func (b *DirectoryBlock) UnmarshalBinary(data []byte) (err error) {
	_, err = b.UnmarshalBinaryData(data)
	return
}

func (b *DirectoryBlock) GetHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "DirectoryBlock.GetHash") }()
	return b.GetFullHash()
}

func (b *DirectoryBlock) GetFullHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "DirectoryBlock.GetFullHash") }()
	binaryDblock, err := b.MarshalBinary()
	if err != nil {
		return nil
	}
	b.DBHash = primitives.Sha(binaryDblock)
	return b.DBHash
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

func NewDirectoryBlock(prev interfaces.IDirectoryBlock) interfaces.IDirectoryBlock {
	newdb := new(DirectoryBlock)

	newdb.Header = new(DBlockHeader)
	newdb.GetHeader().SetVersion(constants.VERSION_0)

	if prev != nil {
		newdb.GetHeader().SetPrevFullHash(prev.GetFullHash())
		newdb.GetHeader().SetPrevKeyMR(prev.GetKeyMR())
		newdb.GetHeader().SetDBHeight(prev.GetHeader().GetDBHeight() + 1)
	} else {
		newdb.GetHeader().SetPrevFullHash(primitives.NewZeroHash())
		newdb.GetHeader().SetPrevKeyMR(primitives.NewZeroHash())
		newdb.GetHeader().SetDBHeight(0)
	}

	newdb.SetDBEntries(make([]interfaces.IDBEntry, 0))

	newdb.AddEntry(primitives.NewHash(constants.ADMIN_CHAINID), primitives.NewZeroHash())
	newdb.AddEntry(primitives.NewHash(constants.EC_CHAINID), primitives.NewZeroHash())
	newdb.AddEntry(primitives.NewHash(constants.FACTOID_CHAINID), primitives.NewZeroHash())

	return newdb
}

func CheckBlockPairIntegrity(block interfaces.IDirectoryBlock, prev interfaces.IDirectoryBlock) error {
	if block == nil {
		return fmt.Errorf("No block specified")
	}

	if prev == nil {
		if block.GetHeader().GetPrevKeyMR().IsZero() == false {
			return fmt.Errorf("Invalid PrevKeyMR")
		}
		if block.GetHeader().GetPrevFullHash().IsZero() == false {
			return fmt.Errorf("Invalid PrevFullHash")
		}
		if block.GetHeader().GetDBHeight() != 0 {
			return fmt.Errorf("Invalid DBHeight")
		}
	} else {
		if block.GetHeader().GetPrevKeyMR().IsSameAs(prev.GetKeyMR()) == false {
			return fmt.Errorf("Invalid PrevKeyMR")
		}
		if block.GetHeader().GetPrevFullHash().IsSameAs(prev.GetFullHash()) == false {
			return fmt.Errorf("Invalid PrevFullHash")
		}
		if block.GetHeader().GetDBHeight() != (prev.GetHeader().GetDBHeight() + 1) {
			return fmt.Errorf("Invalid DBHeight")
		}
	}

	return nil
}

type ExpandedDBlock DirectoryBlock

func (e DirectoryBlock) MarshalJSON() ([]byte, error) {
	e.GetKeyMR()
	e.GetFullHash()

	return json.Marshal(ExpandedDBlock(e))
}
