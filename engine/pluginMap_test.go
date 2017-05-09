package engine_test

import (
	"testing"

	// "github.com/FactomProject/factomd/common/primitives/random"
	. "github.com/FactomProject/factomd/engine"
)

func TestNoPanicCompleteManager(t *testing.T) {
	cm := NewCompletedHeightManager()
	cm.CompleteHeight(10000)

	cm = NewCompletedHeightManager()
	cm.ClearTo(10000)
}

func TestCompleteManager(t *testing.T) {
	cm := NewCompletedHeightManager()

	// Simple interaction
	for i := 0; i < 1000; i++ {
		a := cm.CompleteHeight(i)
		if !a {
			t.Error("Should be true")
		}
	}

	// Should be false
	for i := 0; i < 1000; i++ {
		a := cm.CompleteHeight(i)
		if a {
			t.Error("Should be false")
		}
	}

	for i := 0; i < 100; i++ {
		n := i + 4000

		a := cm.CompleteHeight(n)
		if !a {
			t.Error("Should be true")
		}

		a = cm.CompleteHeight(n)
		if a {
			t.Error("Should be false")
		}
	}

	cm.ClearTo(2500)
	for i := 0; i < 100; i++ {
		n := i + 6000

		a := cm.CompleteHeight(n)
		if !a {
			t.Error("Should be true")
		}

		a = cm.CompleteHeight(n)
		if a {
			t.Error("Should be false")
		}
	}

	// Should be false
	for i := 0; i < 1000; i++ {
		a := cm.CompleteHeight(i)
		if a {
			t.Error("Should be false")
		}
	}

	to := len(cm.Completed) + cm.Base
	cm.ClearTo(len(cm.Completed) + cm.Base)

	// Should be false
	for i := 0; i < to; i++ {
		a := cm.CompleteHeight(i)
		if a {
			t.Error("Should be false")
		}
	}
}
