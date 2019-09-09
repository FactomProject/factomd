package adminBlock_test

import (
	"testing"

	"math/rand"
	"time"

	. "github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/testHelper"
)

// TestCancelCoinbaseDescriptorMarshal creates 100 random CancelCoinbaseDescriptors and checks they can be marshaled and unmarshaled correctly.
// It also tests the an empty CancelCoinbaseDescriptor
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

// TestAddBadCancelCoinbaseDescriptor deliberately corrupts a marshaled CancelCoinbasDescriptor and confirms it errors out upon unmarshalling
func TestAddBadCancelCoinbaseDescriptor(t *testing.T) {
	c1 := NewCancelCoinbaseDescriptor(1000, 2000)
	p, err := c1.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	p[1] = 0xff // replace body length with bad value

	c2 := new(CancelCoinbaseDescriptor)
	err = c2.UnmarshalBinary(p)
	if err == nil {
		t.Error("CancelCoinbaseDescriptor should have errored on unmarshal", c2)
	} else {
		t.Log(err)
	}
}
