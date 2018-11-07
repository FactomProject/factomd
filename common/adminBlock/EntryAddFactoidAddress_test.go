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

		testHelper.TestABlockEntryFunctions(a, b, t)
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

func TestAddBadFactoidAddress(t *testing.T) {
	f1 := NewAddFactoidAddress(primitives.RandomHash(), factoid.RandomAddress())
	p, err := f1.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	p[1] = 0xff // replace body lenght with bad value

	f2 := new(AddFactoidAddress)
	err = f2.UnmarshalBinary(p)
	if err == nil {
		t.Error("AddFactoidAddress should have errored on unmarshal", f2)
	} else {
		t.Log(err)
	}
}
