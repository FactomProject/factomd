// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package adminBlock

import (
	"encoding/json"
	"fmt"
	"os"

	"reflect"
	"sort"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// AdminBlock - Administrative Block
// This is a special block which accompanies this Directory Block.
// It contains the signatures and organizational data needed to validate previous and future Directory Blocks.
// This block is included in the DB body. It appears there with a pair of the Admin AdminChainID:SHA256 of the block.
// For more details, please go to:
// https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#administrative-block
//
type AdminBlock struct {
	//Marshalized
	Header            interfaces.IABlockHeader      `json:"header"`    // the admin block header
	ABEntries         []interfaces.IABEntry         `json:"abentries"` // array of admin block entries
	identityABEntries []interfaces.IIdentityABEntry // This is all identity related entries. They are sorted before added
}

var _ interfaces.IAdminBlock = (*AdminBlock)(nil)
var _ interfaces.Printable = (*AdminBlock)(nil)
var _ interfaces.BinaryMarshallableAndCopyable = (*AdminBlock)(nil)
var _ interfaces.DatabaseBatchable = (*AdminBlock)(nil)

// Init initializes the admin block if the header is nil
func (c *AdminBlock) Init() {
	if c.Header == nil {
		h := new(ABlockHeader)
		h.Init()
		c.Header = h
		c.ABEntries = make([]interfaces.IABEntry, 0)
	}
}

// IsSameAs returns true iff the input admin block header is identical to this admin block header, and the AB entry lengths are the same
func (c *AdminBlock) IsSameAs(b interfaces.IAdminBlock) bool {
	if !c.Header.IsSameAs(b.GetHeader()) {
		return false
	}
	if len(c.ABEntries) != len(b.GetABEntries()) {
		return false
	}
	return true
}

// String writes the admin block to a string
func (c *AdminBlock) String() string {
	c.Init()
	var out primitives.Buffer

	fh, _ := c.BackReferenceHash()
	if fh == nil {
		fh = primitives.NewZeroHash()
	}
	out.WriteString(fmt.Sprintf("%20s %x\n", "Primary Hash:", c.DatabasePrimaryIndex().Bytes()))
	out.WriteString(fmt.Sprintf("%20s %x\n", "512 Sha3:", fh.Bytes()))

	out.WriteString(c.GetHeader().String())
	out.WriteString("entries: \n")
	for _, entry := range c.ABEntries {
		out.WriteString(entry.String() + "\n")
	}

	return (string)(out.DeepCopyBytes())
}

// UpdateState updates the factomd state based on the admin block entries
func (c *AdminBlock) UpdateState(state interfaces.IState) error {
	c.Init()
	if state == nil {
		return fmt.Errorf("No State provided")
	}

	dbSigs := []*DBSignatureEntry{}
	for _, entry := range c.ABEntries {
		if entry.Type() == constants.TYPE_DB_SIGNATURE {
			dbSigs = append(dbSigs, entry.(*DBSignatureEntry))
		} else {
			err := entry.UpdateState(state)
			if err != nil {
				return err
			}
		}
	}

	for _, dbSig := range dbSigs {
		//list.State.ProcessLists.Get(currentDBHeight).DBSignatures
		state.AddDBSig(c.GetDBHeight()-1, dbSig.IdentityAdminChainID, &dbSig.PrevDBSig)
	}

	// Clear any keys that are now too old to be valid
	//state.UpdateAuthSigningKeys(c.GetHeader().GetDBHeight())
	return nil
}

// FetchCoinbaseDescriptor returns the first admin block entry associated with the coinbase descriptor
func (c *AdminBlock) FetchCoinbaseDescriptor() interfaces.IABEntry {
	for _, e := range c.ABEntries {
		if e.Type() == constants.TYPE_COINBASE_DESCRIPTOR {
			return e
		}
	}
	return nil
}

// AddDBSig creates and adds a new DBSignatureEntry to the admin block given the input server identity
func (c *AdminBlock) AddDBSig(serverIdentity interfaces.IHash, sig interfaces.IFullSignature) error {
	if serverIdentity == nil {
		return fmt.Errorf("No serverIdentity provided")
	}
	if sig == nil {
		return fmt.Errorf("No sig provided")
	}

	entry, err := NewDBSignatureEntry(serverIdentity, sig)
	if err != nil {
		return err
	}
	return c.AddEntry(entry)
}

// AddFedServer adds a new entry to add a federated server with the input chain id
func (c *AdminBlock) AddFedServer(identityChainID interfaces.IHash) error {
	c.Init()
	if identityChainID == nil {
		return fmt.Errorf("No identityChainID provided")
	}

	entry := NewAddFederatedServer(identityChainID, c.GetHeader().GetDBHeight()+1) // Goes in the NEXT block
	return c.AddEntry(entry)
}

// AddAuditServer adds a new entry to add an audit server with the input chain id
func (c *AdminBlock) AddAuditServer(identityChainID interfaces.IHash) error {
	c.Init()
	if identityChainID == nil {
		return fmt.Errorf("No identityChainID provided")
	}

	entry := NewAddAuditServer(identityChainID, c.GetHeader().GetDBHeight()+1) // Goes in the NEXT block
	return c.AddEntry(entry)
}

// RemoveFederatedServer adds an entry to remove the federated server with the input chain id
func (c *AdminBlock) RemoveFederatedServer(identityChainID interfaces.IHash) error {
	c.Init()
	if identityChainID == nil {
		return fmt.Errorf("No identityChainID provided")
	}

	entry := NewRemoveFederatedServer(identityChainID, c.GetHeader().GetDBHeight()+1) // Goes in the NEXT block
	return c.AddEntry(entry)
}

// AddCancelCoinbaseDescriptor adds an entry to cancel the a previous coinbase descriptor
func (c *AdminBlock) AddCancelCoinbaseDescriptor(descriptorHeight, index uint32) error {
	c.Init()
	entry := NewCancelCoinbaseDescriptor(descriptorHeight, index)

	return c.AddIdentityEntry(entry)
}

// InsertIdentityABEntries will prepare the identity entries and add them into the adminblock
func (c *AdminBlock) InsertIdentityABEntries() error {
	sort.Sort(interfaces.IIdentityABEntrySort(c.identityABEntries))
	for _, v := range c.identityABEntries {
		err := c.AddEntry(v)
		if err != nil {
			return fmt.Errorf("No identityChainID provided")
		}
	}
	return nil
}

// AddMatryoshkaHash adds an entry to add a Matryoshka hash to the input server identity chain
func (c *AdminBlock) AddMatryoshkaHash(identityChainID interfaces.IHash, mHash interfaces.IHash) error {
	if identityChainID == nil {
		return fmt.Errorf("No identityChainID provided")
	}
	if mHash == nil {
		return fmt.Errorf("No mHash provided")
	}

	entry := NewAddReplaceMatryoshkaHash(identityChainID, mHash)
	return c.AddIdentityEntry(entry)
}

// AddFederatedServerSigningKey adds an entry to add a signing key to the input server identity chain
func (c *AdminBlock) AddFederatedServerSigningKey(identityChainID interfaces.IHash, publicKey [32]byte) error {
	c.Init()
	if identityChainID == nil {
		return fmt.Errorf("No identityChainID provided")
	}

	p := new(primitives.PublicKey)
	err := p.UnmarshalBinary(publicKey[:])
	if err != nil {
		return err
	}
	entry := NewAddFederatedServerSigningKey(identityChainID, byte(0), *p, c.GetHeader().GetDBHeight()+1)
	return c.AddIdentityEntry(entry)
}

// AddFederatedServerBitcoinAnchorKey adds an entry to add a BTC key to the specific server identity chain
func (c *AdminBlock) AddFederatedServerBitcoinAnchorKey(identityChainID interfaces.IHash, keyPriority byte, keyType byte, ecdsaPublicKey [20]byte) error {
	if identityChainID == nil {
		return fmt.Errorf("No identityChainID provided")
	}

	b := new(primitives.ByteSlice20)
	err := b.UnmarshalBinary(ecdsaPublicKey[:])
	if err != nil {
		return err
	}
	entry := NewAddFederatedServerBitcoinAnchorKey(identityChainID, keyPriority, keyType, *b)
	return c.AddIdentityEntry(entry)
}

// AddCoinbaseDescriptor adds an entry to add a coinbase descriptor
func (c *AdminBlock) AddCoinbaseDescriptor(outputs []interfaces.ITransAddress) error {
	c.Init()
	if outputs == nil {
		return fmt.Errorf("No outputs provided")
	}

	entry := NewCoinbaseDescriptor(outputs)
	return c.AddEntry(entry)
}

// AddEfficiency adds an entry to add an efficiency to the input server identity chain
func (c *AdminBlock) AddEfficiency(chain interfaces.IHash, efficiency uint16) error {
	c.Init()
	if chain == nil {
		return fmt.Errorf("No chainid provided")
	}

	entry := NewAddEfficiency(chain, efficiency)
	return c.AddIdentityEntry(entry)
}

// AddCoinbaseAddress adds an entry to add a factoid address for coinbase transactions associated with the input server identity chain
func (c *AdminBlock) AddCoinbaseAddress(chain interfaces.IHash, add interfaces.IAddress) error {
	c.Init()
	if chain == nil {
		return fmt.Errorf("No chainid provided")
	}

	entry := NewAddFactoidAddress(chain, add)
	return c.AddIdentityEntry(entry)
}

// AddIdentityEntry appends a new identity entry to the array
func (c *AdminBlock) AddIdentityEntry(entry interfaces.IIdentityABEntry) error {
	if entry == nil {
		return fmt.Errorf("No entry provided")
	}
	// These get sorted when you call the InsertIdentityABEntries function
	c.identityABEntries = append(c.identityABEntries, entry)
	return nil
}

// AddEntry adds a new admin block entry to the array
func (c *AdminBlock) AddEntry(entry interfaces.IABEntry) error {
	if entry == nil {
		return fmt.Errorf("No entry provided")
	}

	if entry.Type() == constants.TYPE_SERVER_FAULT {
		//Server Faults needs to be ordered in a specific way
		return c.AddServerFault(entry)
	}

	for i := range c.ABEntries {
		//Server Faults are always the last entry in an AdminBlock
		if c.ABEntries[i].Type() == constants.TYPE_SERVER_FAULT {
			c.ABEntries = append(c.ABEntries[:i], append([]interfaces.IABEntry{entry}, c.ABEntries[i:]...)...)
			return nil
		}
	}
	c.ABEntries = append(c.ABEntries, entry)
	return nil
}

// AddServerFault adds an entry for a server fault
func (c *AdminBlock) AddServerFault(serverFault interfaces.IABEntry) error {
	if serverFault == nil {
		return fmt.Errorf("No serverFault provided")
	}

	sf, ok := serverFault.(*ServerFault)
	if ok == false {
		return fmt.Errorf("Entry is not serverFault")
	}

	for i := range c.ABEntries {
		if c.ABEntries[i].Type() == constants.TYPE_SERVER_FAULT {
			//Server Faults need to follow a deterministic order
			if c.ABEntries[i].(*ServerFault).Compare(sf) > 0 {
				c.ABEntries = append(c.ABEntries[:i], append([]interfaces.IABEntry{sf}, c.ABEntries[i:]...)...)
				return nil
			}
		}
	}
	c.ABEntries = append(c.ABEntries, sf)
	return nil
}

// GetHeader returns the admin block header
func (c *AdminBlock) GetHeader() interfaces.IABlockHeader {
	c.Init()
	return c.Header
}

// SetHeader sets the admin block header
func (c *AdminBlock) SetHeader(header interfaces.IABlockHeader) {
	c.Header = header
}

// GetABEntries returns the admin block entry array
func (c *AdminBlock) GetABEntries() []interfaces.IABEntry {
	return c.ABEntries
}

// GetDBHeight returns the directory block height associated with this admin block
func (c *AdminBlock) GetDBHeight() uint32 {
	return c.GetHeader().GetDBHeight()
}

// SetABEntries sets the admin block entry array to the input entries
func (c *AdminBlock) SetABEntries(abentries []interfaces.IABEntry) {
	c.ABEntries = abentries
}

// New returns a new empty admin block
func (c *AdminBlock) New() interfaces.BinaryMarshallableAndCopyable {
	return NewAdminBlock(nil)
}

// GetDatabaseHeight returns the directory block height associated with this admin block
func (c *AdminBlock) GetDatabaseHeight() uint32 {
	return c.GetHeader().GetDBHeight()
}

// GetChainID returns the hardcoded admin block chain id 0x0a
func (c *AdminBlock) GetChainID() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("AdminBlock.DatabasePrimaryIndex() saw an interface that was nil")
		}
	}()
	return c.GetHeader().GetAdminChainID()
}

// DatabasePrimaryIndex returns the SHA256 hash for the admin block
func (c *AdminBlock) DatabasePrimaryIndex() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("AdminBlock.DatabasePrimaryIndex() saw an interface that was nil")
		}
	}()
	key, _ := c.LookupHash()
	return key
}

// DatabaseSecondaryIndex returns the SHA512Half hash for the admin block
func (c *AdminBlock) DatabaseSecondaryIndex() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("AdminBlock.DatabaseSecondaryIndex() saw an interface that was nil")
		}
	}()
	key, _ := c.BackReferenceHash()
	return key
}

// GetHash returns the key Merkle root of the admin block
func (c *AdminBlock) GetHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("AdminBlock.GetHash() saw an interface that was nil")
		}
	}()
	h, _ := c.GetKeyMR()
	return h
}

// GetKeyMR returns the key Merkle root of the admin block
func (c *AdminBlock) GetKeyMR() (interfaces.IHash, error) {
	return c.BackReferenceHash()
}

// BackReferenceHash returns the SHA512Half hash for the admin block
func (c *AdminBlock) BackReferenceHash() (interfaces.IHash, error) {
	var binaryAB []byte
	binaryAB, err := c.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return primitives.Sha512Half(binaryAB), nil
}

// LookupHash returns the SHA256 hash for the admin block
func (c *AdminBlock) LookupHash() (interfaces.IHash, error) {
	var binaryAB []byte
	binaryAB, err := c.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return primitives.Sha(binaryAB), nil
}

// AddABEntry adds an Admin Block entry to the block
func (c *AdminBlock) AddABEntry(e interfaces.IABEntry) error {
	return c.AddEntry(e)
}

// AddFirstABEntry adds an Admin Block entry to the start of the block entries
func (c *AdminBlock) AddFirstABEntry(e interfaces.IABEntry) (err error) {
	c.ABEntries = append(c.ABEntries, nil)
	copy(c.ABEntries[1:], c.ABEntries[:len(c.ABEntries)-1])
	c.ABEntries[0] = e
	return
}

// MarshalBinary marshals the object to binary
func (c *AdminBlock) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "AdminBlock.MarshalBinary err:%v", *pe)
		}
	}(&err)
	c.Init()
	// Marshal all the entries into their own thing (need the size)
	var buf2 primitives.Buffer
	for _, v := range c.ABEntries {
		err := buf2.PushBinaryMarshallable(v)
		if err != nil {
			return nil, err
		}
	}

	c.GetHeader().SetMessageCount(uint32(len(c.ABEntries)))
	c.GetHeader().SetBodySize(uint32(buf2.Len()))

	var buf primitives.Buffer
	err = buf.PushBinaryMarshallable(c.GetHeader())
	if err != nil {
		return nil, err
	}

	// Write the Body out
	buf.Write(buf2.DeepCopyBytes())

	return buf.DeepCopyBytes(), err
}

// UnmarshalABlock unmarshals the input data to a new admin block
func UnmarshalABlock(data []byte) (interfaces.IAdminBlock, error) {
	block := NewAdminBlock(nil)
	err := block.UnmarshalBinary(data)
	if err != nil {
		return nil, err
	}

	return block, nil
}

// UnmarshalBinaryData unmarshals the input data to this admin block
func (c *AdminBlock) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)
	h := new(ABlockHeader)
	err := buf.PopBinaryMarshallable(h)
	if err != nil {
		return nil, err
	}
	c.Header = h

	// msgLimit is the theoretical maximum number of messages possible in the
	// admin block. The limit is the body size divided by the smallest possible
	// message size (2 bytes for a minute message {0x00, 0x0[0-9]})
	msgLimit := c.Header.GetBodySize() / 2
	msgCount := c.Header.GetMessageCount()
	if msgCount > msgLimit {
		return nil, fmt.Errorf(
			"Error: AdminBlock.UnmarshalBinary: message count %d is greater "+
				"than remaining space in buffer %d (uint underflow?)",
			msgCount, msgLimit,
		)
	}

	c.ABEntries = make([]interfaces.IABEntry, int(msgCount))
	for i := uint32(0); i < msgCount; i++ {
		t, err := buf.PeekByte()
		if err != nil {
			return nil, err
		}
		switch t {
		case constants.TYPE_MINUTE_NUM:
			c.ABEntries[i] = new(EndOfMinuteEntry)
		case constants.TYPE_DB_SIGNATURE:
			c.ABEntries[i] = new(DBSignatureEntry)
		case constants.TYPE_REVEAL_MATRYOSHKA:
			c.ABEntries[i] = new(RevealMatryoshkaHash)
		case constants.TYPE_ADD_MATRYOSHKA:
			c.ABEntries[i] = new(AddReplaceMatryoshkaHash)
		case constants.TYPE_ADD_SERVER_COUNT:
			c.ABEntries[i] = new(IncreaseServerCount)
		case constants.TYPE_ADD_FED_SERVER:
			c.ABEntries[i] = new(AddFederatedServer)
		case constants.TYPE_ADD_AUDIT_SERVER:
			c.ABEntries[i] = new(AddAuditServer)
		case constants.TYPE_REMOVE_FED_SERVER:
			c.ABEntries[i] = new(RemoveFederatedServer)
		case constants.TYPE_ADD_FED_SERVER_KEY:
			c.ABEntries[i] = new(AddFederatedServerSigningKey)
		case constants.TYPE_ADD_BTC_ANCHOR_KEY:
			c.ABEntries[i] = new(AddFederatedServerBitcoinAnchorKey)
		case constants.TYPE_SERVER_FAULT:
			c.ABEntries[i] = new(ServerFault)
		case constants.TYPE_COINBASE_DESCRIPTOR:
			c.ABEntries[i] = new(CoinbaseDescriptor)
		case constants.TYPE_COINBASE_DESCRIPTOR_CANCEL:
			c.ABEntries[i] = new(CancelCoinbaseDescriptor)
		case constants.TYPE_ADD_FACTOID_ADDRESS:
			c.ABEntries[i] = new(AddFactoidAddress)
		case constants.TYPE_ADD_FACTOID_EFFICIENCY:
			c.ABEntries[i] = new(AddEfficiency)
		default:
			// Undefined types are > 0x09 and are not defined yet, but we have placeholder code to deal with them.
			// This allows for future updates to the admin block with backwards compatibility
			fmt.Printf("AB UNDEFINED ENTRY %x for block %v. Using Forward Compatible holder\n", t, c.GetHeader().GetDBHeight())
			c.ABEntries[i] = new(ForwardCompatibleEntry)

			//panic("Undefined Admin Block Entry Type")
		}
		err = buf.PopBinaryMarshallable(c.ABEntries[i])
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinary unmarshals the input data into this admin block.
func (c *AdminBlock) UnmarshalBinary(data []byte) (err error) {
	_, err = c.UnmarshalBinaryData(data)
	return
}

// GetDBSignature returns the first DBSignature object in the admin block entry array
func (c *AdminBlock) GetDBSignature() interfaces.IABEntry {
	c.Init()
	for i := uint32(0); i < c.GetHeader().GetMessageCount(); i++ {
		if c.ABEntries[i].Type() == constants.TYPE_DB_SIGNATURE {
			return c.ABEntries[i]
		}
	}

	return nil
}

// JSONByte returns the json encoded byte array
func (c *AdminBlock) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(c)
}

// JSONString returns the json encoded string
func (c *AdminBlock) JSONString() (string, error) {
	return primitives.EncodeJSONString(c)
}

type ExpandedABlock AdminBlock

func (c AdminBlock) MarshalJSON() ([]byte, error) {
	backRefHash, err := c.BackReferenceHash()
	if err != nil {
		return nil, err
	}

	lookupHash, err := c.LookupHash()
	if err != nil {
		return nil, err
	}

	return json.Marshal(struct {
		ExpandedABlock
		BackReferenceHash string `json:"backreferencehash"`
		LookupHash        string `json:"lookuphash"`
	}{
		ExpandedABlock:    ExpandedABlock(c),
		BackReferenceHash: backRefHash.String(),
		LookupHash:        lookupHash.String(),
	})
}

/*********************************************************************
 * Support
 *********************************************************************/

// NewAdminBlock returns a new admin block which comes after the input admin block
func NewAdminBlock(prev interfaces.IAdminBlock) interfaces.IAdminBlock {
	block := new(AdminBlock)
	block.Init()
	if prev != nil {
		block.GetHeader().SetPrevBackRefHash(primitives.NewZeroHash())
		block.GetHeader().SetDBHeight(prev.GetDBHeight() + 1)
	} else {
		block.GetHeader().SetPrevBackRefHash(primitives.NewZeroHash())
	}
	return block
}

// CheckBlockPairIntegrity checks that two admin blocks are sequential with the previous admin block coming before the current
func CheckBlockPairIntegrity(block interfaces.IAdminBlock, prev interfaces.IAdminBlock) error {
	if block == nil {
		return fmt.Errorf("No block specified")
	}

	if prev == nil {
		if block.GetHeader().GetPrevBackRefHash().IsZero() == false {
			return fmt.Errorf("Invalid PrevBackRefHash")
		}
		if block.GetHeader().GetDBHeight() != 0 {
			return fmt.Errorf("Invalid DBHeight")
		}
	} else {
		if block.GetHeader().GetPrevBackRefHash().IsSameAs(prev.GetHash()) == false {
			return fmt.Errorf("Invalid PrevBackRefHash")
		}
		if block.GetHeader().GetDBHeight() != (prev.GetHeader().GetDBHeight() + 1) {
			return fmt.Errorf("Invalid DBHeight")
		}
	}

	return nil
}
