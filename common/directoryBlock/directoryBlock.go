// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package directoryBlock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"errors"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

var _ = fmt.Print

// DirectoryBlock holds all the information related to a directory block, which covers a ~10 minute period of time.
// A directory block is the fundamental unit of the Factom blockchain and contains two main data objects:
// a header and its entries. The header simply contains information about where this directory block fits
// within the larger block chain. While the entries contain specific information entered into the block chain.
// Each directory block always contains an admin block (organizational information for the block chain), and
// EC block (entry credit transactions occuring during the ~10mins), a Factoid block (Factoid transactions occuring
// during the ~10mins), and finally an unknown number of standard entry blocks (Each Entry Block contains all of the
// entries for a particular Chain during a 10 minute period)
type DirectoryBlock struct {
	//Not Marshalized - hashing information related to the main entries below
	DBHash     interfaces.IHash `json:"dbhash"`     // Hash of the 'body' of the DirctoryBlock (entries only)
	KeyMR      interfaces.IHash `json:"keymr"`      // Merkle root of the HeaderHash and DBHash combined
	HeaderHash interfaces.IHash `json:"headerhash"` // Hash of the header of the directory block
	keyMRset   bool             `json:"keymrset"`   // has the key Merkle Root already computed

	//Marshalized - These are the main contents of the DirectoryBlock
	Header    interfaces.IDirectoryBlockHeader `json:"header"`    // Contains the header information for this directory block
	DBEntries []interfaces.IDBEntry            `json:"dbentries"` // A list of directory block entries
	// [0] is always the admin block
	// [1] is always the ec block
	// [2] is always the factoid block
	// [3+] is standard entry blocks
}

var _ interfaces.Printable = (*DirectoryBlock)(nil)
var _ interfaces.BinaryMarshallableAndCopyable = (*DirectoryBlock)(nil)
var _ interfaces.IDirectoryBlock = (*DirectoryBlock)(nil)
var _ interfaces.DatabaseBatchable = (*DirectoryBlock)(nil)
var _ interfaces.DatabaseBlockWithEntries = (*DirectoryBlock)(nil)

// Init creates a new Header for the internal object if its currently nil
func (c *DirectoryBlock) Init() {
	if c.Header == nil {
		h := new(DBlockHeader)
		h.Init()
		c.Header = h
	}
}

// IsSameAs checks the input objects header and its entries are identical to this directory block
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

// SetEntryHash adds a new entry to the directory block, expanding the entry array if needed
// to fit the new entry at the specified index
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

// SetABlockHash sets the input admin block as the first (0th) entry in the directory block
func (c *DirectoryBlock) SetABlockHash(aBlock interfaces.IAdminBlock) error {
	hash := aBlock.DatabasePrimaryIndex()
	c.SetEntryHash(hash, aBlock.GetChainID(), 0)
	return nil
}

// SetECBlockHash sets the input entry credit block to the second entry in the directory block
func (c *DirectoryBlock) SetECBlockHash(ecBlock interfaces.IEntryCreditBlock) error {
	hash := ecBlock.DatabasePrimaryIndex()
	c.SetEntryHash(hash, ecBlock.GetChainID(), 1)
	return nil
}

// SetFBlockHash sets the input Factoid block to the third entry in the directory block
func (c *DirectoryBlock) SetFBlockHash(fBlock interfaces.IFBlock) error {
	hash := fBlock.DatabasePrimaryIndex()
	c.SetEntryHash(hash, fBlock.GetChainID(), 2)
	return nil
}

// GetEntryHashes returns a copy of the entry block array
func (c *DirectoryBlock) GetEntryHashes() []interfaces.IHash {
	entries := c.DBEntries[:]
	answer := make([]interfaces.IHash, len(entries))
	for i, entry := range entries {
		answer[i] = entry.GetKeyMR()
	}
	return answer
}

// GetEntrySigHashes always returns nil
func (c *DirectoryBlock) GetEntrySigHashes() []interfaces.IHash {
	return nil
}

//Sort uses bubble sort to sort the standard chain entries (non admin, ec, or factoid blocks)
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

// GetEntryHashesForBranch expands the entry blocks into a single array consisting of paired hashes
// (chainID,keyMR) resulting in an array that is 2x the length of the DBEntries
func (c *DirectoryBlock) GetEntryHashesForBranch() []interfaces.IHash {
	entries := c.DBEntries[:]
	answer := make([]interfaces.IHash, 2*len(entries))
	for i, entry := range entries {
		answer[2*i] = entry.GetChainID()
		answer[2*i+1] = entry.GetKeyMR()
	}
	return answer
}

// GetDBEntries returns the array of entry blocks
func (c *DirectoryBlock) GetDBEntries() []interfaces.IDBEntry {
	return c.DBEntries
}

// GetEBlockDBEntries returns the non special entry blocks (omitting admin, ec, and factoid blocks)
func (c *DirectoryBlock) GetEBlockDBEntries() []interfaces.IDBEntry {
	answer := []interfaces.IDBEntry{}
	for _, v := range c.DBEntries {
		if v.GetChainID().String() == "000000000000000000000000000000000000000000000000000000000000000a" {
			continue // Ignore the admin block
		}
		if v.GetChainID().String() == "000000000000000000000000000000000000000000000000000000000000000f" {
			continue // Ignore the factoid block
		}
		if v.GetChainID().String() == "000000000000000000000000000000000000000000000000000000000000000c" {
			continue // Ignore the ec block
		}
		answer = append(answer, v)
	}
	return answer
}

// CheckDBEntries checks the first three special entry blocks: admin, ec, and factoid blocks, to make sure
// their chain IDs are properly set
func (c *DirectoryBlock) CheckDBEntries() error {
	if len(c.DBEntries) < 3 {
		return fmt.Errorf("Not enough entries - %v", len(c.DBEntries))
	}
	// Admin block is always in the 0th position
	if c.DBEntries[0].GetChainID().String() != "000000000000000000000000000000000000000000000000000000000000000a" {
		return fmt.Errorf("Invalid ChainID at position 0 - %v", c.DBEntries[0].GetChainID().String())
	}
	// EC block is always in the 1st position
	if c.DBEntries[1].GetChainID().String() != "000000000000000000000000000000000000000000000000000000000000000c" {
		return fmt.Errorf("Invalid ChainID at position 1 - %v", c.DBEntries[1].GetChainID().String())
	}
	// Factoid block is always in the 2nd position
	if c.DBEntries[2].GetChainID().String() != "000000000000000000000000000000000000000000000000000000000000000f" {
		return fmt.Errorf("Invalid ChainID at position 2 - %v", c.DBEntries[2].GetChainID().String())
	}
	return nil
}

// GetKeyMR returns the key Merkle Root of the directory block
func (c *DirectoryBlock) GetKeyMR() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("DirectoryBlock.GetKeyMR() saw an interface that was nil")
		}
	}()
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

// GetHeader returns the directory block header (initializing it if needed so its never nil)
func (c *DirectoryBlock) GetHeader() interfaces.IDirectoryBlockHeader {
	c.Init()
	return c.Header
}

// SetHeader sets the directory block header to the input header
func (c *DirectoryBlock) SetHeader(header interfaces.IDirectoryBlockHeader) {
	c.Header = header
}

// SetDBEntries wholesale replaces the entire internal entry block array with the input array
func (c *DirectoryBlock) SetDBEntries(dbEntries []interfaces.IDBEntry) error {
	if dbEntries == nil {
		return errors.New("dbEntries cannot be nil")
	}

	c.DBEntries = dbEntries
	return nil
}

// New returns a new directory block initialized to empty/zero hashes
func (c *DirectoryBlock) New() interfaces.BinaryMarshallableAndCopyable {
	dBlock := new(DirectoryBlock)
	dBlock.Header = NewDBlockHeader()
	dBlock.DBHash = primitives.NewZeroHash()
	dBlock.KeyMR = primitives.NewZeroHash()
	return dBlock
}

// GetDatabaseHeight returns the block height for this directory block from the header
func (c *DirectoryBlock) GetDatabaseHeight() uint32 {
	c.Init()
	return c.GetHeader().GetDBHeight()
}

// GetChainID returns the directory block chain id
func (c *DirectoryBlock) GetChainID() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("DirectoryBlock.GetChainID() saw an interface that was nil")
		}
	}()
	return primitives.NewHash(constants.D_CHAINID)
}

// DatabasePrimaryIndex returns the key Merkle root of the directory block
func (c *DirectoryBlock) DatabasePrimaryIndex() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("DirectoryBlock.DatabasePrimaryIndex() saw an interface that was nil")
		}
	}()
	return c.GetKeyMR()
}

// DatabaseSecondaryIndex returns the hash of the directory block
func (c *DirectoryBlock) DatabaseSecondaryIndex() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("DirectoryBlock.DatabaseSecondaryIndex() saw an interface that was nil")
		}
	}()
	return c.GetHash()
}

// JSONByte returns the json encoded byte array
func (c *DirectoryBlock) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(c)
}

// JSONString returns the json encoded string
func (c *DirectoryBlock) JSONString() (string, error) {
	return primitives.EncodeJSONString(c)
}

// String returns the DirectoryBlock as a string
func (c *DirectoryBlock) String() string {
	c.Init()
	var out primitives.Buffer

	kmr := c.GetKeyMR()
	out.WriteString(fmt.Sprintf("%20s %v\n", "keymr:", kmr.String()))

	kmr = c.BodyKeyMR()
	out.WriteString(fmt.Sprintf("%20s %v\n", "bodymr:", kmr.String()))

	fh := c.GetFullHash()
	out.WriteString(fmt.Sprintf("%20s %v\n", "fullhash:", fh.String()))

	out.WriteString(c.GetHeader().String())
	out.WriteString("entries:\n")
	for i, entry := range c.DBEntries {
		out.WriteString(fmt.Sprintf("%5d %s", i, entry.String()))
	}

	return (string)(out.DeepCopyBytes())
}

// MarshalBinary marshals the directory block
func (c *DirectoryBlock) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "DirectoryBlock.MarshalBinary err:%v", *pe)
		}
	}(&err)
	c.Init()
	c.Sort()
	_, err = c.BuildBodyMR()
	if err != nil {
		return nil, err
	}

	buf := primitives.NewBuffer(nil)

	err = buf.PushBinaryMarshallable(c.GetHeader())
	if err != nil {
		return nil, err
	}

	for i := uint32(0); i < c.Header.GetBlockCount(); i++ {
		err = buf.PushBinaryMarshallable(c.GetDBEntries()[i])
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), err
}

// BuildBodyMR builds the Merkle root of the 'body' (entries only) of the directory block, and both stores the
// Merkle root in the header and returns its to the user
func (c *DirectoryBlock) BuildBodyMR() (interfaces.IHash, error) {
	count := uint32(len(c.GetDBEntries()))
	c.GetHeader().SetBlockCount(count)
	if count == 0 {
		panic("Zero block size!")
	}

	hashes := make([]interfaces.IHash, len(c.GetDBEntries()))
	for i, entry := range c.GetDBEntries() {
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

	c.GetHeader().SetBodyMR(merkleRoot)

	return merkleRoot, nil
}

// GetHeaderHash returns the hash of the directory block header
func (c *DirectoryBlock) GetHeaderHash() (interfaces.IHash, error) {
	c.Header.SetBlockCount(uint32(len(c.GetDBEntries())))
	return c.Header.GetHeaderHash()
}

// BodyKeyMR builds the Merkle root of the 'body' (entries only) of the directory block, and both stores the
// Merkle root in the header and returns its to the user
func (c *DirectoryBlock) BodyKeyMR() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("DirectoryBlock.BodyKeyMR() saw an interface that was nil")
		}
	}()
	key, _ := c.BuildBodyMR()
	return key
}

// BuildKeyMerkleRoot returns the Merkle root by of the header hash and the DirectoryBlock body hash
// and also stores the Merkle root in the Directory block
func (c *DirectoryBlock) BuildKeyMerkleRoot() (keyMR interfaces.IHash, err error) {
	// Create the Entry Block Key Merkle Root from the hash of Header and the Body Merkle Root

	hashes := make([]interfaces.IHash, 0, 2)
	bodyKeyMR := c.BodyKeyMR() //This needs to be called first to build the header properly!!
	headerHash, err := c.GetHeaderHash()
	if err != nil {
		return nil, err
	}
	hashes = append(hashes, headerHash)
	hashes = append(hashes, bodyKeyMR)
	merkle := primitives.BuildMerkleTreeStore(hashes)
	keyMR = merkle[len(merkle)-1] // MerkleRoot is not marshalized in Dir Block

	c.KeyMR = keyMR

	c.GetFullHash() // Create the Full Hash when we create the keyMR

	return primitives.NewHash(keyMR.Bytes()), nil
}

// UnmarshalDBlock returns a new DirectoryBlock with the input data unmarshalled into it
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

// UnmarshalBinaryData unmarshals the input data into the directory block
func (c *DirectoryBlock) UnmarshalBinaryData(data []byte) ([]byte, error) {
	newData := data
	var err error
	var fbh interfaces.IDirectoryBlockHeader = new(DBlockHeader)

	newData, err = fbh.UnmarshalBinaryData(data)
	if err != nil {
		return nil, err
	}
	c.SetHeader(fbh)

	// entryLimit is the maximum number of 32 byte entries that could fit in the body of the binary dblock
	entryLimit := uint32(len(newData) / 32)
	entryCount := c.GetHeader().GetBlockCount()
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

	err = c.SetDBEntries(entries)
	if err != nil {
		return nil, err
	}

	err = c.CheckDBEntries()
	if err != nil {
		return nil, err
	}

	return newData, nil
}

// GetTimestamp returns the timestamp from the header
func (c *DirectoryBlock) GetTimestamp() interfaces.Timestamp {
	return c.GetHeader().GetTimestamp()
}

// UnmarshalBinary unmarshals the input data into this directory block
func (c *DirectoryBlock) UnmarshalBinary(data []byte) (err error) {
	_, err = c.UnmarshalBinaryData(data)
	return
}

// GetHash marshals the directory block and returns the SHA hash of the marshalled data
func (c *DirectoryBlock) GetHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("DirectoryBlock.GetHash() saw an interface that was nil")
		}
	}()
	return c.GetFullHash()
}

// GetFullHash marshals the directory block and returns the SHA hash of the marshalled data
func (c *DirectoryBlock) GetFullHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("DirectoryBlock.GetFullHash() saw an interface that was nil")
		}
	}()
	binaryDblock, err := c.MarshalBinary()
	if err != nil {
		return nil
	}
	c.DBHash = primitives.Sha(binaryDblock)
	return c.DBHash
}

// AddEntry creates and appends a new entry into the directory block given the input chainID and keyMR
func (c *DirectoryBlock) AddEntry(chainID interfaces.IHash, keyMR interfaces.IHash) error {
	var dbentry interfaces.IDBEntry
	dbentry = new(DBEntry)
	dbentry.SetChainID(chainID)
	dbentry.SetKeyMR(keyMR)

	if c.DBEntries == nil {
		c.DBEntries = []interfaces.IDBEntry{}
	}

	return c.SetDBEntries(append(c.DBEntries, dbentry))
}

/*********************************************************************
 * Support
 *********************************************************************/

// NewDirectoryBlock creates a new directory block whose previous block is the input prev
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

// CheckBlockPairIntegrity returns nil iff the input block's previous pointer matches the input prev values
// and the input block is the 'next' block after the prev
func CheckBlockPairIntegrity(block interfaces.IDirectoryBlock, prev interfaces.IDirectoryBlock) error {
	if block == nil {
		return fmt.Errorf("No block specified")
	}

	if prev == nil { // If we are the first directory block
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

// MarshalJSON marshals the directory block into json
func (c DirectoryBlock) MarshalJSON() ([]byte, error) {
	// Calculate the Merkle Root first
	c.GetKeyMR()
	// Calculate the hash of the directory block
	c.GetFullHash()
	// Now marshal the data
	return json.Marshal(ExpandedDBlock(c))
}
