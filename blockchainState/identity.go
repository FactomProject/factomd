// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockchainState

import (
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/meta"
)

func (bs *BlockchainState) ProcessIdentityEntry(entry interfaces.IEBEntry) error {
	if entry.GetChainID().String() != "888888001750ede0eff4b05f0c3f557890b256450cabbb84cada937f9c258327" {
		return fmt.Errorf("Invalic chainID - expected 888888001750ede0eff4b05f0c3f557890b256450cabbb84cada937f9c258327, got %v", entry.GetChainID().String())
	}

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
		fmt.Printf("%v", ic)
		break
	case "New Bitcoin Key":
		nkb, err := meta.DecodeNewBitcoinKeyStructureFromExtIDs(extIDs)
		if err != nil {
			return err
		}
		fmt.Printf("%v", nkb)
		break
	case "New Block Signing Key":
		nbsk, err := meta.DecodeNewBlockSigningKeyStructFromExtIDs(extIDs)
		if err != nil {
			return err
		}
		fmt.Printf("%v", nbsk)
		break
	case "New Matryoshka Hash":
		nmh, err := meta.DecodeNewMatryoshkaHashStructureFromExtIDs(extIDs)
		if err != nil {
			return err
		}
		fmt.Printf("%v", nmh)
		break
	case "Register Factom Identity":
		rfi, err := meta.DecodeRegisterFactomIdentityStructureFromExtIDs(extIDs)
		if err != nil {
			return err
		}
		fmt.Printf("%v", rfi)
		break
	case "Register Server Management":
		rsm, err := meta.DecodeRegisterServerManagementStructureFromExtIDs(extIDs)
		if err != nil {
			return err
		}
		fmt.Printf("%v", rsm)
		break
	case "Server Management":
		sm, err := meta.DecodeServerManagementStructureFromExtIDs(extIDs)
		if err != nil {
			return err
		}
		fmt.Printf("%v", sm)
		break
	}

	return nil
}
