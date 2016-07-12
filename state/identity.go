package state

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	ed "github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/log"
)

var (
	TWELVE_HOURS_S uint64 = 12 * 60 * 60
	// Time window for identity to require registration: 24hours = 144 blocks
	TIME_WINDOW uint32 = 144
)

type AnchorSigningKey struct {
	BlockChain string
	KeyLevel   byte
	KeyType    byte
	SigningKey []byte //if bytes, it is hex
}

type Identity struct {
	IdentityChainID      interfaces.IHash
	IdentityRegistered   uint32
	IdentityCreated      uint32
	ManagementChainID    interfaces.IHash
	ManagementRegistered uint32
	ManagementCreated    uint32
	MatryoshkaHash       interfaces.IHash
	Key1                 interfaces.IHash
	Key2                 interfaces.IHash
	Key3                 interfaces.IHash
	Key4                 interfaces.IHash
	SigningKey           interfaces.IHash
	Status               int
	AnchorKeys           []AnchorSigningKey
}

var _ interfaces.Printable = (*Identity)(nil)

func (e *Identity) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *Identity) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *Identity) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *Identity) String() string {
	str, _ := e.JSONString()
	return str
}

func LoadIdentityCache(st *State) {
	blockHead, err := st.DB.FetchDirectoryBlockHead()

	if blockHead == nil {
		// new block chain just created.  no id yet
		return
	}
	bHeader := blockHead.GetHeader()
	height := bHeader.GetDBHeight()

	if err != nil {
		log.Printfln("Identity Error:", err)
	}

	var i uint32
	for i = 1; i < height; i++ {
		dblk, err := st.DB.FetchDBlockByHeight(i)
		if err != nil && dblk != nil {
			LoadIdentityByDirectoryBlock(dblk, st, false)
		}
	}

}

func LoadIdentityByDirectoryBlock(dblk interfaces.IDirectoryBlock, st *State, update bool) {
	var ManagementChain interfaces.IHash
	var err error
	if dblk == nil {
		log.Println("DEBUG: Identity Error, DBlock nil, disregard")
		return
	}
	height := dblk.GetDatabaseHeight()
	// TODO: Remove 1 block wait
	/*dblk, err = st.DB.FetchDBlockByHeight(height - 1)
	if err != nil {
		return
	}*/

	ManagementChain, _ = primitives.HexToHash("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")

	entries := dblk.GetDBEntries()

	// Used to hold key update entries to process at end of entry block
	holdEntry := make([]interfaces.IEBEntry, 0)

	for _, eBlk := range entries {
		cid := eBlk.GetChainID()
		if cid == nil {
			continue
		}
		if cid.IsSameAs(ManagementChain) {
			// is it a new one?
			entkmr := eBlk.GetKeyMR() //eBlock Hash
			ecb, err := st.DB.FetchEBlock(entkmr)
			if err != nil || ecb == nil {
				fmt.Println("DEBUG: ECB is nil")
				continue
			}
			entryHashes := ecb.GetEntryHashes()
			for _, eHash := range entryHashes {
				hs := eHash.String()
				if hs[0:10] != "0000000000" { //ignore minute markers
					ent, err := st.DB.FetchEntry(eHash)
					if err != nil {
						continue
					}
					if len(ent.ExternalIDs()) > 2 {
						// This is the Register Factom Identity Message
						if string(ent.ExternalIDs()[1]) == "Register Factom Identity" {
							registerFactomIdentity(ent.ExternalIDs(), cid, st, height)
						}
					}
				}
			}
		} else if cid.String()[0:6] == "888888" {
			entkmr := eBlk.GetKeyMR() //eBlock Hash
			ecb, err := st.DB.FetchEBlock(entkmr)
			if err != nil || ecb == nil {
				fmt.Println("DEBUG: ECB is nil")
				continue
			}
			entryHashes := ecb.GetEntryHashes()
			for _, eHash := range entryHashes {
				hs := eHash.String()
				if hs[0:10] != "0000000000" { //ignore minute markers
					ent, _ := st.DB.FetchEntry(eHash)
					if string(ent.ExternalIDs()[1]) == "Register Server Management" {
						// this is an Identity that should have been registered already with a 0 status.
						//  this registers it with the management chain.  Now it can be assigned as federated or audit.
						//  set it to status 6 - Pending Full
						registerIdentityAsServer(ent.ExternalIDs(), cid, st, height)
					} else if string(ent.ExternalIDs()[1]) == "New Block Signing Key" {
						// this is the Signing Key for this Identity
						if len(ent.ExternalIDs()) == 7 { // update management should have 4 items
							// Hold
							holdEntry = append(holdEntry, ent)
						}

					} else if string(ent.ExternalIDs()[1]) == "New Bitcoin Key" {
						// this is the Signing Key for this Identity
						if len(ent.ExternalIDs()) == 9 { // update management should have 4 items
							// Hold
							holdEntry = append(holdEntry, ent)
						}

					} else if string(ent.ExternalIDs()[1]) == "New Matryoshka Hash" {
						// this is the Signing Key for this Identity
						if len(ent.ExternalIDs()) == 7 { // update management should have 4 items
							// hold
							holdEntry = append(holdEntry, ent)
						}
					} else if len(ent.ExternalIDs()) > 1 && string(ent.ExternalIDs()[1]) == "Identity Chain" {
						// this is a new identity
						addIdentity(ent.ExternalIDs(), cid, st, height)
					} else if len(ent.ExternalIDs()) > 1 && string(ent.ExternalIDs()[1]) == "Server Management" {
						// this is a new identity
						if len(ent.ExternalIDs()) == 4 {
							// update management should have 4 items
							updateManagementKey(ent.ExternalIDs(), cid, st, height)
						}
					}
				}
			}
		} else {
			if cid.String()[0:10] != "0000000000" { //ignore minute markers
				//  not a chain id I care about
			}
		}
	}
	// Process entries that are being held
	if len(holdEntry) > 0 {
		for _, entry := range holdEntry {
			if string(entry.ExternalIDs()[1]) == "New Block Signing Key" {
				if len(entry.ExternalIDs()) == 7 {
					registerBlockSigningKey(entry.ExternalIDs(), entry.GetChainID(), st, update)
				}
			} else if string(entry.ExternalIDs()[1]) == "New Bitcoin Key" {
				if len(entry.ExternalIDs()) == 9 {
					registerAnchorSigningKey(entry.ExternalIDs(), entry.GetChainID(), st, "BTC", update)
				}
			} else if string(entry.ExternalIDs()[1]) == "New Matryoshka Hash" {
				if len(entry.ExternalIDs()) == 7 {
					updateMatryoshkaHash(entry.ExternalIDs(), entry.GetChainID(), st, update)
				}
			}
		}
	}

	if err != nil {
		log.Printfln("Identity Error:", err)
	}
	// Remove Stale Identities
	// if an identity has taken more than 72 blocks (12 hours) to be fully created, remove it from the state identity list.
	var i int
	for i = 0; i < len(st.Identities); i++ {
		cutoff := height - TIME_WINDOW
		if TIME_WINDOW > height {
			cutoff = 0
		}
		if st.Identities[i].IdentityCreated < cutoff && st.Identities[i].IdentityRegistered == 0 && st.Identities[i].IdentityCreated != 0 {
			removeIdentity(i, st)
		} else if st.Identities[i].IdentityRegistered < cutoff && st.Identities[i].IdentityCreated == 0 && st.Identities[i].IdentityRegistered != 0 {
			removeIdentity(i, st)
		} else if st.Identities[i].ManagementCreated < cutoff && st.Identities[i].ManagementRegistered == 0 && st.Identities[i].ManagementCreated != 0 {
			removeIdentity(i, st)
		} else if st.Identities[i].ManagementRegistered < cutoff && st.Identities[i].ManagementCreated == 0 && st.Identities[i].ManagementRegistered != 0 {
			removeIdentity(i, st)
		}
	}
}

func removeIdentity(i int, st *State) {
	fmt.Println("Stale ID Removed")
	var newIDs []Identity
	newIDs = make([]Identity, len(st.Identities)-1)
	var j int
	for j = 0; j < i; j++ {
		newIDs[j] = st.Identities[j]
	}
	// skip removed Identity
	for j = i + 1; j < len(newIDs); j++ {
		newIDs[j-1] = st.Identities[j]
	}
	st.Identities = newIDs
}

func isIdentityChain(cid interfaces.IHash, ids []Identity) int {
	//is this an identity chain
	for i, identityChain := range ids {
		if identityChain.IdentityChainID.IsSameAs(cid) {
			return i
		}
	}

	// or is it an identity management subchain
	for i, identityChain := range ids {
		// might not have been filled in yet
		if identityChain.ManagementChainID != nil {
			if identityChain.ManagementChainID.IsSameAs(cid) {
				return i
			}
		}
	}
	return -1
}

func createFactomIdentity(st *State, chainID interfaces.IHash) int {
	var idnew []Identity
	idnew = make([]Identity, len(st.Identities)+1)

	var oneID Identity

	for i := 0; i < len(st.Identities); i++ {
		idnew[i] = st.Identities[i]
	}
	oneID.IdentityChainID = chainID

	oneID.Status = constants.IDENTITY_PENDING
	oneID.IdentityRegistered = 0
	oneID.IdentityCreated = 0
	oneID.ManagementRegistered = 0
	oneID.ManagementCreated = 0

	idnew[len(st.Identities)] = oneID

	st.Identities = idnew
	return len(st.Identities) - 1
}

func registerFactomIdentity(extIDs [][]byte, chainID interfaces.IHash, st *State, height uint32) {
	if bytes.Compare([]byte{0x00}, extIDs[0]) != 0 || // Version
		!CheckLength(24, extIDs[1]) || // Ascii
		!CheckLength(32, extIDs[2]) || // ID Chain
		!CheckLength(33, extIDs[3]) || // Preimage
		!CheckLength(64, extIDs[4]) { // Signiture
		log.Println("Identity Error Register Identity: Invalid external ID length")
		return
	}

	// find the Identity index from the chain id in the external id.  add this chainID as the management id
	idChain := primitives.NewHash(extIDs[2])
	IdentityIndex := isIdentityChain(idChain, st.Identities)
	if IdentityIndex == -1 {
		IdentityIndex = createFactomIdentity(st, idChain)
	}

	sigmsg, err := AppendExtIDs(extIDs, 0, 2)
	if err != nil {
		log.Printfln("Identity Error:", err)
		return
	} else {
		// Verify Signature
		idKey := st.Identities[IdentityIndex].Key1
		if CheckSig(idKey, extIDs[3][1:33], sigmsg, extIDs[4]) {
			st.Identities[IdentityIndex].ManagementRegistered = height
			checkIdentityInitialStatus(IdentityIndex, st)
		} else {
			log.Println("New Management Chain Register for identity [" + chainID.String()[:10] + "] is invalid. Bad signiture")
		}

	}
	st.Identities[IdentityIndex].IdentityRegistered = height

	checkIdentityInitialStatus(IdentityIndex, st)

	//}
}

func addIdentity(extIDs [][]byte, chainID interfaces.IHash, st *State, height uint32) {
	if bytes.Compare([]byte{0x00}, extIDs[0]) != 0 || // Version
		!CheckLength(14, extIDs[1]) || // Ascii
		!CheckLength(32, extIDs[2]) || // ID Key
		!CheckLength(32, extIDs[3]) || // ID Key
		!CheckLength(32, extIDs[4]) || // ID Key
		!CheckLength(32, extIDs[5]) || // ID Key
		extIDs[6] == nil { // Nonce
		log.Println("Identity Error Create Identity Chain: Invalid external ID length")
		return
	}

	IdentityIndex := isIdentityChain(chainID, st.Identities)

	if IdentityIndex == -1 {
		IdentityIndex = createFactomIdentity(st, chainID)
	}
	h := primitives.NewHash(extIDs[2])
	st.Identities[IdentityIndex].Key1 = h
	h = primitives.NewHash(extIDs[3])
	st.Identities[IdentityIndex].Key2 = h
	h = primitives.NewHash(extIDs[4])
	st.Identities[IdentityIndex].Key3 = h
	h = primitives.NewHash(extIDs[5])
	st.Identities[IdentityIndex].Key4 = h
	st.Identities[IdentityIndex].IdentityCreated = height

	checkIdentityInitialStatus(IdentityIndex, st)
}

func checkIdentityInitialStatus(IdentityIndex int, st *State) {
	// if all needed information is ready for the Identity , set it to IDENTITY_FULL
	dif := st.Identities[IdentityIndex].IdentityCreated - st.Identities[IdentityIndex].IdentityRegistered
	if dif < 0 {
		dif = -dif
	}
	if dif < TIME_WINDOW {
		st.Identities[IdentityIndex].Status = constants.IDENTITY_FULL
	}
}

func updateManagementKey(extIDs [][]byte, chainID interfaces.IHash, st *State, height uint32) {
	if bytes.Compare([]byte{0x00}, extIDs[0]) != 0 || // Version
		!CheckLength(17, extIDs[1]) || // Ascii
		!CheckLength(32, extIDs[2]) || // ID Chain
		extIDs[3] == nil { // Nonce
		log.Println("Identity Error Create Management: Invalid external ID length")
		return
	}
	idChain := primitives.NewHash(extIDs[2])
	IdentityIndex := isIdentityChain(idChain, st.Identities)
	if IdentityIndex == -1 {
		IdentityIndex = createFactomIdentity(st, idChain)
	}

	st.Identities[IdentityIndex].ManagementChainID = chainID
	st.Identities[IdentityIndex].ManagementCreated = height

	checkIdentityInitialStatus(IdentityIndex, st)
}

func registerIdentityAsServer(extIDs [][]byte, chainID interfaces.IHash, st *State, height uint32) {
	if bytes.Compare([]byte{0x00}, extIDs[0]) != 0 || // Version
		!CheckLength(26, extIDs[1]) || // Ascii
		!CheckLength(32, extIDs[2]) || // Sub ID Chain
		!CheckLength(33, extIDs[3]) || // Preimage
		!CheckLength(64, extIDs[4]) { // Signiture
		log.Println("Identity Error Register Identity: Invalid external ID length")
		return
	}
	idChain := primitives.NewHash(extIDs[2])
	IdentityIndex := isIdentityChain(idChain, st.Identities)
	if IdentityIndex == -1 {
		IdentityIndex = createFactomIdentity(st, idChain)
	}

	sigmsg, err := AppendExtIDs(extIDs, 0, 2)
	if err != nil {
		log.Printfln("Identity Error:", err)
	} else {
		// Verify Signature
		idKey := st.Identities[IdentityIndex].Key1
		if CheckSig(idKey, extIDs[3][1:33], sigmsg, extIDs[4]) {
			st.Identities[IdentityIndex].ManagementRegistered = height
			checkIdentityInitialStatus(IdentityIndex, st)
		} else {
			log.Println("New Management Chain Register for identity [" + chainID.String()[:10] + "] is invalid. Bad signiture")
		}

	}
	checkIdentityInitialStatus(IdentityIndex, st)
}

func registerBlockSigningKey(extIDs [][]byte, subChainID interfaces.IHash, st *State, update bool) {
	if bytes.Compare([]byte{0x00}, extIDs[0]) != 0 || // Version
		!CheckLength(21, extIDs[1]) || // Ascii
		!CheckLength(32, extIDs[2]) || // ID Chain
		!CheckLength(32, extIDs[3]) || // New Key
		!CheckLength(8, extIDs[4]) || // Timestamp
		!CheckLength(33, extIDs[5]) || // Preimage
		!CheckLength(64, extIDs[6]) { // Signiture
		log.Println("Identity Error Block Signing Key: Invalid external ID length")
		return
	}

	chainID := new(primitives.Hash)
	chainID.SetBytes(extIDs[2][:32])

	IdentityIndex := isIdentityChain(chainID, st.Identities)
	if IdentityIndex == -1 {
		log.Println("Identity Error: This cannot happen. New block signing key to nonexistent identity")
		return
	}

	if !st.Identities[IdentityIndex].ManagementChainID.IsSameAs(subChainID) {
		log.Println("Identity Error: Entry was not placed in the correct management chain")
		return
	}

	sigmsg, err := AppendExtIDs(extIDs, 0, 4)
	if err != nil {
		log.Printfln("Identity Error:", err)
	} else {
		//verify Signature
		idKey := st.Identities[IdentityIndex].Key1
		if CheckSig(idKey, extIDs[5][1:33], sigmsg, extIDs[6]) {
			// Check block key length
			if len(extIDs[3]) != 32 {
				log.Println("New Block Signing key for identity [" + chainID.String()[:10] + "] is invalid length")
				return
			}
			// Check timestamp of message
			if !CheckTimestamp(extIDs[4]) {
				log.Println("New Block Signing key for identity [" + chainID.String()[:10] + "] timestamp is too old")
				return
			}

			st.Identities[IdentityIndex].SigningKey = primitives.NewHash(extIDs[3])
			// Add to admin block
			status := st.Identities[IdentityIndex].Status
			if update && (status == constants.IDENTITY_FEDERATED_SERVER ||
				status == constants.IDENTITY_AUDIT_SERVER ||
				status == constants.IDENTITY_PENDING_FEDERATED_SERVER ||
				status == constants.IDENTITY_PENDING_AUDIT_SERVER) {
				if st.LeaderPL.VMIndexFor(constants.ADMIN_CHAINID) == st.GetLeaderVM() {
					fmt.Println("DEBUG: Change Server Key Message Sent")
					key := primitives.NewHash(extIDs[3])
					msg := messages.NewChangeServerKeyMsg(st, chainID, constants.TYPE_ADD_FED_SERVER_KEY, 0, 0, key)
					fmt.Println(msg.GetVMIndex())
					st.InMsgQueue() <- msg
				}
				//st.LeaderPL.AdminBlock.AddFederatedServerSigningKey(chainID, &key)
			}
		} else {
			log.Println("New Block Signing key for identity [" + chainID.String()[:10] + "] is invalid. Bad signiture")
		}
	}

}

func updateMatryoshkaHash(extIDs [][]byte, subChainID interfaces.IHash, st *State, update bool) {
	if bytes.Compare([]byte{0x00}, extIDs[0]) != 0 || // Version
		!CheckLength(19, extIDs[1]) || // Ascii
		!CheckLength(32, extIDs[2]) || // ID Chain
		!CheckLength(32, extIDs[3]) || // MHash
		!CheckLength(8, extIDs[4]) || // Timestamp
		!CheckLength(33, extIDs[5]) || // Preimage
		!CheckLength(64, extIDs[6]) { // Signiture
		log.Println("Identity Error MHash: Invalid external ID length")
		return
	}
	chainID := new(primitives.Hash)
	chainID.SetBytes(extIDs[2][:32])

	IdentityIndex := isIdentityChain(chainID, st.Identities)
	if IdentityIndex == -1 {
		log.Println("Identity Error: This cannot happen. New Matryoshka Hash to nonexistent identity")
		return
	}

	if !st.Identities[IdentityIndex].ManagementChainID.IsSameAs(subChainID) {
		log.Println("Identity Error: Entry was not placed in the correct management chain")
		return
	}

	sigmsg, err := AppendExtIDs(extIDs, 0, 4)
	if err != nil {
		log.Printfln("Identity Error:", err)
	} else {
		// Verify Signature
		idKey := st.Identities[IdentityIndex].Key1
		if CheckSig(idKey, extIDs[5][1:33], sigmsg, extIDs[6]) {
			// Check MHash length
			if len(extIDs[3]) != 32 {
				log.Println("New Matryoshka Hash for identity [" + chainID.String()[:10] + "] is invalid length")
				return
			}
			// Check Timestamp of message
			if !CheckTimestamp(extIDs[4]) {
				log.Println("New Matryoshka Hash for identity [" + chainID.String()[:10] + "] timestamp is too old")
				return
			}
			mhash := primitives.NewHash(extIDs[3])
			st.Identities[IdentityIndex].MatryoshkaHash = mhash
			// Add to admin block
			status := st.Identities[IdentityIndex].Status
			if update && (status == constants.IDENTITY_FEDERATED_SERVER ||
				status == constants.IDENTITY_AUDIT_SERVER ||
				status == constants.IDENTITY_PENDING_FEDERATED_SERVER ||
				status == constants.IDENTITY_PENDING_AUDIT_SERVER) {
				if st.LeaderPL.VMIndexFor(constants.ADMIN_CHAINID) == st.GetLeaderVM() {
					msg := messages.NewChangeServerKeyMsg(st, chainID, constants.TYPE_ADD_FED_SERVER_KEY, 0, 0, mhash)
					st.InMsgQueue() <- msg
				}
				//st.LeaderPL.AdminBlock.AddMatryoshkaHash(chainID, mhash)
			}
		} else {
			log.Println("New Matryoshka Hash for identity [" + chainID.String()[:10] + "] is invalid. Bad signiture")
		}

	}
}

func registerAnchorSigningKey(extIDs [][]byte, subChainID interfaces.IHash, st *State, BlockChain string, update bool) {
	if bytes.Compare([]byte{0x00}, extIDs[0]) != 0 || // Version
		!CheckLength(15, extIDs[1]) || // Ascii
		!CheckLength(32, extIDs[2]) || // ID Chain
		!CheckLength(1, extIDs[3]) || // Key Level
		!CheckLength(1, extIDs[4]) || // Key Type
		!CheckLength(20, extIDs[5]) || // Key
		!CheckLength(8, extIDs[6]) || // Timestamp
		!CheckLength(33, extIDs[7]) || // Preimage
		!CheckLength(64, extIDs[8]) { // Signiture
		log.Println("Identity Error Anchor Key: Invalid external ID length")
		return
	}

	chainID := new(primitives.Hash)
	chainID.SetBytes(extIDs[2][:32])

	IdentityIndex := isIdentityChain(chainID, st.Identities)
	if IdentityIndex == -1 {
		log.Println("Identity Error: This cannot happen. New Bitcoin Key to nonexistent identity")
		return
	}

	if !st.Identities[IdentityIndex].ManagementChainID.IsSameAs(subChainID) {
		log.Println("Identity Error: Entry was not placed in the correct management chain")
		return
	}

	var ask []AnchorSigningKey
	var newAsk []AnchorSigningKey
	var oneAsk AnchorSigningKey

	ask = st.Identities[IdentityIndex].AnchorKeys
	newAsk = make([]AnchorSigningKey, len(ask)+1)

	oneAsk.BlockChain = BlockChain
	oneAsk.KeyLevel = extIDs[3][0]
	oneAsk.KeyType = extIDs[4][0]
	oneAsk.SigningKey = extIDs[5]

	contains := false
	for i := 0; i < len(ask); i++ {
		if ask[i].KeyLevel == oneAsk.KeyLevel &&
			strings.Compare(ask[i].BlockChain, oneAsk.BlockChain) == 0 {
			contains = true
			ask[i] = oneAsk
		} else {
			newAsk[i] = ask[i]
		}
	}

	newAsk[len(ask)] = oneAsk
	sigmsg, err := AppendExtIDs(extIDs, 0, 6)
	if err != nil {
		log.Printfln("Identity Error:", err)
	} else {
		// Verify Signature
		idKey := st.Identities[IdentityIndex].Key1
		if CheckSig(idKey, extIDs[7][1:33], sigmsg, extIDs[8]) {
			var key [20]byte
			if len(extIDs[5]) != 20 {
				log.Println("New Anchor key for identity [" + chainID.String()[:10] + "] is invalid length")
				return
			}
			// Check Timestamp of message
			if !CheckTimestamp(extIDs[6]) {
				log.Println("New Anchor key for identity [" + chainID.String()[:10] + "] timestamp is too old")
				return
			}
			if contains {
				st.Identities[IdentityIndex].AnchorKeys = ask
			} else {
				st.Identities[IdentityIndex].AnchorKeys = newAsk
			}
			// Add to admin block
			status := st.Identities[IdentityIndex].Status
			if update && (status == constants.IDENTITY_FEDERATED_SERVER ||
				status == constants.IDENTITY_AUDIT_SERVER ||
				status == constants.IDENTITY_PENDING_FEDERATED_SERVER ||
				status == constants.IDENTITY_PENDING_AUDIT_SERVER) {
				if st.LeaderPL.VMIndexFor(constants.ADMIN_CHAINID) == st.GetLeaderVM() {
					copy(key[:20], extIDs[5][:20])
					extIDs[5] = append(extIDs[5], []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}...)
					key := primitives.NewHash(extIDs[5])
					msg := messages.NewChangeServerKeyMsg(st, chainID, constants.TYPE_ADD_FED_SERVER_KEY, extIDs[3][0], extIDs[4][0], key)
					st.InMsgQueue() <- msg
				}
				//st.LeaderPL.AdminBlock.AddFederatedServerBitcoinAnchorKey(chainID, extIDs[3][0], extIDs[4][0], &key)
			}
		} else {
			log.Println("New Anchor key for identity [" + chainID.String()[:10] + "] is invalid. Bad signiture")
		}
	}
}

//  stub for fake identity entries
func StubIdentityCache(st *State) {
	var id []Identity

	id = make([]Identity, 24)
	id[0] = MakeID("FED1", 1)
	id[1] = MakeID("FED2", 1)
	id[2] = MakeID("FED3", 1)
	id[3] = MakeID("FED4", 1)
	id[4] = MakeID("FED5", 1)
	id[5] = MakeID("FED6", 1)
	id[6] = MakeID("FED7", 1)
	id[7] = MakeID("FED8", 1)
	id[8] = MakeID("AUD1", 2)
	id[9] = MakeID("AUD2", 2)
	id[10] = MakeID("AUD3", 2)
	id[11] = MakeID("AUD4", 2)
	id[12] = MakeID("AUD5", 2)
	id[13] = MakeID("AUD6", 2)
	id[14] = MakeID("AUD7", 2)
	id[15] = MakeID("AUD8", 2)
	id[16] = MakeID("FUL1", 3)
	id[17] = MakeID("FUL2", 3)
	id[18] = MakeID("FUL3", 3)
	id[19] = MakeID("FUL4", 3)
	id[20] = MakeID("FUL5", 3)
	id[21] = MakeID("FUL6", 3)
	id[22] = MakeID("FUL7", 3)
	id[23] = MakeID("FUL8", 3)

	st.Identities = id
}

// Called by AddServer Message
func ProcessIdentityToAdminBlock(st *State, chainID interfaces.IHash, servertype int) bool {
	index := isIdentityChain(chainID, st.Identities)
	if index != -1 {
		id := st.Identities[index]

		if id.SigningKey == nil {
			log.Println("New Fed/Audit server [" + chainID.String()[:10] + "] does not have an Block Signing Key associated to it")
			return false
		} else {
			var pub [32]byte
			copy(pub[:32], id.SigningKey.Bytes()[:32])
			st.LeaderPL.AdminBlock.AddFederatedServerSigningKey(chainID, &pub)
		}

		if id.AnchorKeys == nil {
			log.Println("New Fed/Audit server [" + chainID.String()[:10] + "] does not have an BTC Anchor Key associated to it")
		} else {
			for _, aKey := range id.AnchorKeys {
				if strings.Compare(aKey.BlockChain, "BTC") == 0 {
					var key [20]byte
					copy(key[:20], aKey.SigningKey[:20])
					st.LeaderPL.AdminBlock.AddFederatedServerBitcoinAnchorKey(chainID, aKey.KeyLevel, aKey.KeyType, &key)
				}
			}
		}

		if id.MatryoshkaHash == nil {
			log.Println("New Fed/Audit server [" + chainID.String()[:10] + "] does not have an Matryoshka Hash associated to it")
		} else {
			st.LeaderPL.AdminBlock.AddMatryoshkaHash(chainID, id.MatryoshkaHash)
		}

		if servertype == 0 {
			id.Status = constants.IDENTITY_PENDING_FEDERATED_SERVER
		} else if servertype == 1 {
			id.Status = constants.IDENTITY_PENDING_AUDIT_SERVER
		}

		st.Identities[index] = id
	} else {
		log.Println("New Fed/Audit server [" + chainID.String()[:10] + "] does not have an identity associated to it")
		return false
	}
	return true
}

func MakeID(seed string, ServerType int) Identity {
	var id Identity
	nonce := primitives.Sha([]byte("Nonce")).Bytes()

	id.Key1 = primitives.Sha([]byte(seed))
	id.Key2 = primitives.Sha(id.Key1.Bytes())
	id.Key3 = primitives.Sha(id.Key2.Bytes())
	id.Key4 = primitives.Sha(id.Key3.Bytes())

	// make chainid  not bothering to loop looking for 888888 at start
	Chain := primitives.Sha(id.Key1.Bytes()).Bytes()
	temp := primitives.Sha(id.Key2.Bytes()).Bytes()
	Chain = append(Chain, temp[:]...)
	temp = primitives.Sha(id.Key3.Bytes()).Bytes()
	Chain = append(Chain, temp[:]...)
	temp = primitives.Sha(id.Key4.Bytes()).Bytes()
	Chain = append(Chain, temp[:]...)
	Chain = append(Chain, nonce[:]...)

	id.IdentityChainID = primitives.Sha(Chain)
	id.ManagementChainID = primitives.Sha(Chain)
	id.SigningKey = primitives.Sha(id.Key4.Bytes())
	id.Status = ServerType

	var ak AnchorSigningKey
	ak.BlockChain = "BTC"
	ak.KeyType = 0
	ak.SigningKey = id.SigningKey.Bytes()[0:20]
	id.AnchorKeys = make([]AnchorSigningKey, 1)
	id.AnchorKeys[0] = ak
	return id
}

// Sig is signed message, msg is raw message
func CheckSig(idKey interfaces.IHash, pub []byte, msg []byte, sig []byte) bool {
	var pubFix [32]byte
	var sigFix [64]byte

	copy(pubFix[:], pub[:32])
	copy(sigFix[:], sig[:64])

	pre := make([]byte, 0)
	pre = append(pre, []byte{0x01}...)
	pre = append(pre, pubFix[:]...)
	id := primitives.Shad(pre)

	/*if idKey == nil {
		log.Println("Identity Issue: No identity currently exist to check public key against identity key. Not full validation")
		return ed.VerifyCanonical(&pubFix, msg, &sigFix)
	}*/

	if id.IsSameAs(idKey) {
		return ed.VerifyCanonical(&pubFix, msg, &sigFix)
	} else {
		return false
	}
}

func CheckLength(length int, item []byte) bool {
	if len(item) != length {
		return false
	} else {
		return true
	}
}

func AppendExtIDs(extIDs [][]byte, start int, end int) ([]byte, error) {
	if len(extIDs) < (end + 1) {
		return nil, errors.New("Error: Index out of bound exception in AppendExtIDs()")
	}
	appended := make([]byte, 0)
	for i := start; i <= end; i++ {
		appended = append(appended, extIDs[i][:]...)
	}
	return appended, nil
}

// Makes sure the timestamp is within the designated window to be valid : 12 hours
func CheckTimestamp(time []byte) bool {
	if len(time) < 8 {
		zero := []byte{00}
		add := make([]byte, 0)
		for i := len(time); i <= 8; i++ {
			add = append(add, zero...)
		}
		time = append(add, time...)
	}
	//TODO: get time from State for replaying?
	now := primitives.GetTime()

	ts := binary.BigEndian.Uint64(time)
	var res uint64
	if now > ts {
		res = now - ts
	} else {
		res = ts - now
	}
	if res <= TWELVE_HOURS_S {
		return true
	} else {
		return false
	}
}

// Verifies an identity exists and if it is a federated or audit server
func (st *State) VerifyIdentityAdminInfo(cid interfaces.IHash) bool {
	IdentityIndex := isIdentityChain(cid, st.Identities)
	if IdentityIndex != -1 {
		status := st.Identities[IdentityIndex].Status
		if status == constants.IDENTITY_FEDERATED_SERVER ||
			status == constants.IDENTITY_AUDIT_SERVER ||
			status == constants.IDENTITY_PENDING_FEDERATED_SERVER ||
			status == constants.IDENTITY_PENDING_AUDIT_SERVER {
			// Is valid identity and a fed/audit server
			return true
		}
	}
	return false
}

func UpdateIdentityStatus(ChainID interfaces.IHash, StatusFrom int, StatusTo int, st *State) {
	// if StatusFrom < 0 then it will change from any status to the StatusTo status.
	//  if StatusFrom is > -1 then it will only change the status if it is = to the statusFrom

	IdentityIndex := isIdentityChain(ChainID, st.Identities)
	if IdentityIndex == -1 {
		log.Println("Cannot Update Status for ChainID " + ChainID.String() + ". Chain not found in Identities")
		return
	}
	st.Identities[IdentityIndex].Status = StatusTo
	/*
		if StatusFrom < 0 {
			st.Identities[IdentityIndex].Status = StatusTo
		} else {
			if st.Identities[IdentityIndex].Status == StatusFrom {
				st.Identities[IdentityIndex].Status = StatusTo
			} else {
				log.Println("Cannot Update Status for ChainID " + ChainID.String() + ". Status not equal to expected Current Status.")
			}
		}*/
}
