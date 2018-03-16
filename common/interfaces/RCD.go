// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

/**************************
 * IRCD  Interface for Redeem Condition Datastructures (RCD)
 *
 * https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#factoid-transaction
 **************************/

type IRCD interface {
	BinaryMarshallable
	Printable

	CheckSig(trans ITransaction, sigblk ISignatureBlock) bool
	Clone() IRCD
	CustomMarshalText() ([]byte, error)
	GetAddress() (IAddress, error)
	NumberOfSignatures() int
	IsSameAs(IRCD) bool
}

type IRCD_1 interface {
	IRCD
	GetPublicKey() []byte
}
