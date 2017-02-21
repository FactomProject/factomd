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

//https://github.com/FactomProject/FactomDocs/blob/master/Identity.md#add-new-matryoshka-hash
type NewMatryoshkaHashStructure struct {
	//[0 (version)] [New Matryoshka Hash] [identity ChainID] [new SHA256 M-hash] [timestamp] [identity key preimage] [signature of version through timestamp]

	//It starts with the version
	Version byte
	//and the text "New Matryoshka Hash".
	FunctionName []byte //"New Matryoshka Hash"
	//Next is the root identity chainID.
	RootIdentityChainID interfaces.IHash
	//Forth is the outermost M-hash.
	OutermostMHash interfaces.IHash
	//Fifth is a timestamp.
	Timestamp []byte
	//Sixth is the root identity key preimage.
	PreimageIdentityKey []byte
	//Last is the signature of the version through the timestamp.
	Signature []byte
}

func DecodeNewMatryoshkaHashStructureFromExtIDs(extIDs [][]byte) (*NewMatryoshkaHashStructure, error) {
	nmh := new(NewMatryoshkaHashStructure)
	err := nmh.DecodeFromExtIDs(extIDs)
	if err != nil {
		return nil, err
	}
	return nmh, nil
}

func (nmh *NewMatryoshkaHashStructure) DecodeFromExtIDs(extIDs [][]byte) error {
	if len(extIDs) != 7 {
		return fmt.Errorf("Wrong number of ExtIDs - expected 7, got %v", len(extIDs))
	}
	if CheckExternalIDsLength(extIDs, []int{1, 19, 32, 32, 8, 33, 64}) == false {
		return fmt.Errorf("Wrong lengths of ExtIDs")
	}
	nmh.Version = extIDs[0][0]
	if nmh.Version != 0 {
		return fmt.Errorf("Wrong Version - expected 0, got %v", nmh.Version)
	}
	nmh.FunctionName = extIDs[1]
	if string(nmh.FunctionName) != "New Matryoshka Hash" {
		return fmt.Errorf("Invalid FunctionName - expected 'New Matryoshka Hash', got '%s'", nmh.FunctionName)
	}
	h, err := primitives.NewShaHash(extIDs[2])
	if err != nil {
		return err
	}
	nmh.RootIdentityChainID = h
	h, err = primitives.NewShaHash(extIDs[3])
	if err != nil {
		return err
	}
	nmh.OutermostMHash = h

	nmh.Timestamp = extIDs[4]
	nmh.PreimageIdentityKey = extIDs[5]
	nmh.Signature = extIDs[6]
	//TODO: chech signature
	return nil
}

func (nmh *NewMatryoshkaHashStructure) ToExternalIDs() [][]byte {
	extIDs := [][]byte{}

	extIDs = append(extIDs, []byte{nmh.Version})
	extIDs = append(extIDs, nmh.FunctionName)
	extIDs = append(extIDs, nmh.RootIdentityChainID.Bytes())
	extIDs = append(extIDs, nmh.OutermostMHash.Bytes())
	extIDs = append(extIDs, nmh.Timestamp)
	extIDs = append(extIDs, nmh.PreimageIdentityKey)
	extIDs = append(extIDs, nmh.Signature)

	return extIDs
}

func (nmh *NewMatryoshkaHashStructure) GetChainID() interfaces.IHash {
	extIDs := nmh.ToExternalIDs()

	return entryBlock.ExternalIDsToChainID(extIDs)
}
