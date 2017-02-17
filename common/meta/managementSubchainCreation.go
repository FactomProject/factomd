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

type ManagementSubchainCreationStructure struct {
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

func (msc *ManagementSubchainCreationStructure) DecodeFromExtIDs(extIDs [][]byte) error {
	if len(extIDs) != 4 {
		return fmt.Errorf("Wrong number of ExtIDs - expected 4, got %v", len(extIDs))
	}
	if CheckExternalIDsLength(extIDs[:3], []int{1, 17, 32}) == false {
		return fmt.Errorf("Wrong lengths of ExtIDs")
	}
	msc.Version = extIDs[0][0]
	if msc.Version != 0 {
		return fmt.Errorf("Wrong Version - expected 0, got %v", msc.Version)
	}
	msc.FunctionName = extIDs[1]
	if string(msc.FunctionName) != "Server Management" {
		return fmt.Errorf("Invalid FunctionName - expected 'Server Management', got '%s'", msc.FunctionName)
	}
	h, err := primitives.NewShaHash(extIDs[2])
	if err != nil {
		return err
	}
	msc.RootIdentityChainID = h
	msc.Nonce = extIDs[3]

	chainID := msc.GetChainID()
	if chainID.String()[:6] != "888888" {
		return fmt.Errorf("Invalid ChainID - it should start with '888888', but doesn't - %v", chainID.String())
	}
	return nil
}

func (msc *ManagementSubchainCreationStructure) ToExternalIDs() [][]byte {
	extIDs := [][]byte{}

	extIDs = append(extIDs, []byte{msc.Version})
	extIDs = append(extIDs, msc.FunctionName)
	extIDs = append(extIDs, msc.RootIdentityChainID.Bytes())
	extIDs = append(extIDs, msc.Nonce)

	return extIDs
}

func (msc *ManagementSubchainCreationStructure) GetChainID() interfaces.IHash {
	extIDs := msc.ToExternalIDs()

	return entryBlock.ExternalIDsToChainID(extIDs)
}
