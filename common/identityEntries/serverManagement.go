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

// ExpectedServerManagementExternalIDLengths is a hardcoded slice containing the expected lengths of each entry in an external ID (the fields of ServerManagementStructure)
var ExpectedServerManagementExternalIDLengths = []int{1, 17, 32}

// ServerManagementStructure holds all the information for creating a new server management subchain
// https://github.com/FactomProject/FactomDocs/blob/master/Identity.md#server-management-subchain-creation
type ServerManagementStructure struct {
	//This chain is created after the identity chain is known.
	//The Chain Name first element is a version, 0.
	Version byte
	//The second is the ASCII string "Server Management".
	FunctionName []byte //"Server Management"
	//The 3rd consists of the root identity chainID.
	RootIdentityChainID interfaces.IHash
	////The 4th is a nonce which makes the first 6 bytes of the chainID match 0x888888.
	Nonce []byte
}

// DecodeServerManagementStructureFromExtIDs returns a new object with values from the input external ID
func DecodeServerManagementStructureFromExtIDs(extIDs [][]byte) (*ServerManagementStructure, error) {
	sm := new(ServerManagementStructure)
	err := sm.DecodeFromExtIDs(extIDs)
	if err != nil {
		return nil, err
	}
	return sm, nil
}

// DecodeFromExtIDs places the information from the input external IDs into this object
func (sm *ServerManagementStructure) DecodeFromExtIDs(extIDs [][]byte) error {
	if len(extIDs) != 4 {
		return fmt.Errorf("Wrong number of ExtIDs - expected 4, got %v", len(extIDs))
	}
	if CheckExternalIDsLength(extIDs[:3], ExpectedServerManagementExternalIDLengths) == false {
		return fmt.Errorf("Wrong lengths of ExtIDs")
	}
	sm.Version = extIDs[0][0]
	if sm.Version != 0 {
		return fmt.Errorf("Wrong Version - expected 0, got %v", sm.Version)
	}
	sm.FunctionName = extIDs[1]
	if string(sm.FunctionName) != "Server Management" {
		return fmt.Errorf("Invalid FunctionName - expected 'Server Management', got '%s'", sm.FunctionName)
	}
	h, err := primitives.NewShaHash(extIDs[2])
	if err != nil {
		return err
	}
	sm.RootIdentityChainID = h
	sm.Nonce = extIDs[3]

	chainID := sm.GetChainID()
	if chainID.String()[:6] != "888888" {
		return fmt.Errorf("Invalid ChainID - it should start with '888888', but doesn't - %v", chainID.String())
	}
	return nil
}

// ToExternalIDs returns a 2d byte slice of all the data in this object
func (sm *ServerManagementStructure) ToExternalIDs() [][]byte {
	extIDs := [][]byte{}

	extIDs = append(extIDs, []byte{sm.Version})
	extIDs = append(extIDs, sm.FunctionName)
	extIDs = append(extIDs, sm.RootIdentityChainID.Bytes())
	extIDs = append(extIDs, sm.Nonce)

	return extIDs
}

// GetChainID computes the chain ID associated with this object
func (sm *ServerManagementStructure) GetChainID() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "ServerManagementStructure.GetChainID") }()

	extIDs := sm.ToExternalIDs()

	return entryBlock.ExternalIDsToChainID(extIDs)
}
