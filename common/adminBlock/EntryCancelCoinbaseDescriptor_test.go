package adminBlock_test

import (
	"testing"

	"math/rand"
	"time"

	. "github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/testHelper"
)

func TestCancelCoinbaseDescriptorMarshal(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 100; i++ {
		h := rand.Uint32()
		i := rand.Uint32()
		a := NewCancelCoinbaseDescriptor(h, i)

		b := NewCancelCoinbaseDescriptor(0, 0)
		testHelper.TestMarshaling(a, b, rand.Intn(100), t)

		if !a.IsSameAs(b) {
			t.Errorf("Objects are not the same")
		}

		testHelper.TestABlockEntryFunctions(a, b, t)
	}

	// Test the empty
	{
		a := NewCancelCoinbaseDescriptor(0, 0)

		b := NewCancelCoinbaseDescriptor(0, 0)
		testHelper.TestMarshaling(a, b, rand.Intn(100), t)

	}

}
