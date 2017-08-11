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

//https://github.com/FactomProject/FactomDocs/blob/master/Identity.md#factom-identity-registration
type RegisterFactomIdentityStructure struct {
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

func DecodeRegisterFactomIdentityStructureFromExtIDs(extIDs [][]byte) (*RegisterFactomIdentityStructure, error) {
	rfi := new(RegisterFactomIdentityStructure)
	err := rfi.DecodeFromExtIDs(extIDs)
	if err != nil {
		return nil, err
	}
	return rfi, nil
}

func (rfi *RegisterFactomIdentityStructure) MarshalForSig() []byte {
	answer := []byte{}
	answer = append(answer, rfi.Version)
	answer = append(answer, rfi.FunctionName...)
	answer = append(answer, rfi.IdentityChainID.Bytes()...)
	return answer
}

func (rfi *RegisterFactomIdentityStructure) VerifySignature(key1 interfaces.IHash) error {
	bin := rfi.MarshalForSig()
	pk := new(primitives.PublicKey)
	err := pk.UnmarshalBinary(rfi.PreimageIdentityKey[1:])
	if err != nil {
		return err
	}
	var sig [64]byte
	copy(sig[:], rfi.Signature)
	ok := pk.Verify(bin, &sig)
	if ok == false {
		return fmt.Errorf("Invalid signature")
	}

	if key1 == nil {
		return nil
	}
	hashedKey := primitives.Shad(rfi.PreimageIdentityKey)
	if hashedKey.IsSameAs(key1) == false {
		return fmt.Errorf("PreimageIdentityKey does not equal Key1 - %v vs %v", hashedKey, key1)
	}

	return nil
}

func (rfi *RegisterFactomIdentityStructure) DecodeFromExtIDs(extIDs [][]byte) error {
	if len(extIDs) != 5 {
		return fmt.Errorf("Wrong number of ExtIDs - expected 5, got %v", len(extIDs))
	}
	if CheckExternalIDsLength(extIDs, []int{1, 24, 32, 33, 64}) == false {
		return fmt.Errorf("Wrong lengths of ExtIDs")
	}
	rfi.Version = extIDs[0][0]
	if rfi.Version != 0 {
		return fmt.Errorf("Wrong Version - expected 0, got %v", rfi.Version)
	}
	rfi.FunctionName = extIDs[1]
	if string(rfi.FunctionName) != "Register Factom Identity" {
		return fmt.Errorf("Invalid FunctionName - expected 'Register Factom Identity', got '%s'", rfi.FunctionName)
	}
	h, err := primitives.NewShaHash(extIDs[2])
	if err != nil {
		return err
	}
	rfi.IdentityChainID = h

	rfi.PreimageIdentityKey = extIDs[3]
	if rfi.PreimageIdentityKey[0] != 1 {
		return fmt.Errorf("Invalid PreimageIdentityKey prefix byte - 3xpected 1, got %v", rfi.PreimageIdentityKey[0])
	}
	rfi.Signature = extIDs[4]

	err = rfi.VerifySignature(nil)
	if err != nil {
		return err
	}

	return nil
}

func (rfi *RegisterFactomIdentityStructure) ToExternalIDs() [][]byte {
	extIDs := [][]byte{}

	extIDs = append(extIDs, []byte{rfi.Version})
	extIDs = append(extIDs, rfi.FunctionName)
	extIDs = append(extIDs, rfi.IdentityChainID.Bytes())
	extIDs = append(extIDs, rfi.PreimageIdentityKey)
	extIDs = append(extIDs, rfi.Signature)

	return extIDs
}

func (rfi *RegisterFactomIdentityStructure) GetChainID() interfaces.IHash {
	extIDs := rfi.ToExternalIDs()

	return entryBlock.ExternalIDsToChainID(extIDs)
}
