// +build all 

package blockgen_test

import (
	"testing"

	. "github.com/FactomProject/factomd/Utilities/DatabaseGenerator/blockgen"
)

func TestRange(t *testing.T) {
	testRange := func(ir Range) {
		for i := 0; i < 100; i++ {
			if ir.Amount() > ir.Max {
				t.Error("Amount is over")
			}
			if ir.Amount() < ir.Min {
				t.Error("Amount is under")
			}
		}
	}

	r := Range{999, 1000}
	testRange(r)
	r = Range{5, 100}
	testRange(r)
	r = Range{0, 1000}
	testRange(r)
	r = Range{9000, 10000}
	testRange(r)
	r = Range{55, 500}
	testRange(r)
	r = Range{16, 200}
	testRange(r)
}
