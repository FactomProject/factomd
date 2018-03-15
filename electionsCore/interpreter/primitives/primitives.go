package primitives

import (
	"strings"

	"fmt"

	. "github.com/FactomProject/electiontesting/interpreter/common"
	. "github.com/FactomProject/electiontesting/interpreter/dictionary"
	. "github.com/FactomProject/electiontesting/interpreter/interpreter"
	. "github.com/FactomProject/electiontesting/interpreter/names"
)

type Primitives struct {
	Interpreter
	// Hmm used to be more stuff here but it's all migrated away... guess this is really all on interpreter
}

var executable FlagsStruct = FlagsStruct{Traced: false, Immediate: false, Executable: true}
var immediate FlagsStruct = FlagsStruct{Traced: false, Immediate: true, Executable: true}

func (p *Primitives) AddPrim(dict Dictionary, name string, x interface{}, f FlagsStruct) {
	n := p.GetName(name)
	dict.Add(n.GetRawName(), DictionaryEnrty{n.GetRawName(), f, x})
}

func NewPrimitives() *Primitives {
	p := new(Primitives)
	p.Interpreter = NewInterpreter()

	primitives := NewDictionary()
	p.DictionaryPush(primitives) // Primitives Dictionary

	p.AddPrim(primitives, "primitives", primitives, executable)

	p.AddPrim(primitives, "+", func() { p.Push(p.PopInt() + p.PopInt()) }, executable)  // ToDo: Handle float too
	p.AddPrim(primitives, "-", func() { p.Push(-p.PopInt() + p.PopInt()) }, executable) // ToDo: Handle float too
	p.AddPrim(primitives, "*", func() { p.Push(p.PopInt() * p.PopInt()) }, executable)  // ToDo: Handle float too
	p.AddPrim(primitives, "/", func() { p.Push(p.PopInt() / p.PopInt()) }, executable)  // ToDo: Handle float too
	p.AddPrim(primitives, "&", func() { p.Push(p.PopInt() & p.PopInt()) }, executable)
	p.AddPrim(primitives, "|", func() { p.Push(p.PopInt() | p.PopInt()) }, executable)
	p.AddPrim(primitives, "^", func() { p.Push(p.PopInt() ^ p.PopInt()) }, executable)
	p.AddPrim(primitives, "~", func() { p.Push(p.PopInt() ^ -1) }, executable)
	p.AddPrim(primitives, "&&", func() { p.Push(p.PopBool() && p.PopBool()) }, executable)
	p.AddPrim(primitives, "||", func() { p.Push(p.PopBool() || p.PopBool()) }, executable)
	p.AddPrim(primitives, "!", func() { p.Push(!p.PopBool()) }, executable)
	p.AddPrim(primitives, ".", func() { p.Print(p.Pop()) }, executable)
	p.AddPrim(primitives, "<", func() { p.Push(p.PopInt() < p.PopInt()) }, executable)
	p.AddPrim(primitives, "=", func() { p.Push(p.PopInt() == p.PopInt()) }, executable)
	p.AddPrim(primitives, "!=", func() { p.Push(p.PopInt() != p.PopInt()) }, executable)
	p.AddPrim(primitives, ">", func() { p.Push(p.PopInt() > p.PopInt()) }, executable)
	p.AddPrim(primitives, "0=", func() { p.Push(p.PopInt() > 0) }, executable)
	p.AddPrim(primitives, "?dup",
		func() {
			x := p.Peek().(int)
			if x != 0 {
				p.Push(x)
			}
		}, executable)

	p.AddPrim(primitives, "dup", func() { p.Push(p.Peek()) }, executable)
	p.AddPrim(primitives, "clear", func() { p.Clear() }, executable)
	p.AddPrim(primitives, "pick", func() { p.Push(p.PeekN(p.PopInt())) }, executable)
	p.AddPrim(primitives, "drop", func() { p.Pop() }, executable)
	p.AddPrim(primitives, ".s", func() { p.PStack() }, executable)
	p.AddPrim(primitives, "swap",
		func() {
			x := p.Pop()
			p.Push(p.Pop(), x)
		}, executable)

	p.AddPrim(primitives, "\"", func() { p.Quote() }, immediate)

	// arrays
	p.AddPrim(primitives, "{", func() { p.StartXArray() }, immediate)
	p.AddPrim(primitives, "}", func() { p.EndXArray() }, immediate)

	p.AddPrim(primitives, "[", func() { p.StartArray() }, executable)
	p.AddPrim(primitives, "]", func() { p.EndArray() }, executable)

	p.AddPrim(primitives, "exec", func() { p.Exec() }, executable)
	p.AddPrim(primitives, "def", func() { p.Def() }, executable)

	// Control Structures
	p.AddPrim(primitives, "repeat", func() { p.Repeat() }, executable)
	p.AddPrim(primitives, "forall", func() { p.ForAll() }, executable)
	p.AddPrim(primitives, "for", func() { p.For() }, executable)
	p.AddPrim(primitives, "I", func() { p.I() }, executable)
	p.AddPrim(primitives, "J", func() { p.J() }, executable)
	p.AddPrim(primitives, "K", func() { p.K() }, executable)
	p.AddPrim(primitives, "if", func() { p.If() }, executable)
	p.AddPrim(primitives, "ifelse", func() { p.IfElse() }, executable)

	// Control debug
	p.AddPrim(primitives, "traceon", func() { p.Tracing++ }, executable)
	p.AddPrim(primitives, "traceoff", func() { p.Tracing = 0 }, executable)

	// Control object flags
	p.AddPrim(primitives, "immediate", func() { p.SetImmediate() }, executable)
	p.AddPrim(primitives, "executable", func() { p.SetExecutable() }, executable)
	p.AddPrim(primitives, "trace", func() { p.SetTrace() }, executable)

	userDictionary := NewDictionary()
	p.AddPrim(primitives, "userdict", userDictionary, executable)
	p.DictionaryPush(userDictionary) // User Dictionary

	return p
}

//func (p *Primitives) AddPrim(s string, f func(), immediate bool) {
//	n := p.GetName(s)
//
//}

var mark Mark

func (p *Primitives) Exec() { p.Exec3(p.Pop()) }

func (p *Primitives) SetExecutable() {
	x := p.Pop()
	switch x.(type) {
	case Name:
		p.Push(x.(Name).MakeExecutable())
	case Array:
		a := x.(Array)
		a.Executable = true
		p.Push(a)
	default:
		fmt.Printf("Can't make %T %v executable", x, x)
		p.Push(x)
	}
}
func (p *Primitives) SetImmediate() {
	x := p.Peek()
	y, ok := x.(HasFlags)
	if ok {
		x.(HasFlags).SetFlags(y.GetFlags().SetImmediate(true))
	}

}
func (p *Primitives) SetTrace() {
	x := p.Peek()
	y, ok := x.(HasFlags)
	if ok {
		x.(HasFlags).SetFlags(y.GetFlags().SetTraced(true))
	}

}

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
func (p *Primitives) IfElse() {
	cond := p.Pop()
	x := p.Pop()
	y := p.Pop()
	switch cond.(type) {
	case bool:
		if cond.(bool) {
			p.Exec3(x)
		} else {
			p.Exec3(y)
		}
	case int:
		if cond.(int) != 0 {
			p.Exec3(x)
		} else {
			p.Exec3(y)
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

func (p *Primitives) ForAll() {
	x := p.Pop()
	y := p.PopArray()
	for _, z := range y.Data {
		p.Push(z)
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

// arrays
func (p *Primitives) StartXArray() {
	p.Compiling++
	p.Push(mark)
}

func (p *Primitives) StartArray() {
	p.Push(mark)
}

func (p *Primitives) EndArray() {
	//	fmt.Print("EndArray ")
	//	p.PStack()
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
	p.Push(a)
}

func (p *Primitives) EndXArray() {
	p.Compiling--
	p.EndArray()
	a := p.Pop().(Array)
	a.Executable = true
	p.Push(a)
}

func (p *Primitives) Def() {
	var flags FlagsStruct
	name := p.PopName()
	body := p.Pop()
	switch body.(type) {
	case Array:
		flags = body.(Array).GetFlags()
	case DictionaryEnrty:
		flags = body.(DictionaryEnrty).FlagsStruct
	case Name:
		flags.Executable = body.(Name).IsExecutable()
	}
	p.DictStack[0].Add(name, DictionaryEnrty{name, flags, body})
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
