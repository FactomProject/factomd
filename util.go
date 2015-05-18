// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package simplecoin


import (
    "fmt"
    "bytes"
    "encoding/hex"
    "encoding/binary"
)

func WriteNumber16(out *bytes.Buffer, num uint16) {
    var buf bytes.Buffer
    
    binary.Write(&buf, binary.BigEndian, num)
    str := hex.EncodeToString(buf.Bytes())
    out.WriteString(str)
    
} 

func WriteNumber8(out *bytes.Buffer, num uint8) {
    var buf bytes.Buffer
    
    binary.Write(&buf, binary.BigEndian, num)
    str := hex.EncodeToString(buf.Bytes())
    out.WriteString(str)
    
} 

func Prt (a ...interface{}) {
    fmt.Print(a...)
}

func Prtln (a ...interface{}) {
    fmt.Println(a...)
}

