// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid

import (
    "fmt"
    "bytes"
   
    "math/rand"
    "testing"
)

// func DecodeVarInt(data []byte)                   (uint64, []byte) 
// func EncodeVarInt(out *bytes.Buffer, v uint64)   error 

func Test_Variable_Integers (test *testing.T) {
     
    for i:=0; i<1000; i++ {
        var out bytes.Buffer
        
        v := make([]uint64,10)
        
        for j:=0; j<len(v); j++ {
            sw := rand.Int63()%5                                   // Pick a random choice
            switch sw {
                case 0: v[j] = uint64(rand.Int63() & 0xFF)         // Random byte  
                case 1: v[j] = uint64(rand.Int63() & 0xFFFF)       // Random 16 bit integer
                case 2: v[j] = uint64(rand.Int63() & 0xFFFFFFFF)   // Random 32 bit integer
                case 3: v[j] = uint64(rand.Int63())                // Random 63 bit int, high order zero
                case 4: v[j] = uint64(rand.Int63()<<1)             // Random 63 bit int, low order zero
            }
        }
                    
        for j:=0; j<len(v); j++ {               // Encode our entire array of numbers
            err := EncodeVarInt(&out,v[j])
            if err != nil {
                fmt.Println(err)
                test.Fail()
                return
            }
//              fmt.Printf("%x ",v[j])
        }
//          fmt.Println( "Length: ",out.Len())
        
        data := out.Bytes()
        
//          PrtData(data) 
//          fmt.Println()
        sdata := data                           // Decode our entire array of numbers, and 
        var dv uint64                           // check we got them back correctly.
        for k:=0; k<1000; k++ {
            data = sdata
            for j:=0; j<len(v); j++ {
                dv, data = DecodeVarInt(data) 
                if ( dv != v[j] ) {
                    fmt.Printf("Values don't match: decode:%x expected:%x (%d)\n",dv,v[j], j)
                    test.Fail()
                    return
                }
            }
        }
    }
}
    