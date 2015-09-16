// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid

import (
	"fmt"
)

/**************************
 * IRCD  Interface for Redeem Condition Datastructures (RCD)
 *
 * https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#factoid-transaction
 **************************/

type IRCD interface {
	IBlock
	GetAddress() (IAddress, error)
	Clone() IRCD
	NumberOfSignatures() int
	CheckSig(trans ITransaction, sigblk ISignatureBlock) bool
}

/***********************
 * Helper Functions
 ***********************/

func UnmarshalBinaryAuth(data []byte) (a IRCD, newData []byte, err error) {

	t := data[0]

	var auth IRCD
	switch int(t) {
	case 1:
		auth = new(RCD_1)
	case 2:
		auth = new(RCD_2)
	default:
		return nil, nil, fmt.Errorf("Invalid type byte for authorizations: %x ", int(t))
	}
	data, err = auth.UnmarshalBinaryData(data)
	return auth, data, err
}

func NewRCD_1(publicKey []byte) IRCD {
	if len(publicKey) != ADDRESS_LENGTH {
		panic("Bad publickey.  This should not happen")
	}
	a := new(RCD_1)
	copy(a.publicKey[:], publicKey)
	return a
}

func NewRCD_2(n int, m int, addresses []IAddress) (IRCD, error) {
	if len(addresses) != m {
		return nil, fmt.Errorf("Improper number of addresses.  m = %d n = %d #addresses = %d", m, n, len(addresses))
	}

	au := new(RCD_2)
	au.n = n
	au.m = m
	au.n_addresses = make([]IAddress, len(addresses), len(addresses))
	copy(au.n_addresses, addresses)

	return au, nil
}

func CreateRCD(data []byte) IRCD {
	switch data[0] {
	case 1:
		return new(RCD_1)
	case 2:
		return new(RCD_2)
	default:
		panic("Bad Data encountered by CreateRCD.  Should never happen")
	}
}
