package state

import (
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/log"
	//ed "github.com/FactomProject/ed25519"
)

type AnchorSigningKey struct {
	BlockChain string
	KeyLevel   string
	KeyType    string
	SigningKey string //if bytes, it is hex
}
type Identity struct {
	IdentityChainID   interfaces.IHash
	ManagementChainID interfaces.IHash
	MatryoshkaHash    interfaces.IHash
	Key1              interfaces.IHash
	Key2              interfaces.IHash
	Key3              interfaces.IHash
	Key4              interfaces.IHash
	SigningKey        interfaces.IHash
	Status            int
	AnchorKeys        []AnchorSigningKey
}

func LoadIdentityCache(st *State) {

	// var s State
	blockHead, err := st.DB.FetchDirectoryBlockHead()

	if blockHead == nil {
		// new block chain just created.  no id yet
		return
	}
	bHeader := blockHead.GetHeader()
	height := bHeader.GetDBHeight()

	if err != nil {
		fmt.Println("ERR:", err)
		fmt.Println("############################################################################")

	}
	var i uint32
	for i = 1; i < height; i++ {

		LoadIdentityByDirectoryBlockHeight(i, st)
		if i == 1281 {
			fmt.Println("added:", i, ":", st.Identities)
		}
	}

}

func LoadIdentityByDirectoryBlockHeight(height uint32, st *State) {
	var id []Identity
	id = st.Identities

	dblk, err := st.DB.FetchDBlockByHeight(uint32(height))
	if err != nil {
		fmt.Println("ERR:", err)
		fmt.Println("############################################################################")

	}
	var ManagementChain interfaces.IHash
	ManagementChain, _ = primitives.HexToHash("5a77d1e9612d350b3734f6282259b7ff0a3f87d62cfef5f35e91a5604c0490a3")

	entries := dblk.GetDBEntries()
	for _, eBlk := range entries {

		cid := eBlk.GetChainID()
		if cid.IsSameAs(ManagementChain) {
			// is it a new one?
			entkmr := eBlk.GetKeyMR() //eBlock Hash
			ecb, _ := st.DB.FetchEBlockByKeyMR(entkmr)
			entryHashes := ecb.GetEntryHashes()
			for _, eHash := range entryHashes {

				hs := eHash.String()
				if hs[0:10] != "0000000000" { //ignore minute markers
					ent, _ := st.DB.FetchEntryByHash(eHash)
					if len(ent.ExternalIDs()) > 2 {
						fmt.Println("Federated Management Chain:", string(ent.ExternalIDs()[1]))
					}
				}
			}

		} else if cid.String()[0:2] == "88" {
			IdentityIndex := isIdentityChain(cid, id)
			if IdentityIndex > -1 {
				// is it in the list already?
				// if so, what kind of entry is it?

				entkmr := eBlk.GetKeyMR() //eBlock Hash
				ecb, _ := st.DB.FetchEBlockByKeyMR(entkmr)
				entryHashes := ecb.GetEntryHashes()
				for _, eHash := range entryHashes {

					hs := eHash.String()
					if hs[0:10] != "0000000000" { //ignore minute markers

						ent, _ := st.DB.FetchEntryByHash(eHash)

						if string(ent.ExternalIDs()[1]) == "Register Server Management" {
							// this is an Identity that should have been registered already with a 0 status.
							//  this registers it with the management chain.  Now it can be assigned as federated or audit.
							//  set it to status 6 - Pending Full
							registerIdentityAsServer(IdentityIndex, cid, st)
						} else if string(ent.ExternalIDs()[1]) == "New Block Signing Key" {
							// this is the Signing Key for this Identity
							if len(ent.ExternalIDs()) == 7 { // update management should have 4 items
								registerBlockSigningKey(ent.ExternalIDs(), cid, st)
							}

						} else if string(ent.ExternalIDs()[1]) == "New Bitcoin Key" {
							// this is the Signing Key for this Identity
							if len(ent.ExternalIDs()) == 9 { // update management should have 4 items
								registerAnchorSigningKey(ent.ExternalIDs(), cid, st, "BTC")
							}

						} else if string(ent.ExternalIDs()[1]) == "New Matryoshka Hash" {
							// this is the Signing Key for this Identity
							if len(ent.ExternalIDs()) == 7 { // update management should have 4 items
								updateMatryoshkaHash(ent.ExternalIDs(), cid, st)
							}

						}

					}
				}

			} else {

				// this identity is not in the
				// read the entry and see if it looks like an initial Identity Chain Creation
				entkmr := eBlk.GetKeyMR() //eBlock Hash
				ecb, _ := st.DB.FetchEBlockByKeyMR(entkmr)
				if ecb != nil {
					entryHashes := ecb.GetEntryHashes()
					for _, eHash := range entryHashes {
						hs := eHash.String()
						if hs[0:10] != "0000000000" {
							//ignore minute markers

							ent, _ := st.DB.FetchEntryByHash(eHash)

							if len(ent.ExternalIDs()) > 1 && string(ent.ExternalIDs()[1]) == "Identity Chain" {
								// this is a new identity
								addIdentity(ent.ExternalIDs(), cid, st)
							} else if len(ent.ExternalIDs()) > 1 && string(ent.ExternalIDs()[1]) == "Server Management" {
								// this is a new identity
								if len(ent.ExternalIDs()) == 4 {
									// update management should have 4 items
									updateManagementKey(ent.ExternalIDs(), cid, st)
								}
							}
						}
					}
				}
			}
		} else {

		}
	}

	if err != nil {
		fmt.Println("ERR:", err)
		fmt.Println("############################################################################")

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

func addIdentity(extIDs [][]byte, chainID interfaces.IHash, st *State) {
	var id []Identity
	var idnew []Identity
	var oneID Identity
	id = st.Identities
	idnew = make([]Identity, len(id)+1)
	for i := 0; i < len(id); i++ {
		idnew[i] = id[i]
	}
	oneID.IdentityChainID = chainID
	h := primitives.NewHash(extIDs[2])
	oneID.Key1 = h
	h = primitives.NewHash(extIDs[3])
	oneID.Key2 = h
	h = primitives.NewHash(extIDs[4])
	oneID.Key3 = h
	h = primitives.NewHash(extIDs[5])
	oneID.Key4 = h
	oneID.Status = constants.IDENTITY_UNASSIGNED // new identity.
	idnew[len(id)] = oneID

	//sigmsg := appendbytes(extIDs[0],extIDs[1])
	//sigmsg = appendbytes (sigmsg,extIDs[2])
	//verify Signature
	//if ed.Verify(oneID.Key4,sigmsg,extIDs[4]){
	st.Identities = idnew
	//}
}

func updateManagementKey(extIDs [][]byte, chainID interfaces.IHash, st *State) {
	// find the Identity index from the chain id in the external id.  add this chainID as the management id
	idChain := primitives.NewHash(extIDs[2])
	IdentityIndex := isIdentityChain(idChain, st.Identities)

	//sigmsg := append(extIDs[0],extIDs[1],extIDs[2])
	//verify Signature
	//if ed.Verify(st.Identities[IdentityIndex].Key1,sigmsg,extIDs[4]){
	st.Identities[IdentityIndex].ManagementChainID = chainID
	//}

}

func registerIdentityAsServer(IdentityIndex int, chainID interfaces.IHash, st *State) {
	st.Identities[IdentityIndex].Status = constants.IDENTITY_PENDING_FULL
}

func registerBlockSigningKey(extIDs [][]byte, chainID interfaces.IHash, st *State) {
	IdentityIndex := isIdentityChain(chainID, st.Identities)

	//sigmsg := append(extIDs[0],extIDs[1],extIDs[2],extIDs[3],extIDs[4])
	//verify Signature
	//if ed.Verify(st.Identities[IdentityIndex].Key1,sigmsg,extIDs[6]){
	st.Identities[IdentityIndex].SigningKey = primitives.NewHash(extIDs[3])
	//}

}
func updateMatryoshkaHash(extIDs [][]byte, chainID interfaces.IHash, st *State) {
	IdentityIndex := isIdentityChain(chainID, st.Identities)

	//sigmsg := append(extIDs[0],extIDs[1],extIDs[2],extIDs[3],extIDs[4])
	//verify Signature
	//if ed.Verify(st.Identities[IdentityIndex].Key1,sigmsg,extIDs[6]){
	st.Identities[IdentityIndex].MatryoshkaHash = primitives.NewHash(extIDs[3])
	//}

}

func registerAnchorSigningKey(extIDs [][]byte, chainID interfaces.IHash, st *State, BlockChain string) {
	IdentityIndex := isIdentityChain(chainID, st.Identities)

	var ask []AnchorSigningKey
	var newAsk []AnchorSigningKey
	var oneAsk AnchorSigningKey

	ask = st.Identities[IdentityIndex].AnchorKeys
	newAsk = make([]AnchorSigningKey, len(ask)+1)

	for i := 0; i < len(ask); i++ {
		newAsk[i] = ask[i]
	}

	oneAsk.BlockChain = BlockChain
	oneAsk.KeyLevel = hex.EncodeToString(extIDs[3])
	oneAsk.KeyType = hex.EncodeToString(extIDs[4])
	oneAsk.SigningKey = hex.EncodeToString(extIDs[5])

	newAsk[len(ask)] = oneAsk

	//sigmsg := append(extIDs[0],extIDs[1],extIDs[2],extIDs[3],extIDs[4],extIDs[5],extIDs[6])
	//verify Signature
	//if ed.Verify(st.Identities[IdentityIndex].Key1,sigmsg,extIDs[8]){
	st.Identities[IdentityIndex].AnchorKeys = newAsk
	//}
}

func newSigningKey(extIDs [][]byte, chainID interfaces.IHash, st *State) error {
	idChain := primitives.NewHash(extIDs[2])
	newKey := primitives.NewHash(extIDs[3])
	chainIndex := isIdentityChain(primitives.NewHash(extIDs[2]), st.Identities)
	if chainIndex == -1 {
		log.Printfln("New Signing Key Error: Identity Chain Not Found. %s (changing signing pub key to %s)", idChain, newKey)
		return fmt.Errorf("New Signing Key Error: Identity Chain Not Found. %s (changing signing pub key to %s)", idChain, newKey)
	}

	// sigmsg := append(extIDs[0],extIDs[1],extIDs[2],extIDs[3],extIDs[4])
	//verify Signature
	//  if ed.Verify(st.Identities[IdentityIndex].Key1,sigmsg,extIDs[6]){
	st.Identities[chainIndex].SigningKey = newKey
	// }

	return nil
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
	Chain = appendbytes(Chain, temp)
	temp = primitives.Sha(id.Key3.Bytes()).Bytes()
	Chain = appendbytes(Chain, temp)
	temp = primitives.Sha(id.Key4.Bytes()).Bytes()
	Chain = appendbytes(Chain, temp)
	Chain = appendbytes(Chain, nonce)

	id.IdentityChainID = primitives.Sha(Chain)
	id.ManagementChainID = primitives.Sha(Chain)
	id.SigningKey = primitives.Sha(id.Key4.Bytes())
	id.Status = ServerType

	var ak AnchorSigningKey
	ak.BlockChain = "BTC"
	ak.KeyType = "P2PKH"
	ak.SigningKey = hex.EncodeToString(id.SigningKey.Bytes()[0:20])
	id.AnchorKeys = make([]AnchorSigningKey, 1)
	id.AnchorKeys[0] = ak
	return id
}

func appendbytes(first []byte, second []byte) []byte {
	for i := 0; i < len(second); i++ {
		first = append(first, second[i])
	}
	return first
}
