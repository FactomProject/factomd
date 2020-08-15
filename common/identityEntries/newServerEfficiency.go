// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identityEntries

import (
	"encoding/binary"
	"fmt"

	"github.com/PaulSnow/factom2d/common/entryBlock"
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/primitives"
)

//https://github.com/FactomProject/FactomDocs/blob/master/Identity.md#add-server-efficiency
type NewServerEfficiencyStruct struct {
	//The message is a Factom Entry with several extIDs holding the various parts.
	//[0 (version)] [Server Efficiency] [identity ChainID] [new efficiency] [timestamp] [identity key preimage] [signature of version through timestamp]

	//The first part is a version binary string 0.
	Version byte
	//The second is the ASCII string "Server Efficiency".
	FunctionName []byte //"Server Efficiency"
	//The third is the root identity ChainID.
	RootIdentityChainID interfaces.IHash
	//Forth is the new efficiency
	Efficiency uint16
	//5th is the timestamp with an 8 byte epoch time.
	Timestamp []byte
	//6th is the identity key preimage.
	PreimageIdentityKey []byte
	//7th is the signature of the serialized version through timestamp.
	Signature []byte
}

func DecodeNewServerEfficiencyStructFromExtIDs(extIDs [][]byte) (*NewServerEfficiencyStruct, error) {
	nbsk := new(NewServerEfficiencyStruct)
	err := nbsk.DecodeFromExtIDs(extIDs)
	if err != nil {
		return nil, err
	}
	return nbsk, nil
}

func (nses *NewServerEfficiencyStruct) SetFunctionName() {
	nses.FunctionName = []byte("Server Efficiency")
}

func (nses *NewServerEfficiencyStruct) MarshalForSig() []byte {
	answer := []byte{}
	efficiency := make([]byte, 2)
	binary.BigEndian.PutUint16(efficiency, nses.Efficiency)

	answer = append(answer, nses.Version)
	answer = append(answer, nses.FunctionName...)
	answer = append(answer, nses.RootIdentityChainID.Bytes()...)
	answer = append(answer, efficiency...)
	answer = append(answer, nses.Timestamp...)
	return answer
}

func (nses *NewServerEfficiencyStruct) VerifySignature(key1 interfaces.IHash) error {
	bin := nses.MarshalForSig()
	pk := new(primitives.PublicKey)
	err := pk.UnmarshalBinary(nses.PreimageIdentityKey[1:])
	if err != nil {
		return err
	}
	var sig [64]byte
	copy(sig[:], nses.Signature)
	ok := pk.Verify(bin, &sig)
	if ok == false {
		return fmt.Errorf("Invalid signature")
	}

	if key1 == nil {
		return nil
	}
	hashedKey := primitives.Shad(nses.PreimageIdentityKey)
	if hashedKey.IsSameAs(key1) == false {
		return fmt.Errorf("PreimageIdentityKey does not equal Key1 - %v vs %v", hashedKey, key1)
	}

	return nil
}

func (nses *NewServerEfficiencyStruct) DecodeFromExtIDs(extIDs [][]byte) error {
	if len(extIDs) != 7 {
		return fmt.Errorf("Wrong number of ExtIDs - expected 7, got %v", len(extIDs))
	}
	if CheckExternalIDsLength(extIDs, []int{1, 17, 32, 2, 8, 33, 64}) == false {
		return fmt.Errorf("Wrong lengths of ExtIDs")
	}
	nses.Version = extIDs[0][0]
	if nses.Version != 0 {
		return fmt.Errorf("Wrong Version - expected 0, got %v", nses.Version)
	}
	nses.FunctionName = extIDs[1]
	if string(nses.FunctionName) != "Server Efficiency" {
		return fmt.Errorf("Invalid FunctionName - expected 'Server Efficiency', got '%s'", nses.FunctionName)
	}
	h, err := primitives.NewShaHash(extIDs[2])
	if err != nil {
		return err
	}
	nses.RootIdentityChainID = h
	nses.Efficiency = binary.BigEndian.Uint16(extIDs[3])

	nses.Timestamp = extIDs[4]
	nses.PreimageIdentityKey = extIDs[5]
	nses.Signature = extIDs[6]

	err = nses.VerifySignature(nil)
	if err != nil {
		return err
	}

	return nil
}

func (nses *NewServerEfficiencyStruct) ToExternalIDs() [][]byte {
	extIDs := [][]byte{}

	efficiency := make([]byte, 2)
	binary.BigEndian.PutUint16(efficiency, nses.Efficiency)

	extIDs = append(extIDs, []byte{nses.Version})
	extIDs = append(extIDs, nses.FunctionName)
	extIDs = append(extIDs, nses.RootIdentityChainID.Bytes())
	extIDs = append(extIDs, efficiency)
	extIDs = append(extIDs, nses.Timestamp)
	extIDs = append(extIDs, nses.PreimageIdentityKey)
	extIDs = append(extIDs, nses.Signature)

	return extIDs
}

func (nses *NewServerEfficiencyStruct) GetChainID() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "NewServerEfficiencyStruct.GetChainID") }()

	extIDs := nses.ToExternalIDs()

	return entryBlock.ExternalIDsToChainID(extIDs)
}
