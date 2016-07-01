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
	SigningKey        primitives.PublicKey
	Status            int
	AnchorKeys        []AnchorSigningKey
	// add key history?
}

// TODO: Remove function
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

// TODO: Remove function
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
		//var AuthorityIndex int
		for _, e := range ABlock.GetABEntries() {
			st.UpdateAuthorityFromABEntry(e)
		}
	} else {
		log.Printfln("ERR:", err)
	}

}

func (st *State) UpdateAuthorityFromABEntry(entry interfaces.IABEntry) error {
	var AuthorityIndex int
	data, err := entry.MarshalBinary()
	if err != nil {
		return err
	}
	switch entry.Type() {
	case constants.TYPE_MINUTE_NUM:
		// Does not affect Authority.
	case constants.TYPE_DB_SIGNATURE:
		// Does not affect Authority
	case constants.TYPE_REVEAL_MATRYOSHKA:
		r := new(adminBlock.RevealMatryoshkaHash)
		err := r.UnmarshalBinary(data)
		if err != nil {
			return err
		}
		// Does nothing for authority right now
	case constants.TYPE_ADD_MATRYOSHKA:
		m := new(adminBlock.AddReplaceMatryoshkaHash)
		err := m.UnmarshalBinary(data)
		if err != nil {
			return err
		}

		AuthorityIndex = isAuthorityChain(m.IdentityChainID, st.Authorities)
		if AuthorityIndex == -1 {
			log.Println("Invalid Authority Chain ID. Add MatryoshkaHash " + m.IdentityChainID.String())
			break
		}
		st.Authorities[AuthorityIndex].MatryoshkaHash = m.MHash
	case constants.TYPE_ADD_SERVER_COUNT:
		s := new(adminBlock.IncreaseServerCount)
		err := s.UnmarshalBinary(data)
		if err != nil {
			return err
		}

		st.AuthorityServerCount = st.AuthorityServerCount + int(s.Amount)
	case constants.TYPE_ADD_FED_SERVER:
		f := new(adminBlock.AddFederatedServer)
		err := f.UnmarshalBinary(data)
		if err != nil {
			return err
		}

		AuthorityIndex = isAuthorityChain(f.IdentityChainID, st.Authorities)
		if AuthorityIndex == -1 {
			//log.Println(f.IdentityChainID.String() + " being added to Federated Server List AdminBlock Height:" + string(height))
			err = AddIdentityFromChainID(f.IdentityChainID, st)
			if err != nil {
				log.Printfln(err.Error())
				return err
			} else {
				AuthorityIndex = addAuthority(st, f.IdentityChainID)
			}
		} else {
			//log.Println(f.IdentityChainID.String() + " being promoted to Federated Server AdminBlock Height:" + string(height))
		}
		st.Authorities[AuthorityIndex].Status = constants.IDENTITY_FEDERATED_SERVER
		// check Identity status
		UpdateIdentityStatus(f.IdentityChainID, constants.IDENTITY_PENDING_FEDERATED_SERVER, constants.IDENTITY_FEDERATED_SERVER, st)
	case constants.TYPE_ADD_AUDIT_SERVER:
		a := new(adminBlock.AddAuditServer)
		err := a.UnmarshalBinary(data)
		if err != nil {
			return err
		}

		AuthorityIndex = isAuthorityChain(a.IdentityChainID, st.Authorities)
		if AuthorityIndex == -1 {
			//log.Println(a.IdentityChainID.String() + " being added to Federated Server List AdminBlock Height:" + string(height))
			err = AddIdentityFromChainID(a.IdentityChainID, st)
			if err != nil {
				log.Printfln(err.Error())
			} else {
				AuthorityIndex = addAuthority(st, a.IdentityChainID)
			}
		} else {
			//log.Println(a.IdentityChainID.String() + " being promoted to Federated Server AdminBlock Height:" + string(height))
		}
		st.Authorities[AuthorityIndex].Status = constants.IDENTITY_AUDIT_SERVER
		// check Identity status
		UpdateIdentityStatus(a.IdentityChainID, constants.IDENTITY_PENDING_AUDIT_SERVER, constants.IDENTITY_AUDIT_SERVER, st)
	case constants.TYPE_REMOVE_FED_SERVER:
		f := new(adminBlock.RemoveFederatedServer)
		err := f.UnmarshalBinary(data)
		if err != nil {
			return err
		}

		AuthorityIndex = isAuthorityChain(f.IdentityChainID, st.Authorities)
		if AuthorityIndex == -1 {
			//Add Identity as Federated Server
			log.Println(f.IdentityChainID.String() + " Cannot be removed.  Not in Authorities List.")
		} else {
			//log.Println(f.IdentityChainID.String() + " being removed from Authorities List:" + string(height))
			removeAuthority(AuthorityIndex, st)
		}
	case constants.TYPE_ADD_FED_SERVER_KEY:
		f := new(adminBlock.AddFederatedServerSigningKey)
		err := f.UnmarshalBinary(data)
		if err != nil {
			return err
		}
		keyBytes, err := f.PublicKey.MarshalBinary()
		if err != nil {
			return err
		}
		key := new(primitives.Hash)
		err = key.SetBytes(keyBytes)
		if err != nil {
			return err
		}
		addServerSigningKey(f.IdentityChainID, key, st)
	case constants.TYPE_ADD_BTC_ANCHOR_KEY:
		b := new(adminBlock.AddFederatedServerBitcoinAnchorKey)
		err := b.UnmarshalBinary(data)
		if err != nil {
			return err
		}

		AuthorityIndex = isAuthorityChain(b.IdentityChainID, st.Authorities)
		if AuthorityIndex == -1 {
			//Add Identity as Federated Server
			log.Println(b.IdentityChainID.String() + " Cannot Update Signing Key.  Not in Authorities List.")
		} else {
			//log.Println(b.IdentityChainID.String() + " Updating Signing Key. AdminBlock Height:" + string(height))
			pubKey, err := b.ECDSAPublicKey.MarshalBinary()
			if err != nil {
				return err
			}
			registerAuthAnchor(AuthorityIndex, pubKey, b.KeyType, b.KeyPriority, st, "BTC")
		}
	}
	return nil
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

	idIndex := isIdentityChain(chainID, st.Identities)
	if idIndex != -1 && st.Identities[idIndex].ManagementChainID != nil {
		oneAuth.ManagementChainID = st.Identities[idIndex].ManagementChainID
		// Keys should only update from admin block
		/*if st.Identities[idIndex].SigningKey != nil {
			oneAuth.SigningKey = primitives.PubKeyFromString(st.Identities[idIndex].SigningKey.String())
		}*/
		/*if st.Identities[idIndex].MatryoshkaHash != nil {
			oneAuth.MatryoshkaHash = st.Identities[idIndex].MatryoshkaHash
		}*/
		/*if len(st.Identities[idIndex].AnchorKeys) > 0 {
			oneAuth.AnchorKeys = make([]AnchorSigningKey, 0)
			oneAuth.AnchorKeys = append(oneAuth.AnchorKeys[:], st.Identities[idIndex].AnchorKeys[:]...)
		}*/
	} else {
		log.Println("Authority Error: " + chainID.String()[:10] + " No management chain found from identities.")
	}

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

func addServerSigningKey(ChainID interfaces.IHash, key interfaces.IHash, st *State) {
	var AuthorityIndex int
	AuthorityIndex = isAuthorityChain(ChainID, st.Authorities)
	if AuthorityIndex == -1 {
		log.Println(ChainID.String() + " Cannot Update Signing Key.  Not in Authorities List.")
	} else {
		//log.Println(ChainID.String() + " Updating Signing Key. AdminBlock Height:" + string(height))
		st.Authorities[AuthorityIndex].SigningKey = primitives.PubKeyFromString(key.String())
	}
}
