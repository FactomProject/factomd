// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identityEntries

import (
	"fmt"

	"encoding/binary"

	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// https://github.com/FactomProject/FactomDocs/blob/master/Identity.md#coinbase-cancel
type NewCoinbaseCancelStruct struct {
	//The message is a Factom Entry with several extIDs holding the various parts.
	//[0 (version)] [Coinbase Cancel] [identity ChainID] [Descriptor height] [Desciptor index] [identity key preimage] [signature of version through timestamp]

	//The first part is a version binary string 0.
	Version byte
	//The second is the ASCII string "Coinbase Address".
	FunctionName []byte //"Server Efficiency"
	//The third is the root identity ChainID.
	RootIdentityChainID interfaces.IHash
	//Forth is coinbase descriptor height
	CoinbaseDescriptorHeight uint32
	// Fifth is the descriptor index
	CoinbaseDescriptorIndex uint32
	//6th is the identity key preimage.
	PreimageIdentityKey []byte
	//7th is the signature of the serialized version, text, chainID, new key, and the timestamp.
	Signature []byte
}

func DecodeNewCoinbaseCancelStructFromExtIDs(extIDs [][]byte) (*NewCoinbaseCancelStruct, error) {
	nbsk := new(NewCoinbaseCancelStruct)
	err := nbsk.DecodeFromExtIDs(extIDs)
	if err != nil {
		return nil, err
	}
	return nbsk, nil
}

func (ncas *NewCoinbaseCancelStruct) SetFunctionName() {
	ncas.FunctionName = []byte("Coinbase Cancel")
}

func (ncas *NewCoinbaseCancelStruct) MarshalForSig() []byte {
	answer := []byte{}

	ht := make([]byte, 4)
	binary.BigEndian.PutUint32(ht, ncas.CoinbaseDescriptorHeight)

	id := make([]byte, 4)
	binary.BigEndian.PutUint32(id, ncas.CoinbaseDescriptorIndex)

	answer = append(answer, ncas.Version)
	answer = append(answer, ncas.FunctionName...)
	answer = append(answer, ncas.RootIdentityChainID.Bytes()...)
	answer = append(answer, ht...)
	answer = append(answer, id...)
	return answer
}

func (ncas *NewCoinbaseCancelStruct) VerifySignature(key1 interfaces.IHash) error {
	bin := ncas.MarshalForSig()
	pk := new(primitives.PublicKey)
	err := pk.UnmarshalBinary(ncas.PreimageIdentityKey[1:])
	if err != nil {
		return err
	}
	var sig [64]byte
	copy(sig[:], ncas.Signature)
	ok := pk.Verify(bin, &sig)
	if ok == false {
		return fmt.Errorf("Invalid signature")
	}

	if key1 == nil {
		return nil
	}
	hashedKey := primitives.Shad(ncas.PreimageIdentityKey)
	if hashedKey.IsSameAs(key1) == false {
		return fmt.Errorf("PreimageIdentityKey does not equal Key1 - %v vs %v", hashedKey, key1)
	}

	return nil
}

func (ncas *NewCoinbaseCancelStruct) DecodeFromExtIDs(extIDs [][]byte) error {
	if len(extIDs) != 7 {
		return fmt.Errorf("Wrong number of ExtIDs - expected 7, got %v", len(extIDs))
	}
	if CheckExternalIDsLength(extIDs, []int{1, 15, 32, 4, 4, 33, 64}) == false {
		return fmt.Errorf("Wrong lengths of ExtIDs")
	}
	ncas.Version = extIDs[0][0]
	if ncas.Version != 0 {
		return fmt.Errorf("Wrong Version - expected 0, got %v", ncas.Version)
	}
	ncas.FunctionName = extIDs[1]
	if string(ncas.FunctionName) != "Coinbase Cancel" {
		return fmt.Errorf("Invalid FunctionName - expected 'Coinbase Address', got '%s'", ncas.FunctionName)
	}
	h, err := primitives.NewShaHash(extIDs[2])
	if err != nil {
		return err
	}
	ncas.RootIdentityChainID = h

	ncas.CoinbaseDescriptorHeight = binary.BigEndian.Uint32(extIDs[3])
	ncas.CoinbaseDescriptorIndex = binary.BigEndian.Uint32(extIDs[4])

	ncas.PreimageIdentityKey = extIDs[5]
	ncas.Signature = extIDs[6]

	err = ncas.VerifySignature(nil)
	if err != nil {
		return err
	}

	return nil
}

func (ncas *NewCoinbaseCancelStruct) ToExternalIDs() [][]byte {
	extIDs := [][]byte{}

	ht := make([]byte, 4)
	binary.BigEndian.PutUint32(ht, ncas.CoinbaseDescriptorHeight)

	id := make([]byte, 4)
	binary.BigEndian.PutUint32(id, ncas.CoinbaseDescriptorIndex)

	extIDs = append(extIDs, []byte{ncas.Version})
	extIDs = append(extIDs, ncas.FunctionName)
	extIDs = append(extIDs, ncas.RootIdentityChainID.Bytes())
	extIDs = append(extIDs, ht)
	extIDs = append(extIDs, id)
	extIDs = append(extIDs, ncas.PreimageIdentityKey)
	extIDs = append(extIDs, ncas.Signature)

	return extIDs
}

func (ncas *NewCoinbaseCancelStruct) GetChainID() interfaces.IHash {
	extIDs := ncas.ToExternalIDs()

	return entryBlock.ExternalIDsToChainID(extIDs)
}
