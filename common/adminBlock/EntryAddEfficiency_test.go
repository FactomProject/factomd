package adminBlock_test

import (
	"testing"

	"math/rand"
	"time"

	. "github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/testHelper"
)

func TestAddEfficiency(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 100; i++ {
		a := NewAddEfficiency(primitives.RandomHash(), uint16(rand.Intn(10000)))

		b := NewAddEfficiency(primitives.RandomHash(), 0)
		testHelper.TestMarshaling(a, b, rand.Intn(100), t)

		if !a.IsSameAs(b) {
			t.Errorf("Objects are not the same")
		}

		testHelper.TestABlockEntryFunctions(a, b, t)
	}

	// Test the empty
	{
		a := NewAddEfficiency(primitives.RandomHash(), uint16(rand.Intn(10000)))

		b := NewAddEfficiency(primitives.RandomHash(), 0)
		testHelper.TestMarshaling(a, b, rand.Intn(100), t)

		if !a.IsSameAs(b) {
			t.Errorf("Objects are not the same")
		}
	}
}
