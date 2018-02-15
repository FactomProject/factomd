package stack_test

import (
	"testing"
	. "github.com/FactomProject/electiontesting/interpreter/stack"
)

func TestStack(t *testing.T) {

	s := NewStack()

	s.Push(1, 2)
	x := s.Pop()
	y := s.Pop()
	if y!=1 {
		t.Errorf("expected 1 got %d", y)
	}
	if x!=2 {
		t.Errorf("expected 2 got %d", x)
	}


}
