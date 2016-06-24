package state

import (
	//	"bytes"
	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/log"
)

type Authority struct {
	AuthorityChainID  interfaces.IHash
	ManagementChainID interfaces.IHash
	MatryoshkaHash    interfaces.IHash
	SigningKey        interfaces.IHash
	Status            int
	AnchorKeys        []AnchorSigningKey
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

		LoadAuthorityByAdminBlockHeight(i, st, false)

	}

}

func LoadAuthorityByAdminBlockHeight(height uint32, st *State, update bool) {

	dblk, _ := st.DB.FetchDBlockByHeight(uint32(height))
	if dblk == nil {
		log.Println("Invalid Admin Block Height:" + string(height))
		return
	}

	msg, err := st.LoadDBState(height)
	if err == nil && msg != nil {
		dsmsg := msg.(*messages.DBStateMsg)
		ABlock := dsmsg.AdminBlock
		//abBytes, _ := ABlock.MarshalBinary()
		//abBytes = abBytes[32:] //remove admin chain id
		//abBytes = abBytes[32:] //Previous Hash
		//abBytes = abBytes[4:]  //remove Block Height
		//ext, abBytes := primitives.DecodeVarIntBTC(abBytes)
		//	don't care about header expansion at this time.  skip it.
		//abBytes = abBytes[ext:] //remove admin chain id
		//abBytes = abBytes[4:]   //remove message count (we are parsing bytes instead of messages)
		//abBytes = abBytes[4:]   //remove body size
		// the rest is admin byte messages

		//ChainID := new(primitives.Hash)
		var AuthorityIndex int
		//var testBytes []byte

		/*
			TYPE_MINUTE_NUM         uint8 = iota // 0
			TYPE_DB_SIGNATURE                    // 1
			TYPE_REVEAL_MATRYOSHKA               // 2
			TYPE_ADD_MATRYOSHKA                  // 3
			TYPE_ADD_SERVER_COUNT                // 4
			TYPE_ADD_FED_SERVER                  // 5
			TYPE_ADD_AUDIT_SERVER                // 6
			TYPE_REMOVE_FED_SERVER               // 7
			TYPE_ADD_FED_SERVER_KEY              // 8
			TYPE_ADD_BTC_ANCHOR_KEY              // 9
		*/
		for _, e := range ABlock.GetABEntries() {
			data, err := e.MarshalBinary()
			if err != nil {
				continue
			}
			switch e.Type() {
			case constants.TYPE_MINUTE_NUM:

			case constants.TYPE_DB_SIGNATURE:
				// Does not affect Authority
			case constants.TYPE_REVEAL_MATRYOSHKA:
				r := new(adminBlock.RevealMatryoshkaHash)
				err := r.UnmarshalBinary(data)
				if err != nil {
					break
				}
				// Does nothing for authority right now
			case constants.TYPE_ADD_MATRYOSHKA:
				m := new(adminBlock.AddReplaceMatryoshkaHash)
				err := m.UnmarshalBinary(data)
				if err != nil {
					break
				}

				AuthorityIndex = isAuthorityChain(m.IdentityChainID, st.Authorities)
				if AuthorityIndex == -1 {
					log.Println("Invalid Authority Chain ID. Add MatryoshkaHash AdminBlock Height:" + string(height) + " " + m.IdentityChainID.String())
					// dont return, just ignote this item
				}
				st.Authorities[AuthorityIndex].MatryoshkaHash = m.MHash
			case constants.TYPE_ADD_SERVER_COUNT:
				s := new(adminBlock.IncreaseServerCount)
				err := s.UnmarshalBinary(data)
				if err != nil {
					break
				}

				st.AuthorityServerCount = st.AuthorityServerCount + int(s.Amount)
			case constants.TYPE_ADD_FED_SERVER:
				f := new(adminBlock.AddFederatedServer)
				err := f.UnmarshalBinary(data)
				if err != nil {
					break
				}

				AuthorityIndex = isAuthorityChain(f.IdentityChainID, st.Authorities)
				if AuthorityIndex == -1 {
					//Add Identity as Federated Server
					log.Println(f.IdentityChainID.String() + " being added to Federated Server List AdminBlock Height:" + string(height))
					AuthorityIndex = addAuthority(st, f.IdentityChainID)
				} else {
					log.Println(f.IdentityChainID.String() + " being promoted to Federated Server AdminBlock Height:" + string(height))
				}
				st.Authorities[AuthorityIndex].Status = constants.IDENTITY_FEDERATED_SERVER
				// check Identity status
				UpdateIdentityStatus(f.IdentityChainID, constants.IDENTITY_PENDING_FEDERATED_SERVER, constants.IDENTITY_FEDERATED_SERVER, st)
			case constants.TYPE_ADD_AUDIT_SERVER:
				a := new(adminBlock.AddAuditServer)
				err := a.UnmarshalBinary(data)
				if err != nil {
					break
				}

				AuthorityIndex = isAuthorityChain(a.IdentityChainID, st.Authorities)
				if AuthorityIndex == -1 {
					//Add Identity as Federated Server
					log.Println(a.IdentityChainID.String() + " being added to Federated Server List AdminBlock Height:" + string(height))
					AuthorityIndex = addAuthority(st, a.IdentityChainID)
				} else {
					log.Println(a.IdentityChainID.String() + " being promoted to Federated Server AdminBlock Height:" + string(height))
				}
				st.Authorities[AuthorityIndex].Status = constants.IDENTITY_AUDIT_SERVER
				// check Identity status
				UpdateIdentityStatus(a.IdentityChainID, constants.IDENTITY_PENDING_AUDIT_SERVER, constants.IDENTITY_AUDIT_SERVER, st)
			case constants.TYPE_REMOVE_FED_SERVER:
				f := new(adminBlock.RemoveFederatedServer)
				err := f.UnmarshalBinary(data)
				if err != nil {
					break
				}

				AuthorityIndex = isAuthorityChain(f.IdentityChainID, st.Authorities)
				if AuthorityIndex == -1 {
					//Add Identity as Federated Server
					log.Println(f.IdentityChainID.String() + " Cannot be removed.  Not in Authorities List. AdminBlock Height:" + string(height))
				} else {
					log.Println(f.IdentityChainID.String() + " being removed from Authorities List:" + string(height))
					removeAuthority(AuthorityIndex, st)
				}
			case constants.TYPE_ADD_FED_SERVER_KEY:
				f := new(adminBlock.AddFederatedServerSigningKey)
				err := f.UnmarshalBinary(data)
				if err != nil {
					break
				}
				keyBytes, err := f.PublicKey.MarshalBinary()
				if err != nil {
					break
				}
				key := new(primitives.Hash)
				err = key.SetBytes(keyBytes)
				if err != nil {
					break
				}
				addServerSigningKey(f.IdentityChainID, key, height, st)
			case constants.TYPE_ADD_BTC_ANCHOR_KEY:
				b := new(adminBlock.AddFederatedServerBitcoinAnchorKey)
				err := b.UnmarshalBinary(data)
				if err != nil {
					break
				}

				AuthorityIndex = isAuthorityChain(b.IdentityChainID, st.Authorities)
				if AuthorityIndex == -1 {
					//Add Identity as Federated Server
					log.Println(b.IdentityChainID.String() + " Cannot Update Signing Key.  Not in Authorities List. AdminBlock Height:" + string(height))
				} else {
					log.Println(b.IdentityChainID.String() + " Updating Signing Key. AdminBlock Height:" + string(height))
					pubKey, err := b.ECDSAPublicKey.MarshalBinary()
					if err != nil {
						break
					}
					registerAuthAnchor(AuthorityIndex, pubKey, b.KeyType, b.KeyPriority, st, "BTC")
				}
			}
		}
	} else {
		log.Printfln("ERR:", err)
	}

}

func isAuthorityChain(cid interfaces.IHash, ids []Authority) int {
	//is this an identity chain
	for i, authorityChain := range ids {
		if authorityChain.AuthorityChainID.IsSameAs(cid) {
			return i
		}
	}

	return -1
}

func addAuthority(st *State, chainID interfaces.IHash) int {

	var authnew []Authority
	authnew = make([]Authority, len(st.Authorities)+1)

	var oneAuth Authority

	for i := 0; i < len(st.Authorities); i++ {
		authnew[i] = st.Authorities[i]
	}
	oneAuth.AuthorityChainID = chainID

	oneAuth.Status = constants.IDENTITY_PENDING

	authnew[len(st.Authorities)] = oneAuth

	st.Authorities = authnew
	return len(st.Authorities) - 1
}

func removeAuthority(i int, st *State) {
	var newIDs []Authority
	newIDs = make([]Authority, len(st.Authorities)-1)
	var j int
	for j = 0; j < i; j++ {
		newIDs[j] = st.Authorities[j]
	}
	// skip removed Identity
	for j = i + 1; j < len(newIDs); j++ {
		newIDs[j-1] = st.Authorities[j]
	}
	st.Authorities = newIDs
}

func registerAuthAnchor(AuthorityIndex int, signingKey []byte, keyType byte, keyLevel byte, st *State, BlockChain string) {
	var ask []AnchorSigningKey
	var newASK []AnchorSigningKey
	var oneASK AnchorSigningKey

	ask = st.Authorities[AuthorityIndex].AnchorKeys
	newASK = make([]AnchorSigningKey, len(ask)+1)

	for i := 0; i < len(ask); i++ {
		newASK[i] = ask[i]
	}

	oneASK.BlockChain = BlockChain
	oneASK.KeyLevel = keyLevel
	oneASK.KeyType = keyType
	oneASK.SigningKey = signingKey

	newASK[len(ask)] = oneASK
	st.Authorities[AuthorityIndex].AnchorKeys = newASK
}

func addServerSigningKey(ChainID interfaces.IHash, key interfaces.IHash, height uint32, st *State) {
	var AuthorityIndex int
	AuthorityIndex = isAuthorityChain(ChainID, st.Authorities)
	if AuthorityIndex == -1 {
		//Add Identity as Federated Server
		log.Println(ChainID.String() + " Cannot Update Signing Key.  Not in Authorities List. AdminBlock Height:" + string(height))
	} else {
		log.Println(ChainID.String() + " Updating Signing Key. AdminBlock Height:" + string(height))
		st.Authorities[AuthorityIndex].SigningKey = key
	}
}
