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

// ExpectedBlockSigningKeyExternalIDLengths is a hardcoded slice containing the expected lengths of each entry in an external ID (the fields of BlockSigningKey)
var ExpectedBlockSigningKeyExternalIDLengths = []int{1, 21, 32, 32, 8, 33, 64}

// NewBlockSigningKeyStruct holds all the information for adding a new block signing key to a server subchain
// https://github.com/FactomProject/FactomDocs/blob/master/Identity.md#add-new-block-signing-key
type NewBlockSigningKeyStruct struct {
	//The message is a Factom Entry with several extIDs holding the various parts.
	//[0 (version)] [New Block Signing Key] [identity ChainID] [new key] [timestamp] [identity key preimage] [signature of version through timestamp]

	//The first part is a version binary string 0.
	Version byte
	//The second is the ASCII string "New Block Signing Key".
	FunctionName []byte //"New Block Signing Key"
	//The third is the root identity ChainID.
	RootIdentityChainID interfaces.IHash
	//Forth is the new public key being asserted.
	NewPublicKey []byte
	//5th is the timestamp with an 8 byte epoch time.
	Timestamp []byte
	//Next is the identity key preimage.
	PreimageIdentityKey []byte
	//7th is the signature of the serialized version, text, chainID, new key, and the timestamp.
	Signature []byte
}

// DecodeNewBlockSigningKeyStructFromExtIDs returns a new object with values from the input external ID
func DecodeNewBlockSigningKeyStructFromExtIDs(extIDs [][]byte) (*NewBlockSigningKeyStruct, error) {
	nbsk := new(NewBlockSigningKeyStruct)
	err := nbsk.DecodeFromExtIDs(extIDs)
	if err != nil {
		return nil, err
	}
	return nbsk, nil
}

// MarshalForSig marshals the object without its signature
func (nbsk *NewBlockSigningKeyStruct) MarshalForSig() []byte {
	answer := []byte{}
	answer = append(answer, nbsk.Version)
	answer = append(answer, nbsk.FunctionName...)
	answer = append(answer, nbsk.RootIdentityChainID.Bytes()...)
	answer = append(answer, nbsk.NewPublicKey...)
	answer = append(answer, nbsk.Timestamp...)
	return answer
}

// VerifySignature marshals the object without its signature and verifies the marshaled data with the signature, and verifies the input key
func (nbsk *NewBlockSigningKeyStruct) VerifySignature(key1 interfaces.IHash) error {
	bin := nbsk.MarshalForSig()
	pk := new(primitives.PublicKey)
	err := pk.UnmarshalBinary(nbsk.PreimageIdentityKey[1:])
	if err != nil {
		return err
	}
	var sig [64]byte
	copy(sig[:], nbsk.Signature)
	ok := pk.Verify(bin, &sig)
	if ok == false {
		return fmt.Errorf("Invalid signature")
	}

	if key1 == nil {
		return nil
	}
	hashedKey := primitives.Shad(nbsk.PreimageIdentityKey)
	if hashedKey.IsSameAs(key1) == false {
		return fmt.Errorf("PreimageIdentityKey does not equal Key1 - %v vs %v", hashedKey, key1)
	}

	return nil
}

// DecodeFromExtIDs places the information from the input external IDs into this object
func (nbsk *NewBlockSigningKeyStruct) DecodeFromExtIDs(extIDs [][]byte) error {
	if len(extIDs) != 7 {
		return fmt.Errorf("Wrong number of ExtIDs - expected 7, got %v", len(extIDs))
	}
	if CheckExternalIDsLength(extIDs, ExpectedBlockSigningKeyExternalIDLengths) == false {
		return fmt.Errorf("Wrong lengths of ExtIDs")
	}
	nbsk.Version = extIDs[0][0]
	if nbsk.Version != 0 {
		return fmt.Errorf("Wrong Version - expected 0, got %v", nbsk.Version)
	}
	nbsk.FunctionName = extIDs[1]
	if string(nbsk.FunctionName) != "New Block Signing Key" {
		return fmt.Errorf("Invalid FunctionName - expected 'New Block Signing Key', got '%s'", nbsk.FunctionName)
	}
	h, err := primitives.NewShaHash(extIDs[2])
	if err != nil {
		return err
	}
	nbsk.RootIdentityChainID = h
	nbsk.NewPublicKey = extIDs[3]
	nbsk.Timestamp = extIDs[4]
	nbsk.PreimageIdentityKey = extIDs[5]
	nbsk.Signature = extIDs[6]

	err = nbsk.VerifySignature(nil)
	if err != nil {
		return err
	}

	return nil
}

// ToExternalIDs returns a 2d byte slice of all the data in this object
func (nbsk *NewBlockSigningKeyStruct) ToExternalIDs() [][]byte {
	extIDs := [][]byte{}

	extIDs = append(extIDs, []byte{nbsk.Version})
	extIDs = append(extIDs, nbsk.FunctionName)
	extIDs = append(extIDs, nbsk.RootIdentityChainID.Bytes())
	extIDs = append(extIDs, nbsk.NewPublicKey)
	extIDs = append(extIDs, nbsk.Timestamp)
	extIDs = append(extIDs, nbsk.PreimageIdentityKey)
	extIDs = append(extIDs, nbsk.Signature)

	return extIDs
}

// GetChainID computes the chain ID associated with this object
func (nbsk *NewBlockSigningKeyStruct) GetChainID() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "NewBlockSigningKeyStruct.GetChainID") }()

	extIDs := nbsk.ToExternalIDs()

	return entryBlock.ExternalIDsToChainID(extIDs)
}
