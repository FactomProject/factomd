// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package meta

import (
	"fmt"

	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type IdentityRegistrationStructure struct {
	//The registration message has 5 ExtIDs.
	//The first ExtID is a binary string 0 signifying the version.
	Version byte
	//The second ExtID has 24 ASCII bytes "Register Factom Identity".
	FunctionName []byte //"Register Factom Identity"
	//The third ExtID is the binary encoded ChainID of the identity. It will start with 888888.
	IdentityChainID interfaces.IHash //888888...
	//The 4th ExtID is the preimage to the identity key. It includes the type prefix (0x01) and the raw ed25519 pubkey.
	PreimageIdentityKey []byte
	//The 5th ExtID is the signature of the first, second, and third ExtIDs serialized together.
	Signature []byte
}

func (irs *IdentityRegistrationStructure) DecodeFromExtIDs(extIDs [][]byte) error {
	if len(extIDs) != 5 {
		return fmt.Errorf("Wrong number of ExtIDs - expected 5, got %v", len(extIDs))
	}
	if CheckExternalIDsLength(extIDs, []int{1, 24, 32, 33, 64}) == false {
		return fmt.Errorf("Wrong lengths of ExtIDs")
	}
	irs.Version = extIDs[0][0]
	if irs.Version != 0 {
		return fmt.Errorf("Wrong Version - expected 0, got %v", irs.Version)
	}
	irs.FunctionName = extIDs[1]
	if string(irs.FunctionName) != "Register Factom Identity" {
		return fmt.Errorf("Invalid FunctionName - expected 'Register Factom Identity', got '%s'", irs.FunctionName)
	}
	h, err := primitives.NewShaHash(extIDs[2])
	if err != nil {
		return err
	}
	irs.IdentityChainID = h

	irs.PreimageIdentityKey = extIDs[3]
	if irs.PreimageIdentityKey[0] != 1 {
		return fmt.Errorf("Invalid PreimageIdentityKey prefix byte - 3xpected 1, got %v", irs.PreimageIdentityKey[0])
	}
	irs.Signature = extIDs[4]
	//TODO: test signature
	return nil
}

func (irs *IdentityRegistrationStructure) ToExternalIDs() [][]byte {
	extIDs := [][]byte{}

	extIDs = append(extIDs, []byte{irs.Version})
	extIDs = append(extIDs, irs.FunctionName)
	extIDs = append(extIDs, irs.IdentityChainID.Bytes())
	extIDs = append(extIDs, irs.PreimageIdentityKey)
	extIDs = append(extIDs, irs.Signature)

	return extIDs
}

func (irs *IdentityRegistrationStructure) GetChainID() interfaces.IHash {
	extIDs := irs.ToExternalIDs()

	return entryBlock.ExternalIDsToChainID(extIDs)
}
