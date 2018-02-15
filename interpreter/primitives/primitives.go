package primitives

import (
	. "github.com/FactomProject/electiontesting/interpreter/interpreter"
	. "github.com/FactomProject/electiontesting/interpreter/dictionary"
	"fmt"
	. "github.com/FactomProject/electiontesting/interpreter/common"
)

type Primitives struct {
	Interpreter
}

func NewPrimitives() *Primitives {
	p := new(Primitives)
	dict := NewDictionary()
	p.DictionaryPush(dict)            // Primitives Dictionary
	p.DictionaryPush(NewDictionary()) // User Dictionary

	dict.Add("+", func() { p.Push(p.PopInt() + p.PopInt()) }) // ToDo: Handle float too
	dict.Add("-", func() { p.Push(p.PopInt() - p.PopInt()) }) // ToDo: Handle float too
	dict.Add("*", func() { p.Push(p.PopInt() * p.PopInt()) }) // ToDo: Handle float too
	dict.Add("/", func() { p.Push(p.PopInt() / p.PopInt()) }) // ToDo: Handle float too
	dict.Add("&", func() { p.Push(p.PopInt() & p.PopInt()) })
	dict.Add("|", func() { p.Push(p.PopInt() | p.PopInt()) })
	dict.Add("^", func() { p.Push(p.PopInt() ^ p.PopInt()) })
	dict.Add("~", func() { p.Push(p.PopInt() ^ -1) })
	dict.Add("&&", func() { p.Push(p.PopBool() && p.PopBool()) })
	dict.Add("||", func() { p.Push(p.PopBool() || p.PopBool()) })
	dict.Add("!", func() { p.Push(!p.PopBool()) })
	dict.Add(".", func() { fmt.Printf("%v\n", p.Pop()) })
	dict.Add("<", func() { p.Push(p.PopInt() < p.PopInt()) })
	dict.Add("=", func() { p.Push(p.PopInt() == p.PopInt()) })
	dict.Add("!=", func() { p.Push(p.PopInt() != p.PopInt()) })
	dict.Add(">", func() { p.Push(p.PopInt() > p.PopInt()) })
	dict.Add("0=", func() { p.Push(p.PopInt() > 0) })
	dict.Add("?dup", func() {
		x := p.Peek().(int)
		if x != 0 {
			p.Push(x)
		}
	})
	dict.Add("drop", func() { p.Pop() })
	dict.Add("swap", func() {
		x := p.Pop()
		y := p.Pop()
		p.Push(y, x)
	})

	// executable array
	dict.Add("{", func() { p.StartXArray() })
	dict.Add("}", ImmediateFunc{Flags{Immediate: true, Executable: true}, func() { p.EndXArray() }})
	dict.Add("exec", func() {
		p.Exec3(p.Pop())
	})

	return p
}

var mark Mark

func (p *Primitives) StartXArray() {
	p.Compiling = true
	p.Push()
}

func (p *Primitives) EndXArray() {
	var a Array
	for p.Peek() != mark {
		a.Data = append(a.Data, p.Pop())
	}
	p.Pop() // drop mark
	a.Executable = true
	p.Push(a)
}

/*

R          N1 -
Push N1 onto the return stack.

!     N1 ADDR -
Store N1 at location ADDR in program memory.

+!    N1 ADDR -
Add N1 to the value pointed to by ADDR.

:             -
Define the start of a subroutine.  The primitive
[CALL] is compiled every time this subroutine is
reference by other definitions.

;             -
Perform a subroutine return and end the definition
of a subroutine.  The primitive [EXIT] is compiled.

@       ADDR  - N1<
Fetch the value at location ADDR in program memory,
returning N1.

ABS        N1 - N2
Take the absolute value of N1 and return the result N2.

AND     N1 N2 - N3
Perform a bitwise AND on N1 and N2, giving result N3.

BRANCH        -
Perform an unconditional branch to the compiled in-line
address.

I             - N1
Return the index of the currently active loop.

I'            - N1
Return the limit of the currently active loop.

J             - N1
Return the index of the outer loop in a nested loop structure.

LEAVE         -
Set the loop counter on the return stack equal to the
loop limit to force an exit from the loop.

LIT           - N1
Treat the compiled in-line value as an integer constant,
and push it onto the stack as N1.

OVER    N1 N2 - N1 N2 N1
Push a copy of the second element on the stack, N1, onto
the top of the stack.

PICK   ... N1 - ... N2
Copy the N1'th element deep in the data stack to the top.
In Forth-83, 0 PICK is equivalent to DUP , and 1 PICK
is equivalent to OVER .

ROLL   ... N1 - ... N2
Pull the N1'th element deep in the data stack to the top,
closing the hole left in the stack.  In Forth-83, 1 ROLL
is equivalent to SWAP , and 2 ROLL is equivalent to ROT.

ROT  N1 N2 N3 - N2 N3 N1
Pull the third element down in the stack onto the top of
the stack.

 */
