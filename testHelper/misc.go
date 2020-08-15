package testHelper

import (
	"github.com/PaulSnow/factom2d/common/constants"
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/primitives"
)

func NewRepeatingHash(b byte) interfaces.IHash {
	tmp := make([]byte, constants.HASH_LENGTH)
	for i := range tmp {
		tmp[i] = b
	}
	return primitives.NewHash(tmp)
}
