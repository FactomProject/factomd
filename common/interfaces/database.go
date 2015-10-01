// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

import ()

type IDatabase interface {
	Close() error
	Put(bucket, key []byte, data BinaryMarshallable) error
	Get(bucket, key []byte, destination BinaryMarshallable) (BinaryMarshallable, error)
	Delete(bucket, key []byte) error
	ListAllKeys(bucket []byte) ([][]byte, error)
	Clear(bucket []byte) error
}
