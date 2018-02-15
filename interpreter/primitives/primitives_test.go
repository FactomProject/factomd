package primitives_test

import (
	"testing"
	"strings"
	. "github.com/FactomProject/electiontesting/interpreter/primitives"
)

func TestPrimitives(t *testing.T) {

	p := NewPrimitives()
//	p.Interpret(strings.NewReader("1 2 + ."))
	p.Interpret(strings.NewReader("{ 1 2 + . } exec"))

}
