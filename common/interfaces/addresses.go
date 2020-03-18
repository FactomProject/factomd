// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

// IAddress is the same as IHash
type IAddress interface {
	IHash
}

// ITransAddress is an interface to an object interpretable as a TransAddress (an address associated with a transaction and a transaction amount)
// See factomd/common/factoid/transaddress.go
type ITransAddress interface {
	BinaryMarshallable // Must be marshallable

	GetAmount() uint64           // Get the amount associated with this transaction
	SetAmount(uint64)            // Set the amount associated with this transaction
	GetAddress() IAddress        // Return the address associated with this transaction
	SetAddress(IAddress)         // Set the address associated with this transaction
	IsSameAs(ITransAddress) bool // Compares two ITransAddress's

	CustomMarshalTextInput() ([]byte, error)    // Convert this object into text bytes
	CustomMarshalTextOutput() ([]byte, error)   // Convert this object into text bytes
	CustomMarshalTextECOutput() ([]byte, error) // Convert this object into text bytes

	StringInput() string    // Convert this object to a string
	StringOutput() string   // Convert this object to a string
	StringECOutput() string // Conver tthis object to a string

	GetUserAddress() string // Returns the human readable address for this transaction
	SetUserAddress(string)  // Sets the human readable address for this transaction
}
