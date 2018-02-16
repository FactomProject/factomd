package primitives_test

import (
	"strings"
	"testing"

	. "github.com/FactomProject/electiontesting/interpreter/primitives"
)

func TestPrimitives(t *testing.T) {

	p := NewPrimitives()
	//p.Interpret(strings.NewReader("1 0 + ."))
	//p.Interpret(strings.NewReader("{ 1 1 + . } exec 4 ."))
	//p.Interpret(strings.NewReader("1 1 5 { I . } for 5 ."))
	//p.Interpret(strings.NewReader("\" abc def\" 6 . ."))
	//p.Interpret(strings.NewReader("{ 1 1 5 { I . } for 5 } \" test\" def test"))
	p.Interpret(strings.NewReader("6 { dup . 1 - } 5 repeat"))

}
