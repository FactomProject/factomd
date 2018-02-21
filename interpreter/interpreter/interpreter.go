package interpreter

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	. "github.com/FactomProject/electiontesting/interpreter/common"
	. "github.com/FactomProject/electiontesting/interpreter/dictionary"
	. "github.com/FactomProject/electiontesting/interpreter/names"
	. "github.com/FactomProject/electiontesting/interpreter/stack"
)

type Interpreter struct {
	Stack     // Data stack is integral
	C         Stack
	Compiling int
	DictStack []Dictionary
	Input     *bufio.Reader
	Line      string
	NameManager
}

func NewInterpreter() Interpreter {
	var i Interpreter
	i.Stack = NewStack()
	i.C = NewStack()
	i.DictStack = make([]Dictionary, 0)
	i.NameManager = NewNameManager()
	return i
}

// Convert a string to a name
func (i *Interpreter) Lookup(s string) Name {
	n := i.GetName(s)
	return n
}

// Push a dictionary on the stack
func (i *Interpreter) DictionaryPush(d Dictionary) {
	i.DictStack = append([]Dictionary{d}, i.DictStack...)
}
func (i *Interpreter) DictionaryPop() { i.DictStack = i.DictStack[1:] }

func (p *Interpreter) String(x interface{}) (s string) {
	switch x.(type) {
	case Mark:
		s = "MARK"
	case Name:
		s = p.GetString(x.(Name))
	case Array:
		if x.(Array).Executable {
			s = "{ "
			for _, y := range x.(Array).Data {
				s += p.String(y) + " "

			} // for all elements of the executable array
			s += "}"
		} else {
			s = "[ "
			for _, y := range x.(Array).Data {
				s += p.String(y) + " "

			} // for all elements of the executable array
			s += "]"
		}
	default:
		s = fmt.Sprintf("%v", x)
	}
	return s
}

func (p *Interpreter) Print(x interface{}) { fmt.Print(p.String(x) + " ") }

func (p *Interpreter) PStack() {
	fmt.Print("-TOS- ")
	for i := 0; i < p.Ptr; i++ {
		p.Print(p.PeekN(i))
	}
	fmt.Println("|\n")
}

func (i *Interpreter) executeArray(a Array) {
	flags := a.GetFlags()
	if flags.Immediate || (flags.Executable && i.Compiling == 0) {
		for _, y := range a.Data {
			switch y.(type) {
			case Array:
				i.Push(y)
			default:
				i.Exec3(y)
			}
		} // for all elements of the executable array
	} else {
		i.Push(a)
	}
} // executeArray(){...}

func (i *Interpreter) findInDictStack(n Name) (*DictionaryEnrty, bool) {
	// find the name in the dict stack
	for _, d := range i.DictStack {
		e, ok := d[n.GetRawName()]
		if ok {
			return &e, ok
		}
	}
	return nil, false
}

func (i *Interpreter) executeName(n Name) {
	s := i.GetString(n)
	_ = s
	if n.IsExecutable() {
		dictEntry, ok := i.findInDictStack(n)

		if !ok {
			panic("executeName(): Undefined " + i.GetString(n))
		}

		if dictEntry.Immediate || (dictEntry.Executable && i.Compiling == 0) {
			switch dictEntry.E.(type) {
			case func():
				dictEntry.E.(func())()
				return
			case Array:
				i.executeArray(dictEntry.E.(Array))
				return
			default:
				i.Push(dictEntry.E)
				return
			}
		}
	}
	// if not executing the name push then it
	i.Push(n)
} // executeName90{...}

func (i *Interpreter) Exec3(x interface{}) {

	//	fmt.Printf("Exec3(%v) ", i.String(x))
	//	i.PStack()

	// Got an executable thing and I want to execute it
	switch x.(type) {
	case Array:
		i.executeArray(x.(Array))
		return
	case func():
		panic("primitive in exec3()")
	case Name:
		i.executeName(x.(Name))
		return
	default:
		i.Push(x)
	} // switch on type
}

// execute one thing
func (i *Interpreter) InterpretString(s string) {
	//	fmt.Printf("Exec2(\"%s\")\n", s)
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

func (i *Interpreter) InterpretLine(line string) {
	//fmt.Printf("Interpret(\"%s\")\n", line)
	defer func() { i.Line = i.Line }()
	i.Line = line

	comment := false
	var s string
	for {
		// Scan a string from the current line (possible modified by execution)
		line := i.Line
		line = strings.TrimSpace(line)
		n, err := fmt.Sscan(line, &s)
		if n == 1 {
			line = line[len(s):] // Trim off the string and the ws following
			i.Line = line
			if s != "" {
				// # is a comment
				if s == "/*" {
					comment = true
					continue
				}
				if s == "*/" {
					comment = false
					continue
				}
				if comment {
					continue
				}
				i.InterpretString(s) // execute the string
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
	fmt.Println()
} // till EOF or error

func (i *Interpreter) Interpret(source io.Reader) {
	defer func() { i.Input = i.Input }() // Reset i.Input when we exit
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error:", r)
			i.Compiling = 0
		}
	}()

	i.Input = bufio.NewReader(source) // save the source off for primitives that need it
	for {
		var line string
		for {
			chunk, isPrefix, err := i.Input.ReadLine()
			line += string(chunk) // append this piece of the line
			if err == io.EOF || line == "" {
				return
			}
			if err != nil {
				panic(err)
			}
			if !isPrefix {
				break
			}
		} // Until we get a whole line
		i.InterpretLine(line)
	}
}
