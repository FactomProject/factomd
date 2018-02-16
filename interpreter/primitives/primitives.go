package primitives

import (
	"fmt"

	"strings"

	. "github.com/FactomProject/electiontesting/interpreter/common"
	. "github.com/FactomProject/electiontesting/interpreter/dictionary"
	. "github.com/FactomProject/electiontesting/interpreter/interpreter"
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
	dict.Add("-", func() {
		a := p.PopInt()
		p.Push(p.PopInt() - a)
	})                                                        // ToDo: Handle float too
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
	dict.Add("dup", func() { p.Push(p.Peek()) })
	dict.Add("pick", func() { p.Push(p.PeekN(p.PopInt())) })
	dict.Add("drop", func() { p.Pop() })
	dict.Add(".s", func() { p.PStack() })
	dict.Add("swap", func() {
		x := p.Pop()
		y := p.Pop()
		p.Push(y, x)
	})

	dict.Add("\"", func() { p.Quote() })

	// executable array
	dict.Add("{", ImmediateFunc{FlagsStruct{Immediate: true, Executable: true}, func() { p.StartXArray() }})
	dict.Add("}", ImmediateFunc{FlagsStruct{Immediate: true, Executable: true}, func() { p.EndXArray() }})
	dict.Add("exec", func() { p.Exec() })
	dict.Add("def", func() { p.Def() })

	// Control Structures
	dict.Add("repeat", func() { p.Repeat() })
	dict.Add("for", func() { p.For() })
	dict.Add("I", func() { p.I() })
	dict.Add("J", func() { p.J() })
	dict.Add("K", func() { p.K() })
	dict.Add("if", func() { p.If() })

	return p
}

//func (p *Primitives) AddPrim(s string, f func(), immediate bool) {
//	n := p.GetName(s)
//
//}

var mark Mark

func (p *Primitives) Exec() { p.Exec3(p.Pop()) }

func (p *Primitives) If() {
	cond := p.Pop()
	x := p.Pop()
	switch cond.(type) {
	case bool:
		if cond.(bool) {
			p.Exec3(x)
		}
	case int:
		if cond.(int) != 0 {
			p.Exec3(x)
		}
	}
}

func (p *Primitives) Repeat() {
	count := p.PopInt()
	x := p.Pop()
	for i := 0; i < count; i++ {
		p.Exec3(x)
	}
}

func (p *Primitives) For() {
	x := p.Pop() // Get what we execute
	end := p.PopInt()
	increment := p.PopInt()
	start := p.PopInt()

	for i := start; i < end; i += increment {
		p.C.Push(i) // publish I
		p.Exec3(x)
		p.C.Pop() // remove old I
	}
}

func (p *Primitives) I() { p.Push(p.C.Peek()) }   // Copy I to data stack
func (p *Primitives) J() { p.Push(p.C.PeekN(1)) } // Copy J to data stack
func (p *Primitives) K() { p.Push(p.C.PeekN(2)) } // Copy K to data stack

// executable arrays
func (p *Primitives) StartXArray() {
	p.Compiling++
	p.Push(mark)
}

func (p *Primitives) EndXArray() {
	//	fmt.Print("EndXArray ")
	//	p.PStack()
	p.Compiling--

	var a Array = NewArray()
	// count how far down the stack my mark is
	var i int = 0
	var start int = p.Ptr
	for {
		x := p.PeekN(i)
		if x == mark {
			break
		}
		start--
		i++
	}

	a.Data = append(a.Data, p.Data[start:p.Ptr]...)

	p.PopN(i + 1) // drop everything to the mark
	a.Flags.Executable = true
	p.Push(a)
}

func (p *Primitives) Def() {
	name := p.PopString()
	body := p.Pop()
	p.DictStack[0].Add(name, body)
}

// Strings
func (p *Primitives) Quote() {
	line := p.Line[1:] // remove space after leading quote
	n := strings.IndexByte(line, []byte("\"")[0])
	if n == -1 {
		panic("No closing \"")
	}
	s := line[:n]       // exclude the leading whitespace and the trailing "
	p.Line = line[n+1:] // remove the scanned string and quote
	p.Push(s)           // Push the string
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
