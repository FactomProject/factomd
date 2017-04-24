// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package meta

import (
	ed "github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

//https://github.com/FactomProject/FactomDocs/blob/master/Identity.md

type AnchorSigningKey struct {
	BlockChain string   `json:"blockchain"`
	KeyLevel   byte     `json:"level"`
	KeyType    byte     `json:"keytype"`
	SigningKey [20]byte `json:"key"` //if bytes, it is hex
}

type Identity struct {
	IdentityChainID      interfaces.IHash
	IdentityRegistered   uint32
	IdentityCreated      uint32
	ManagementChainID    interfaces.IHash
	ManagementRegistered uint32
	ManagementCreated    uint32
	MatryoshkaHash       interfaces.IHash
	Key1                 interfaces.IHash
	Key2                 interfaces.IHash
	Key3                 interfaces.IHash
	Key4                 interfaces.IHash
	SigningKey           interfaces.IHash
	Status               int
	AnchorKeys           []AnchorSigningKey
}

var _ interfaces.Printable = (*Identity)(nil)

func (id *Identity) VerifySignature(msg []byte, sig *[constants.SIGNATURE_LENGTH]byte) (bool, error) {
	//return true, nil // Testing
	var pub [32]byte
	tmp, err := id.SigningKey.MarshalBinary()
	if err != nil {
		return false, err
	} else {
		copy(pub[:], tmp)
		valid := ed.VerifyCanonical(&pub, msg, sig)
		if !valid {
		} else {
			return true, nil
		}
	}
	return false, nil
}

func (e *Identity) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *Identity) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *Identity) String() string {
	str, _ := e.JSONString()
	return str
}

func (id *Identity) IsFull() bool {
	zero := primitives.NewZeroHash()
	if id.IdentityChainID.IsSameAs(zero) {
		return false
	}
	if id.ManagementChainID.IsSameAs(zero) {
		return false
	}
	if id.MatryoshkaHash.IsSameAs(zero) {
		return false
	}
	if id.Key1.IsSameAs(zero) {
		return false
	}
	if id.Key2.IsSameAs(zero) {
		return false
	}
	if id.Key3.IsSameAs(zero) {
		return false
	}
	if id.Key4.IsSameAs(zero) {
		return false
	}
	if id.SigningKey.IsSameAs(zero) {
		return false
	}
	if len(id.AnchorKeys) == 0 {
		return false
	}
	return true
}
