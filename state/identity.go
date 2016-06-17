package state

import (
	"encoding/binary"
	"errors"
	"fmt"
	ed "github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/log"
	"strings"
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

		LoadIdentityByDirectoryBlockHeight(i, st, false)
	}

}

func LoadIdentityByDirectoryBlockHeight(height uint32, st *State, update bool) {
	dblk, err := st.DB.FetchDBlockByHeight(uint32(height))
	if err != nil {
		log.Printfln("Identity Error:", err)
	}
	var ManagementChain interfaces.IHash
	ManagementChain, _ = primitives.HexToHash("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")

	entries := dblk.GetDBEntries()
	for _, eBlk := range entries {
		cid := eBlk.GetChainID()
		if cid.IsSameAs(ManagementChain) {
			// is it a new one?
			entkmr := eBlk.GetKeyMR() //eBlock Hash
			ecb, _ := st.DB.FetchEBlock(entkmr)
			entryHashes := ecb.GetEntryHashes()
			for _, eHash := range entryHashes {
				hs := eHash.String()
				if hs[0:10] != "0000000000" { //ignore minute markers
					ent, _ := st.DB.FetchEntry(eHash)
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
			ecb, _ := st.DB.FetchEBlock(entkmr)
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
							registerBlockSigningKey(ent.ExternalIDs(), cid, st, update)
						}

					} else if string(ent.ExternalIDs()[1]) == "New Bitcoin Key" {
						// this is the Signing Key for this Identity
						if len(ent.ExternalIDs()) == 9 { // update management should have 4 items
							registerAnchorSigningKey(ent.ExternalIDs(), cid, st, "BTC", update)
						}

					} else if string(ent.ExternalIDs()[1]) == "New Matryoshka Hash" {
						// this is the Signing Key for this Identity
						if len(ent.ExternalIDs()) == 7 { // update management should have 4 items
							updateMatryoshkaHash(ent.ExternalIDs(), cid, st, update)
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

	if err != nil {
		log.Printfln("Identity Error:", err)
	}
	// Remove Stale Identities
	// if an identity has taken more than 72 blocks (12 hours) to be fully created, remove it from the state identity list.
	var i int
	for i = 0; i < len(st.Identities); i++ {
		if st.Identities[i].IdentityCreated < height-TIME_WINDOW && st.Identities[i].IdentityRegistered == 0 && st.Identities[i].IdentityCreated != 0 {
			removeIdentity(i, st)
		} else if st.Identities[i].IdentityRegistered < height-TIME_WINDOW && st.Identities[i].IdentityCreated == 0 && st.Identities[i].IdentityRegistered != 0 {
			removeIdentity(i, st)
		} else if st.Identities[i].ManagementCreated < height-TIME_WINDOW && st.Identities[i].ManagementRegistered == 0 && st.Identities[i].ManagementCreated != 0 {
			removeIdentity(i, st)
		} else if st.Identities[i].ManagementRegistered < height-TIME_WINDOW && st.Identities[i].ManagementCreated == 0 && st.Identities[i].ManagementRegistered != 0 {
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
	// find the Identity index from the chain id in the external id.  add this chainID as the management id
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
		if checkSig(idKey, extIDs[3][1:33], sigmsg, extIDs[4]) {
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
		if checkSig(idKey, extIDs[3][1:33], sigmsg, extIDs[4]) {
			st.Identities[IdentityIndex].ManagementRegistered = height
			checkIdentityInitialStatus(IdentityIndex, st)
		} else {
			log.Println("New Management Chain Register for identity [" + chainID.String()[:10] + "] is invalid. Bad signiture")
		}

	}
	checkIdentityInitialStatus(IdentityIndex, st)
}

func registerBlockSigningKey(extIDs [][]byte, chainID interfaces.IHash, st *State, update bool) {
	IdentityIndex := isIdentityChain(chainID, st.Identities)
	if IdentityIndex == -1 {
		log.Println("Identity_Error: This cannot happen. New block signing key to nonexistent identity")
		return
	}

	sigmsg, err := AppendExtIDs(extIDs, 0, 4)
	if err != nil {
		log.Printfln("Identity Error:", err)
	} else {
		//verify Signature
		idKey := st.Identities[IdentityIndex].Key1
		if checkSig(idKey, extIDs[5][1:33], sigmsg, extIDs[6]) {
			// Check block key length
			var key [32]byte
			if len(extIDs[3]) != 32 {
				log.Println("New Block Signing key for identity [" + chainID.String()[:10] + "] is invalid length")
				return
			}
			// Check timestamp of message
			if !checkTimeStamp(extIDs[4]) {
				log.Println("New Block Signing key for identity [" + chainID.String()[:10] + "] timestamp is too old")
				return
			}

			st.Identities[IdentityIndex].SigningKey = primitives.NewHash(extIDs[3])
			// Add to admin block
			status := st.Identities[IdentityIndex].Status
			if update && (status == constants.IDENTITY_FEDERATED_SERVER || status == constants.IDENTITY_AUDIT_SERVER) {
				copy(key[:32], extIDs[3][:32])
				st.LeaderPL.AdminBlock.AddFederatedServerSigningKey(chainID, &key)
			}
		} else {
			log.Println("New Block Signing key for identity [" + chainID.String()[:10] + "] is invalid. Bad signiture")
		}
	}

}

func updateMatryoshkaHash(extIDs [][]byte, chainID interfaces.IHash, st *State, update bool) {
	IdentityIndex := isIdentityChain(chainID, st.Identities)
	if IdentityIndex == -1 {
		log.Println("Identity_Error: This cannot happen. New Matryoshka Hash to nonexistent identity")
		return
	}

	sigmsg, err := AppendExtIDs(extIDs, 0, 4)
	if err != nil {
		log.Printfln("Identity Error:", err)
	} else {
		// Verify Signature
		idKey := st.Identities[IdentityIndex].Key1
		if checkSig(idKey, extIDs[5][1:33], sigmsg, extIDs[6]) {
			// Check MHash length
			if len(extIDs[3]) != 32 {
				log.Println("New Matryoshka Hash for identity [" + chainID.String()[:10] + "] is invalid length")
				return
			}
			// Check Timestamp of message
			if !checkTimeStamp(extIDs[4]) {
				log.Println("New Matryoshka Hash for identity [" + chainID.String()[:10] + "] timestamp is too old")
				return
			}
			mhash := primitives.NewHash(extIDs[3])
			st.Identities[IdentityIndex].MatryoshkaHash = mhash
			// Add to admin block
			status := st.Identities[IdentityIndex].Status
			if update && (status == constants.IDENTITY_FEDERATED_SERVER || status == constants.IDENTITY_AUDIT_SERVER) {
				st.LeaderPL.AdminBlock.AddMatryoshkaHash(chainID, mhash)
			}
		} else {
			log.Println("New Matryoshka Hash for identity [" + chainID.String()[:10] + "] is invalid. Bad signiture")
		}

	}
}

func registerAnchorSigningKey(extIDs [][]byte, chainID interfaces.IHash, st *State, BlockChain string, update bool) {
	IdentityIndex := isIdentityChain(chainID, st.Identities)
	if IdentityIndex == -1 {
		log.Println("Identity_Error: This cannot happen. New Bitcoin Key to nonexistent identity")
		return
	}

	var ask []AnchorSigningKey
	var newAsk []AnchorSigningKey
	var oneAsk AnchorSigningKey

	ask = st.Identities[IdentityIndex].AnchorKeys
	newAsk = make([]AnchorSigningKey, len(ask)+1)

	for i := 0; i < len(ask); i++ {
		newAsk[i] = ask[i]
	}

	oneAsk.BlockChain = BlockChain
	oneAsk.KeyLevel = extIDs[3][0]
	oneAsk.KeyType = extIDs[4][0]
	oneAsk.SigningKey = extIDs[5]

	newAsk[len(ask)] = oneAsk

	sigmsg, err := AppendExtIDs(extIDs, 0, 6)
	if err != nil {
		log.Printfln("Identity Error:", err)
	} else {
		// Verify Signature
		idKey := st.Identities[IdentityIndex].Key1
		if checkSig(idKey, extIDs[7][1:33], sigmsg, extIDs[8]) {
			var key [20]byte
			if len(extIDs[5]) != 20 {
				log.Println("New Anchor key for identity [" + chainID.String()[:10] + "] is invalid length")
				return
			}
			// Check Timestamp of message
			if !checkTimeStamp(extIDs[6]) {
				log.Println("New Anchor key for identity [" + chainID.String()[:10] + "] timestamp is too old")
				return
			}
			st.Identities[IdentityIndex].AnchorKeys = newAsk
			// Add to admin block
			status := st.Identities[IdentityIndex].Status
			if update && (status == constants.IDENTITY_FEDERATED_SERVER || status == constants.IDENTITY_AUDIT_SERVER) {
				copy(key[:20], extIDs[5][:20])
				st.LeaderPL.AdminBlock.AddFederatedServerBitcoinAnchorKey(chainID, extIDs[3][0], extIDs[4][0], &key)
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
func checkSig(idKey interfaces.IHash, pub []byte, msg []byte, sig []byte) bool {
	var pubFix [32]byte
	var sigFix [64]byte

	copy(pubFix[:], pub[:32])
	copy(sigFix[:], sig[:64])

	pre := make([]byte, 0)
	pre = append(pre, []byte{0x01}...)
	pre = append(pre, pubFix[:]...)
	id := primitives.Shad(pre)

	if id.IsSameAs(idKey) {
		return ed.Verify(&pubFix, msg, &sigFix)
	} else {
		return false
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
func checkTimeStamp(time []byte) bool {
	if len(time) < 8 {
		zero := []byte{00}
		add := make([]byte, 0)
		for i := len(time); i <= 8; i++ {
			add = append(add, zero...)
		}
		time = append(add, time...)
	}
	now := interfaces.GetTime()

	ts := binary.BigEndian.Uint64(time)
	res := now - ts
	if res < 0 {
		res = -res
	}
	if res <= TWELVE_HOURS_S {
		return true
	} else {
		return false
	}
	return true

}
