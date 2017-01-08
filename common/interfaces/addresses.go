// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

import ()

type IAddress interface {
	IHash
}

type ITransAddress interface {
	IBlock
	GetAmount() uint64
	SetAmount(uint64)
	GetAddress() IAddress
	SetAddress(IAddress)
	CustomMarshalText2(string) ([]byte, error)

	GetUserAddress() string
	SetUserAddress(string)
}

type IInAddress interface {
	ITransAddress
}

type IOutAddress interface {
	ITransAddress
}

type IOutECAddress interface {
	ITransAddress
}
