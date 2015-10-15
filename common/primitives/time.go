package primitives

import (
	"time"
)

func GetTimeMilli() uint64 {
	return uint64(time.Now().UnixNano()) / 1000000 // 10^-9 >> 10^-3
}

func GetTime() uint64 {
	return uint64(time.Now().Unix())
}
