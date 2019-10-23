// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package random

import (
	"crypto/rand"
	"math"
	"math/big"
	"strconv"
)

// UintSize is either 32 or 64. Bitwise complement 0 to all 1's, bitshift 32.
// a) Will be 0 on 32 bit systems causing 0 & 1 to be zero, leaving 32.
// b) Will be 1 on 64 bit systems causing 1 & 1 to be 1, shifting 32 to 64
const UintSize = 32 << (^uint(0) >> 32 & 1) // 32 or 64

// Define maximum int sizes (correct on 32 bit or 64 bit systems)
const (
	MaxInt  = 1<<(UintSize-1) - 1 // 1<<31 - 1 or 1<<63 - 1
	MinInt  = -MaxInt - 1         // -1 << 31 or -1 << 63
	MaxUint = 1<<UintSize - 1     // 1<<32 - 1 or 1<<64 - 1
)

// RandUInt64 returns a random number between [0,MaxUint64)
func RandUInt64() uint64 {
	return RandUInt64Between(0, math.MaxUint64)
}

// RandUInt32 returns a random number between [0,MaxUint32)
func RandUInt32() uint32 {
	return uint32(RandUInt64Between(0, math.MaxUint32))
}

// RandUInt8 returns a random number between [0,MaxUint8)
func RandUInt8() uint8 {
	return uint8(RandUInt64Between(0, math.MaxUint8))
}

// RandByte returns a random byte
func RandByte() byte {
	return RandByteSliceOfLen(1)[0]
}

// RandUInt64Between returns a random number between [min,max). If min>=max then returns 0.
func RandUInt64Between(min, max uint64) uint64 {
	if min >= max {
		return 0
	}
	uint64max := big.NewInt(0)
	uint64max.SetUint64(max - min)
	randnum, _ := rand.Int(rand.Reader, uint64max)
	m := big.NewInt(0)
	m.SetUint64(min)
	randnum = randnum.Add(randnum, m)
	return randnum.Uint64()
}

// RandInt64 returns a random number between [0,MaxInt64). Note - even though the return int64 is signed,
// the underlying function call should never return a negative value.
func RandInt64() int64 {
	int64max := big.NewInt(math.MaxInt64)
	randnum, _ := rand.Int(rand.Reader, int64max)
	return randnum.Int64()
}

// RandInt64Between returns a random number between [min,max). If min>=max then returns 0. It takes special care
// to ensure the int64 difference between max-min doesn't overflow. Worst case int64 can support from -N to N-1
// which if input into this system as max/min creates (N-1)-(-N) = 2*N-1 number which would overflow int64. To
// deal with this, we make max-min a uint64 and get a random uint64. If the random number is larger than int64
// maximum value, then we first add the max int64 to the min value, and subtract the max int64 from the random
// uint64 number, thus breaking the addition into two smaller pieces which each fit within the int64 size
func RandInt64Between(min, max int64) int64 {
	if min >= max {
		return 0
	}

	// A Uint64 allows the full width of max-min to be stored
	therange := big.NewInt(0)
	therange.SetUint64(uint64(max - min))
	randnum, _ := rand.Int(rand.Reader, therange) // randnum is potentially larger than int64 can handle
	m := big.NewInt(0)
	m.SetInt64(min)
	if randnum.Uint64() > math.MaxInt64 {
		// The random number is larger than the maximum int64 and must use uint64 math
		// Note: The only way this is possible is if the min is negative (necessary but not sufficient to be here)
		maxInt64 := big.NewInt(math.MaxInt64)
		m = m.Add(maxInt64, m)                   // First add the MaxInt64 number to the input min
		randnum = randnum.Sub(randnum, maxInt64) // Then make randnum the residual (subtract MaxInt64)
	}
	randnum = randnum.Add(randnum, m) // We either directly add the randnum (or the residual from large number)

	return randnum.Int64()
}

// RandIntBetween returns a random number between [min,max). If min>=max then returns 0. The same special care
// is used as explained in the above RandInt64Between function. Here we never allow something larger than
// MaxInt to be added to the int, which preserves safe addition if we are on a 32 bit or 64 bit system.
func RandIntBetween(min, max int) int {
	if min >= max {
		return 0
	}

	// A Uint64 allows the full width of max-min to be stored
	therange := big.NewInt(0)
	therange.SetUint64(uint64(max - min))
	randnum, _ := rand.Int(rand.Reader, therange) // randnum is potentially larger than int can handle
	m := big.NewInt(0)
	m.SetInt64(int64(min))
	if randnum.Uint64() > MaxInt {
		// The random number is larger than the maximum int and so we use careful math to avoid integer oveflow
		maxInt := big.NewInt(int64(MaxInt))
		m = m.Add(maxInt, m)                   // First add the MaxInt number to the input min
		randnum = randnum.Sub(randnum, maxInt) // Then make randnum the residual (subtract MaxInt)
	}
	randnum = randnum.Add(randnum, m) // We either directly add the randnum (or the residual from large number)

	// To go back from a big.Int to a standard int, we write to a string and then convert it back to an int
	randstringnum := randnum.String()
	randnumint, err := strconv.Atoi(randstringnum)
	if err != nil {
		panic("RandIntBetween: Unable to convert string to int")
	}
	return randnumint
}

// RandInt returns a random integer between [0,MaxInt64)
func RandInt() int {
	return RandIntBetween(0, MaxInt)
}

// RandByteSlice returns a random byte slice of length 0 <= len <= 63
func RandByteSlice() []byte {
	l := RandInt() % 64
	return RandByteSliceOfLen(l)
}

// RandNonEmptyByteSlice returns a random byte slice of length 1 <= len <= 63 (guaranteed not zero length)
func RandNonEmptyByteSlice() []byte {
	l := RandInt()%63 + 1
	return RandByteSliceOfLen(l)
}

// RandByteSliceOfLen returns a random byte slice of specified length
func RandByteSliceOfLen(l int) []byte {
	if l <= 0 {
		return nil
	}
	answer := make([]byte, l)
	_, err := rand.Read(answer)
	if err != nil {
		return nil
	}
	return answer
}

// StringAlphabet is the valid characters supported by RandomString() below
const StringAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()"

// RandomString creates a random string of length 0 <= len < 128 consisting only of characters within
// the set of lower and uppercase letters, numbers 0-9, and the special characters associated with 'shift+<num>'
// on your keyboard
func RandomString() string {
	l := RandIntBetween(0, 128)
	answer := []byte{}
	for i := 0; i < l; i++ {
		answer = append(answer, StringAlphabet[RandIntBetween(0, len(StringAlphabet)-1)])
	}
	return string(answer)
}
