// +build all 

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

func TestCompleteManagerBadInput(t *testing.T) {
	cm := NewCompletedHeightManager()
	// Test the underflow
	if cm.CompleteHeight(-1) {
		t.Error("Should be false")
	}
	if !cm.CompleteHeight(0) {
		t.Error("Should be true")
	}

	// Test on the end
	for i := 0; i < 10; i++ {
		if !cm.CompleteHeight(len(cm.Completed)) {
			t.Error("Should be true")
		}
	}

	// Move the base
	cm.ClearTo(1000)

	// Test on the end + move
	for i := 0; i < 10; i++ {
		if !cm.CompleteHeight(1000 + len(cm.Completed)) {
			t.Error("Should be true")
		}
	}

	// First one is true, so throw that first for each checking
	if !cm.CompleteHeight(1000 + len(cm.Completed) - 1) {
		t.Error("Should be true")
	}

	// Test on the end -1
	for i := 0; i < 10; i++ {
		if cm.CompleteHeight(1000 + len(cm.Completed) - 1) {
			t.Error("Should be false")
		}
	}
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
