// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package meta

import (
	"encoding/json"

	ed "github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type Authority struct {
	AuthorityChainID  interfaces.IHash
	ManagementChainID interfaces.IHash
	MatryoshkaHash    interfaces.IHash
	SigningKey        primitives.PublicKey
	Status            int
	AnchorKeys        []AnchorSigningKey
	KeyHistory        []struct {
		ActiveDBHeight uint32
		SigningKey     primitives.PublicKey
	}
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
	//return true, nil // Testing
	var pub [32]byte
	tmp, err := auth.SigningKey.MarshalBinary()
	if err != nil {
		return false, err
	} else {
		copy(pub[:], tmp)
		valid := ed.VerifyCanonical(&pub, msg, sig)
		if !valid {
			for _, histKey := range auth.KeyHistory {
				histTemp, err := histKey.SigningKey.MarshalBinary()
				if err != nil {
					continue
				}
				copy(pub[:], histTemp)
				if ed.VerifyCanonical(&pub, msg, sig) {
					return true, nil
				}
			}
		} else {
			return true, nil
		}
	}
	return false, nil
}

func (auth *Authority) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		AuthorityChainID  interfaces.IHash   `json:"chainid"`
		ManagementChainID interfaces.IHash   `json:"manageid"`
		MatryoshkaHash    interfaces.IHash   `json:"matroyshka"`
		SigningKey        string             `json:"signingkey"`
		Status            string             `json:"status"`
		AnchorKeys        []AnchorSigningKey `json:"anchorkeys"`
	}{
		AuthorityChainID:  auth.AuthorityChainID,
		ManagementChainID: auth.ManagementChainID,
		MatryoshkaHash:    auth.MatryoshkaHash,
		SigningKey:        auth.SigningKey.String(),
		Status:            statusToJSONString(auth.Status),
		AnchorKeys:        auth.AnchorKeys,
	})
}

// Only used for marshaling JSON
func statusToJSONString(status int) string {
	switch status {
	case constants.IDENTITY_UNASSIGNED:
		return "none"
	case constants.IDENTITY_FEDERATED_SERVER:
		return "federated"
	case constants.IDENTITY_AUDIT_SERVER:
		return "audit"
	case constants.IDENTITY_FULL:
		return "none"
	case constants.IDENTITY_PENDING_FEDERATED_SERVER:
		return "federated"
	case constants.IDENTITY_PENDING_AUDIT_SERVER:
		return "audit"
	case constants.IDENTITY_PENDING_FULL:
		return "none"
	case constants.IDENTITY_SKELETON:
		return "skeleton"
	}
	return "NA"
}
