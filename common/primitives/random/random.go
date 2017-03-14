// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package random

import (
	"crypto/rand"
	"math"
	"math/big"
)

func RandUInt64() uint64 {
	return RandUInt64Between(0, math.MaxUint64)
}

func RandUInt32() uint32 {
	return uint32(RandUInt64Between(0, math.MaxUint32))
}

func RandUInt8() uint8 {
	return uint8(RandUInt64Between(0, math.MaxUint8))
}

func RandByte() byte {
	return RandByteSliceOfLen(1)[0]
}

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

func RandInt64() int64 {
	int64max := big.NewInt(math.MaxInt64)
	randnum, _ := rand.Int(rand.Reader, int64max)
	return randnum.Int64()
}

func RandInt64Between(min, max int64) int64 {
	if min >= max {
		return 0
	}
	int64max := big.NewInt(max - min)
	randnum, _ := rand.Int(rand.Reader, int64max)
	m := big.NewInt(min)
	randnum = randnum.Add(randnum, m)
	return randnum.Int64()
}

func RandInt() int {
	return int(RandInt64())
}

func RandIntBetween(min, max int) int {
	return int(RandInt64Between(int64(min), int64(max)))
}

func RandByteSlice() []byte {
	l := RandInt() % 64
	answer := make([]byte, l)
	_, err := rand.Read(answer)
	if err != nil {
		return nil
	}
	return answer
}

func RandNonEmptyByteSlice() []byte {
	l := RandInt()%63 + 1
	answer := make([]byte, l)
	_, err := rand.Read(answer)
	if err != nil {
		return nil
	}
	return answer
}

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

const StringAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()"

func RandomString() string {
	l := RandIntBetween(0, 128)
	answer := []byte{}
	for i := 0; i < l; i++ {
		answer = append(answer, StringAlphabet[RandIntBetween(0, len(StringAlphabet)-1)])
	}
	return string(answer)
}
