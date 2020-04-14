// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identity

import (
	"fmt"

	"bytes"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// ProcessABlockEntry processes the input admin block entry based on type
func (im *IdentityManager) ProcessABlockEntry(entry interfaces.IABEntry, st interfaces.IState) error {
	switch entry.Type() {
	case constants.TYPE_REVEAL_MATRYOSHKA:
		return im.ApplyRevealMatryoshkaHash(entry)
	case constants.TYPE_ADD_MATRYOSHKA:
		return im.ApplyAddReplaceMatryoshkaHash(entry)
	case constants.TYPE_ADD_SERVER_COUNT:
		return im.ApplyIncreaseServerCount(entry)
	case constants.TYPE_ADD_FED_SERVER:
		return im.ApplyAddFederatedServer(entry, st)
	case constants.TYPE_ADD_AUDIT_SERVER:
		return im.ApplyAddAuditServer(entry, st)
	case constants.TYPE_REMOVE_FED_SERVER:
		return im.ApplyRemoveFederatedServer(entry)
	case constants.TYPE_ADD_FED_SERVER_KEY:
		return im.ApplyAddFederatedServerSigningKey(entry)
	case constants.TYPE_ADD_BTC_ANCHOR_KEY:
		return im.ApplyAddFederatedServerBitcoinAnchorKey(entry)
	case constants.TYPE_SERVER_FAULT:
		return im.ApplyServerFault(entry)
	case constants.TYPE_ADD_FACTOID_ADDRESS:
		im.ApplyAddFactoidAddress(entry)
	case constants.TYPE_ADD_FACTOID_EFFICIENCY:
		im.ApplyAddEfficiency(entry)
	case constants.TYPE_COINBASE_DESCRIPTOR_CANCEL:
		im.ApplyCancelCoinbaseDescriptor(entry)
	case constants.TYPE_COINBASE_DESCRIPTOR:
		// This does nothing. The coinbase code looks back in the database
		// for this entry. In the present, it does not do anything.
	}
	return nil
}

//func (im *IdentityManager) () {

//}

// ApplyCancelCoinbaseDescriptor transfers the cancel coinbase outputs from the entry into the identity manager
func (im *IdentityManager) ApplyCancelCoinbaseDescriptor(entry interfaces.IABEntry) error {
	e := entry.(*adminBlock.CancelCoinbaseDescriptor)

	// Add the descriptor and index to the list of cancelled outputs.
	//	This will be checked and garbage collected on payout
	var list []uint32
	var ok bool
	if list, ok = im.CanceledCoinbaseOutputs[e.DescriptorHeight]; !ok {
		im.CanceledCoinbaseOutputs[e.DescriptorHeight] = make([]uint32, 0)
	}

	list = append(list, e.DescriptorIndex)
	list = BubbleSortUint32(list)
	im.CanceledCoinbaseOutputs[e.DescriptorHeight] = list

	return nil
}

// ApplyRevealMatryoshkaHash is a no-op function
func (im *IdentityManager) ApplyRevealMatryoshkaHash(entry interfaces.IABEntry) error {
	//e:=entry.(*adminBlock.RevealMatryoshkaHash)
	// Does nothing for authority right now
	return nil
}

// ApplyAddReplaceMatryoshkaHash grabs the authority server from the entry's chain id, and sets its Matryoshka Hash to the entry's Matryoshka Hash
func (im *IdentityManager) ApplyAddReplaceMatryoshkaHash(entry interfaces.IABEntry) error {
	e := entry.(*adminBlock.AddReplaceMatryoshkaHash)

	auth := im.GetAuthority(e.IdentityChainID)
	if auth == nil {
		return fmt.Errorf("Authority %v not found", e.IdentityChainID.String())
	}
	auth.MatryoshkaHash = e.MHash.(*primitives.Hash)
	im.SetAuthority(e.IdentityChainID, auth)

	return nil
}

// ApplyIncreaseServerCount increases the maximum authority server count by the amount in the input entry
func (im *IdentityManager) ApplyIncreaseServerCount(entry interfaces.IABEntry) error {
	e := entry.(*adminBlock.IncreaseServerCount)
	im.MaxAuthorityServerCount = im.MaxAuthorityServerCount + int(e.Amount)
	return nil
}

// ApplyAddFederatedServer gets the authority server from the entry's chain id, and sets its status to a federated server. If no existing authority
// server can be found, it adds a new authority server with the requisite properties
func (im *IdentityManager) ApplyAddFederatedServer(entry interfaces.IABEntry, st interfaces.IState) error {
	e := entry.(*adminBlock.AddFederatedServer)

	// New server. Check if the identity exists, and create it if it does not
	id := im.GetIdentity(e.IdentityChainID)
	if id == nil {
		st.AddIdentityFromChainID(e.IdentityChainID)
		id = im.GetIdentity(e.IdentityChainID)
	}

	auth := im.GetAuthority(e.IdentityChainID)
	if auth == nil {
		auth = NewAuthority()
	}

	auth.Status = constants.IDENTITY_FEDERATED_SERVER
	auth.AuthorityChainID = e.IdentityChainID.(*primitives.Hash)

	if id != nil {
		id.Status = constants.IDENTITY_FEDERATED_SERVER
		im.SetIdentity(id.IdentityChainID, id)
	}

	im.SetAuthority(e.IdentityChainID, auth)
	return nil
}

// ApplyAddAuditServer gets the authority server from the entry's chain id, and sets its status to an audit server. If no existing authority
// server can be found, it adds a new authroity server with the requisite properties
func (im *IdentityManager) ApplyAddAuditServer(entry interfaces.IABEntry, st interfaces.IState) error {
	e := entry.(*adminBlock.AddAuditServer)
	// New server. Check if the identity exists, and create it if it does not
	id := im.GetIdentity(e.IdentityChainID)
	if id == nil {
		st.AddIdentityFromChainID(e.IdentityChainID)
		id = im.GetIdentity(e.IdentityChainID)
	}

	auth := im.GetAuthority(e.IdentityChainID)
	if auth == nil {
		auth = NewAuthority()
	}

	auth.Status = constants.IDENTITY_AUDIT_SERVER
	auth.AuthorityChainID = e.IdentityChainID.(*primitives.Hash)

	if id != nil {
		id.Status = constants.IDENTITY_AUDIT_SERVER
		im.SetIdentity(id.IdentityChainID, id)
	}

	im.SetAuthority(e.IdentityChainID, auth)

	return nil
}

// ApplyRemoveFederatedServer removes the authority server and identity with the entry's chain id
func (im *IdentityManager) ApplyRemoveFederatedServer(entry interfaces.IABEntry) error {
	e := entry.(*adminBlock.RemoveFederatedServer)
	im.RemoveAuthority(e.IdentityChainID)
	im.RemoveIdentity(e.IdentityChainID)
	return nil
}

// ApplyAddFederatedServerSigningKey gets the authority server from the entry's chain id, and adds a new signing key to the authority, placing its
// old signing key in its KeyHistory for posterity
func (im *IdentityManager) ApplyAddFederatedServerSigningKey(entry interfaces.IABEntry) error {
	e := entry.(*adminBlock.AddFederatedServerSigningKey)

	auth := im.GetAuthority(e.IdentityChainID)
	if auth == nil {
		return fmt.Errorf("Authority %v not found!", e.IdentityChainID.String())
	}

	auth.KeyHistory = append(auth.KeyHistory, struct {
		ActiveDBHeight uint32
		SigningKey     primitives.PublicKey
	}{e.DBHeight, auth.SigningKey})

	b, err := e.PublicKey.MarshalBinary()
	if err != nil {
		return err
	}
	err = auth.SigningKey.UnmarshalBinary(b)
	if err != nil {
		return err
	}

	im.SetAuthority(e.IdentityChainID, auth)
	return nil
}

// ApplyAddFederatedServerBitcoinAnchorKey adds the AnchorSigningKey to the federated server
func (im *IdentityManager) ApplyAddFederatedServerBitcoinAnchorKey(entry interfaces.IABEntry) error {
	e := entry.(*adminBlock.AddFederatedServerBitcoinAnchorKey)

	auth := im.GetAuthority(e.IdentityChainID)
	if auth == nil {
		return fmt.Errorf("Authority %v not found", e.IdentityChainID.String())
	}

	var ask AnchorSigningKey
	ask.SigningKey = e.ECDSAPublicKey
	ask.KeyLevel = e.KeyPriority
	ask.KeyType = e.KeyType
	ask.BlockChain = "BTC"

	written := false

	for i, a := range auth.AnchorKeys {
		// We are only dealing with bitcoin keys, so no need to check blockchain
		if a.KeyLevel == ask.KeyLevel && a.KeyType == ask.KeyType {
			if bytes.Compare(a.SigningKey[:], ask.SigningKey[:]) == 0 {
				return nil // Key already exists in authority
			}
			// Overwrite
			written = true
			auth.AnchorKeys[i] = ask
			break
		}
	}

	if !written {
		auth.AnchorKeys = append(auth.AnchorKeys, ask)
	}

	im.SetAuthority(e.IdentityChainID, auth)
	return nil
}

// ApplyServerFault is a no-op function
func (im *IdentityManager) ApplyServerFault(entry interfaces.IABEntry) error {
	//	e := entry.(*adminBlock.ServerFault)
	return nil
}

// ApplyAddFactoidAddress processes the input admin block entry to add a Factoid address to an authority server
func (im *IdentityManager) ApplyAddFactoidAddress(entry interfaces.IABEntry) error {
	e := entry.(*adminBlock.AddFactoidAddress)

	auth := im.GetAuthority(e.IdentityChainID)
	if auth == nil {
		return fmt.Errorf("Authority %v not found!", e.IdentityChainID.String())
	}

	auth.CoinbaseAddress = e.FactoidAddress

	im.SetAuthority(auth.AuthorityChainID, auth)
	return nil
}

// ApplyAddEfficiency processes the input admin block entry to change efficiency of an authority server
func (im *IdentityManager) ApplyAddEfficiency(entry interfaces.IABEntry) error {
	e := entry.(*adminBlock.AddEfficiency)

	auth := im.GetAuthority(e.IdentityChainID)
	if auth == nil {
		return fmt.Errorf("Authority %v not found!", e.IdentityChainID.String())
	}

	auth.Efficiency = e.Efficiency

	im.SetAuthority(auth.AuthorityChainID, auth)
	return nil
}
