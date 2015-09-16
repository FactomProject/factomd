// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package database

import (
	"encoding/binary"
	"fmt"
	"bytes"
	"github.com/FactomProject/ed25519"
	"math/rand"
	"testing"
)

var _ = fmt.Printf
var _ = ed25519.Sign
var _ = rand.New
var _ = binary.Read



func Test_bytestore(t *testing.T) {
	b1 := new(ByteStore)
	d := []byte("ksdafljljkaglajkgljkagljkRW")
	b1.SetBytes(d)
	data,err := b1.MarshalBinary()
	if err != nil {
		t.Fail()
		return
	}
	
	b2 := new(ByteStore)
	err = b2.UnmarshalBinary(data)
	if err != nil {
		t.Fail()
		return
	}
	
	if b1.IsEqual(b2) != nil {
		t.Fail()
		return
	}
	
	if !bytes.Equal(b2.Bytes(),d) {
		fmt.Println(string(b2.Bytes()))
		t.Fail()
		return
	}
		
}