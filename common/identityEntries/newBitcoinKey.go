// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identityEntries

import (
	"fmt"

	"github.com/PaulSnow/factom2d/common/entryBlock"
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/primitives"
)

//https://github.com/FactomProject/FactomDocs/blob/master/Identity.md#add-new-bitcoin-key
type NewBitcoinKeyStructure struct {
	//The message is an entry with multiple extIDs.
	//[0 (version)] [New Bitcoin Key] [identity ChainID] [Bitcoin key level] [type byte] [new key] [timestamp] [identity key preimage] [signature of version through timestamp]

	//It has a version
	Version byte
	//a text string saying "New Bitcoin Key"
	FunctionName []byte //"New Bitcoin Key"
	//It specifies the root identity chainID.
	RootIdentityChainID interfaces.IHash
	//Next is a byte signifying the bitcoin key level.
	//It is 0 origin indexed, so with only 1 key it would be 0x00.
	BitcoinKeyLevel byte
	//The next extID specifies what type of Bitcoin key is used, P2PKH or P2SH.
	KeyType byte
	//Next is the 20 byte Bitcoin key.
	NewKey [20]byte
	//Seventh is a timestamp, which prevents replay attacks.
	Timestamp []byte
	//Eighth is the identity key preimage.
	PreimageIdentityKey []byte
	//Last is the signature of the serialized version through the timestamp.
	Signature []byte
}

func DecodeNewBitcoinKeyStructureFromExtIDs(extIDs [][]byte) (*NewBitcoinKeyStructure, error) {
	nbks := new(NewBitcoinKeyStructure)
	err := nbks.DecodeFromExtIDs(extIDs)
	if err != nil {
		return nil, err
	}
	return nbks, nil
}

func (nbks *NewBitcoinKeyStructure) MarshalForSig() []byte {
	answer := []byte{}
	answer = append(answer, nbks.Version)
	answer = append(answer, nbks.FunctionName...)
	answer = append(answer, nbks.RootIdentityChainID.Bytes()...)
	answer = append(answer, nbks.BitcoinKeyLevel)
	answer = append(answer, nbks.KeyType)
	answer = append(answer, nbks.NewKey[:]...)
	answer = append(answer, nbks.Timestamp...)
	return answer
}

func (nbks *NewBitcoinKeyStructure) VerifySignature(key1 interfaces.IHash) error {
	bin := nbks.MarshalForSig()
	pk := new(primitives.PublicKey)
	err := pk.UnmarshalBinary(nbks.PreimageIdentityKey[1:])
	if err != nil {
		return err
	}
	var sig [64]byte
	copy(sig[:], nbks.Signature)
	ok := pk.Verify(bin, &sig)
	if ok == false {
		return fmt.Errorf("Invalid signature")
	}

	if key1 == nil {
		return nil
	}
	hashedKey := primitives.Shad(nbks.PreimageIdentityKey)
	if hashedKey.IsSameAs(key1) == false {
		return fmt.Errorf("PreimageIdentityKey does not equal Key1 - %v vs %v", hashedKey, key1)
	}

	return nil
}

func (nbks *NewBitcoinKeyStructure) DecodeFromExtIDs(extIDs [][]byte) error {
	if len(extIDs) != 9 {
		return fmt.Errorf("Wrong number of ExtIDs - expected 9, got %v", len(extIDs))
	}
	if CheckExternalIDsLength(extIDs, []int{1, 15, 32, 1, 1, 20, 8, 33, 64}) == false {
		return fmt.Errorf("Wrong lengths of ExtIDs")
	}
	nbks.Version = extIDs[0][0]
	if nbks.Version != 0 {
		return fmt.Errorf("Wrong Version - expected 0, got %v", nbks.Version)
	}
	nbks.FunctionName = extIDs[1]
	if string(nbks.FunctionName) != "New Bitcoin Key" {
		return fmt.Errorf("Invalid FunctionName - expected 'New Bitcoin Key', got '%s'", nbks.FunctionName)
	}
	h, err := primitives.NewShaHash(extIDs[2])
	if err != nil {
		return err
	}
	nbks.RootIdentityChainID = h

	nbks.BitcoinKeyLevel = extIDs[3][0]
	nbks.KeyType = extIDs[4][0]
	copy(nbks.NewKey[:], extIDs[5])
	nbks.Timestamp = extIDs[6]
	nbks.PreimageIdentityKey = extIDs[7]
	nbks.Signature = extIDs[8]

	err = nbks.VerifySignature(nil)
	if err != nil {
		return err
	}

	return nil
}

func (nbks *NewBitcoinKeyStructure) ToExternalIDs() [][]byte {
	extIDs := [][]byte{}

	extIDs = append(extIDs, []byte{nbks.Version})
	extIDs = append(extIDs, nbks.FunctionName)
	extIDs = append(extIDs, nbks.RootIdentityChainID.Bytes())
	extIDs = append(extIDs, []byte{nbks.BitcoinKeyLevel})
	extIDs = append(extIDs, []byte{nbks.KeyType})
	extIDs = append(extIDs, nbks.NewKey[:])
	extIDs = append(extIDs, nbks.Timestamp)
	extIDs = append(extIDs, nbks.PreimageIdentityKey)
	extIDs = append(extIDs, nbks.Signature)

	return extIDs
}

func (nbks *NewBitcoinKeyStructure) GetChainID() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "NewBitcoinKeyStructure.GetChainID") }()

	extIDs := nbks.ToExternalIDs()

	return entryBlock.ExternalIDsToChainID(extIDs)
}
