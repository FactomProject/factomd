// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package simplecoin

import (
	// "fmt"
	"encoding"
)

type IBlock interface {
    
    encoding.BinaryMarshaler                        // Easy to support this, just drop the slice. 
	encoding.BinaryUnmarshaler                      // And once in Binary, it must come back.
	encoding.TextMarshaler                          // Using this mostly for debugging

	UnmarshalBinaryData(data []byte) ([]byte,error) // We need the progress through the slice.
                                                    //   so we really can't use the stock spec
                                                    //   for the UnmarshalBinary() method from
                                                    //   encode.
	newBlock(...interface{}) IBlock
	IsEqual(IBlock) bool                            // Check if this block is the same as itself   
	GetName() string                                // Debugging thing
	SetName(string)                                 // Debugging thing
	SetIndex(int)                                   // Debugging thing
    
}
