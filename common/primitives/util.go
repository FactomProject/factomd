// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/btcsuitereleases/btcutil/base58"
)

func CalculateCoinbasePayout(efficiency uint16) uint64 {
	// Keep is the percentage of the coinbase kept for the authority
	//		(Percentage * 100)
	keep := 10000 - uint64(efficiency)

	// The amount of factoshis in the payout
	payout := keep * constants.COINBASE_PAYOUT_AMOUNT

	// Put the percentage back into the correct scale
	// 		(10000 == 100%, so divide by 10000)
	payout = payout / 10000
	return payout
}

func EfficiencyToString(eff uint16) string {
	return fmt.Sprintf("%d.%02d", eff/100, eff%100)
}

/*********************************
 * Print helpers
 ********************************/

func AddCommas(v int64) (ret string) {
	pos := true
	if v < 0 {
		pos = false
		v = -v
	}
	finish := func() {
		if pos {
			return
		}
		ret = "-" + ret
		return
	}
	defer finish()
	for {
		nxt := v / 1000
		this := v % 1000
		switch {
		case nxt == 0:
			ret = fmt.Sprintf("%d%s", this, ret)
			return
		default:
			ret = fmt.Sprintf(",%03d%s", this, ret)
			v = v / 1000
		}
	}

	return
}

/*********************************
 * Marshalling helper functions
 *********************************/

func WriteNumber64(out *Buffer, num uint64) {
	var buf Buffer

	binary.Write(&buf, binary.BigEndian, num)
	str := hex.EncodeToString(buf.DeepCopyBytes())
	out.WriteString(str)

}

func WriteNumber32(out *Buffer, num uint32) {
	var buf Buffer

	binary.Write(&buf, binary.BigEndian, num)
	str := hex.EncodeToString(buf.DeepCopyBytes())
	out.WriteString(str)

}

func WriteNumber16(out *Buffer, num uint16) {
	var buf Buffer

	binary.Write(&buf, binary.BigEndian, num)
	str := hex.EncodeToString(buf.DeepCopyBytes())
	out.WriteString(str)

}

func WriteNumber8(out *Buffer, num uint8) {
	var buf Buffer

	binary.Write(&buf, binary.BigEndian, num)
	str := hex.EncodeToString(buf.DeepCopyBytes())
	out.WriteString(str)

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

var FactoidPrefix = []byte{0x5f, 0xb1}
var EntryCreditPrefix = []byte{0x59, 0x2a}
var FactoidPrivatePrefix = []byte{0x64, 0x78}
var EntryCreditPrivatePrefix = []byte{0x5d, 0xb6}

// Converts factoshis to floating point factoids
func ConvertDecimalToFloat(v uint64) float64 {
	f := float64(v)
	f = f / 100000000.0
	return f
}

// Converts factoshis to floating point string
func ConvertDecimalToString(v uint64) string {
	f := ConvertDecimalToFloat(v)
	return fmt.Sprintf("%.8f", f)
}

// Take fixed point data and produce a nice decimial point
// sort of output that users can handle.
func ConvertDecimalToPaddedString(v uint64) string {
	tv := v / 100000000
	bv := v - (tv * 100000000)
	var str string

	// Count zeros to lop off
	var cnt int
	for cnt = 0; cnt < 7; cnt++ {
		if (bv/10)*10 != bv {
			break
		}
		bv = bv / 10
	}
	// Print the proper format string
	fstr := fmt.Sprintf(" %s%dv.%s0%vd", "%", 12, "%", 8-cnt)
	// Use the format string to print our Factoid balance
	str = fmt.Sprintf(fstr, tv, bv)

	return str
}

// Convert Decimal point input to FixedPoint (no decimal point)
// output suitable for Factom to chew on.
func ConvertFixedPoint(amt string) (string, error) {
	var v int64
	var err error
	index := strings.Index(amt, ".")
	if index == 0 {
		amt = "0" + amt
		index++
	}
	if index < 0 {
		v, err = strconv.ParseInt(amt, 10, 64)
		if err != nil {
			return "", err
		}
		v *= 100000000 // Convert to Factoshis
	} else {
		tp := amt[:index]
		v, err = strconv.ParseInt(tp, 10, 64)
		if err != nil {
			return "", err
		}
		v = v * 100000000 // Convert to Factoshis

		bp := amt[index+1:]
		if len(bp) > 8 {
			bp = bp[:8]
		}
		bpv, err := strconv.ParseInt(bp, 10, 64)
		if err != nil {
			return "", err
		}
		for i := 0; i < 8-len(bp); i++ {
			bpv *= 10
		}
		v += bpv
	}
	return strconv.FormatInt(v, 10), nil
}

//  Convert Factoid and Entry Credit addresses to their more user
//  friendly and human readable formats.
//
//  Creates the binary form.  Just needs the conversion to base58
//  for display.
func ConvertAddressToUser(prefix []byte, addr interfaces.IAddress) []byte {
	dat := prefix
	dat = append(dat, addr.Bytes()...)
	sha256d := Sha(Sha(dat).Bytes()).Bytes()
	userd := prefix
	userd = append(userd, addr.Bytes()...)
	userd = append(userd, sha256d[:4]...)
	return userd
}

// Convert Factoid Addresses
func ConvertFctAddressToUserStr(addr interfaces.IAddress) string {
	//NOTE: This converts the final hash into user-readable string, NOT the public key!
	//In practical terms, you'll need to convert the public key into RCD,
	//then hash it before using this function!
	userd := ConvertAddressToUser(FactoidPrefix, addr)
	return base58.Encode(userd)
}

// Convert Factoid Private Key
func ConvertFctPrivateToUserStr(addr interfaces.IAddress) string {
	userd := ConvertAddressToUser(FactoidPrivatePrefix, addr)
	return base58.Encode(userd)
}

// Convert Entry Credits
func ConvertECAddressToUserStr(addr interfaces.IAddress) string {
	userd := ConvertAddressToUser(EntryCreditPrefix, addr)
	return base58.Encode(userd)
}

// Convert Entry Credit Private key
func ConvertECPrivateToUserStr(addr interfaces.IAddress) string {
	userd := ConvertAddressToUser(EntryCreditPrivatePrefix, addr)
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
	if len(v) < 3 {
		return false
	}

	if bytes.Compare(prefix, v[:2]) != 0 {
		return false

	}

	sha256d := Sha(Sha(v[:34]).Bytes()).Bytes()
	if bytes.Compare(sha256d[:4], v[34:]) != 0 {
		return false
	}

	return true
}

// Validate Factoids
func ValidateFUserStr(userFAddr string) bool {
	return validateUserStr(FactoidPrefix, userFAddr)
}

// Validate Factoid Private Key
func ValidateFPrivateUserStr(userFAddr string) bool {
	return validateUserStr(FactoidPrivatePrefix, userFAddr)
}

// Validate Entry Credits
func ValidateECUserStr(userFAddr string) bool {
	return validateUserStr(EntryCreditPrefix, userFAddr)
}

// Validate Entry Credit Private Key
func ValidateECPrivateUserStr(userFAddr string) bool {
	return validateUserStr(EntryCreditPrivatePrefix, userFAddr)
}

// Convert a User facing Factoid or Entry Credit address
// or their Private Key representations
// to the regular form.  Note validation must be done
// separately!
func ConvertUserStrToAddress(userFAddr string) []byte {
	v := base58.Decode(userFAddr)
	return v[2:34]
}
