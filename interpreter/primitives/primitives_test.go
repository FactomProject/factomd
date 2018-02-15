package primitives_test

import (
	. "github.com/FactomProject/electiontesting/interpreter/primitives"
	"strings"
	"testing"
)

func TestPrimitives(t *testing.T) {

	p := NewPrimitives()
	//	p.Interpret(strings.NewReader("1 2 + ."))
	p.Interpret(strings.NewReader("{ 1 2 + . } exec"))

}
