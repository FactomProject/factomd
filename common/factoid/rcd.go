// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
)

/**************************
 * interfaces.IRCD  Interface for Redeem Condition Datastructures (RCD)
 *
 * https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#factoid-transaction
 **************************/

/***********************
 * Helper Functions
 ***********************/

// UnmarshalBinaryAuth takes the input byte slice, determines whether its RCD 1 or 2, and
// unmarshals it into a new RCD of that type
func UnmarshalBinaryAuth(data []byte) (a interfaces.IRCD, newData []byte, err error) {
	if data == nil || len(data) < 1 {
		return nil, nil, fmt.Errorf("Not enough data to unmarshal")
	}
	t := data[0]

	var auth interfaces.IRCD
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

// NewRCD_1 creates a new RCD of type 1
func NewRCD_1(publicKey []byte) interfaces.IRCD {
	if len(publicKey) != constants.ADDRESS_LENGTH {
		panic("Bad publickey.  This should not happen")
	}
	a := new(RCD_1)
	copy(a.PublicKey[:], publicKey)
	return a
}

// NewRCD_2 creates a new RCD of type 2
func NewRCD_2(n int, m int, addresses []interfaces.IAddress) (interfaces.IRCD, error) {
	if len(addresses) != m {
		return nil, fmt.Errorf("Improper number of addresses.  m = %d n = %d #addresses = %d", m, n, len(addresses))
	}

	au := new(RCD_2)
	au.N = n
	au.M = m
	au.N_Addresses = make([]interfaces.IAddress, len(addresses), len(addresses))
	copy(au.N_Addresses, addresses)

	return au, nil
}

// CreateRCD creates a new RCD based on the input type
func CreateRCD(data []byte) interfaces.IRCD {
	switch data[0] {
	case 1:
		return new(RCD_1)
	case 2:
		return new(RCD_2)
	default:
		panic("Bad Data encountered by CreateRCD.  Should never happen")
	}
}
