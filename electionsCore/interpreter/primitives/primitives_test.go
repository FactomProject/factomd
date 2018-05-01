package primitives_test

import (
	"strings"
	"testing"

	. "github.com/FactomProject/factomd/electionsCore/interpreter/primitives"
)

func TestPrimitives(t *testing.T) {

	// TODO: Need to make this so it actually checks the result instead of printing it :-)
	p := NewPrimitives()
	p.Interpret(strings.NewReader("1 0 + ."))
	p.Interpret(strings.NewReader("1 0 7 8 clear .s"))
	p.Interpret(strings.NewReader("{ 1 1 + . } exec "))
	p.Interpret(strings.NewReader("3 1 5 { I . } for 5 ."))
	p.Interpret(strings.NewReader(`" six" .`))
	p.Interpret(strings.NewReader(`{ 7 1 10 { I . } for } /test def test`))
	p.Interpret(strings.NewReader(`11 true if false { drop 0 } if .`))
	p.Interpret(strings.NewReader("12 { dup . 1 + } 3 repeat drop"))
	p.Interpret(strings.NewReader(`[ 1 2 3 4 5 ] { . } forall`))
	p.Interpret(strings.NewReader(`{ " thirteen" } exec .`))
	p.Interpret(strings.NewReader(`0 [ 1 6 7 ] dup . { + } forall . `))
	p.Interpret(strings.NewReader(`[ 15 ] /foo def foo . `))
	p.Interpret(strings.NewReader(`[ 1 5 5 5 /+ executable  dup dup ] executable exec .  `))

}
