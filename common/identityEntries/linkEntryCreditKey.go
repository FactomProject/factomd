// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identityEntries

//************************ NOT USED YET! *************************************//

/*
import (
	"fmt"

	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

//https://github.com/FactomProject/FactomDocs/blob/master/Identity.md#link-entry-credit-key-to-identity
type LinkEntryCreditKeyStructure struct {
	//The link message would consist of 5 extIDs.
	//[0 (version)] [Link Entry Credit Key] [identity ChainID] [Entry Credit public key] [signature of version through ChainID]

	//The first is the version, with a single byte of 0.
	Version byte
	//The second is 21 bytes of ASCII text "Link Entry Credit Key".
	FunctionName []byte //"Link Entry Credit Key"
	//The third is the identity ChainID.
	RootIdentityChainID interfaces.IHash

	//Some random byte goes here?! TODO: check!

	//The 4th is the Entry Credit public key doing the signing, which should be linked to the specified identity.
	EntryCreditPublicKey []byte
	//The 5th is the signature of the first, second, and third ExtIDs serialized together.
	Signature []byte
}

func (leck *LinkEntryCreditKeyStructure) DecodeFromExtIDs(extIDs [][]byte) error {
	if len(extIDs) != 5 {
		return fmt.Errorf("Wrong number of ExtIDs - expected 5, got %v", len(extIDs))
	}
	if CheckExternalIDsLength(extIDs, []int{1, 24, 32, 33, 64}) == false {
		return fmt.Errorf("Wrong lengths of ExtIDs")
	}
	leck.Version = extIDs[0][0]
	if leck.Version != 0 {
		return fmt.Errorf("Wrong Version - expected 0, got %v", leck.Version)
	}
	leck.FunctionName = extIDs[1]
	if string(leck.FunctionName) != "Link Entry Credit Key" {
		return fmt.Errorf("Invalid FunctionName - expected 'Link Entry Credit Key', got '%s'", leck.FunctionName)
	}
	h, err := primitives.NewShaHash(extIDs[2])
	if err != nil {
		return err
	}
	leck.RootIdentityChainID = h

	leck.EntryCreditPublicKey = extIDs[3]
	leck.Signature = extIDs[4]
	//TODO: chech signature
	return nil
}

func (leck *LinkEntryCreditKeyStructure) ToExternalIDs() [][]byte {
	extIDs := [][]byte{}

	extIDs = append(extIDs, []byte{leck.Version})
	extIDs = append(extIDs, leck.FunctionName)
	extIDs = append(extIDs, leck.RootIdentityChainID.Bytes())
	extIDs = append(extIDs, leck.EntryCreditPublicKey)
	extIDs = append(extIDs, leck.Signature)

	return extIDs
}

func (leck *LinkEntryCreditKeyStructure) GetChainID()(rval interfaces.IHash) {
defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
		rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("LinkEntryCreditKeyStructure.GetChainID() returned a nil for IHash")
		}
	}()

	extIDs := leck.ToExternalIDs()

	return entryBlock.ExternalIDsToChainID(extIDs)
}
*/
