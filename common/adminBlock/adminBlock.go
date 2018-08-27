// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package adminBlock

import (
	"encoding/json"
	"fmt"
	"os"

	"sort"

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
	Header            interfaces.IABlockHeader      `json:"header"`
	ABEntries         []interfaces.IABEntry         `json:"abentries"` //Interface
	identityABEntries []interfaces.IIdentityABEntry // This is all identity related entries. They are sorted before added
}

var _ interfaces.IAdminBlock = (*AdminBlock)(nil)
var _ interfaces.Printable = (*AdminBlock)(nil)
var _ interfaces.BinaryMarshallableAndCopyable = (*AdminBlock)(nil)
var _ interfaces.DatabaseBatchable = (*AdminBlock)(nil)

func (c *AdminBlock) Init() {
	if c.Header == nil {
		h := new(ABlockHeader)
		h.Init()
		c.Header = h
		c.ABEntries = make([]interfaces.IABEntry, 0)
	}
}

func (a *AdminBlock) IsSameAs(b interfaces.IAdminBlock) bool {
	if !a.Header.IsSameAs(b.GetHeader()) {
		return false
	}
	if len(a.ABEntries) != len(b.GetABEntries()) {
		return false
	}
	return true
}

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

func (c *AdminBlock) FetchCoinbaseDescriptor() interfaces.IABEntry {
	for _, e := range c.ABEntries {
		if e.Type() == constants.TYPE_COINBASE_DESCRIPTOR {
			return e
		}
	}
	return nil
}

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

func (c *AdminBlock) AddFedServer(identityChainID interfaces.IHash) error {
	c.Init()
	if identityChainID == nil {
		return fmt.Errorf("No identityChainID provided")
	}

	entry := NewAddFederatedServer(identityChainID, c.GetHeader().GetDBHeight()+1) // Goes in the NEXT block
	return c.AddEntry(entry)
}

func (c *AdminBlock) AddAuditServer(identityChainID interfaces.IHash) error {
	c.Init()
	if identityChainID == nil {
		return fmt.Errorf("No identityChainID provided")
	}

	entry := NewAddAuditServer(identityChainID, c.GetHeader().GetDBHeight()+1) // Goes in the NEXT block
	return c.AddEntry(entry)
}

func (c *AdminBlock) RemoveFederatedServer(identityChainID interfaces.IHash) error {
	c.Init()
	if identityChainID == nil {
		return fmt.Errorf("No identityChainID provided")
	}

	entry := NewRemoveFederatedServer(identityChainID, c.GetHeader().GetDBHeight()+1) // Goes in the NEXT block
	return c.AddEntry(entry)
}

func (c *AdminBlock) AddCancelCoinbaseDescriptor(descriptorHeight, index uint32) error {
	c.Init()
	entry := NewCancelCoinbaseDescriptor(descriptorHeight, index)

	return c.AddIdentityEntry(entry)
}

// InsertIdentityABEntries will prepare the identity entries and add them into the adminblock
func (a *AdminBlock) InsertIdentityABEntries() error {
	sort.Sort(interfaces.IIdentityABEntrySort(a.identityABEntries))
	for _, v := range a.identityABEntries {
		err := a.AddEntry(v)
		if err != nil {
			return fmt.Errorf("No identityChainID provided")
		}
	}
	return nil
}

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

func (c *AdminBlock) AddCoinbaseDescriptor(outputs []interfaces.ITransAddress) error {
	c.Init()
	if outputs == nil {
		return fmt.Errorf("No outputs provided")
	}

	entry := NewCoinbaseDescriptor(outputs)
	return c.AddEntry(entry)
}

func (c *AdminBlock) AddEfficiency(chain interfaces.IHash, efficiency uint16) error {
	c.Init()
	if chain == nil {
		return fmt.Errorf("No chainid provided")
	}

	entry := NewAddEfficiency(chain, efficiency)
	return c.AddIdentityEntry(entry)
}

func (c *AdminBlock) AddCoinbaseAddress(chain interfaces.IHash, add interfaces.IAddress) error {
	c.Init()
	if chain == nil {
		return fmt.Errorf("No chainid provided")
	}

	entry := NewAddFactoidAddress(chain, add)
	return c.AddIdentityEntry(entry)
}

func (a *AdminBlock) AddIdentityEntry(entry interfaces.IIdentityABEntry) error {
	if entry == nil {
		return fmt.Errorf("No entry provided")
	}
	// These get sorted when you call the InsertIdentityABEntries function
	a.identityABEntries = append(a.identityABEntries, entry)
	return nil
}

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

func (c *AdminBlock) GetHeader() interfaces.IABlockHeader {
	c.Init()
	return c.Header
}

func (c *AdminBlock) SetHeader(header interfaces.IABlockHeader) {
	c.Header = header
}

func (c *AdminBlock) GetABEntries() []interfaces.IABEntry {
	return c.ABEntries
}

func (c *AdminBlock) GetDBHeight() uint32 {
	return c.GetHeader().GetDBHeight()
}

func (c *AdminBlock) SetABEntries(abentries []interfaces.IABEntry) {
	c.ABEntries = abentries
}

func (c *AdminBlock) New() interfaces.BinaryMarshallableAndCopyable {
	return NewAdminBlock(nil)
}

func (c *AdminBlock) GetDatabaseHeight() uint32 {
	return c.GetHeader().GetDBHeight()
}

func (c *AdminBlock) GetChainID() interfaces.IHash {
	return c.GetHeader().GetAdminChainID()
}

func (c *AdminBlock) DatabasePrimaryIndex() interfaces.IHash {
	key, _ := c.LookupHash()
	return key
}

func (c *AdminBlock) DatabaseSecondaryIndex() interfaces.IHash {
	key, _ := c.BackReferenceHash()
	return key
}

func (c *AdminBlock) GetHash() interfaces.IHash {
	h, _ := c.GetKeyMR()
	return h
}

func (c *AdminBlock) GetKeyMR() (interfaces.IHash, error) {
	return c.BackReferenceHash()
}

// Returns the SHA512Half hash for the admin block
func (b *AdminBlock) BackReferenceHash() (interfaces.IHash, error) {
	var binaryAB []byte
	binaryAB, err := b.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return primitives.Sha512Half(binaryAB), nil
}

// Returns the SHA256 hash for the admin block
func (b *AdminBlock) LookupHash() (interfaces.IHash, error) {
	var binaryAB []byte
	binaryAB, err := b.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return primitives.Sha(binaryAB), nil
}

// Add an Admin Block entry to the block
func (b *AdminBlock) AddABEntry(e interfaces.IABEntry) error {
	return b.AddEntry(e)
}

// Add an Admin Block entry to the start of the block entries
func (b *AdminBlock) AddFirstABEntry(e interfaces.IABEntry) (err error) {
	b.ABEntries = append(b.ABEntries, nil)
	copy(b.ABEntries[1:], b.ABEntries[:len(b.ABEntries)-1])
	b.ABEntries[0] = e
	return
}

// Write out the AdminBlock to binary.
func (b *AdminBlock) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "AdminBlock.MarshalBinary err:%v", *pe)
		}
	}(&err)
	b.Init()
	// Marshal all the entries into their own thing (need the size)
	var buf2 primitives.Buffer
	for _, v := range b.ABEntries {
		err := buf2.PushBinaryMarshallable(v)
		if err != nil {
			return nil, err
		}
	}

	b.GetHeader().SetMessageCount(uint32(len(b.ABEntries)))
	b.GetHeader().SetBodySize(uint32(buf2.Len()))

	var buf primitives.Buffer
	err = buf.PushBinaryMarshallable(b.GetHeader())
	if err != nil {
		return nil, err
	}

	// Write the Body out
	buf.Write(buf2.DeepCopyBytes())

	return buf.DeepCopyBytes(), err
}

func UnmarshalABlock(data []byte) (interfaces.IAdminBlock, error) {
	block := NewAdminBlock(nil)
	err := block.UnmarshalBinary(data)
	if err != nil {
		return nil, err
	}

	return block, nil
}

func (b *AdminBlock) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)
	h := new(ABlockHeader)
	err := buf.PopBinaryMarshallable(h)
	if err != nil {
		return nil, err
	}
	b.Header = h

	b.ABEntries = make([]interfaces.IABEntry, int(b.GetHeader().GetMessageCount()))
	for i := uint32(0); i < b.GetHeader().GetMessageCount(); i++ {
		t, err := buf.PeekByte()
		if err != nil {
			return nil, err
		}
		switch t {
		case constants.TYPE_MINUTE_NUM:
			b.ABEntries[i] = new(EndOfMinuteEntry)
		case constants.TYPE_DB_SIGNATURE:
			b.ABEntries[i] = new(DBSignatureEntry)
		case constants.TYPE_REVEAL_MATRYOSHKA:
			b.ABEntries[i] = new(RevealMatryoshkaHash)
		case constants.TYPE_ADD_MATRYOSHKA:
			b.ABEntries[i] = new(AddReplaceMatryoshkaHash)
		case constants.TYPE_ADD_SERVER_COUNT:
			b.ABEntries[i] = new(IncreaseServerCount)
		case constants.TYPE_ADD_FED_SERVER:
			b.ABEntries[i] = new(AddFederatedServer)
		case constants.TYPE_ADD_AUDIT_SERVER:
			b.ABEntries[i] = new(AddAuditServer)
		case constants.TYPE_REMOVE_FED_SERVER:
			b.ABEntries[i] = new(RemoveFederatedServer)
		case constants.TYPE_ADD_FED_SERVER_KEY:
			b.ABEntries[i] = new(AddFederatedServerSigningKey)
		case constants.TYPE_ADD_BTC_ANCHOR_KEY:
			b.ABEntries[i] = new(AddFederatedServerBitcoinAnchorKey)
		case constants.TYPE_SERVER_FAULT:
			b.ABEntries[i] = new(ServerFault)
		case constants.TYPE_COINBASE_DESCRIPTOR:
			b.ABEntries[i] = new(CoinbaseDescriptor)
		case constants.TYPE_COINBASE_DESCRIPTOR_CANCEL:
			b.ABEntries[i] = new(CancelCoinbaseDescriptor)
		case constants.TYPE_ADD_FACTOID_ADDRESS:
			b.ABEntries[i] = new(AddFactoidAddress)
		case constants.TYPE_ADD_FACTOID_EFFICIENCY:
			b.ABEntries[i] = new(AddEfficiency)
		default:
			// Undefined types are > 0x09 and are not defined yet, but we have placeholder code to deal with them.
			// This allows for future updates to the admin block with backwards compatibility
			fmt.Printf("AB UNDEFINED ENTRY %x for block %v. Using Forward Compatible holder\n", t, b.GetHeader().GetDBHeight())
			b.ABEntries[i] = new(ForwardCompatibleEntry)

			//panic("Undefined Admin Block Entry Type")
		}
		err = buf.PopBinaryMarshallable(b.ABEntries[i])
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), nil
}

// Read in the binary into the Admin block.
func (b *AdminBlock) UnmarshalBinary(data []byte) (err error) {
	_, err = b.UnmarshalBinaryData(data)
	return
}

// Read in the binary into the Admin block.
func (b *AdminBlock) GetDBSignature() interfaces.IABEntry {
	b.Init()
	for i := uint32(0); i < b.GetHeader().GetMessageCount(); i++ {
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

type ExpandedABlock AdminBlock

func (e AdminBlock) MarshalJSON() ([]byte, error) {
	backRefHash, err := e.BackReferenceHash()
	if err != nil {
		return nil, err
	}

	lookupHash, err := e.LookupHash()
	if err != nil {
		return nil, err
	}

	return json.Marshal(struct {
		ExpandedABlock
		BackReferenceHash string `json:"backreferencehash"`
		LookupHash        string `json:"lookuphash"`
	}{
		ExpandedABlock:    ExpandedABlock(e),
		BackReferenceHash: backRefHash.String(),
		LookupHash:        lookupHash.String(),
	})
}

/*********************************************************************
 * Support
 *********************************************************************/

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
