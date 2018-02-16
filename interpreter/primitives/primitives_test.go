package primitives_test

import (
	"strings"
	"testing"

	. "github.com/FactomProject/electiontesting/interpreter/primitives"
)

func TestPrimitives(t *testing.T) {

	p := NewPrimitives()
	p.Interpret(strings.NewReader("1 0 + ."))
	p.Interpret(strings.NewReader("{ 1 1 + . } exec "))
	p.Interpret(strings.NewReader("3 1 5 { I . } for 5 ."))
	p.Interpret(strings.NewReader(`" six" .`))
	p.Interpret(strings.NewReader(`{ 7 1 10 { I . } for } " test" def test`))
	p.Interpret(strings.NewReader(`11 true if false { drop 0 } if .`))
	p.Interpret(strings.NewReader("12 { dup . 1 + } 3 repeat drop"))
	p.Interpret(strings.NewReader(`{ " thirteen" } exec .`))

}
