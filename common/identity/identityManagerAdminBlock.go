// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identity

import (
	"fmt"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

func (im *IdentityManager) ProcessABlockEntry(entry interfaces.IABEntry) error {
	switch entry.Type() {
	case constants.TYPE_REVEAL_MATRYOSHKA:
		return im.ApplyRevealMatryoshkaHash(entry)
	case constants.TYPE_ADD_MATRYOSHKA:
		return im.ApplyAddReplaceMatryoshkaHash(entry)
	case constants.TYPE_ADD_SERVER_COUNT:
		return im.ApplyIncreaseServerCount(entry)
	case constants.TYPE_ADD_FED_SERVER:
		return im.ApplyAddFederatedServer(entry)
	case constants.TYPE_ADD_AUDIT_SERVER:
		return im.ApplyAddAuditServer(entry)
	case constants.TYPE_REMOVE_FED_SERVER:
		return im.ApplyRemoveFederatedServer(entry)
	case constants.TYPE_ADD_FED_SERVER_KEY:
		return im.ApplyAddFederatedServerSigningKey(entry)
	case constants.TYPE_ADD_BTC_ANCHOR_KEY:
		return im.ApplyAddFederatedServerBitcoinAnchorKey(entry)
	case constants.TYPE_SERVER_FAULT:
		return im.ApplyServerFault(entry)
	}
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

func (im *IdentityManager) ApplyAddFederatedServer(entry interfaces.IABEntry) error {
	e := entry.(*adminBlock.AddFederatedServer)

	auth := im.GetAuthority(e.IdentityChainID)
	if auth == nil {
		auth = new(Authority)
	}

	auth.Status.Store(constants.IDENTITY_FEDERATED_SERVER)
	auth.AuthorityChainID = e.IdentityChainID.(*primitives.Hash)

	im.SetAuthority(e.IdentityChainID, auth)
	return nil
}

func (im *IdentityManager) ApplyAddAuditServer(entry interfaces.IABEntry) error {
	e := entry.(*adminBlock.AddAuditServer)

	auth := im.GetAuthority(e.IdentityChainID)
	if auth == nil {
		auth = new(Authority)
	}

	auth.Status.Store(constants.IDENTITY_AUDIT_SERVER)
	auth.AuthorityChainID = e.IdentityChainID.(*primitives.Hash)

	im.SetAuthority(e.IdentityChainID, auth)

	return nil
}

func (im *IdentityManager) ApplyRemoveFederatedServer(entry interfaces.IABEntry) error {
	e := entry.(*adminBlock.RemoveFederatedServer)
	im.RemoveAuthority(e.IdentityChainID)
	return nil
}

func (im *IdentityManager) ApplyAddFederatedServerSigningKey(entry interfaces.IABEntry) error {
	e := entry.(*adminBlock.AddFederatedServerSigningKey)

	auth := im.GetAuthority(e.IdentityChainID)
	if auth == nil {
		return fmt.Errorf("Authority %v not found!", e.IdentityChainID.String())
	}

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
	//ask.BlockChain = e.

	auth.AnchorKeys = append(auth.AnchorKeys, ask)

	im.SetAuthority(e.IdentityChainID, auth)
	return nil
}

func (im *IdentityManager) ApplyServerFault(entry interfaces.IABEntry) error {
	//	e := entry.(*adminBlock.ServerFault)
	return nil
}
