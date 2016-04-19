package testHelper

import (
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

func NewRepeatingHash(b byte) interfaces.IHash {
	tmp := make([]byte, constants.HASH_LENGTH)
	for i := range tmp {
		tmp[i] = b
	}
	return primitives.NewHash(tmp)
}
