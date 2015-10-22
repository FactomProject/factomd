// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

import ()

type IDirectoryBlock interface {
	BinaryMarshallable
	
	Header() 							IDirectoryBlockHeader
	SetHeader(IDirectoryBlockHeader)
	DBEntries() 						[]IDBEntry
	SetDBEntries([] IDBEntry)
	BuildKeyMerkleRoot()				(IHash, error)
	KeyMR()								IHash
}

type IDirectoryBlockHeader interface {
	BinaryMarshallable
	
	Version()							byte
	SetVersion(byte)	
	PrevLedgerKeyMR() 					IHash
	SetPrevLedgerKeyMR(IHash)
	BodyMR()							IHash
	SetBodyMR(IHash)
	PrevKeyMR()							IHash
	SetPrevKeyMR(IHash)
	DBHeight()							uint32
	SetDBHeight(uint32)
	BlockCount()						uint32
	SetBlockCount(uint32)
	
}

type IDBEntry interface {
	BinaryMarshallable
}
