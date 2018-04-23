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

func TestAddFactoidAddressMarshal(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 100; i++ {
		a := NewAddFactoidAddress(primitives.RandomHash(), factoid.RandomAddress())

		b := NewAddFactoidAddress(nil, nil)
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
