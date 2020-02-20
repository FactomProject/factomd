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

//https://github.com/FactomProject/FactomDocs/blob/master/Identity.md#server-management-subchain-registration
type RegisterServerManagementStructure struct {
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

func DecodeRegisterServerManagementStructureFromExtIDs(extIDs [][]byte) (*RegisterServerManagementStructure, error) {
	rsm := new(RegisterServerManagementStructure)
	err := rsm.DecodeFromExtIDs(extIDs)
	if err != nil {
		return nil, err
	}
	return rsm, nil
}

func (rsm *RegisterServerManagementStructure) MarshalForSig() []byte {
	answer := []byte{}
	answer = append(answer, rsm.Version)
	answer = append(answer, rsm.FunctionName...)
	answer = append(answer, rsm.SubchainChainID.Bytes()...)
	return answer
}

func (rsm *RegisterServerManagementStructure) VerifySignature(key1 interfaces.IHash) error {
	bin := rsm.MarshalForSig()
	pk := new(primitives.PublicKey)
	err := pk.UnmarshalBinary(rsm.PreimageIdentityKey[1:])
	if err != nil {
		return err
	}
	var sig [64]byte
	copy(sig[:], rsm.Signature)
	ok := pk.Verify(bin, &sig)
	if ok == false {
		return fmt.Errorf("Invalid signature")
	}

	if key1 == nil {
		return nil
	}
	hashedKey := primitives.Shad(rsm.PreimageIdentityKey)
	if hashedKey.IsSameAs(key1) == false {
		return fmt.Errorf("PreimageIdentityKey does not equal Key1 - %v vs %v", hashedKey, key1)
	}

	return nil
}

func (rsm *RegisterServerManagementStructure) DecodeFromExtIDs(extIDs [][]byte) error {
	if len(extIDs) != 5 {
		return fmt.Errorf("Wrong number of ExtIDs - expected 5, got %v", len(extIDs))
	}
	if CheckExternalIDsLength(extIDs, []int{1, 26, 32, 33, 64}) == false {
		return fmt.Errorf("Wrong lengths of ExtIDs")
	}
	rsm.Version = extIDs[0][0]
	if rsm.Version != 0 {
		return fmt.Errorf("Wrong Version - expected 0, got %v", rsm.Version)
	}
	rsm.FunctionName = extIDs[1]
	if string(rsm.FunctionName) != "Register Server Management" {
		return fmt.Errorf("Invalid FunctionName - expected 'Register Server Management', got '%s'", rsm.FunctionName)
	}
	h, err := primitives.NewShaHash(extIDs[2])
	if err != nil {
		return err
	}
	rsm.SubchainChainID = h

	rsm.PreimageIdentityKey = extIDs[3]
	if rsm.PreimageIdentityKey[0] != 1 {
		return fmt.Errorf("Invalid PreimageIdentityKey prefix byte - 3xpected 1, got %v", rsm.PreimageIdentityKey[0])
	}
	rsm.Signature = extIDs[4]

	err = rsm.VerifySignature(nil)
	if err != nil {
		return err
	}

	return nil
}

func (rsm *RegisterServerManagementStructure) ToExternalIDs() [][]byte {
	extIDs := [][]byte{}

	extIDs = append(extIDs, []byte{rsm.Version})
	extIDs = append(extIDs, rsm.FunctionName)
	extIDs = append(extIDs, rsm.SubchainChainID.Bytes())
	extIDs = append(extIDs, rsm.PreimageIdentityKey)
	extIDs = append(extIDs, rsm.Signature)

	return extIDs
}

func (rsm *RegisterServerManagementStructure) GetChainID() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "RegisterServerManagementStructure.GetChainID") }()

	extIDs := rsm.ToExternalIDs()

	return entryBlock.ExternalIDsToChainID(extIDs)
}
