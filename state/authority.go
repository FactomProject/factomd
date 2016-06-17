package state

import (
	"encoding/hex"
	"fmt"
	//"github.com/FactomProject/factomd/common/constants"
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
type Authority struct {
	AuthorityChainID      interfaces.IHash
	ManagementChainID    interfaces.IHash
	MatryoshkaHash       interfaces.IHash
	SigningKey           interfaces.IHash
	Status               int
	AnchorKeys           []AnchorSigningKey
	// add key history?
}

func LoadAuthorityCache(st *State) {

	// var s State
	blockHead, err := st.DB.FetchDirectoryBlockHead()

	if blockHead == nil {
		// new block chain just created.  no id yet
		return
	}
	bHeader := blockHead.GetHeader()
	height := bHeader.GetDBHeight()

	if err != nil {
		log.Printfln("ERR:", err)

	}
	var i uint32
	for i = 1; i < height; i++ {

		LoadAuthorityByAdminBlockHeight(i, st)

	}

}

func LoadAuthorityByAdminBlockHeight(height uint32, st *State) {
	var id []Authority
	id = st.Authorities

	dblk, err := st.DB.FetchABlockByHeight(uint32(height))
	if err != nil {
		log.Printfln("ERR:", err)

	}
	var ManagementChain interfaces.IHash
	ManagementChain, _ = primitives.HexToHash("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")

	entries := dblk.GetABEntries()
	for _, aBlk := range entries {

             fmt.Println("ABlock:",aBlk.Printable)

		}
	}


//  stub for fake Authority entries

func StubAuthorityCache(st *State) {

	var id []Authority
	id = make([]Authority, 24)
	id[0] = MakeAuth("FED1", 1)
	id[1] = MakeAuth("FED2", 1)
	id[2] = MakeAuth("FED3", 1)
	id[3] = MakeAuth("FED4", 1)
	id[4] = MakeAuth("FED5", 1)
	id[5] = MakeAuth("FED6", 1)
	id[6] = MakeAuth("FED7", 1)
	id[7] = MakeAuth("FED8", 1)
	id[8] = MakeAuth("AUD1", 2)
	id[9] = MakeAuth("AUD2", 2)
	id[10] = MakeAuth("AUD3", 2)
	id[11] = MakeAuth("AUD4", 2)
	id[12] = MakeAuth("AUD5", 2)
	id[13] = MakeAuth("AUD6", 2)
	id[14] = MakeAuth("AUD7", 2)
	id[15] = MakeAuth("AUD8", 2)
	id[16] = MakeAuth("FUL1", 3)
	id[17] = MakeAuth("FUL2", 3)
	id[18] = MakeAuth("FUL3", 3)
	id[19] = MakeAuth("FUL4", 3)
	id[20] = MakeAuth("FUL5", 3)
	id[21] = MakeAuth("FUL6", 3)
	id[22] = MakeAuth("FUL7", 3)
	id[23] = MakeAuth("FUL8", 3)

	st.Authorities = id

}

func MakeAuth(seed string, ServerType int) Authority {

	var id Authority
	nonce := primitives.Sha([]byte("Nonce")).Bytes()


	// make chainid  not bothering to loop looking for 888888 at start


	id.AuthorityChainID = primitives.Sha(nonce)
	id.ManagementChainID = primitives.Sha(id.AuthorityChainID)
	id.SigningKey = primitives.Sha(id.ManagementChainID)
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