package interpreter

import (
	"bufio"
	"fmt"
	"strconv"

	"io"

	"strings"

	. "github.com/FactomProject/electiontesting/interpreter/common"
	. "github.com/FactomProject/electiontesting/interpreter/dictionary"
	. "github.com/FactomProject/electiontesting/interpreter/stack"
)

type Interpreter struct {
	Stack     // Data stack is integral
	C         Stack
	Compiling int
	DictStack []Dictionary
	Input     *bufio.Reader
	Line      string
}

func NewInterpreter() Interpreter {
	var i Interpreter
	i.Stack = NewStack()
	i.C = NewStack()
	i.DictStack = make([]Dictionary, 0)
	return i
}

func (i *Interpreter) Lookup(s string) interface{} {
	for _, d := range i.DictStack {
		e, ok := d[s]
		if ok {
			return e
		}
	}
	panic("Undefined " + s)
	return nil
}

// Push a dictionary on the stack
func (i *Interpreter) DictionaryPush(d Dictionary) {
	i.DictStack = append([]Dictionary{d}, i.DictStack...)
}
func (i *Interpreter) DictionaryPop() { i.DictStack = i.DictStack[1:] }

func (i *Interpreter) Exec3(x interface{}) {
	var flags FlagsStruct

	fmt.Printf("Exec3(%v) ", x)
	i.PStack()

	f, executable := x.(func()) // is it a Go Function? Then it's executable
	if executable {
		if i.Compiling == 0 {
			f()
			return
		} else {
			i.Push(x) // GoFuncs are never immediate
			return
		}
		return
	}

	flagSrc, ok := x.(HasFlags) // Does it have flags (Array et.al.)?
	if ok {
		flags = flagSrc.GetFlags() // get them so we know what to do...
	}

	// If it's a literal or we are compiling and it's not immediate just push it
	if !flags.Executable || (!flags.Immediate && i.Compiling > 0) {
		i.Push(x)
		return
	}

	// Got an executable thing
	switch x.(type) {
	case Array:
		for _, y := range x.(Array).Data {

			switch y.(type) {
			case Array:
				i.Push(y)
			default:
				i.Exec3(y)
			}

		} // for all elements of the executable array
	case ImmediateFunc:
		x.(ImmediateFunc).Func()
	case func():
		x.(func())() // execute the primitive
	default:
		panic(fmt.Sprintf("Exec3 of %#+v", x))
	} // switch on type

}

// execute one thing (can recurse)
func (i *Interpreter) Exec2(s string) {
	if ii, err := strconv.Atoi(s); err == nil {
		i.Exec3(ii)
	} else if b, err := strconv.ParseBool(s); err == nil {
		i.Exec3(b)
	} else if f, err := strconv.ParseFloat(s, 64); err == nil {
		i.Exec3(f)
	} else if ii, err := strconv.ParseInt(s, 10, 64); err == nil {
		i.Exec3(ii)
	} else if u, err := strconv.ParseUint(s, 10, 64); err == nil {
		i.Exec3(u)
	} else {
		// Wasn't a literal
		e := i.Lookup(s)
		i.Exec3(e)
	}
}

// execute one thing (can't recurse, catches panics)
func (i *Interpreter) ExecString(s string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error:", r)
		}
	}()
	i.Exec2(s)
}

func (i *Interpreter) Interpret(source io.Reader) {
	var prevInput *bufio.Reader = i.Input
	var prevLine = i.Line
	var s string
	i.Line = ""
	i.Input = bufio.NewReader(source) // save the source off for primitives that need it

readline:
	for {
		for {
			line, isPrefix, err := i.Input.ReadLine()
			i.Line += string(line) // append this piece of the line
			if err == io.EOF || i.Line == "" {
				break readline
			}
			if err != nil {
				panic(err)
			}
			if !isPrefix {
				break
			}
		} // Until we get a whole line
		i.Line = i.Line + " " // Insure there is training whitespace
		for {
			// Scan a string from the current line (possible modified by execution)
			line := i.Line
			n, err := fmt.Sscan(line, &s)
			if n == 1 {
				n := strings.Index(line, s)
				line = line[n+len(s)+1:] // Trim off the string and the ws following
				i.Line = line
				if s != "" {
					i.ExecString(s) // execute the string
				}
			}
			if i.Line == "" {
				break
			}
			if err == io.EOF {
				return
			}
			if err != nil {
				panic(err)
			}
		} // Until we have done all the strings on the line
	} // till EOF or error
	i.Input = prevInput
	i.Line = prevLine
}
