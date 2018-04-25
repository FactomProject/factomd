package adminBlock_test

import (
	"testing"

	"math/rand"
	"time"

	. "github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/testHelper"
)

func TestCoinbaseDescriptorMarshal(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 100; i++ {
		var outputs []interfaces.ITransAddress
		for c := 0; c < rand.Intn(64); c++ {
			outputs = append(outputs, factoid.RandomTransAddress())
		}
		a := NewCoinbaseDescriptor(outputs)

		b := NewCoinbaseDescriptor(nil)
		testHelper.TestMarshaling(a, b, rand.Intn(100), t)

		if !a.IsSameAs(b) {
			t.Errorf("Objects are not the same")
		}
	}

	// Test the empty
	{
		var outputs []interfaces.ITransAddress
		a := NewCoinbaseDescriptor(outputs)

		b := NewCoinbaseDescriptor(nil)
		testHelper.TestMarshaling(a, b, rand.Intn(100), t)

		if !a.IsSameAs(b) {
			t.Errorf("Objects are not the same")
		}
	}
}
