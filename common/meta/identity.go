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

//https://github.com/FactomProject/FactomDocs/blob/master/Identity.md

//TODO:
//- Add conversion to human-readible private / public key

const (
	IdentityPrivateKeyPrefix1 = "4db6c9"
	IdentityPrivateKeyPrefix2 = "4db6e7"
	IdentityPrivateKeyPrefix3 = "4db705"
	IdentityPrivateKeyPrefix4 = "4db723"

	IdentityPublicKeyPrefix1 = "3fbeba"
	IdentityPublicKeyPrefix2 = "3fbed8"
	IdentityPublicKeyPrefix3 = "3fbef6"
	IdentityPublicKeyPrefix4 = "3fbf14"
)

type IdentityChainStructure struct {
	Version      byte
	FunctionName []byte //"Identity Chain"
	Key1         interfaces.IHash
	Key2         interfaces.IHash
	Key3         interfaces.IHash
	Key4         interfaces.IHash
	Nonce        []byte
}

func (ics *IdentityChainStructure) DecodeFromExtIDs(extIDs [][]byte) error {
	if len(extIDs) != 7 {
		return fmt.Errorf("Wrong number of ExtIDs - expected 7, got %v", len(extIDs))
	}
	if CheckExternalIDsLength(extIDs[:6], []int{1, 14, 32, 32, 32, 32}) == false {
		return fmt.Errorf("Wrong lengths of ExtIDs")
	}
	ics.Version = extIDs[0][0]
	ics.FunctionName = extIDs[1]
	if string(ics.FunctionName) != "Identity Chain" {
		return fmt.Errorf("Invalid FunctionName - expected 'Identity Chain', got '%s'", ics.FunctionName)
	}
	h, err := primitives.NewShaHash(extIDs[2])
	if err != nil {
		return err
	}
	ics.Key1 = h
	h, err = primitives.NewShaHash(extIDs[3])
	if err != nil {
		return err
	}
	ics.Key2 = h
	h, err = primitives.NewShaHash(extIDs[4])
	if err != nil {
		return err
	}
	ics.Key3 = h
	h, err = primitives.NewShaHash(extIDs[5])
	if err != nil {
		return err
	}
	ics.Key4 = h
	ics.Nonce = extIDs[6]
	return nil
}

func (ics *IdentityChainStructure) GetChainID() interfaces.IHash {
	extIDs := [][]byte{}

	extIDs = append(extIDs, []byte{ics.Version})
	extIDs = append(extIDs, ics.FunctionName)
	extIDs = append(extIDs, ics.Key1.Bytes())
	extIDs = append(extIDs, ics.Key2.Bytes())
	extIDs = append(extIDs, ics.Key3.Bytes())
	extIDs = append(extIDs, ics.Key4.Bytes())
	extIDs = append(extIDs, ics.Nonce)

	return entryBlock.ExternalIDsToChainID(extIDs)
}

// Checking the external ids if they match the needed lengths
func CheckExternalIDsLength(extIDs [][]byte, lengths []int) bool {
	if len(extIDs) != len(lengths) {
		return false
	}
	for i := range extIDs {
		if lengths[i] != len(extIDs[i]) {
			return false
		}
	}
	return true
}
