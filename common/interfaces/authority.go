// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

import (
	"github.com/FactomProject/factomd/common/constants"
)

type IAuthority interface {
	Type() int
	VerifySignature([]byte, *[constants.SIGNATURE_LENGTH]byte) (bool, error)
	GetAuthorityChainID() IHash
}
