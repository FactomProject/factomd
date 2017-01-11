package random

import (
	"crypto/rand"
	"math"
	"math/big"
)

func RandUInt64() uint64 {
	uint64max := big.NewInt(0)
	uint64max.SetUint64(math.MaxUint64)
	randnum, _ := rand.Int(rand.Reader, uint64max)
	return randnum.Uint64()
}

func RandInt64() int64 {
	int64max := big.NewInt(math.MaxInt64)
	randnum, _ := rand.Int(rand.Reader, int64max)
	return randnum.Int64()
}

func RandInt() int {
	return int(RandInt64())
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
