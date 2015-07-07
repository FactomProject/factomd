// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid

import (
    "strings"
	"bytes"
    "strconv"
    "encoding/binary"
	"encoding/hex"
	"fmt"
	"runtime/debug"
    "github.com/btcsuitereleases/btcutil/base58"
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

/************************************************
 * Helper Functions for User Address handling
 ************************************************/

// Factoid Address
// 
// 
// Add a prefix of 0x5fb1 at the start, and the first 4 bytes of a SHA256d to
// the end.  Using zeros for the address, this might look like:
// 
//     5fb10000000000000000000000000000000000000000000000000000000000000000d48a8e32
// 
// A typical Factoid Address:
//
//     FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu
// 
// Entry credits only differ by the prefix of 0x592a and typically look like:
//
//     EC3htx3MxKqKTrTMYj4ApWD8T3nYBCQw99veRvH1FLFdjgN6GuNK
//
// More words on this can be found here:
//
// https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#human-readable-addresses
//

var FactoidPrefix = []byte{ 0x5f, 0xb1 }
var EntryCreditPrefix = []byte{ 0x59, 0x2a }

// Take fixed point data and produce a nice decimial point
// sort of output that users can handle.
func ConvertDecimal(v uint64) string {   
    tv := v/100000000
    bv := v-(tv*100000000)
    var str string
    
    // Count zeros to lop off
    var cnt int 
    for cnt=0; cnt<7;cnt++ {
        if (bv/10)*10 != bv {break}
        bv = bv/10
    }
    // Print the proper format string
    fstr := fmt.Sprintf(" %s%dv.%s0%vd","%",12,"%",8-cnt)
    // Use the format string to print our Factoid balance
    str = fmt.Sprintf(fstr,tv,bv)
    
    return str
}    

// Convert Decimal point input to FixedPoint (no decimal point)
// output suitable for Factom to chew on.
func ConvertFixedPoint(amt string) (string, error) {
    var v int64
    var err error
    index := strings.Index(amt,".");
    if  index < 0 {
        v, err =strconv.ParseInt(amt,10,64)
        if err != nil {
            return "", err
        }
        v *= 100000000      // Convert to Factoshis
    }else{
        tp := amt[:index]
        v, err =strconv.ParseInt(tp,10,64)
        if err != nil {
            return "", err
        }
        v = v*100000000      // Convert to Factoshis
        
        bp := amt[index+1:]
        if len(bp)>8 {
            return "", fmt.Errorf("Format Error in amount")
        }
        bpv, err :=strconv.ParseInt(bp,10,64)
        if err != nil {
            return "", err
        }
        for i:=0; i<8-len(bp); i++ {
            bpv *= 10
        }
        v += bpv
    }
    return strconv.FormatInt(v,10),nil
}



//  Convert Factoid and Entry Credit addresses to their more user
//  friendly and human readable formats.
//
//  Creates the binary form.  Just needs the conversion to base58
//  for display.
func ConvertAddressToUser(prefix []byte, addr IAddress) []byte {
    sha256d := Sha(Sha(addr.Bytes()).Bytes()).Bytes()
    userd := make([]byte,0,32)
    userd = append(userd, prefix...)
    userd = append(userd, addr.Bytes()...)
    userd = append(userd, sha256d[:4]...)
    return userd
}

// Convert Factoid Addresses
func ConvertFctAddressToUserStr(addr IAddress) string {
    userd := ConvertAddressToUser(FactoidPrefix, addr)
    return base58.Encode(userd)
}

// Convert Entry Credits
func ConvertECAddressToUserStr(addr IAddress) string {
    userd := ConvertAddressToUser(EntryCreditPrefix, addr)
    return base58.Encode(userd)
}


//
// Validates a User representation of a Factom and 
// Entry Credit addresses.
//
// Returns false if the length is wrong.
// Returns false if the prefix is wrong.  
// Returns false if the checksum is wrong.
//
func validateUserStr(prefix []byte, userFAddr string) bool {
    if len(userFAddr) != 52 {  
        return false 
        
    }
    v := base58.Decode(userFAddr)
    if bytes.Compare(prefix, v[:2]) != 0 { 
        return false 
        
    }
    sha256d := Sha(Sha(v[2:34]).Bytes()).Bytes()
    if bytes.Compare (sha256d[:4],v[34:]) != 0 {
        return false 
    }
    return true
}

// Validate Factoids
func ValidateFUserStr(userFAddr string) bool {
    return validateUserStr(FactoidPrefix, userFAddr)
}

// Validate Entry Credits
func ValidateECUserStr(userFAddr string) bool {
    return validateUserStr(EntryCreditPrefix, userFAddr)
}

// Convert a User facing Factoid or Entry Credit address
// to the regular form.  Note validation must be done
// separately!
func ConvertUserStrToAddress(userFAddr string) []byte {
    v := base58.Decode(userFAddr)
    return v[2:34]
}



