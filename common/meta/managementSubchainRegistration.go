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

type ManagementSubchainRegistrationStructure struct {
	//It is very similar to the Factom identity registration message.
	//[0 (version)] [Register Server Management] [subchain ChainID] [identity key preimage] [signature of version through ChainID]

	//The first ExtID is a binary string 0 signifying the version.
	Version byte
	//The second ExtID has 26 ASCII bytes "Register Server Management".
	FunctionName []byte //"Register Server Management"
	//The third ExtID is the binary encoded ChainID of the identity. It will start with 888888.
	SubchainChainID interfaces.IHash //888888...
	//The 4th ExtID is the preimage to the identity key. It includes the type prefix (0x01) and the raw ed25519 pubkey.
	PreimageIdentityKey []byte
	//The 5th ExtID is the signature of the first, second, and third ExtIDs serialized together.
	Signature []byte
}

func (msr *ManagementSubchainRegistrationStructure) DecodeFromExtIDs(extIDs [][]byte) error {
	if len(extIDs) != 5 {
		return fmt.Errorf("Wrong number of ExtIDs - expected 5, got %v", len(extIDs))
	}
	if CheckExternalIDsLength(extIDs, []int{1, 26, 32, 33, 64}) == false {
		return fmt.Errorf("Wrong lengths of ExtIDs")
	}
	msr.Version = extIDs[0][0]
	if msr.Version != 0 {
		return fmt.Errorf("Wrong Version - expected 0, got %v", msr.Version)
	}
	msr.FunctionName = extIDs[1]
	if string(msr.FunctionName) != "Register Server Management" {
		return fmt.Errorf("Invalid FunctionName - expected 'Register Server Management', got '%s'", msr.FunctionName)
	}
	h, err := primitives.NewShaHash(extIDs[2])
	if err != nil {
		return err
	}
	msr.SubchainChainID = h

	msr.PreimageIdentityKey = extIDs[3]
	if msr.PreimageIdentityKey[0] != 1 {
		return fmt.Errorf("Invalid PreimageIdentityKey prefix byte - 3xpected 1, got %v", msr.PreimageIdentityKey[0])
	}
	msr.Signature = extIDs[4]
	//TODO: test signature
	return nil
}

func (msr *ManagementSubchainRegistrationStructure) ToExternalIDs() [][]byte {
	extIDs := [][]byte{}

	extIDs = append(extIDs, []byte{msr.Version})
	extIDs = append(extIDs, msr.FunctionName)
	extIDs = append(extIDs, msr.SubchainChainID.Bytes())
	extIDs = append(extIDs, msr.PreimageIdentityKey)
	extIDs = append(extIDs, msr.Signature)

	return extIDs
}

func (msr *ManagementSubchainRegistrationStructure) GetChainID() interfaces.IHash {
	extIDs := msr.ToExternalIDs()

	return entryBlock.ExternalIDsToChainID(extIDs)
}
