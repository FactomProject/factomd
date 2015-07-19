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


func TestConversions (test *testing.T) {
    v, err := ConvertFixedPoint(".999"); if err != nil || v != "99900000" {fmt.Println("1",v,err);test.Fail()}
    v, err = ConvertFixedPoint("0.999"); if err != nil || v != "99900000" {fmt.Println("2",v,err);test.Fail()}
    v, err = ConvertFixedPoint("10.999"); if err != nil || v != "1099900000" {fmt.Println("3",v,err);test.Fail()}
    v, err = ConvertFixedPoint(".99999999999999"); if err != nil || v != "99999999" {fmt.Println("4",v,err);test.Fail()}
}


// func DecodeVarInt(data []byte)                   (uint64, []byte) 
// func EncodeVarInt(out *bytes.Buffer, v uint64)   error 

func TestVariable_Integers (test *testing.T) {
     
    for i:=0; i<1000; i++ {
        var out bytes.Buffer
        
        v := make([]uint64,10)
        
        for j:=0; j<len(v); j++ {
            var m uint64                        // 64 bit mask
            sw := rand.Int63()%4                // Pick a random choice
            switch sw {
                case 0: m = 0xFF                // Random byte  
                case 1: m = 0xFFFF              // Random 16 bit integer
                case 2: m = 0xFFFFFFFF          // Random 32 bit integer
                case 3: m = 0xFFFFFFFFFFFFFFFF  // Random 64 bit integer
            }
            n := uint64(rand.Int63() + (rand.Int63() << 32))
            v[j] = n & m 
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
    