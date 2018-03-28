// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

type IConfirmation interface {
	IMsg

	DBHeight() int      // Directory Block Height
	ChainID() []byte    // ChainID of the sending server
	ListHeight() int    // Height in the Process List
	SerialHash() []byte // Serial Hash so far
	//Signature() []byte  // Signature
}
