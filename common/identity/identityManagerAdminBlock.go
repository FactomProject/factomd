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

func (im *IdentityManager) ApplyRevealMatryoshkaHash(entry interfaces.IABEntry) error {
	//e:=entry.(*adminBlock.RevealMatryoshkaHash)
	// Does nothing for authority right now
	return nil
}

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

func (im *IdentityManager) ApplyIncreaseServerCount(entry interfaces.IABEntry) error {
	e := entry.(*adminBlock.IncreaseServerCount)
	im.AuthorityServerCount = im.AuthorityServerCount + int(e.Amount)
	return nil
}

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

func (im *IdentityManager) ApplyRemoveFederatedServer(entry interfaces.IABEntry) error {
	e := entry.(*adminBlock.RemoveFederatedServer)
	im.RemoveAuthority(e.IdentityChainID)
	im.RemoveIdentity(e.IdentityChainID)
	return nil
}

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
			} else {
				// Overwrite
				written = true
				auth.AnchorKeys[i] = ask
				break
			}
		}
	}

	if !written {
		auth.AnchorKeys = append(auth.AnchorKeys, ask)
	}

	im.SetAuthority(e.IdentityChainID, auth)
	return nil
}

func (im *IdentityManager) ApplyServerFault(entry interfaces.IABEntry) error {
	//	e := entry.(*adminBlock.ServerFault)
	return nil
}

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
