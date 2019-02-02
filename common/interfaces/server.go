// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

// This object will hold the public keys for servers that are not
// us, and maybe other information about servers.
type IServer interface {
	BinaryMarshallable
	Printable

	GetChainID() IHash
	GetName() string
	IsOnline() bool
	SetOnline(bool)
	LeaderToReplace() IHash
	SetReplace(IHash)
	IsSameAs(b IServer) bool
}
