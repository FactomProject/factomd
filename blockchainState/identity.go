// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockchainState

import (
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/meta"
)

func (bs *BlockchainState) ProcessIdentityEntry(entry interfaces.IEBEntry, dBlockHeight uint32, dBlockTimestamp interfaces.Timestamp) error {
	if entry.GetChainID().String() != "888888001750ede0eff4b05f0c3f557890b256450cabbb84cada937f9c258327" {
		return fmt.Errorf("Invalic chainID - expected 888888001750ede0eff4b05f0c3f557890b256450cabbb84cada937f9c258327, got %v", entry.GetChainID().String())
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
	if extIDs[0][0] != 1 {
		//We only support version 1
		return fmt.Errorf("Invalid Identity Chain Entry version")
	}
	switch string(extIDs[1]) {
	case "Identity Chain":
		ic, err := meta.DecodeIdentityChainStructureFromExtIDs(extIDs)
		if err != nil {
			return err
		}
		err = bs.IdentityManager.ApplyIdentityChainStructure(ic, chainID, dBlockHeight)
		if err != nil {
			return err
		}
		break
	case "New Bitcoin Key":
		nkb, err := meta.DecodeNewBitcoinKeyStructureFromExtIDs(extIDs)
		if err != nil {
			return err
		}
		err = bs.IdentityManager.ApplyNewBitcoinKeyStructure(nkb, chainID, "???????????????", dBlockTimestamp)
		if err != nil {
			return err
		}
		break
	case "New Block Signing Key":
		nbsk, err := meta.DecodeNewBlockSigningKeyStructFromExtIDs(extIDs)
		if err != nil {
			return err
		}
		err = bs.IdentityManager.ApplyNewBlockSigningKeyStruct(nbsk, chainID, dBlockTimestamp)
		if err != nil {
			return err
		}
		break
	case "New Matryoshka Hash":
		nmh, err := meta.DecodeNewMatryoshkaHashStructureFromExtIDs(extIDs)
		if err != nil {
			return err
		}
		err = bs.IdentityManager.ApplyNewMatryoshkaHashStructure(nmh, dBlockTimestamp)
		if err != nil {
			return err
		}
		break
	case "Register Factom Identity":
		rfi, err := meta.DecodeRegisterFactomIdentityStructureFromExtIDs(extIDs)
		if err != nil {
			return err
		}
		err = bs.IdentityManager.ApplyRegisterFactomIdentityStructure(rfi, dBlockHeight)
		if err != nil {
			return err
		}
		break
	case "Register Server Management":
		rsm, err := meta.DecodeRegisterServerManagementStructureFromExtIDs(extIDs)
		if err != nil {
			return err
		}
		err = bs.IdentityManager.ApplyRegisterServerManagementStructure(rsm, chainID, dBlockHeight)
		if err != nil {
			return err
		}
		break
	case "Server Management":
		sm, err := meta.DecodeServerManagementStructureFromExtIDs(extIDs)
		if err != nil {
			return err
		}
		err = bs.IdentityManager.ApplyServerManagementStructure(sm, chainID, dBlockHeight)
		if err != nil {
			return err
		}
		break
	}

	return nil
}
