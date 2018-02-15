package interpreter

import (
	"bufio"
	"fmt"
	. "github.com/FactomProject/electiontesting/interpreter/common"
	. "github.com/FactomProject/electiontesting/interpreter/dictionary"
	. "github.com/FactomProject/electiontesting/interpreter/stack"
	"io"
	"strconv"
)

type Interpreter struct {
	Stack     // Data stack is integral
	C         Stack
	Compiling bool
	DictStack []Dictionary
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

	f, executable := x.(func()) // is it a Go Function?
	if executable {
		if i.Compiling == false {
			f()
		} else {
			i.Push(x) // Never immediate
		}
		return
	}

	flags, ok := x.(Flags) // Does it have flags ?
	if !ok || !flags.Executable {
		i.Push(x)
		return
	}

	immediateFunc, executable := x.(ImmediateFunc) // Should not have to manually check this!!!
	if executable {
		immediateFunc.Func()
		return
	}

	// Got an executable thing
	switch x.(type) {
	case Array:
		if flags.Immediate || i.Compiling == false {
			for _, y := range x.(Array).Data {
				i.Exec3(y)
			} // for all elements of the executable array
		} else {
			i.Push(x)
		}
	case func():
		if flags.Immediate || i.Compiling == false {
			x.(func())() // execute the primitive
		} else {
			i.Push(x)
		}
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
	scanner := bufio.NewScanner(source)
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		s := scanner.Text()
		i.ExecString(s)
	}

}
