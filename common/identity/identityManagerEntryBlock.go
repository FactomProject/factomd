// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identity

import (
	"fmt"
	"strings"

	. "github.com/FactomProject/factomd/common/identityEntries"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

func (im *IdentityManager) ProcessIdentityEntry(entry interfaces.IEBEntry, dBlockHeight uint32, dBlockTimestamp interfaces.Timestamp, newEntry bool) error {
	if entry.GetChainID().String()[:6] != "888888" {
		return fmt.Errorf("Invalic chainID - expected 888888..., got %v", entry.GetChainID().String())
	}
	if entry.GetHash().String() == "172eb5cb84a49280c9ad0baf13bea779a624def8d10adab80c3d007fe8bce9ec" {
		//First entry, can ignore
		return nil
	}

	chainID := entry.GetChainID()

	extIDs := entry.ExternalIDs()
	if len(extIDs) < 2 {
		//Invalid Identity Chain Entry
		return fmt.Errorf("Invalid Identity Chain Entry")
	}
	if len(extIDs[0]) == 0 {
		return fmt.Errorf("Invalid Identity Chain Entry")
	}
	if extIDs[0][0] != 0 {
		//We only support version 0
		return fmt.Errorf("Invalid Identity Chain Entry version")
	}
	switch string(extIDs[1]) {
	case "Identity Chain":
		ic, err := DecodeIdentityChainStructureFromExtIDs(extIDs)
		if err != nil {
			return err
		}
		tryAgain, err := im.ApplyIdentityChainStructure(ic, chainID, dBlockHeight)
		if tryAgain == true && newEntry == true {
			//if it's a new entry, push it and return nil
			return im.PushEntryForLater(entry, dBlockHeight, dBlockTimestamp)
		}
		//if it's an old entry, return error to signify the entry has not been processed and should be kept
		if err != nil {
			return err
		}
		break
	case "New Bitcoin Key":
		nkb, err := DecodeNewBitcoinKeyStructureFromExtIDs(extIDs)
		if err != nil {
			return err
		}
		tryAgain, err := im.ApplyNewBitcoinKeyStructure(nkb, chainID, "???????????????", dBlockTimestamp)
		if tryAgain == true && newEntry == true {
			//if it's a new entry, push it and return nil
			return im.PushEntryForLater(entry, dBlockHeight, dBlockTimestamp)
		}
		//if it's an old entry, return error to signify the entry has not been processed and should be kept
		if err != nil {
			return err
		}
		break
	case "New Block Signing Key":
		nbsk, err := DecodeNewBlockSigningKeyStructFromExtIDs(extIDs)
		if err != nil {
			return err
		}
		tryAgain, err := im.ApplyNewBlockSigningKeyStruct(nbsk, chainID, dBlockTimestamp)
		if tryAgain == true && newEntry == true {
			//if it's a new entry, push it and return nil
			return im.PushEntryForLater(entry, dBlockHeight, dBlockTimestamp)
		}
		//if it's an old entry, return error to signify the entry has not been processed and should be kept
		if err != nil {
			return err
		}
		break
	case "New Matryoshka Hash":
		nmh, err := DecodeNewMatryoshkaHashStructureFromExtIDs(extIDs)
		if err != nil {
			return err
		}
		tryAgain, err := im.ApplyNewMatryoshkaHashStructure(nmh, dBlockTimestamp)
		if tryAgain == true && newEntry == true {
			//if it's a new entry, push it and return nil
			return im.PushEntryForLater(entry, dBlockHeight, dBlockTimestamp)
		}
		//if it's an old entry, return error to signify the entry has not been processed and should be kept
		if err != nil {
			return err
		}
		break
	case "Register Factom Identity":
		rfi, err := DecodeRegisterFactomIdentityStructureFromExtIDs(extIDs)
		if err != nil {
			return err
		}
		tryAgain, err := im.ApplyRegisterFactomIdentityStructure(rfi, dBlockHeight)
		if tryAgain == true && newEntry == true {
			//if it's a new entry, push it and return nil
			return im.PushEntryForLater(entry, dBlockHeight, dBlockTimestamp)
		}
		//if it's an old entry, return error to signify the entry has not been processed and should be kept
		if err != nil {
			return err
		}
		break
	case "Register Server Management":
		rsm, err := DecodeRegisterServerManagementStructureFromExtIDs(extIDs)
		if err != nil {
			return err
		}
		tryAgain, err := im.ApplyRegisterServerManagementStructure(rsm, chainID, dBlockHeight)
		if tryAgain == true && newEntry == true {
			//if it's a new entry, push it and return nil
			return im.PushEntryForLater(entry, dBlockHeight, dBlockTimestamp)
		}
		//if it's an old entry, return error to signify the entry has not been processed and should be kept
		if err != nil {
			return err
		}
		break
	case "Server Management":
		sm, err := DecodeServerManagementStructureFromExtIDs(extIDs)
		if err != nil {
			return err
		}
		tryAgain, err := im.ApplyServerManagementStructure(sm, chainID, dBlockHeight)
		if tryAgain == true && newEntry == true {
			//if it's a new entry, push it and return nil
			return im.PushEntryForLater(entry, dBlockHeight, dBlockTimestamp)
		}
		//if it's an old entry, return error to signify the entry has not been processed and should be kept
		if err != nil {
			return err
		}
		break
	}

	return nil
}

func (im *IdentityManager) ApplyIdentityChainStructure(ic *IdentityChainStructure, chainID interfaces.IHash, dBlockHeight uint32) (bool, error) {
	id := im.GetIdentity(chainID)
	if id != nil {
		return false, fmt.Errorf("ChainID already exists! %v", chainID.String())
	}

	id = new(Identity)

	id.Key1 = ic.Key1.(*primitives.Hash)
	id.Key2 = ic.Key2.(*primitives.Hash)
	id.Key3 = ic.Key3.(*primitives.Hash)
	id.Key4 = ic.Key4.(*primitives.Hash)

	id.IdentityCreated = dBlockHeight

	id.IdentityChainID = chainID.(*primitives.Hash)

	im.SetIdentity(chainID, id)
	return false, nil
}

func (im *IdentityManager) ApplyNewBitcoinKeyStructure(bnk *NewBitcoinKeyStructure, subChainID interfaces.IHash, BlockChain string, dBlockTimestamp interfaces.Timestamp) (bool, error) {
	chainID := bnk.RootIdentityChainID

	id := im.GetIdentity(chainID)
	if id == nil {
		return true, fmt.Errorf("ChainID doesn't exists! %v", chainID.String())
	}
	err := bnk.VerifySignature(id.Key1)
	if err != nil {
		return false, err
	}

	if id.ManagementChainID.IsSameAs(subChainID) == false {
		return false, fmt.Errorf("Identity Error: Entry was not placed in the correct management chain - %v vs %v", subChainID.String(), id.ManagementChainID.String())
	}

	var ask []AnchorSigningKey
	var newAsk []AnchorSigningKey
	var oneAsk AnchorSigningKey

	ask = id.AnchorKeys
	newAsk = make([]AnchorSigningKey, len(ask)+1)

	oneAsk.BlockChain = BlockChain
	oneAsk.KeyLevel = bnk.BitcoinKeyLevel
	oneAsk.KeyType = bnk.KeyType
	oneAsk.SigningKey = bnk.NewKey

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

	if contains {
		id.AnchorKeys = ask
	} else {
		id.AnchorKeys = newAsk
	}

	/*
		dbase := st.GetAndLockDB()
		dblk, err := dbase.FetchDBlockByHeight(height)

		if err == nil && dblk != nil && dblk.GetHeader().GetTimestamp().GetTimeSeconds() != 0 {
			if !CheckTimestamp(extIDs[6], dblk.GetHeader().GetTimestamp().GetTimeSeconds()) {
				return errors.New("New Anchor key for identity [" + chainID.String()[:10] + "] timestamp is too old")
			}
		} else {
			if !CheckTimestamp(extIDs[6], st.GetTimestamp().GetTimeSeconds()) {
				return errors.New("New Anchor key for identity [" + chainID.String()[:10] + "] timestamp is too old")
			}
		}
	*/

	/*
		// Add to admin block
		status := st.Identities[IdentityIndex].Status
		if !initial && statusIsFedOrAudit(status) && st.GetLeaderVM() == st.ComputeVMIndex(entry.GetChainID().Bytes()) {
			copy(key[:20], extIDs[5][:20])
			extIDs[5] = append(extIDs[5], []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}...)
			key := primitives.NewHash(extIDs[5])
			msg := messages.NewChangeServerKeyMsg(st, chainID, constants.TYPE_ADD_BTC_ANCHOR_KEY, extIDs[3][0], extIDs[4][0], key)
			err := msg.(*messages.ChangeServerKeyMsg).Sign(st.serverPrivKey)
			if err != nil {
				return errors.New("New Block Signing key for identity [" + chainID.String()[:10] + "] Error: cannot sign msg")
			}
			st.InMsgQueue() <- msg
		}
	*/

	im.SetIdentity(chainID, id)
	return false, nil
}

func (im *IdentityManager) ApplyNewBlockSigningKeyStruct(nbsk *NewBlockSigningKeyStruct, subchainID interfaces.IHash, dBlockTimestamp interfaces.Timestamp) (bool, error) {
	chainID := nbsk.RootIdentityChainID
	id := im.GetIdentity(chainID)
	if id == nil {
		return true, fmt.Errorf("ChainID doesn't exists! %v", nbsk.RootIdentityChainID.String())
	}
	err := nbsk.VerifySignature(id.Key1)
	if err != nil {
		return false, err
	}

	if id.ManagementChainID.IsSameAs(subchainID) == false {
		return false, fmt.Errorf("Identity Error: Entry was not placed in the correct management chain - %v vs %v", id.ManagementChainID.String(), subchainID.String())
	}

	//Check Timestamp??

	key := primitives.NewZeroHash()
	err = key.UnmarshalBinary(nbsk.NewPublicKey)
	if err != nil {
		return false, err
	}
	id.SigningKey = key.(*primitives.Hash)

	im.SetIdentity(nbsk.RootIdentityChainID, id)
	return false, nil

	/*

		func RegisterBlockSigningKey(entry interfaces.IEBEntry, initial bool, height uint32, st *State) error {





		dbase := st.GetAndLockDB()
		dblk, err := dbase.FetchDBlockByHeight(height)


		if err == nil && dblk != nil && dblk.GetHeader().GetTimestamp().GetTimeSeconds() != 0 {
			if !CheckTimestamp(extIDs[4], dblk.GetHeader().GetTimestamp().GetTimeSeconds()) {
				return errors.New("New Block Signing key for identity  [" + chainID.String()[:10] + "] timestamp is too old")
			}
		} else {
			if !CheckTimestamp(extIDs[4], st.GetTimestamp().GetTimeSeconds()) {
				return errors.New("New Block Signing key for identity  [" + chainID.String()[:10] + "] timestamp is too old")
			}
		}
	*/

	/*
		st.Identities[IdentityIndex].SigningKey = primitives.NewHash(extIDs[3])
		// Add to admin block if the following:
		//		Not the initial load
		//		A Federated or Audit server
		//		This node is charge of admin block
		status := st.Identities[IdentityIndex].Status
		if !initial && statusIsFedOrAudit(status) && st.GetLeaderVM() == st.ComputeVMIndex(entry.GetChainID().Bytes()) {
			key := primitives.NewHash(extIDs[3])
			msg := messages.NewChangeServerKeyMsg(st, chainID, constants.TYPE_ADD_FED_SERVER_KEY, 0, 0, key)
			err := msg.(*messages.ChangeServerKeyMsg).Sign(st.serverPrivKey)
			if err != nil {
				return errors.New("New Block Signing key for identity [" + chainID.String()[:10] + "] Error: cannot sign msg")
			}
			st.InMsgQueue() <- msg

	*/
}

func (im *IdentityManager) ApplyNewMatryoshkaHashStructure(nmh *NewMatryoshkaHashStructure, dBlockTimestamp interfaces.Timestamp) (bool, error) {
	id := im.GetIdentity(nmh.RootIdentityChainID)
	if id == nil {
		return true, fmt.Errorf("ChainID doesn't exists! %v", nmh.RootIdentityChainID.String())
	}
	err := nmh.VerifySignature(id.Key1)
	if err != nil {
		return false, err
	}
	//Check Timestamp??

	id.MatryoshkaHash = nmh.OutermostMHash.(*primitives.Hash)

	im.SetIdentity(nmh.RootIdentityChainID, id)
	return false, nil
}

func (im *IdentityManager) ApplyRegisterFactomIdentityStructure(rfi *RegisterFactomIdentityStructure, dBlockHeight uint32) (bool, error) {
	id := im.GetIdentity(rfi.IdentityChainID)
	if id == nil {
		return true, fmt.Errorf("ChainID doesn't exists! %v", rfi.IdentityChainID.String())
	}

	err := rfi.VerifySignature(id.Key1)
	if err != nil {
		return false, err
	}

	id.IdentityRegistered = dBlockHeight

	im.SetIdentity(id.IdentityChainID, id)
	return false, nil
}

func (im *IdentityManager) ApplyRegisterServerManagementStructure(rsm *RegisterServerManagementStructure, chainID interfaces.IHash, dBlockHeight uint32) (bool, error) {
	id := im.GetIdentity(chainID)
	if id == nil {
		return true, fmt.Errorf("ChainID doesn't exists! %v", chainID.String())
	}

	err := rsm.VerifySignature(id.Key1)
	if err != nil {
		return false, err
	}

	if id.ManagementRegistered == 0 {
		id.ManagementRegistered = dBlockHeight
	}
	id.ManagementChainID = rsm.SubchainChainID.(*primitives.Hash)

	im.SetIdentity(id.IdentityChainID, id)
	return false, nil
}

func (im *IdentityManager) ApplyServerManagementStructure(sm *ServerManagementStructure, chainID interfaces.IHash, dBlockHeight uint32) (bool, error) {
	id := im.GetIdentity(chainID)
	if id == nil {
		return true, fmt.Errorf("ChainID doesn't exists! %v", chainID.String())
	}

	id.ManagementCreated = dBlockHeight

	im.SetIdentity(chainID, id)
	return false, nil
}
