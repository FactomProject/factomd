// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package meta

import (
	ed "github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/primitives"
)

type Authority struct {
	AuthorityChainID  *primitives.Hash
	ManagementChainID *primitives.Hash
	MatryoshkaHash    *primitives.Hash
	SigningKey        primitives.PublicKey
	Status            int
	AnchorKeys        []AnchorSigningKey
}

// 1 if fed, 0 if audit, -1 if neither
func (auth *Authority) Type() int {
	if auth.Status == constants.IDENTITY_FEDERATED_SERVER {
		return 1
	} else if auth.Status == constants.IDENTITY_AUDIT_SERVER {
		return 0
	}
	return -1
}

func (auth *Authority) VerifySignature(msg []byte, sig *[constants.SIGNATURE_LENGTH]byte) (bool, error) {
	var pub [32]byte
	tmp, err := auth.SigningKey.MarshalBinary()
	if err != nil {
		return false, err
	} else {
		copy(pub[:], tmp)
		valid := ed.VerifyCanonical(&pub, msg, sig)
		return valid, nil
	}
	return false, nil
}
