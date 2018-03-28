// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

type IAddress interface {
	IHash
}

type ITransAddress interface {
	BinaryMarshallable

	GetAmount() uint64
	SetAmount(uint64)
	GetAddress() IAddress
	SetAddress(IAddress)
	IsSameAs(ITransAddress) bool

	CustomMarshalTextInput() ([]byte, error)
	CustomMarshalTextOutput() ([]byte, error)
	CustomMarshalTextECOutput() ([]byte, error)

	StringInput() string
	StringOutput() string
	StringECOutput() string

	GetUserAddress() string
	SetUserAddress(string)
}
