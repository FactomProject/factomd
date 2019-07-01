// +build all 

package adminBlock_test

import (
	"testing"

	"math/rand"
	"time"

	. "github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/testHelper"
)

func TestCoinbaseDescriptorMarshal(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 100; i++ {
		outputs := make([]interfaces.ITransAddress, 0)
		for c := 0; c < rand.Intn(64); c++ {
			outputs = append(outputs, factoid.RandomTransAddress())
			outputs[c].SetUserAddress(primitives.ConvertFctAddressToUserStr(outputs[c].GetAddress()))
		}
		a := NewCoinbaseDescriptor(outputs)

		b := NewCoinbaseDescriptor(make([]interfaces.ITransAddress, 0))
		testHelper.TestMarshaling(a, b, rand.Intn(100), t)

		if !a.IsSameAs(b) {
			t.Errorf("Objects are not the same")
		}

		testHelper.TestABlockEntryFunctions(a, b, t)
	}

	// Test the empty
	{
		outputs := make([]interfaces.ITransAddress, 0)
		a := NewCoinbaseDescriptor(outputs)

		b := NewCoinbaseDescriptor(make([]interfaces.ITransAddress, 0))
		testHelper.TestMarshaling(a, b, rand.Intn(100), t)

		if !a.IsSameAs(b) {
			t.Errorf("Objects are not the same")
		}
		testHelper.TestABlockEntryFunctions(a, b, t)
	}
}
