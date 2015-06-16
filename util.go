// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"runtime/debug"
)

/*********************************
 * Marshalling helper functions
 *********************************/

func WriteNumber64(out *bytes.Buffer, num uint64) {
	var buf bytes.Buffer

	binary.Write(&buf, binary.BigEndian, num)
	str := hex.EncodeToString(buf.Bytes())
	out.WriteString(str)

}

func WriteNumber32(out *bytes.Buffer, num uint32) {
	var buf bytes.Buffer

	binary.Write(&buf, binary.BigEndian, num)
	str := hex.EncodeToString(buf.Bytes())
	out.WriteString(str)

}

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

/**************************************
 * Printing Helper Functions for debugging
 **************************************/

func PrtStk() {
	Prtln()
	debug.PrintStack()
}

func Prt(a ...interface{}) {
	fmt.Print(a...)
}

func Prtln(a ...interface{}) {
	fmt.Println(a...)
}

func PrtData(data []byte) {
	if data == nil || len(data) == 0 {
		fmt.Print("No Data Here")
	} else {
		var nl string = "\n"
		for i, b := range data {
			fmt.Print(nl)
			nl = ""
			fmt.Printf("%2.2X ", int(b))
			if i%32 == 31 {
				nl = "\n"
			} else if i%8 == 7 {
				fmt.Print(" | ")
			}
		}
	}
}
func PrtDataL(title string, data []byte) {
	fmt.Println()
	fmt.Println(title)
	fmt.Print("========================-+-========================-+-========================-+-========================")
	PrtData(data)
	fmt.Println("\n========================-+-========================-+-========================-+-========================")
}

// Does a new line, then indents as specified. DON'T end
// a Print with a CR!
func CR(level int) {
	Prtln()
	PrtIndent(level)
}

func PrtIndent(level int) {
	for i := 0; i < level && i < 10; i++ { // Indent up to 10 levels.
		Prt("    ") //   by printing leading spaces
	}
}

/**********************************
 * User Addresses
 **********************************/
/*
 * Factoid Address
 * 
 Factoids are sent to an RCD Hash. Inside the computer, the RCD hash 
 is represented as a 32 byte number. The user sees Factoid addresses 
 as a 52 ch*aracter string starting with FA.
 
 A User Factoid address is constructed like this (using a zero RCD 
 hash for clarity:
 Concatenate 0x5fb1 and the RCD Hash bytewise
 5fb10000000000000000000000000000000000000000000000000000000000000000 
 
 Take the SHA256d of the above data. Append the first 4 bytes of this 
 SHA256d to the end of the above value bytewise.
 5fb10000000000000000000000000000000000000000000000000000000000000000d48a8e32
 Convert the above value from base 256 to base 58. Use standard 
 Bitcoin base58 encoding to display the number.
 
 FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu
 Factoid addresses will range between 
    FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu 
 and 
    FA3upjWMKHmStAHR5ZgKVK4zVHPb8U74L2wzKaaSDQEonHajiLeq
*/

func getFUserAddress(address IAddress) string {
    user := make([]byte,0,2+32+4)   // 2 byte prefix + 32 byte addr + 4 byte chksum
    shad := Sha(Sha(address.Bytes()).Bytes())
    user = append(user, []byte{0x5f, 0xb1}...)
    user = append(user,address.Bytes()  ...)
    user = append(user,shad.Bytes()[:4]...)
    
    
    return ""
    }
func getFAddressFromUser(userAdr string) (IAddress, error) {

    return nil, nil
}

