// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

import ()

type Database interface {
	Close() error
	Save(bucket, key string, data BinaryMarshallable) error
	Load(bucket, key string, destination BinaryMarshallable) (BinaryMarshallable, error)
	Delete(bucket, key string) error
	ListAllKeys(bucket string) ([]string, error)
}
