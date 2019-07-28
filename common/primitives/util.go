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

// CalculateCoinbasePayout computes the payout amount for an authority server at the input
// efficiency. Input efficiency is percentage * 100. So 60% inputs as 6000.
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

// EfficiencyToString returns the input efficiency to two decimal places (ie, 62.25)
func EfficiencyToString(eff uint16) string {
	return fmt.Sprintf("%d.%02d", eff/100, eff%100)
}

/*********************************
 * Print helpers
 ********************************/

// AddCommas returns a string of the input number formatted with commas every 3 significant digits (ie, "66,000")
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

// WriteNumber64 writes the input number to the buffer
func WriteNumber64(out *Buffer, num uint64) {
	var buf Buffer

	binary.Write(&buf, binary.BigEndian, num)
	str := hex.EncodeToString(buf.DeepCopyBytes())
	out.WriteString(str)

}

// WriteNumber32 writes the input number to the buffer
func WriteNumber32(out *Buffer, num uint32) {
	var buf Buffer

	binary.Write(&buf, binary.BigEndian, num)
	str := hex.EncodeToString(buf.DeepCopyBytes())
	out.WriteString(str)

}

// WriteNumber16 writes the input number to the buffer
func WriteNumber16(out *Buffer, num uint16) {
	var buf Buffer

	binary.Write(&buf, binary.BigEndian, num)
	str := hex.EncodeToString(buf.DeepCopyBytes())
	out.WriteString(str)

}

// WriteNumber8 writes the input number to the buffer
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

// User facing addresses (both public and private) are prefixed with the following letters depending on
// the address type

// FactoidPrefix = FA or 0x5fb1
var FactoidPrefix = []byte{0x5f, 0xb1}

// EntryCreditPrefix = EC or 0x592a
var EntryCreditPrefix = []byte{0x59, 0x2a}

// FactoidPrivatePrefix = Fs or 0x6478
var FactoidPrivatePrefix = []byte{0x64, 0x78}

// EntryCreditPrivatePrefix = Es or 0x5db6
var EntryCreditPrivatePrefix = []byte{0x5d, 0xb6}

// The flow for creating the secret (Fs) Factoid address
// 1) Create a random 256 bit (32 byte) number as a private key
// 2) Add prefix 'Fs' to the 32 byte number
// 3) Double SHA the combined string from step 2, take first 4 bytes of number and postfix #2
// 4) You now have "prefix-32bytenumber-postfix" which is the Fs address

// The flow for creating the public (FA) Factoid address
// 1) Create the public key from the private key using ed25519
// 2) Concatenate the RCD mechanism type <int> with the public key "int-publickey"
// 3) Double SHA the combined string from step 2 to get hash
// 4) Add prefix 'FA' to the hash from step 3
// 5) Double SHA the string from #4, take first 4 bytes of number and postfix #4
// 4) You now have "prefix-rcahash-postfix"

// Note: The above two flows can be modified by replacing the input prefix to the proper 'EC' or 'Es'
// to create entry credit addresses instead of factoid addresses

// ConvertDecimalToFloat converts factoshis to floating point factoids
func ConvertDecimalToFloat(v uint64) float64 {
	f := float64(v)
	f = f / 100000000.0 // 1e8 factoshis = 1 Factoid
	return f
}

// ConvertDecimalToString converts factoshis to floating point factoids as a string with 8 decimal precision
func ConvertDecimalToString(v uint64) string {
	f := ConvertDecimalToFloat(v)
	return fmt.Sprintf("%.8f", f)
}

// ConvertDecimalToPaddedString converts factoshis to floating point factoids as a string with up to 8 decimal precision,
// having trailing 0's removed from the decimal side
func ConvertDecimalToPaddedString(v uint64) string {
	tv := v / 100000000 // 1e8 factoshis = 1 Factoid
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
	// Print the proper format string: " %12v.%(8-cnt)" (ie, remove trailing 0's from decimal region)
	fstr := fmt.Sprintf(" %s%dv.%s0%vd", "%", 12, "%", 8-cnt)
	// Use the format string to print our Factoid balance
	str = fmt.Sprintf(fstr, tv, bv)

	return str
}

// ConvertFixedPoint converts floating point Factoid string to Factoshis,
// output suitable for Factom to chew on.
func ConvertFixedPoint(amt string) (string, error) {
	var v int64
	var err error
	index := strings.Index(amt, ".")
	// Add a leading 0 if missing from a decimal entry (ie, .75 -> 0.75)
	if index == 0 {
		amt = "0" + amt
		index++
	}
	if index < 0 { // Could not find decimal point convert from Factoids to Factoshis
		v, err = strconv.ParseInt(amt, 10, 64)
		if err != nil {
			return "", err
		}
		v *= 100000000 // Convert to Factoshis
	} else { // Decimal point found somewhere in the string
		// Convert pre-decimal point whole Factoids to Factoshis
		tp := amt[:index]
		v, err = strconv.ParseInt(tp, 10, 64)
		if err != nil {
			return "", err
		}
		v = v * 100000000 // Convert to Factoshis

		// Convert post-decimal point fractional Factoids to Factoshis
		bp := amt[index+1:]
		if len(bp) > 8 { // We keep only 8 decimal places of Factoids (smallest unit is a Factoshi)
			bp = bp[:8]
		}
		bpv, err := strconv.ParseInt(bp, 10, 64)
		if err != nil {
			return "", err
		}
		for i := 0; i < 8-len(bp); i++ { // If residual decimal integer is not 8 digits long, we keep potentially missing trailing 0's
			bpv *= 10
		}
		v += bpv
	}
	return strconv.FormatInt(v, 10), nil
}

// ConvertAddressToUser does the following:
//  Convert Factoid and Entry Credit addresses to their more user
//  friendly and human readable formats.
//
//  Creates the binary form.  Just needs the conversion to base58
//  for display.
func ConvertAddressToUser(prefix []byte, addr interfaces.IAddress) []byte {
	dat := make([]byte, 0, 64)
	dat = append(dat, prefix...)
	dat = append(dat, addr.Bytes()...)
	sha256d := Sha(Sha(dat).Bytes()).Bytes()
	userd := prefix
	userd = append(userd, addr.Bytes()...)
	userd = append(userd, sha256d[:4]...)
	return userd
}

// ConvertFctAddressToUserStr converts the input Factoid RCD hash to a user facing public Factoid Address (FA)
func ConvertFctAddressToUserStr(addr interfaces.IAddress) string {
	//NOTE: This converts the final hash into user-readable string, NOT the public key!
	//In practical terms, you'll need to convert the public key into RCD,
	//then hash it before using this function!
	userd := ConvertAddressToUser(FactoidPrefix, addr)
	return base58.Encode(userd)
}

// ConvertFctPrivateToUserStr converts the Factoid Private Key to a user facing private Factoid Address (Fs)
func ConvertFctPrivateToUserStr(addr interfaces.IAddress) string {
	userd := ConvertAddressToUser(FactoidPrivatePrefix, addr)
	return base58.Encode(userd)
}

// ConvertECAddressToUserStr converts the input Entry Credit RCD hash to a user facing public Entry Credit address (EC)
func ConvertECAddressToUserStr(addr interfaces.IAddress) string {
	userd := ConvertAddressToUser(EntryCreditPrefix, addr)
	return base58.Encode(userd)
}

// ConvertECPrivateToUserStr converts the Entry Credit Private key to a user facing private Entry Credit address (Es)
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

// ValidateFUserStr returns true iff the input public Factoid Address (FA) is valid (length, prefix, and 4 byte checksum)
func ValidateFUserStr(userFAddr string) bool {
	return validateUserStr(FactoidPrefix, userFAddr)
}

// ValidateFPrivateUserStr returns true iff the input private Factoid Address (Fs) is valid (length, prefix, and 4 byte checksum)
func ValidateFPrivateUserStr(userFAddr string) bool {
	return validateUserStr(FactoidPrivatePrefix, userFAddr)
}

// ValidateECUserStr returns true iff the input public Entry Credit Address (EC) is valid (length, prefix, and 4 byte checksum)
func ValidateECUserStr(userFAddr string) bool {
	return validateUserStr(EntryCreditPrefix, userFAddr)
}

// ValidateECPrivateUserStr returns true iff the input private Entry Credit Address (Es) is valid (length, prefix, and 4 byte checksum)
func ValidateECPrivateUserStr(userFAddr string) bool {
	return validateUserStr(EntryCreditPrivatePrefix, userFAddr)
}

// ConvertUserStrToAddress converts a User facing Factoid or Entry Credit address
// to their Private Key representations by converting from base58 to binary and
// removing their prefix two letter characters and the 4 byte checksum at the end
// Note validation must be done separately!
func ConvertUserStrToAddress(userFAddr string) []byte {
	v := base58.Decode(userFAddr)
	return v[2:34]
}
