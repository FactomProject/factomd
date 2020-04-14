// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identityEntries

import (
	"fmt"

	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// ExpectedIdentityChainExternalIDLengths is a hardcoded slice containing the expected lengths of each entry in an external ID (the fields of IdentityChainStructure)
var ExpectedIdentityChainExternalIDLengths = []int{1, 14, 32, 32, 32, 32}

// IdentityChainStructure contains all the elements for forming an identity on the Factom blockchain
// https://github.com/FactomProject/FactomDocs/blob/master/Identity.md#factom-identity-chain-creation
type IdentityChainStructure struct {
	//A Chain Name is constructed with 7 elements.
	//The first element is a binary string 0 signifying the version.
	Version byte
	//The second element is ASCII bytes "Identity Chain".
	FunctionName []byte //"Identity Chain"
	//The third element is the level 1 identity key in binary form.
	Key1 interfaces.IHash // Level 1 key is the lowest security online key
	//Elements 4-6 are levels 2-4.
	Key2 interfaces.IHash
	Key3 interfaces.IHash
	Key4 interfaces.IHash
	//The 7th element is a nonce which is iterated until the first 3 bytes match 0x888888.
	Nonce []byte
}

// DecodeIdentityChainStructureFromExtIDs returns a new object with values from the input external ID
func DecodeIdentityChainStructureFromExtIDs(extIDs [][]byte) (*IdentityChainStructure, error) {
	ics := new(IdentityChainStructure)
	err := ics.DecodeFromExtIDs(extIDs)
	if err != nil {
		return nil, err
	}
	return ics, nil
}

// DecodeFromExtIDs sets the values from the input 2d byte slice into this object
func (ics *IdentityChainStructure) DecodeFromExtIDs(extIDs [][]byte) error {
	// An external id contains the 7 fields of the IdentityChainStructure
	if len(extIDs) != 7 {
		return fmt.Errorf("Wrong number of ExtIDs - expected 7, got %v", len(extIDs))
	}
	// Check each fields length matches the expected size
	if CheckExternalIDsLength(extIDs[:6], ExpectedIdentityChainExternalIDLengths) == false {
		return fmt.Errorf("Wrong lengths of ExtIDs")
	}
	ics.Version = extIDs[0][0]
	if ics.Version != 0 {
		return fmt.Errorf("Wrong Version - expected 0, got %v", ics.Version)
	}
	ics.FunctionName = extIDs[1]
	if string(ics.FunctionName) != "Identity Chain" {
		return fmt.Errorf("Invalid FunctionName - expected 'Identity Chain', got '%s'", ics.FunctionName)
	}
	h, err := primitives.NewShaHash(extIDs[2])
	if err != nil {
		return err
	}
	ics.Key1 = h
	h, err = primitives.NewShaHash(extIDs[3])
	if err != nil {
		return err
	}
	ics.Key2 = h
	h, err = primitives.NewShaHash(extIDs[4])
	if err != nil {
		return err
	}
	ics.Key3 = h
	h, err = primitives.NewShaHash(extIDs[5])
	if err != nil {
		return err
	}
	ics.Key4 = h
	ics.Nonce = extIDs[6]

	chainID := ics.GetChainID()
	if chainID.String()[:6] != "888888" {
		return fmt.Errorf("Invalid ChainID - it should start with '888888', but doesn't - %v", chainID.String())
	}
	return nil
}

// ToExternalIDs returns a 2d byte slice of all the data in this object
func (ics *IdentityChainStructure) ToExternalIDs() [][]byte {
	extIDs := [][]byte{}

	extIDs = append(extIDs, []byte{ics.Version})
	extIDs = append(extIDs, ics.FunctionName)
	extIDs = append(extIDs, ics.Key1.Bytes())
	extIDs = append(extIDs, ics.Key2.Bytes())
	extIDs = append(extIDs, ics.Key3.Bytes())
	extIDs = append(extIDs, ics.Key4.Bytes())
	extIDs = append(extIDs, ics.Nonce)

	return extIDs
}

// GetChainID computes the chain ID associated with this object
func (ics *IdentityChainStructure) GetChainID() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "IdentityChainStructure.GetChainID") }()

	extIDs := ics.ToExternalIDs()

	return entryBlock.ExternalIDsToChainID(extIDs)
}
