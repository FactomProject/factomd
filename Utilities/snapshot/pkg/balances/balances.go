package balances

import (
	"fmt"
	"io"

	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/primitives"
)

type Balances struct {
	Height uint32
	// Use int64s because temporarily during the math, we might have a negative balance
	FCTAddressMap map[[32]byte]int64
	ECAddressMap  map[[32]byte]int64
}

func NewBalances() *Balances {
	return &Balances{
		FCTAddressMap: make(map[[32]byte]int64),
		ECAddressMap:  make(map[[32]byte]int64),
	}
}

func (bs *Balances) Dump(w io.Writer) error {
	height := bs.Height
	_, _ = fmt.Fprintf(w, "height %d\n", height)
	for k, v := range bs.FCTAddressMap {
		_, err := fmt.Fprintf(w, "%s: %d\n", primitives.ConvertFctAddressToUserStr(factoid.NewAddress(k[:])), v)
		if err != nil {
			return fmt.Errorf("write fct addr: %w", err)
		}
	}
	for k, v := range bs.ECAddressMap {
		_, err := fmt.Fprintf(w, "%s: %d\n", primitives.ConvertECAddressToUserStr(factoid.NewAddress(k[:])), v)
		if err != nil {
			return fmt.Errorf("write ec addr: %w", err)
		}
	}

	return nil
}
