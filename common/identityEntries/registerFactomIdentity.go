// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identityEntries

import (
	"bytes"
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
)

// ExpectedRegisterFactomIdentityExternalIDLengths is a hardcoded slice containing the expected lengths of each entry in an external ID (the fields of RegisterFactomIdentityStructure)
var ExpectedRegisterFactomIdentityExternalIDLengths = []int{1, 24, 32, 33, 64}

// RegisterFactomIdentityStructureSort is a slice of RegisterFactomIdentityStructures usef for sorting them
type RegisterFactomIdentityStructureSort []*RegisterFactomIdentityStructure

// Len returns the length of the slice
func (p RegisterFactomIdentityStructureSort) Len() int {
	return len(p)
}

// Swap swaps the data located at the input indices 'i' and 'j'
func (p RegisterFactomIdentityStructureSort) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// Less is a comparison function used for sorting. Checks if bytes of the identity chain id of the 'ith' index is less than
// the 'jth' index, and returns the result.
func (p RegisterFactomIdentityStructureSort) Less(i, j int) bool {
	return bytes.Compare(p[i].IdentityChainID.Bytes(), p[j].IdentityChainID.Bytes()) < 0
}

// RegisterFactomIdentityStructure holds all the information for registering a factom identity
// https://github.com/FactomProject/FactomDocs/blob/master/Identity.md#factom-identity-registration
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

// RandomRegisterFactomIdentityStructure returns a randomly created object
func RandomRegisterFactomIdentityStructure() *RegisterFactomIdentityStructure {
	r := new(RegisterFactomIdentityStructure)
	r.Version = random.RandByteSliceOfLen(1)[0]
	r.FunctionName = random.RandByteSliceOfLen(100)
	r.IdentityChainID = primitives.RandomHash()
	r.PreimageIdentityKey = random.RandByteSliceOfLen(100)
	r.Signature = random.RandByteSliceOfLen(100)

	return r
}

// UnmarshalBinary unmarshals the input data into this object
func (rfi *RegisterFactomIdentityStructure) UnmarshalBinary(p []byte) error {
	_, err := rfi.UnmarshalBinaryData(p)
	return err
}

// UnmarshalBinaryData unmarshals the input data into this object
func (rfi *RegisterFactomIdentityStructure) UnmarshalBinaryData(p []byte) (newData []byte, err error) {
	buf := primitives.NewBuffer(p)
	newData = p

	rfi.Version, err = buf.PopByte()
	if err != nil {
		return
	}

	rfi.FunctionName, err = buf.PopBytes()
	if err != nil {
		return
	}

	rfi.IdentityChainID, err = buf.PopIHash()
	if err != nil {
		return
	}

	rfi.PreimageIdentityKey, err = buf.PopBytes()
	if err != nil {
		return
	}

	rfi.Signature, err = buf.PopBytes()
	if err != nil {
		return
	}

	newData = buf.DeepCopyBytes()
	return
}

// MarshalBinary marshals this object
func (rfi *RegisterFactomIdentityStructure) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "RegisterFactomIdentityStructure.MarshalBinary err:%v", *pe)
		}
	}(&err)
	buf := primitives.NewBuffer(nil)

	err = buf.PushByte(rfi.Version)
	if err != nil {
		return nil, err
	}

	err = buf.PushBytes(rfi.FunctionName)
	if err != nil {
		return nil, err
	}

	err = buf.PushIHash(rfi.IdentityChainID)
	if err != nil {
		return nil, err
	}

	err = buf.PushBytes(rfi.PreimageIdentityKey)
	if err != nil {
		return nil, err
	}

	err = buf.PushBytes(rfi.Signature)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// IsSameAs returns true iff the input object is identical to this object
func (rfi *RegisterFactomIdentityStructure) IsSameAs(b *RegisterFactomIdentityStructure) bool {
	la := rfi.ToExternalIDs()
	lb := b.ToExternalIDs()

	if len(la) != len(lb) {
		return false
	}

	for i := range la {
		if bytes.Compare(la[i], lb[i]) != 0 {
			return false
		}
	}
	return true
}

// DecodeRegisterFactomIdentityStructureFromExtIDs returns a new object with values from the input external ID
func DecodeRegisterFactomIdentityStructureFromExtIDs(extIDs [][]byte) (*RegisterFactomIdentityStructure, error) {
	rfi := new(RegisterFactomIdentityStructure)
	err := rfi.DecodeFromExtIDs(extIDs)
	if err != nil {
		return nil, err
	}
	return rfi, nil
}

// MarshalForSig marshals the object without its signature
func (rfi *RegisterFactomIdentityStructure) MarshalForSig() []byte {
	answer := []byte{}
	answer = append(answer, rfi.Version)
	answer = append(answer, rfi.FunctionName...)
	answer = append(answer, rfi.IdentityChainID.Bytes()...)
	return answer
}

// VerifySignature marshals the object without its signature and verifies the marshaled data with the signature, and verifies the input key
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

// DecodeFromExtIDs places the information from the input external IDs into this object
func (rfi *RegisterFactomIdentityStructure) DecodeFromExtIDs(extIDs [][]byte) error {
	if len(extIDs) != 5 {
		return fmt.Errorf("Wrong number of ExtIDs - expected 5, got %v", len(extIDs))
	}
	if CheckExternalIDsLength(extIDs, ExpectedRegisterFactomIdentityExternalIDLengths) == false {
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

// ToExternalIDs returns a 2d byte slice of all the data in this object
func (rfi *RegisterFactomIdentityStructure) ToExternalIDs() [][]byte {
	extIDs := [][]byte{}

	extIDs = append(extIDs, []byte{rfi.Version})
	extIDs = append(extIDs, rfi.FunctionName)
	extIDs = append(extIDs, rfi.IdentityChainID.Bytes())
	extIDs = append(extIDs, rfi.PreimageIdentityKey)
	extIDs = append(extIDs, rfi.Signature)

	return extIDs
}

// GetChainID computes the chain ID associated with this object
func (rfi *RegisterFactomIdentityStructure) GetChainID() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "RegisterFactomIdentityStructure.GetChainID") }()

	extIDs := rfi.ToExternalIDs()

	return entryBlock.ExternalIDsToChainID(extIDs)
}
