// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

type IAnchor interface {
	InitRPCClient() error
	UpdateDirBlockInfoMap(dirBlockInfo IDirBlockInfo)
}

type IAnchorRecord interface {
	Marshal() ([]byte, error)
	MarshalAndSign(priv Signer) ([]byte, error)
	Unmarshal(data []byte) error
}
