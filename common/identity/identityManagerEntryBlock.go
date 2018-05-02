// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identity

import (
	"fmt"

	"bytes"

	"github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/identityEntries"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

func (im *IdentityManager) ProcessIdentityEntry(entry interfaces.IEBEntry, dBlockHeight uint32, dBlockTimestamp interfaces.Timestamp, newEntry bool) (bool, error) {
	return im.ProcessIdentityEntryWithABlockUpdate(entry, dBlockHeight, dBlockTimestamp, nil, newEntry)
}

// ProcessIdentityEntryWithABlockUpdate will process an entry and update an entry. It will also update the admin block
// with any changes.
// There are some special parameters:
//		Params:
//			entry
//			dBlockHeight
//			dBlockTimestamp
//			d					DBState		If not nil, it means to update the admin block with changes
//			newEntry			bool		Setting this to true means it can be put into the oldEntries queue to be reprocesses (helps for out of order entries)
//
//		Returns
//			change				bool		If a key has been changed
//			err					error
func (im *IdentityManager) ProcessIdentityEntryWithABlockUpdate(entry interfaces.IEBEntry, dBlockHeight uint32, dBlockTimestamp interfaces.Timestamp, a interfaces.IAdminBlock, newEntry bool) (bool, error) {
	if entry == nil {
		return false, fmt.Errorf("Entry is nil")
	}

	if bytes.Compare(entry.GetChainID().Bytes()[:3], []byte{0x88, 0x88, 0x88}) != 0 {
		return false, fmt.Errorf("Invalic chainID - expected 888888..., got %v", entry.GetChainID().String())
	}

	chainID := entry.GetChainID()
	extIDs := entry.ExternalIDs()
	var change, tryAgain bool
	if len(extIDs) < 2 {
		//Invalid Identity Chain Entry
		return false, fmt.Errorf("Invalid Identity Chain Entry")
	}

	if string(extIDs[0]) == "Factom Identity Registration Chain" {
		//First entry, can ignore
		return false, nil
	}

	if len(extIDs[0]) == 0 {
		return false, fmt.Errorf("Invalid Identity Chain Entry")
	}
	if extIDs[0][0] != 0 {
		//We only support version 0
		return false, fmt.Errorf("Invalid Identity Chain Entry version")
	}
	switch string(extIDs[1]) {
	case "Identity Chain":
		ic, err := DecodeIdentityChainStructureFromExtIDs(extIDs)
		if err != nil {
			return false, err
		}
		tryAgain, err = im.ApplyIdentityChainStructure(ic, chainID, dBlockHeight)
		if tryAgain == true && newEntry == true {
			//if it's a new entry, push it and return nil
			return false, im.PushEntryForLater(entry, dBlockHeight, dBlockTimestamp)
		}
		//if it's an old entry, return error to signify the entry has not been processed and should be kept
		if err != nil {
			return false, err
		}
		break
	case "New Bitcoin Key":
		nkb, err := DecodeNewBitcoinKeyStructureFromExtIDs(extIDs)
		if err != nil {
			return false, err
		}
		// hard code BTC because it is a "New bitcoin Key"
		change, tryAgain, err = im.ApplyNewBitcoinKeyStructure(nkb, chainID, "BTC", dBlockTimestamp, a)
		if tryAgain == true && newEntry == true {
			//if it's a new entry, push it and return nil
			return false, im.PushEntryForLater(entry, dBlockHeight, dBlockTimestamp)
		}
		//if it's an old entry, return error to signify the entry has not been processed and should be kept
		if err != nil {
			return false, err
		}
		break
	case "New Block Signing Key":
		nbsk, err := DecodeNewBlockSigningKeyStructFromExtIDs(extIDs)
		if err != nil {
			return false, err
		}
		change, tryAgain, err = im.ApplyNewBlockSigningKeyStruct(nbsk, chainID, dBlockTimestamp, a)
		if tryAgain == true && newEntry == true {
			//if it's a new entry, push it and return nil
			return false, im.PushEntryForLater(entry, dBlockHeight, dBlockTimestamp)
		}
		//if it's an old entry, return error to signify the entry has not been processed and should be kept
		if err != nil {
			return false, err
		}
		break
	case "New Matryoshka Hash":
		nmh, err := DecodeNewMatryoshkaHashStructureFromExtIDs(extIDs)
		if err != nil {
			return false, err
		}
		change, tryAgain, err = im.ApplyNewMatryoshkaHashStructure(nmh, dBlockTimestamp, a)
		if tryAgain == true && newEntry == true {
			//if it's a new entry, push it and return nil
			return false, im.PushEntryForLater(entry, dBlockHeight, dBlockTimestamp)
		}
		//if it's an old entry, return error to signify the entry has not been processed and should be kept
		if err != nil {
			return false, err
		}
		break
	case "Register Factom Identity":
		rfi, err := DecodeRegisterFactomIdentityStructureFromExtIDs(extIDs)
		if err != nil {
			return false, err
		}
		tryAgain, err := im.ApplyRegisterFactomIdentityStructure(rfi, dBlockHeight)
		if tryAgain == true && newEntry == true {
			//if it's a new entry, push it and return nil
			return false, im.PushEntryForLater(entry, dBlockHeight, dBlockTimestamp)
		}
		//if it's an old entry, return error to signify the entry has not been processed and should be kept
		if err != nil {
			return false, err
		}
		break
	case "Register Server Management":
		rsm, err := DecodeRegisterServerManagementStructureFromExtIDs(extIDs)
		if err != nil {
			return false, err
		}
		tryAgain, err := im.ApplyRegisterServerManagementStructure(rsm, chainID, dBlockHeight)
		if tryAgain == true && newEntry == true {
			//if it's a new entry, push it and return nil
			return false, im.PushEntryForLater(entry, dBlockHeight, dBlockTimestamp)
		}
		//if it's an old entry, return error to signify the entry has not been processed and should be kept
		if err != nil {
			return false, err
		}
		break
	case "Server Management":
		sm, err := DecodeServerManagementStructureFromExtIDs(extIDs)
		if err != nil {
			return false, err
		}

		tryAgain, err := im.ApplyServerManagementStructure(sm, chainID, dBlockHeight)
		if tryAgain == true && newEntry == true {
			//if it's a new entry, push it and return nil
			return false, im.PushEntryForLater(entry, dBlockHeight, dBlockTimestamp)
		}
		//if it's an old entry, return error to signify the entry has not been processed and should be kept
		if err != nil {
			return false, err
		}
		break
	case "Server Efficiency":
		sm, err := DecodeNewServerEfficiencyStructFromExtIDs(extIDs)
		if err != nil {
			return false, err
		}

		tryAgain, change, err = im.ApplyNewServerEfficiencyStruct(sm, chainID, dBlockTimestamp, a)
		if tryAgain == true && newEntry == true {
			//if it's a new entry, push it and return nil
			return false, im.PushEntryForLater(entry, dBlockHeight, dBlockTimestamp)
		}
		//if it's an old entry, return error to signify the entry has not been processed and should be kept
		if err != nil {
			return false, err
		}
		break
	case "Coinbase Address":
		sm, err := DecodeNewNewCoinbaseAddressStructFromExtIDs(extIDs)
		if err != nil {
			return false, err
		}

		tryAgain, change, err = im.ApplyNewCoinbaseAddressStruct(sm, chainID, dBlockTimestamp, a)
		if tryAgain == true && newEntry == true {
			//if it's a new entry, push it and return nil
			return false, im.PushEntryForLater(entry, dBlockHeight, dBlockTimestamp)
		}
		//if it's an old entry, return error to signify the entry has not been processed and should be kept
		if err != nil {
			return false, err
		}
		break
	case "Coinbase Cancel":
		cc, err := DecodeNewCoinbaseCancelStructFromExtIDs(extIDs)
		if err != nil {
			return false, err
		}
		tryAgain, change, err = im.ApplyNewCoinbaseCancelStruct(cc, chainID, dBlockHeight, a)
		if tryAgain == true && newEntry == true {
			//if it's a new entry, push it and return nil
			return false, im.PushEntryForLater(entry, dBlockHeight, dBlockTimestamp)
		}
		//if it's an old entry, return error to signify the entry has not been processed and should be kept
		if err != nil {
			return false, err
		}
	}

	return change, nil
}

func (im *IdentityManager) ApplyIdentityChainStructure(ic *IdentityChainStructure, chainID interfaces.IHash, dBlockHeight uint32) (bool, error) {
	id := im.GetIdentity(chainID)
	if id == nil {
		id = NewIdentity()
	}

	id.Keys[0] = ic.Key1.(*primitives.Hash)
	id.Keys[1] = ic.Key2.(*primitives.Hash)
	id.Keys[2] = ic.Key3.(*primitives.Hash)
	id.Keys[3] = ic.Key4.(*primitives.Hash)

	id.IdentityCreated = dBlockHeight

	id.IdentityChainID = chainID.(*primitives.Hash)

	im.SetIdentity(chainID, id)

	// The registration could have been parsed earlier, double check.
	if rfi := im.IdentityRegistrations[id.IdentityChainID.Fixed()]; rfi != nil {
		im.ApplyRegisterFactomIdentityStructure(rfi, dBlockHeight)
	}

	return false, nil
}

//		Returns
//			bool	change		If a key has been changed
//			bool	tryagain	If this is set to true, this entry can be reprocessed if it is *new*
//			error	err			Any errors
func (im *IdentityManager) ApplyNewBitcoinKeyStructure(bnk *NewBitcoinKeyStructure, subChainID interfaces.IHash, BlockChain string, dBlockTimestamp interfaces.Timestamp, a interfaces.IAdminBlock) (bool, bool, error) {
	chainID := bnk.RootIdentityChainID

	id := im.GetIdentity(chainID)
	if id == nil {
		return false, true, fmt.Errorf("ChainID doesn't exists! %v", chainID.String())
	}
	err := bnk.VerifySignature(id.Keys[0]) // Key 1
	if err != nil {
		return false, false, err
	}

	if id.ManagementChainID.IsSameAs(subChainID) == false {
		return false, false, fmt.Errorf("Identity Error: Entry was not placed in the correct management chain - %v vs %v", subChainID.String(), id.ManagementChainID.String())
	}

	// Check Timestamp
	if !CheckTimestamp(bnk.Timestamp, dBlockTimestamp.GetTimeSeconds()) {
		return false, false, fmt.Errorf("New Bitcoin key for Identity [%x]is too old", chainID.Bytes()[:5])
	}

	// New Key to add
	var oneAsk AnchorSigningKey
	oneAsk.BlockChain = BlockChain
	oneAsk.KeyLevel = bnk.BitcoinKeyLevel
	oneAsk.KeyType = bnk.KeyType
	oneAsk.SigningKey = bnk.NewKey

	written := false
	for i, a := range id.AnchorKeys {
		// We are only dealing with bitcoin keys, so no need to check blockchain
		if a.KeyLevel == bnk.BitcoinKeyLevel && a.KeyType == bnk.KeyType {
			if bytes.Compare(a.SigningKey[:], bnk.NewKey[:]) == 0 {
				im.SetIdentity(chainID, id)
				return false, false, nil // Key already exists in identity
			} else {
				// Keylevel and keytype exist already. Overwrite
				id.AnchorKeys[i] = oneAsk
				written = true
				break
			}
		}
	}

	if !written {
		id.AnchorKeys = append(id.AnchorKeys, oneAsk)
	}
	im.SetIdentity(chainID, id)

	// Check if we need to update admin block
	if a != nil && im.GetAuthority(chainID) != nil { // Verify is authority
		err = a.AddFederatedServerBitcoinAnchorKey(id.IdentityChainID, oneAsk.KeyLevel, oneAsk.KeyType, oneAsk.SigningKey)
	}

	return true, false, err
}

// ApplyNewBlockSigningKeyStruct will parse a new block signing key and attempt to add the signing to the proper identity.
//		Returns
//			bool	change		If a key has been changed
//			bool	tryagain	If this is set to true, this entry can be reprocessed if it is *new*
//			error	err			Any errors
func (im *IdentityManager) ApplyNewBlockSigningKeyStruct(nbsk *NewBlockSigningKeyStruct, subchainID interfaces.IHash, dBlockTimestamp interfaces.Timestamp, a interfaces.IAdminBlock) (bool, bool, error) {
	chainID := nbsk.RootIdentityChainID
	id := im.GetIdentity(chainID)
	if id == nil {
		return false, true, fmt.Errorf("ChainID doesn't exists! %v", nbsk.RootIdentityChainID.String())
	}

	key := primitives.NewZeroHash()
	err := key.UnmarshalBinary(nbsk.NewPublicKey)
	if err != nil {
		return false, false, err
	}

	if id.SigningKey.IsSameAs(key) {
		return false, false, nil
	}

	err = nbsk.VerifySignature(id.Keys[0])
	if err != nil {
		return false, false, err
	}

	if id.ManagementChainID.IsSameAs(subchainID) == false {
		return false, false, fmt.Errorf("Identity Error: Entry was not placed in the correct management chain - %v vs %v", id.ManagementChainID.String(), subchainID.String())
	}

	// Check Timestamp
	if !CheckTimestamp(nbsk.Timestamp, dBlockTimestamp.GetTimeSeconds()) {
		return false, false, fmt.Errorf("New Block Signing key for Identity [%x]is too old", chainID.Bytes()[:5])
	}

	id.SigningKey = key.(*primitives.Hash)

	im.SetIdentity(nbsk.RootIdentityChainID, id)

	// Check if we need to update admin block
	if a != nil && im.GetAuthority(nbsk.RootIdentityChainID) != nil { // Verify is authority
		err = a.AddFederatedServerSigningKey(id.IdentityChainID, key.Fixed())
	}

	return true, false, err
}

// ApplyNewMatryoshkaHashStructure will parse a new matryoshka hash and attempt to add the signing to the proper identity.
//		Returns
//			bool	change		If a key has been changed
//			bool	tryagain	If this is set to true, this entry can be reprocessed if it is *new*
//			error	err			Any errors
func (im *IdentityManager) ApplyNewMatryoshkaHashStructure(nmh *NewMatryoshkaHashStructure, dBlockTimestamp interfaces.Timestamp, a interfaces.IAdminBlock) (bool, bool, error) {
	id := im.GetIdentity(nmh.RootIdentityChainID)
	if id == nil {
		return false, true, fmt.Errorf("ChainID doesn't exists! %v", nmh.RootIdentityChainID.String())
	}

	if nmh.OutermostMHash.IsSameAs(id.MatryoshkaHash) {
		return false, false, nil
	}

	err := nmh.VerifySignature(id.Keys[0])
	if err != nil {
		return false, false, err
	}

	// Check Timestamp
	if !CheckTimestamp(nmh.Timestamp, dBlockTimestamp.GetTimeSeconds()) {
		return false, false, fmt.Errorf("New MHash for Identity [%x]is too old", nmh.RootIdentityChainID.Bytes()[:5])
	}

	id.MatryoshkaHash = nmh.OutermostMHash.(*primitives.Hash)

	im.SetIdentity(nmh.RootIdentityChainID, id)

	// Check if we need to update admin block
	if a != nil && im.GetAuthority(nmh.RootIdentityChainID) != nil { // Verify is authority
		err = a.AddMatryoshkaHash(id.IdentityChainID, id.MatryoshkaHash)
	}

	return true, false, err
}

func (im *IdentityManager) ApplyRegisterFactomIdentityStructure(rfi *RegisterFactomIdentityStructure, dBlockHeight uint32) (bool, error) {
	im.IdentityRegistrations[rfi.IdentityChainID.Fixed()] = rfi

	id := im.GetIdentity(rfi.IdentityChainID)
	if id == nil {
		return true, fmt.Errorf("ChainID doesn't exists! %v", rfi.IdentityChainID.String())
	}

	err := rfi.VerifySignature(id.Keys[0])
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

	err := rsm.VerifySignature(id.Keys[0])
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

// ApplyServerManagementStructure is the first entry in the management chain
//	DO NOT set the management chain in the identity, as it will be set on the register
//		"Server Management"
func (im *IdentityManager) ApplyServerManagementStructure(sm *ServerManagementStructure, chainID interfaces.IHash, dBlockHeight uint32) (bool, error) {
	id := im.GetIdentity(sm.RootIdentityChainID)
	if id == nil {
		return true, fmt.Errorf("ChainID doesn't exists! %v", chainID.String())
	}

	if id.ManagementCreated != 0 {
		return true, fmt.Errorf("ManagementCreated is already set.")
	}
	id.ManagementCreated = dBlockHeight

	im.SetIdentity(sm.RootIdentityChainID, id)
	return false, nil
}

// ApplyNewServerEfficiencyStruct will parse a new server efficiency and attempt to add the signing to the proper identity.
//		Returns
//			bool	change		If a key has been changed
//			bool	tryagain	If this is set to true, this entry can be reprocessed if it is *new*
//			error	err			Any errors
func (im *IdentityManager) ApplyNewServerEfficiencyStruct(nses *NewServerEfficiencyStruct, subchainID interfaces.IHash, dBlockTimestamp interfaces.Timestamp, a interfaces.IAdminBlock) (bool, bool, error) {
	chainID := nses.RootIdentityChainID
	id := im.GetIdentity(chainID)
	if id == nil {
		return false, true, fmt.Errorf("ChainID doesn't exists! %v", nses.RootIdentityChainID.String())
	}

	if id.Efficiency == nses.Efficiency {
		return false, false, nil
	}

	err := nses.VerifySignature(id.Keys[0])
	if err != nil {
		return false, false, err
	}

	if id.ManagementChainID.IsSameAs(subchainID) == false {
		return false, false, fmt.Errorf("Identity Error: Entry was not placed in the correct management chain - %v vs %v", id.ManagementChainID.String(), subchainID.String())
	}

	// Check Timestamp
	if !CheckTimestamp(nses.Timestamp, dBlockTimestamp.GetTimeSeconds()) {
		return false, false, fmt.Errorf("New Server Efficiency for Identity [%x]is too old", chainID.Bytes()[:5])
	}

	if nses.Efficiency > 10000 {
		nses.Efficiency = 10000
	}

	id.Efficiency = nses.Efficiency

	im.SetIdentity(nses.RootIdentityChainID, id)

	// Check if we need to update admin block
	if a != nil && im.GetAuthority(nses.RootIdentityChainID) != nil { // Verify is authority
		err = a.AddEfficiency(nses.RootIdentityChainID, nses.Efficiency)
	}

	return true, false, err
}

// ApplyNewCoinbaseAddressStruct will parse a new coinbase address and attempt to add the signing to the proper identity.
//		Returns
//			bool	change		If a key has been changed
//			bool	tryagain	If this is set to true, this entry can be reprocessed if it is *new*
//			error	err			Any errors
func (im *IdentityManager) ApplyNewCoinbaseAddressStruct(ncas *NewCoinbaseAddressStruct, rootchainID interfaces.IHash, dBlockTimestamp interfaces.Timestamp, a interfaces.IAdminBlock) (bool, bool, error) {
	root := ncas.RootIdentityChainID
	id := im.GetIdentity(root)
	if id == nil {
		return false, true, fmt.Errorf("(coinbase address) ChainID doesn't exists! %v", ncas.RootIdentityChainID.String())
	}

	if !rootchainID.IsSameAs(ncas.RootIdentityChainID) {
		return false, true, fmt.Errorf("(coinbase address) ChainID of entry should match root chain id.")
	}

	if id.CoinbaseAddress.IsSameAs(ncas.CoinbaseAddress) {
		return false, false, nil
	}

	err := ncas.VerifySignature(id.Keys[0])
	if err != nil {
		return false, false, err
	}

	// Check Timestamp
	if !CheckTimestamp(ncas.Timestamp, dBlockTimestamp.GetTimeSeconds()) {
		return false, false, fmt.Errorf("New Server Efficiency for Identity [%x]is too old", root.Bytes()[:5])
	}

	id.CoinbaseAddress = ncas.CoinbaseAddress

	im.SetIdentity(ncas.RootIdentityChainID, id)

	// Check if we need to update admin block
	if a != nil && im.GetAuthority(ncas.RootIdentityChainID) != nil { // Verify is authority
		err = a.AddCoinbaseAddress(ncas.RootIdentityChainID, ncas.CoinbaseAddress)
	}

	return true, false, err
}

// ApplyNewCoinbaseCancelStruct will parse a new coinbase cancel
//		Validation Difference:
//			Most entries check the timestamps to ensure entry is within 12 hours of dblock. That does not apply to coinbase, it can be replayed
//			as long as it is between the block window.
//		Returns
//			bool	change		If a key has been changed
//			bool	tryagain	If this is set to true, this entry can be reprocessed if it is *new*
//			error	err			Any errors
func (im *IdentityManager) ApplyNewCoinbaseCancelStruct(nccs *NewCoinbaseCancelStruct, managechain interfaces.IHash, dblockHeight uint32, a interfaces.IAdminBlock) (bool, bool, error) {
	// Validate Block window
	//		If the descriptor to cancel has already been applied, then this entry is no longer valid
	//		Descriptor height + Declaration is the block the coinbase is added
	if dblockHeight > nccs.CoinbaseDescriptorHeight+constants.COINBASE_DECLARATION {
		return false, false, nil
	}

	root := nccs.RootIdentityChainID
	id := im.GetIdentity(root)
	if id == nil {
		return false, true, fmt.Errorf("(coinbase cancel) ChainID doesn't exists! %v", nccs.RootIdentityChainID.String())
	}

	if !managechain.IsSameAs(id.ManagementChainID) {
		return false, true, fmt.Errorf("(coinbase cancel) ChainID of entry should match manage chain id.")
	}

	// TODO: Check if this cancel is a repeat. If it is, we can ignore it as it's already been counted
	if false {
		return false, false, nil
	}

	err := nccs.VerifySignature(id.Keys[0])
	if err != nil {
		return false, false, err
	}

	// TODO: Add count to cancel counting

	// Check if we need to update admin block
	if a != nil {
		// TODO: Check if it reaches critical mass
		// 		TODO: Add to admin block if it does
		// err = a.AddCoinbaseAddress(ncas.RootIdentityChainID, ncas.CoinbaseAddress)
	}
	return false, false, nil
}
