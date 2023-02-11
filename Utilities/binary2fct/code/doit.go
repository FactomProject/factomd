package code

import (
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/primitives"
)

func Doit(s string) string {
	h, err := primitives.NewShaHashFromStr(s)
	if err != nil {
		panic("")
	}
	add := factoid.CreateAddress(h)
	converted := primitives.ConvertFctAddressToUserStr(add)
	return converted
}
