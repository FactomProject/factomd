// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"

	"github.com/FactomProject/factomd/common/messages/msgbase"
	log "github.com/sirupsen/logrus"
)

// Communicate a Directory Block State

type DBStateMsg struct {
	msgbase.MessageBase
	Timestamp interfaces.Timestamp

	//TODO: handle malformed DBStates!
	DirectoryBlock   interfaces.IDirectoryBlock
	AdminBlock       interfaces.IAdminBlock
	FactoidBlock     interfaces.IFBlock
	EntryCreditBlock interfaces.IEntryCreditBlock

	EBlocks []interfaces.IEntryBlock
	Entries []interfaces.IEBEntry

	SignatureList SigList

	//Not marshalled
	IgnoreSigs bool
	Sent       interfaces.Timestamp
	IsInDB     bool
	IsLast     bool // Flag from state.LoadDatabase() that this is the last saved block loaded at boot.
}

var _ interfaces.IMsg = (*DBStateMsg)(nil)

func (a *DBStateMsg) IsSameAs(b *DBStateMsg) bool {
	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()

	if b == nil {
		return false
	}

	if a.Timestamp.GetTimeMilli() != b.Timestamp.GetTimeMilli() {
		return false
	}

	ok, err := primitives.AreBinaryMarshallablesEqual(a.DirectoryBlock, b.DirectoryBlock)
	if err != nil || ok == false {
		return false
	}

	ok, err = primitives.AreBinaryMarshallablesEqual(a.AdminBlock, b.AdminBlock)
	if err != nil || ok == false {
		return false
	}

	ok, err = primitives.AreBinaryMarshallablesEqual(a.FactoidBlock, b.FactoidBlock)
	if err != nil || ok == false {
		return false
	}

	ok, err = primitives.AreBinaryMarshallablesEqual(a.EntryCreditBlock, b.EntryCreditBlock)
	if err != nil || ok == false {
		return false
	}

	if len(a.EBlocks) != len(b.EBlocks) || (len(a.Entries) != len(b.Entries)) {
		return false
	}

	for i := range a.EBlocks {
		ok, err = primitives.AreBinaryMarshallablesEqual(a.EBlocks[i], b.EBlocks[i])
		if err != nil || ok == false {
			return false
		}
	}

	for i := range a.Entries {
		ok, err = primitives.AreBinaryMarshallablesEqual(a.Entries[i], b.Entries[i])
		if err != nil || ok == false {
			return false
		}
	}

	return true
}

func (m *DBStateMsg) GetRepeatHash() interfaces.IHash {
	return m.DirectoryBlock.GetHash()
}

func (m *DBStateMsg) GetHash() interfaces.IHash {
	//	data, _ := m.MarshalBinary()
	//	return primitives.Sha(data)

	// These two calls do the same thing.  It should be standardized.
	return m.GetMsgHash()
}

func (m *DBStateMsg) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *DBStateMsg) Type() byte {
	return constants.DBSTATE_MSG
}

func (m *DBStateMsg) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  1   -- Message is valid
// NOTE! Do no return 0, that sticks this message in the holding map, vs the DBStateList
// 			ValidateSignatures is called when actually applying the DBState.
func (m *DBStateMsg) Validate(state interfaces.IState) int {
	// No matter what, a block has to have what a block has to have.
	if m.DirectoryBlock == nil || m.AdminBlock == nil || m.FactoidBlock == nil || m.EntryCreditBlock == nil {
		state.AddStatus(fmt.Sprintf("DBStateMsg.Validate() Fail  Doesn't have all the blocks"))
		//We need the basic block types
		return -1
	}

	if m.IsInDB {
		return 1
	}

	dbheight := m.DirectoryBlock.GetHeader().GetDBHeight()

	// Just accept the genesis block
	if dbheight == 0 {
		return 1
	}

	if state.GetNetworkID() != m.DirectoryBlock.GetHeader().GetNetworkID() {
		state.AddStatus(fmt.Sprintf("DBStateMsg.Validate() Fail  ht: %d Expecting NetworkID %x and found %x",
			dbheight, state.GetNetworkID(), m.DirectoryBlock.GetHeader().GetNetworkID()))
		//Wrong network ID
		return -1
	}

	// Difference of completed blocks, rather than just highest DBlock (might be missing entries)
	diff := int(dbheight) - (int(state.GetEntryDBHeightComplete()))

	// Look at saved heights if not too far from what we have saved.
	if diff < -1 {
		state.AddStatus(fmt.Sprintf("DBStateMsg.Validate() Fail dbstate dbht: %d Highest Saved %d diff %d",
			dbheight, state.GetEntryDBHeightComplete(), diff))
		return -1
	}

	if m.DirectoryBlock.GetHeader().GetNetworkID() == constants.MAIN_NETWORK_ID {
		key := constants.CheckPoints[dbheight]
		if key != "" {
			if key != m.DirectoryBlock.DatabasePrimaryIndex().String() {
				state.AddStatus(fmt.Sprintf("DBStateMsg.Validate() Fail  ht: %d checkpoint failure. Had %s Expected %s",
					dbheight, m.DirectoryBlock.DatabasePrimaryIndex().String(), key))
				//Key does not match checkpoint
				return -1
			}
		}
	}

	return 1
}

func (m *DBStateMsg) ValidateSignatures(state interfaces.IState) int {
	// Validate Signatures

	// If this is the next block that we need, we can validate it by signatures. If it is a past block
	// we can validate by prevKeyMr of the block that follows this one
	if m.DirectoryBlock.GetDatabaseHeight() == state.GetHighestSavedBlk()+1 {
		// Fed count of this height -1, as we may not have the height itself
		fedCount := len(state.GetFedServers(m.DirectoryBlock.GetDatabaseHeight()))
		tally := m.SigTally(state)
		if tally >= (fedCount/2 + 1) {
			// This has all the signatures it needs
			goto ValidSignatures
		} else {
			// If we remove servers, there will not be as many signatures. We need to accomadate for this
			// by reducing our needed. This will only get called if we fall short on signatures.
			aes := m.AdminBlock.GetABEntries()
			for _, adminEntry := range aes {
				switch adminEntry.Type() {
				case constants.TYPE_REMOVE_FED_SERVER:
					// Double check the entry is a real remove fed server message
					_, ok := adminEntry.(*adminBlock.RemoveFederatedServer)
					if !ok {
						continue
					}
					// Reduce our total fed servers
					fedCount--
				}
			}
			if tally >= (fedCount/2 + 1) {
				// This has all the signatures it needs
				goto ValidSignatures
			}
		}

		// It does not pass have enough signatures. It will not obtain more signatures.
		return -1
	} else { // Alternative to signatures passing by checking our DB

		// This is a future block. We cannot determine at this time
		if m.DirectoryBlock.GetDatabaseHeight() > state.GetHighestSavedBlk()+1 {
			return 0
		}

		// This block is not the next block we need. Check this block +1 and check it's prevKeyMr
		next := state.GetDirectoryBlockByHeight(m.DirectoryBlock.GetDatabaseHeight() + 1)
		if next == nil {
			// Do not have the next directory block, so we cannot tell by this method.
			// We should have this dblock though, so maybe we should return -1?
			return 0
		}

		// If the prevKeyMr of the next matches this one, we know it is valid.
		if next.GetHeader().GetPrevKeyMR().IsSameAs(m.DirectoryBlock.GetKeyMR()) {
			goto ValidSignatures
		} else {
			// The KeyMR does not match, this block is invalid
			return -1
		}
	}
ValidSignatures: // Goto here if signatures pass

	// ValidateData will ensure all the data given matches the DBlock
	return m.ValidateData(state)
}

// ValidateData will check the data attached to the DBState against the directory block it contains.
// This is ensure no additional junk is attached to a valid DBState
func (m *DBStateMsg) ValidateData(state interfaces.IState) int {
	// Checking the content of the DBState against the directoryblock contained
	// Map of Entries and Eblocks in this DBState dblock
	// A value of true indicates a repeat. Repeats are not enforce though
	eblocks := make(map[[32]byte]bool) //, len(m.EBlocks))
	ents := make(map[[32]byte]bool)    //, len(m.Entries))

	// Ensure blocks in the DBlock matches blocks in DBState
	for _, b := range m.DirectoryBlock.GetEBlockDBEntries() {
		switch {
		case bytes.Compare(b.GetChainID().Bytes(), constants.ADMIN_CHAINID) == 0:
			// Validate ABlock
			goodKeyMr, err := m.AdminBlock.GetKeyMR()
			if err != nil {
				return -1
			}
			if !b.GetKeyMR().IsSameAs(goodKeyMr) {
				return -1
			}
		case bytes.Compare(b.GetChainID().Bytes(), constants.FACTOID_CHAINID) == 0:
			// Validate FBlock
			if !b.GetKeyMR().IsSameAs(m.FactoidBlock.GetKeyMR()) {
				return -1
			}
		case bytes.Compare(b.GetChainID().Bytes(), constants.EC_CHAINID) == 0:
			// Validate ECBlock
			if !b.GetKeyMR().IsSameAs(m.EntryCreditBlock.GetHash()) {
				return -1
			}
		default: // EBLOCK
			// Eblocks in the DBlock. Not only check if the Eblocks in DBState list are good, but also entries
			eblocks[b.GetKeyMR().Fixed()] = false
		}
	}

	// Loop over eblocks and see if they fall in the map
	for _, eb := range m.EBlocks {
		keymr, err := eb.KeyMR()
		if err != nil {
			return -1
		}

		// If the eblock does not exist in our map, it doesn't exist in the directory block
		if _, ok := eblocks[keymr.Fixed()]; !ok {
			return -1
		}
		eblocks[keymr.Fixed()] = true

		for _, e := range eb.GetEntryHashes() {
			ents[e.Fixed()] = false
		}
	}

	for _, e := range m.Entries {
		// Although we can check for repeated entries, a directory block is allowed to contain duplicate entries
		// So just check if the entry is in the DBlock
		if _, ok := ents[e.GetHash().Fixed()]; !ok {
			return -1
		}
		ents[e.GetHash().Fixed()] = true
	}

	return 1
}

func (m *DBStateMsg) SigTally(state interfaces.IState) int {
	dbheight := m.DirectoryBlock.GetHeader().GetDBHeight()

	validSigCount := 0
	validSigCount += m.checkpointFix()

	data, err := m.DirectoryBlock.GetHeader().MarshalBinary()
	if err != nil {
		state.AddStatus(fmt.Sprint("Debug: DBState Signature Error, Marshal binary errored"))
		return validSigCount
	}

	// Signatures that are not valid by current fed list
	var remainingSig []interfaces.IFullSignature

	// If there is a repeat signature, we do not count it twice
	sigmap := make(map[string]bool)
	for _, sig := range m.SignatureList.List {
		// check expected signature
		if sigmap[fmt.Sprintf("%x", sig.GetSignature()[:])] {
			continue // Toss duplicate signatures
		}
		sigmap[fmt.Sprintf("%x", sig.GetSignature()[:])] = true
		check, err := state.VerifyAuthoritySignature(data, sig.GetSignature(), dbheight)
		if err == nil && check >= 0 {
			validSigCount++
			continue
		}
		// it was not the expected signature check the boot strap
		//Check signature against the Skeleton key
		authoritativeKey := state.GetNetworkBootStrapKey()
		if authoritativeKey != nil {
			if bytes.Compare(sig.GetKey(), authoritativeKey.Bytes()) == 0 {
				if sig.Verify(data) {
					validSigCount++
					continue
				}
			}
		}

		// save the unverified sig so we can check for leadership changes later on

		if sig.Verify(data) {
			remainingSig = append(remainingSig, sig)
		}
	}

	// If promotions have occurred this block, we need to account for their signatures to be
	// valid. We will only pay for this overhead if there are signatures left, meaning most blocks will
	// not enter this loop
	if len(remainingSig) > 0 {
		type tempAuthority struct {
			Key      []byte
			Promoted bool
		}

		// Map of new federated servers, and their keys
		newSigners := make(map[string]*tempAuthority)

		// Search for promotions and their keys and populate the map:
		aes := m.AdminBlock.GetABEntries()
		for _, adminEntry := range aes {
			switch adminEntry.Type() {
			case constants.TYPE_ADD_FED_SERVER:
				// New federated server that can sign blocks
				r, ok := adminEntry.(*adminBlock.AddFederatedServer)
				if !ok {
					// This shouldn't fail, as we checked the type
					continue
				}

				if _, ok := newSigners[r.IdentityChainID.String()]; !ok {
					newSigners[r.IdentityChainID.String()] = new(tempAuthority)
				}
				newSigners[r.IdentityChainID.String()].Promoted = true
			case constants.TYPE_REMOVE_FED_SERVER:
				// A remove will remove this server from our list of servers that can sign,
				// but we need them to still be valid, so we will allow them to still sign
				r, ok := adminEntry.(*adminBlock.RemoveFederatedServer)
				if !ok {
					// This shouldn't fail, as we checked the type
					continue
				}

				if _, ok := newSigners[r.IdentityChainID.String()]; !ok {
					newSigners[r.IdentityChainID.String()] = new(tempAuthority)
				}
				newSigners[r.IdentityChainID.String()].Promoted = true
			case constants.TYPE_ADD_FED_SERVER_KEY:
				r, ok := adminEntry.(*adminBlock.AddFederatedServerSigningKey)
				if !ok {
					// This shouldn't fail, as we checked the type
					continue
				}

				// We need to grab the signing key from the admin block if provided
				if _, ok := newSigners[r.IdentityChainID.String()]; !ok {
					newSigners[r.IdentityChainID.String()] = new(tempAuthority)
				}
				keybytes, err := r.PublicKey.MarshalBinary()
				if err != nil {
					continue
				}
				newSigners[r.IdentityChainID.String()].Key = keybytes
			}
		}

		// There is a chance their signing keys won't be located in the admin block. If they came from being an audit server,
		// their key might be in the authority list, or identity list.
		for i, v := range newSigners {
			if v.Key != nil {
				continue
			}

			idHash, err := primitives.HexToHash(i)
			if err != nil {
				continue
			}
			signingkey, status := state.GetSigningKey(idHash)
			if status >= 0 {
				v.Key = signingkey.Bytes()
			}
		}

		// These signatures that did not validate with current set of authorities
		for _, sig := range remainingSig {
		InnerSingerLoop:
			for _, signer := range newSigners {
				if bytes.Compare(sig.GetKey(), signer.Key) == 0 {
					validSigCount++
					break InnerSingerLoop
				}
			}
		}
	}

	return validSigCount
}

func (m *DBStateMsg) checkpointFix() int {
	returnAmt := 0
	dbheight := m.DirectoryBlock.GetDatabaseHeight()

	allow := func(str string, siglist SigList) int {
		amt := 0
		goodSig, _ := hex.DecodeString(str)
		for _, s := range siglist.List {
			if bytes.Compare(s.Bytes(), goodSig) == 0 {
				amt++
				break
			}
		}
		return amt
	}

	switch dbheight {
	case 75893:
		returnAmt += allow("8066fc4222eff67470ffaca15bdb5d6d15b65daf3cc86c121b872d7485b388b3cb4b7bbbd0248076065262d54699bab68e7d5be96e137aa3428b903916e4180a", m.SignatureList)
	case 76720:
		returnAmt += allow("ab429576ee93485cfffe0c778d429073f24ce76d3014f2ddecd6e90e87a5e912b849842597cae23a66beee203ee455bd44fe4073747ce6c099a21f4525c3d901", m.SignatureList)
	case 76792:
		returnAmt += allow("9f86122d624400b3036e60105f3db4e99199ae9217cbeb1462811426319983dc0e4f5e5cd16996cc3cf2940ead765ce00fc699e23b459395569c10e1df4c650b", m.SignatureList)
	case 87624:
		returnAmt += allow("cb67b7f8ed2b2845b9941264f3631a639685e2b7a47b8c353461ee9b197c2307aaab27419b72f5478f0e2a0d610ef4f16cbfaa5e9889c114415a63b8cb54f000", m.SignatureList)
	default:
		return 0
	}

	return returnAmt
}

func (m *DBStateMsg) ComputeVMIndex(state interfaces.IState) {}

// Execute the leader functions of the given message
func (m *DBStateMsg) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

func (m *DBStateMsg) FollowerExecute(state interfaces.IState) {
	state.FollowerExecuteDBState(m)
}

// Acknowledgements do not go into the process list.
func (e *DBStateMsg) Process(dbheight uint32, state interfaces.IState) bool {
	panic("DBStatemsg should never have its Process() method called")
}

func (e *DBStateMsg) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *DBStateMsg) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (m *DBStateMsg) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling Directory Block State Message: %v", r)
		}
	}()

	newData = data
	if newData[0] != m.Type() {
		return nil, fmt.Errorf("Invalid Message type")
	}
	newData = newData[1:]

	m.Peer2Peer = true

	m.Timestamp = new(primitives.Timestamp)
	newData, err = m.Timestamp.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.DirectoryBlock = new(directoryBlock.DirectoryBlock)
	newData, err = m.DirectoryBlock.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.AdminBlock = new(adminBlock.AdminBlock)
	newData, err = m.AdminBlock.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.FactoidBlock = new(factoid.FBlock)
	newData, err = m.FactoidBlock.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.EntryCreditBlock = entryCreditBlock.NewECBlock()
	newData, err = m.EntryCreditBlock.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	eBlockCount, newData := binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	for i := uint32(0); i < eBlockCount; i++ {
		eBlock := entryBlock.NewEBlock()
		newData, err = eBlock.UnmarshalBinaryData(newData)
		if err != nil {
			panic(err.Error())
		}
		m.EBlocks = append(m.EBlocks, eBlock)
	}

	entryCount, newData := binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	for i := uint32(0); i < entryCount; i++ {
		var entrySize uint32
		entrySize, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
		entry := entryBlock.NewEntry()
		newData, err = newData[int(entrySize):], entry.UnmarshalBinary(newData[:int(entrySize)])
		if err != nil {
			panic(err.Error())
		}
		m.Entries = append(m.Entries, entry)
	}

	newData, err = m.SignatureList.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	return
}

func (m *DBStateMsg) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *DBStateMsg) MarshalBinary() ([]byte, error) {
	var buf primitives.Buffer

	binary.Write(&buf, binary.BigEndian, m.Type())

	t := m.GetTimestamp()
	data, err := t.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	data, err = m.DirectoryBlock.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	data, err = m.AdminBlock.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	data, err = m.FactoidBlock.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	data, err = m.EntryCreditBlock.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	eBlockCount := uint32(len(m.EBlocks))
	binary.Write(&buf, binary.BigEndian, eBlockCount)
	for _, eb := range m.EBlocks {
		bin, err := eb.MarshalBinary()
		if err != nil {
			return nil, err
		}
		buf.Write(bin)
	}

	entryCount := uint32(len(m.Entries))
	binary.Write(&buf, binary.BigEndian, entryCount)
	for _, e := range m.Entries {
		bin, err := e.MarshalBinary()
		if err != nil || bin == nil || len(bin) == 0 {
			return nil, err
		}
		entrySize := uint32(len(bin))
		binary.Write(&buf, binary.BigEndian, entrySize)
		buf.Write(bin)
	}

	if d, err := m.SignatureList.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	return buf.DeepCopyBytes(), nil
}

func (m *DBStateMsg) String() string {
	data, _ := m.MarshalBinary()
	return fmt.Sprintf("DBState: dbht:%3d [size: %11s] dblock %6x admin %6x fb %6x ec %6x hash %6x",
		m.DirectoryBlock.GetHeader().GetDBHeight(),
		primitives.AddCommas(int64(len(data))),
		m.DirectoryBlock.GetKeyMR().Bytes()[:3],
		m.AdminBlock.GetHash().Bytes()[:3],
		m.FactoidBlock.GetHash().Bytes()[:3],
		m.EntryCreditBlock.GetHash().Bytes()[:3],
		m.GetHash().Bytes()[:3])
}

func (m *DBStateMsg) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "dbstate",
		"dbheight":    m.DirectoryBlock.GetHeader().GetDBHeight(),
		"dblockhash":  m.DirectoryBlock.GetKeyMR().String(),
		"ablockhash":  m.AdminBlock.GetHash().String(),
		"fblockhash":  m.FactoidBlock.GetHash().String(),
		"ecblockhash": m.EntryCreditBlock.GetHash().String(),
		"hash":        m.GetHash().String()}
}

func NewDBStateMsg(timestamp interfaces.Timestamp,
	d interfaces.IDirectoryBlock,
	a interfaces.IAdminBlock,
	f interfaces.IFBlock,
	e interfaces.IEntryCreditBlock,
	eBlocks []interfaces.IEntryBlock,
	entries []interfaces.IEBEntry,
	sigList []interfaces.IFullSignature) interfaces.IMsg {
	msg := new(DBStateMsg)
	msg.NoResend = true

	msg.Peer2Peer = true

	msg.Timestamp = timestamp

	msg.DirectoryBlock = d
	msg.AdminBlock = a
	msg.FactoidBlock = f
	msg.EntryCreditBlock = e

	msg.EBlocks = eBlocks
	msg.Entries = entries

	sl := new(SigList)
	sl.Length = uint32(len(sigList))
	sl.List = sigList

	msg.SignatureList = *sl

	return msg
}
