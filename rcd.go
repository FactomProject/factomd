// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package simplecoin

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
	GetHash() IHash // This is what the world uses as an address
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
		PrtStk()
		return nil, nil, fmt.Errorf("Invalid type byte for authorizations: %x ", int(t))
	}
	data, err = auth.UnmarshalBinaryData(data)
	return auth, data, err
}

func NewSignature1(publicKey []byte) (IRCD, error) {
	if len(publicKey) != ADDRESS_LENGTH {
		panic("Bad publickey.  This should not happen")
	}
	a := new(RCD_1)
	a.publicKey = make([]byte, len(publicKey), len(publicKey))
	copy(a.publicKey[:], publicKey)
	return a, nil
}

func NewSignature2(n int, m int, addresses []IAddress) (IRCD, error) {
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
